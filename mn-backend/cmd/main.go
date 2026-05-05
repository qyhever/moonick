package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	router "moonick/internal/api"
	"moonick/internal/config"
	"moonick/internal/logger"
	"moonick/internal/pkg/cache"
	"moonick/internal/repository/mysql"
	"moonick/internal/task"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Printf("❌ init config failed, err: %v\n", err)
		return
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("❌ init logger failed, err:%v\n", err)
		return
	}

	cfg := config.GetConfig()
	if cfg != nil && cfg.Redis.Enabled {
		redisClient := cache.NewRedisClient(cfg.Redis)
		if err := redisClient.Ping(context.Background()); err != nil {
			fmt.Printf("⚠️ init redis failed, fallback to memory rate limit: %v\n", err)
			_ = redisClient.Close()
		} else {
			cache.SetRedis(redisClient)
			defer func() {
				_ = redisClient.Close()
			}()
		}
	}

	// 注册路由
	r := router.SetupRouter()
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 显示启动信息
	addr := config.GetServerAddr()
	fmt.Printf("\n🚀 服务正在启动...\n")
	fmt.Printf("🔗 地址: http://localhost%s\n", addr)

	tripRepo := mysql.NewTripRepository()
	expireTask := task.NewTripExpireTask(tripRepo)
	go task.NewScheduler(task.SchedulerConfig{
		Interval: time.Minute,
		Run:      expireTask.Run,
	}).Run(rootCtx)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrCh:
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		fmt.Printf("❌ 启动服务器失败: %v\n", err)
		return
	case <-rootCtx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("❌ 关闭服务器失败: %v\n", err)
	}
}
