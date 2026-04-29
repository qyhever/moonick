# 行程过期任务调度链路实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 让 `mn-backend` 在进程启动后自动执行并周期性执行行程过期扫描任务，在服务退出时随主进程优雅停止。

**架构：** 保持现有 `TripExpireTask` 作为业务任务入口，补充“执行一次并返回处理结果”的能力；在 `internal/task` 中新增轻量调度器，负责启动即跑一次、后续按固定间隔执行；`cmd/main.go` 改为显式管理 `http.Server` 和根 `context`，把调度器挂到应用生命周期里。

**技术栈：** Go、Gin、标准库 `context/signal/http/time`、现有 MySQL repository 与 zap 全局日志

---

### 任务 1：为过期任务补可观察结果

**文件：**
- 修改：`mn-backend/internal/task/trip_expire_task.go`
- 创建：`mn-backend/internal/task/trip_expire_task_test.go`
- 测试：`mn-backend/internal/task/trip_expire_task_test.go`

- [ ] **步骤 1：编写失败的测试**

```go
func TestTripExpireTask_RunReturnsExpiredCount(t *testing.T) {
	repo := &stubTripExpireRepository{count: 2}
	task := NewTripExpireTask(repo)

	count, err := task.Run(context.Background())

	if err != nil {
		t.Fatalf("run task: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 expired trips, got %d", count)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task -run TestTripExpireTask_RunReturnsExpiredCount`

预期：FAIL，报错 `assignment mismatch` 或 `cannot use ... as ...`，说明 `Run` 还未返回处理数量。

- [ ] **步骤 3：编写最少实现代码**

```go
type tripExpireRepository interface {
	ExpireTripsBefore(ctx context.Context, before time.Time) (int64, error)
}

func (t *TripExpireTask) Run(ctx context.Context) (int64, error) {
	return t.tripRepo.ExpireTripsBefore(ctx, time.Now())
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task -run TestTripExpireTask_RunReturnsExpiredCount`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/task/trip_expire_task.go mn-backend/internal/task/trip_expire_task_test.go
git commit -m "feat: expose trip expire task result"
```

### 任务 2：为周期调度补失败测试并实现

**文件：**
- 修改：`mn-backend/internal/task/trip_expire_task.go`
- 修改：`mn-backend/internal/task/trip_expire_task_test.go`
- 测试：`mn-backend/internal/task/trip_expire_task_test.go`

- [ ] **步骤 1：编写失败的测试**

```go
func TestScheduler_RunExecutesImmediatelyAndOnTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newFakeTicker()
	var runs atomic.Int64

	scheduler := NewScheduler(SchedulerConfig{
		Interval:   time.Minute,
		NewTicker:  func(time.Duration) ticker { return ticker },
		Run: func(context.Context) (int64, error) {
			return runs.Add(1), nil
		},
	})

	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	waitForRuns(t, &runs, 1)
	ticker.Tick(time.Now())
	waitForRuns(t, &runs, 2)
	cancel()
	<-done
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task -run TestScheduler_RunExecutesImmediatelyAndOnTick`

预期：FAIL，报错 `undefined: NewScheduler` 或等价未实现错误。

- [ ] **步骤 3：编写最少实现代码**

```go
type SchedulerConfig struct {
	Interval  time.Duration
	NewTicker func(time.Duration) Ticker
	Run       func(context.Context) (int64, error)
}

func (s *Scheduler) Run(ctx context.Context) {
	s.runOnce(ctx)
	ticker := s.newTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C():
			s.runOnce(ctx)
		}
	}
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task -run 'TestTripExpireTask_RunReturnsExpiredCount|TestScheduler_RunExecutesImmediatelyAndOnTick'`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/task/trip_expire_task.go mn-backend/internal/task/trip_expire_task_test.go
git commit -m "feat: add trip expire scheduler"
```

### 任务 3：把调度器接入主进程生命周期

**文件：**
- 修改：`mn-backend/cmd/main.go`
- 测试：`mn-backend/internal/task/trip_expire_task_test.go`

- [ ] **步骤 1：编写失败的测试**

```go
func TestScheduler_RunStopsAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := newFakeTicker()

	scheduler := NewScheduler(SchedulerConfig{
		Interval:  time.Minute,
		NewTicker: func(time.Duration) ticker { return ticker },
		Run: func(context.Context) (int64, error) {
			return 0, nil
		},
	})

	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancel")
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task -run TestScheduler_RunStopsAfterContextCancel`

预期：FAIL，说明调度器还未正确处理退出。

- [ ] **步骤 3：编写最少实现代码**

```go
rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

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
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/task ./...`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/cmd/main.go mn-backend/internal/task/trip_expire_task.go mn-backend/internal/task/trip_expire_task_test.go
git commit -m "feat: wire trip expire scheduler into backend lifecycle"
```
