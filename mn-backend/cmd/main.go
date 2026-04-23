package main

import (
	"fmt"

	router "moonick/internal/api"
	"moonick/internal/config"
	"moonick/internal/logger"
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

	// 注册路由
	r := router.SetupRouter()

	// 显示启动信息
	addr := config.GetServerAddr()
	fmt.Printf("\n🚀 服务正在启动...\n")
	fmt.Printf("🔗 地址: http://localhost%s\n", addr)

	// 启动HTTP服务，使用配置文件中的端口
	err := r.Run(addr)
	if err != nil {
		fmt.Printf("❌ 启动服务器失败: %v\n", err)
		return
	}
}
