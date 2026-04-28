# 明叶同行 P1 MySQL 落库与后台完整编辑实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 `mn-backend` 的用户、管理员、行程、收藏仓储切换为真实 MySQL 持久化，补齐后台完整行程编辑链路，并保持 H5/Admin 当前闭环可用。

**架构：** 后端继续保持 `controller -> service -> repository` 分层，不引入 ORM 和迁移框架。通过新增手动初始化 SQL、补充数据库连接入口、替换 `internal/repository/mysql` 的内存实现、升级后台 `PUT /api/admin/v1/trips/:id` 完整编辑语义，完成真实落库与后台编辑打通。前端只对 Admin 编辑页做必要升级，H5 维持现有契约兼容。

**技术栈：** Go + Gin + `database/sql` + MySQL，React + Vite + Ant Design，Vitest，手动初始化 SQL

---

## 文件结构

### 后端数据库与连接

- 创建：`mn-backend/docs/sql/001_init.sql`
  责任：手动初始化 `users`、`admins`、`trips`、`trip_favorites` 表和索引。
- 修改：`mn-backend/internal/config/config.go`
  责任：确认 MySQL 配置读取完整，补数据库连接初始化需要的结构。
- 创建：`mn-backend/internal/repository/mysql/db.go`
  责任：提供共享 `*sql.DB` 初始化与获取入口。
- 修改：`mn-backend/internal/api/router.go`
  责任：在应用启动时初始化 DB、执行管理员 seed upsert、注入真实 MySQL repository。

### 后端 repository

- 修改：`mn-backend/internal/repository/mysql/user_repository.go`
  责任：用户真实落库 CRUD。
- 修改：`mn-backend/internal/repository/mysql/admin_repository.go`
  责任：管理员查询与 seed upsert。
- 修改：`mn-backend/internal/repository/mysql/trip_repository.go`
  责任：行程真实落库 CRUD 与列表查询。
- 修改：`mn-backend/internal/repository/mysql/favorite_repository.go`
  责任：收藏关系真实落库。
- 修改：`mn-backend/internal/repository/mysql/user_repository_test.go`
- 修改：`mn-backend/internal/repository/mysql/admin_repository_test.go`
- 创建：`mn-backend/internal/repository/mysql/trip_repository_test.go`
- 创建：`mn-backend/internal/repository/mysql/favorite_repository_test.go`
  责任：repository 落库测试。

### 后端后台完整编辑

- 修改：`mn-backend/internal/model/request/trip.go`
  责任：新增 `AdminUpdateTripDetailRequest`。
- 修改：`mn-backend/internal/model/entity/trip.go`
  责任：补 `PriceAmount`、`Remark`、`ClosedReason` 等真实落库字段。
- 修改：`mn-backend/internal/model/response/trip.go`
  责任：补后台与前台兼容所需返回字段。
- 修改：`mn-backend/internal/service/trip_service.go`
  责任：补行程字段转换与校验兼容。
- 修改：`mn-backend/internal/service/admin_service.go`
  责任：新增后台完整编辑逻辑。
- 修改：`mn-backend/internal/controller/admin_trip_controller.go`
  责任：将 `PUT /api/admin/v1/trips/:id` 升级为完整字段编辑。
- 修改：`mn-backend/internal/service/admin_service_test.go`
- 修改：`mn-backend/internal/controller/trip_controller_test.go`
  责任：后台完整编辑相关测试。

### Admin 前端完整编辑

- 修改：`mn-frontend-admin/src/features/trips/api.ts`
  责任：后台行程详情与更新 payload 补齐完整字段。
- 修改：`mn-frontend-admin/src/features/trips/TripEditPage.tsx`
  责任：从状态编辑升级为完整表单编辑。
- 修改：`mn-frontend-admin/src/features/trips/TripDetailPage.tsx`
  责任：展示新增字段。
- 修改：`mn-frontend-admin/src/test/trip-edit.test.tsx`
  责任：覆盖完整 payload、确认弹窗和保存行为。

### 文档

- 修改：`mn-backend/README.md`
- 修改：`README.md`
- 修改：`docs/technical/api-checklist.md`
  责任：补 SQL 导入、MySQL 启动要求和人工联调路径。

---

### 任务 1：新增初始化 SQL 与数据库连接入口

**文件：**
- 创建：`mn-backend/docs/sql/001_init.sql`
- 创建：`mn-backend/internal/repository/mysql/db.go`
- 修改：`mn-backend/internal/config/config.go`
- 修改：`mn-backend/internal/api/router.go`

- [ ] **步骤 1：编写失败的数据库连接测试或编译入口断言**

```go
func TestOpenDBRequiresConfiguredDSN(t *testing.T) {
	_, err := OpenDB(Config{})
	if err == nil {
		t.Fatal("expected empty config to fail")
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run TestOpenDBRequiresConfiguredDSN -v`

预期：FAIL，报错包含 `undefined: OpenDB`

- [ ] **步骤 3：编写最少实现代码**

```go
func OpenDB(cfg config.MySQLConfig) (*sql.DB, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, errors.New("mysql addr is required")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		cfg.User, cfg.Password, cfg.Addr, cfg.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
```

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(20) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL,
  avatar_url VARCHAR(512) NOT NULL DEFAULT '',
  default_wechat VARCHAR(64) NOT NULL DEFAULT '',
  default_phone VARCHAR(20) NOT NULL DEFAULT '',
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_phone (phone)
);
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run TestOpenDBRequiresConfiguredDSN -v`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/docs/sql/001_init.sql mn-backend/internal/repository/mysql/db.go mn-backend/internal/config/config.go mn-backend/internal/api/router.go
git commit -m "feat: 补充 MySQL 初始化脚本和数据库连接入口"
```

### 任务 2：将用户与管理员仓储切换为真实 MySQL

**文件：**
- 修改：`mn-backend/internal/repository/mysql/user_repository.go`
- 修改：`mn-backend/internal/repository/mysql/admin_repository.go`
- 修改：`mn-backend/internal/repository/mysql/user_repository_test.go`
- 修改：`mn-backend/internal/repository/mysql/admin_repository_test.go`
- 修改：`mn-backend/internal/api/router.go`

- [ ] **步骤 1：编写失败的落库测试**

```go
func TestUserRepository_CreateRejectsDuplicatePhone(t *testing.T) {
	repo := newUserRepositoryForDBTest(t)

	_, err := repo.Create(context.Background(), entity.User{
		Phone:        "13800138000",
		PasswordHash: "hash-1",
		Nickname:     "用户A",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("first create returned error: %v", err)
	}

	_, err = repo.Create(context.Background(), entity.User{
		Phone:        "13800138000",
		PasswordHash: "hash-2",
		Nickname:     "用户B",
		Status:       "active",
	})
	if !errors.Is(err, ErrUserPhoneAlreadyExists) {
		t.Fatalf("expected duplicate phone error, got %v", err)
	}
}
```

```go
func TestAdminRepository_UpsertSeedAdmin(t *testing.T) {
	repo := newAdminRepositoryForDBTest(t)
	admin := entity.Admin{
		ID:           1,
		Username:     "admin",
		PasswordHash: "hash",
		Name:         "管理员",
		Status:       "active",
	}

	if err := repo.Upsert(context.Background(), admin); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}

	got, err := repo.FindByUsername(context.Background(), "admin")
	if err != nil || got == nil {
		t.Fatalf("find by username failed, err=%v got=%v", err, got)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestUserRepository_CreateRejectsDuplicatePhone|TestAdminRepository_UpsertSeedAdmin' -v`

预期：FAIL，报错包含 `newUserRepositoryForDBTest is undefined` 或 `repo.Upsert undefined`

- [ ] **步骤 3：编写最少实现代码**

```go
func (r *UserRepository) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (phone, password_hash, nickname, avatar_url, default_wechat, default_phone, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Phone, user.PasswordHash, user.Nickname, user.AvatarURL,
		user.DefaultWechat, user.DefaultPhone, user.Status,
	)
	if isDuplicateKey(err) {
		return nil, ErrUserPhoneAlreadyExists
	}
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.FindByID(ctx, id)
}
```

```go
func (r *AdminRepository) Upsert(ctx context.Context, admin entity.Admin) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admins (id, username, password_hash, display_name, status)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			password_hash = VALUES(password_hash),
			display_name = VALUES(display_name),
			status = VALUES(status)`,
		admin.ID, admin.Username, admin.PasswordHash, admin.Name, admin.Status,
	)
	return err
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestUserRepository_CreateRejectsDuplicatePhone|TestAdminRepository_UpsertSeedAdmin' -v`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/repository/mysql/user_repository.go mn-backend/internal/repository/mysql/admin_repository.go mn-backend/internal/repository/mysql/user_repository_test.go mn-backend/internal/repository/mysql/admin_repository_test.go mn-backend/internal/api/router.go
git commit -m "feat: 将用户和管理员仓储切换到 MySQL"
```

### 任务 3：将行程与收藏仓储切换为真实 MySQL

**文件：**
- 修改：`mn-backend/internal/model/entity/trip.go`
- 修改：`mn-backend/internal/repository/mysql/trip_repository.go`
- 修改：`mn-backend/internal/repository/mysql/favorite_repository.go`
- 创建：`mn-backend/internal/repository/mysql/trip_repository_test.go`
- 创建：`mn-backend/internal/repository/mysql/favorite_repository_test.go`

- [ ] **步骤 1：编写失败的 CRUD 与列表测试**

```go
func TestTripRepository_CreateAndFindByID(t *testing.T) {
	repo := newTripRepositoryForDBTest(t)
	created, err := repo.Create(context.Background(), entity.Trip{
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 4, 30, 9, 0, 0, 0, time.Local),
		SeatCount:         2,
		PriceAmount:       decimalPtr("68.00"),
		IsPriceNegotiable: false,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)
	if err != nil || got == nil {
		t.Fatalf("find trip failed, err=%v got=%v", err, got)
	}
}
```

```go
func TestFavoriteRepository_CreateDeleteAndCount(t *testing.T) {
	repo := newFavoriteRepositoryForDBTest(t)
	if err := repo.Create(context.Background(), 1001, 2001); err != nil {
		t.Fatalf("create favorite: %v", err)
	}
	exists, _ := repo.Exists(context.Background(), 1001, 2001)
	if !exists {
		t.Fatal("expected favorite exists")
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestTripRepository_CreateAndFindByID|TestFavoriteRepository_CreateDeleteAndCount' -v`

预期：FAIL，报错包含 `unknown field PriceAmount` 或 `newTripRepositoryForDBTest is undefined`

- [ ] **步骤 3：编写最少实现代码**

```go
type Trip struct {
	ID                int64
	UserID            int64
	TripType          string
	FromText          string
	ToText            string
	DepartureAt       time.Time
	SeatCount         int
	PriceAmount       *float64
	IsPriceNegotiable bool
	ContactWechat     string
	ContactPhone      string
	Remark            string
	Status            string
	ClosedReason      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
```

```go
func (r *TripRepository) Update(ctx context.Context, trip entity.Trip) (*entity.Trip, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE trips
		SET trip_type=?, from_city_text=?, to_city_text=?, departure_date=?, departure_time=?, departure_at=?,
		    seat_count=?, price_amount=?, is_price_negotiable=?, contact_wechat=?, contact_phone=?,
		    remark=?, status=?, closed_reason=?, updated_at=NOW()
		WHERE id=? AND deleted_at IS NULL`,
		trip.TripType, trip.FromText, trip.ToText,
		trip.DepartureAt.Format(time.DateOnly), trip.DepartureAt.Format("15:04:05"), trip.DepartureAt,
		trip.SeatCount, trip.PriceAmount, trip.IsPriceNegotiable, trip.ContactWechat, trip.ContactPhone,
		trip.Remark, trip.Status, trip.ClosedReason, trip.ID,
	)
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, ErrTripNotFound
	}
	return r.FindByID(ctx, trip.ID)
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestTripRepository_CreateAndFindByID|TestFavoriteRepository_CreateDeleteAndCount' -v`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/model/entity/trip.go mn-backend/internal/repository/mysql/trip_repository.go mn-backend/internal/repository/mysql/favorite_repository.go mn-backend/internal/repository/mysql/trip_repository_test.go mn-backend/internal/repository/mysql/favorite_repository_test.go
git commit -m "feat: 将行程和收藏仓储切换到 MySQL"
```

### 任务 4：升级后台完整行程编辑后端接口

**文件：**
- 修改：`mn-backend/internal/model/request/trip.go`
- 修改：`mn-backend/internal/model/response/trip.go`
- 修改：`mn-backend/internal/service/trip_service.go`
- 修改：`mn-backend/internal/service/admin_service.go`
- 修改：`mn-backend/internal/controller/admin_trip_controller.go`
- 修改：`mn-backend/internal/service/admin_service_test.go`
- 修改：`mn-backend/internal/controller/trip_controller_test.go`

- [ ] **步骤 1：编写失败的后台完整编辑测试**

```go
func TestAdminService_UpdateTripDetail(t *testing.T) {
	svc := newAdminServiceForTest(...)
	resp, err := svc.UpdateTripDetail(context.Background(), 2001, request.AdminUpdateTripDetailRequest{
		TripType:          "passenger_post",
		FromText:          "苏州",
		ToText:            "上海",
		DepartureDate:     "2026-05-01",
		DepartureTime:     "08:30",
		SeatCount:         3,
		PriceAmount:       "88.00",
		IsPriceNegotiable: false,
		ContactPhone:      "13900139000",
		ContactWechat:     "trip-admin",
		Remark:            "后台改过备注",
		Status:            "full",
	})
	if err != nil {
		t.Fatalf("update trip detail: %v", err)
	}
	if resp.FromText != "苏州" || resp.Status != "full" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/service ./internal/controller -run TestAdminService_UpdateTripDetail -v`

预期：FAIL，报错包含 `undefined: request.AdminUpdateTripDetailRequest` 或 `svc.UpdateTripDetail undefined`

- [ ] **步骤 3：编写最少实现代码**

```go
type AdminUpdateTripDetailRequest struct {
	TripType          string  `json:"tripType" binding:"required"`
	FromText          string  `json:"fromText" binding:"required"`
	ToText            string  `json:"toText" binding:"required"`
	DepartureDate     string  `json:"departureDate" binding:"required"`
	DepartureTime     string  `json:"departureTime" binding:"required"`
	SeatCount         int     `json:"seatCount"`
	PriceAmount       *float64 `json:"priceAmount"`
	IsPriceNegotiable bool    `json:"isPriceNegotiable"`
	ContactWechat     string  `json:"contactWechat"`
	ContactPhone      string  `json:"contactPhone"`
	Remark            string  `json:"remark"`
	Status            string  `json:"status" binding:"required"`
}
```

```go
func (s *AdminService) UpdateTripDetail(ctx context.Context, tripID int64, req request.AdminUpdateTripDetailRequest) (*response.TripDetail, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	departureAt, err := parseAdminTripDeparture(req)
	if err != nil {
		return nil, err
	}
	trip.TripType = strings.TrimSpace(req.TripType)
	trip.FromText = strings.TrimSpace(req.FromText)
	trip.ToText = strings.TrimSpace(req.ToText)
	trip.DepartureAt = departureAt
	trip.SeatCount = req.SeatCount
	trip.PriceAmount = req.PriceAmount
	trip.IsPriceNegotiable = req.IsPriceNegotiable
	trip.ContactWechat = strings.TrimSpace(req.ContactWechat)
	trip.ContactPhone = strings.TrimSpace(req.ContactPhone)
	trip.Remark = strings.TrimSpace(req.Remark)
	trip.Status = strings.TrimSpace(req.Status)
	updated, err := s.tripRepo.Update(ctx, *trip)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return toTripDetail(updated, false), nil
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/service ./internal/controller -run TestAdminService_UpdateTripDetail -v`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/model/request/trip.go mn-backend/internal/model/response/trip.go mn-backend/internal/service/trip_service.go mn-backend/internal/service/admin_service.go mn-backend/internal/controller/admin_trip_controller.go mn-backend/internal/service/admin_service_test.go mn-backend/internal/controller/trip_controller_test.go
git commit -m "feat: 支持后台完整编辑行程字段"
```

### 任务 5：升级 Admin 行程编辑页为完整表单

**文件：**
- 修改：`mn-frontend-admin/src/features/trips/api.ts`
- 修改：`mn-frontend-admin/src/features/trips/TripEditPage.tsx`
- 修改：`mn-frontend-admin/src/features/trips/TripDetailPage.tsx`
- 修改：`mn-frontend-admin/src/test/trip-edit.test.tsx`

- [ ] **步骤 1：编写失败的完整编辑前端测试**

```tsx
it("submits full admin trip payload after confirmation", async () => {
  mockGetAdminTripDetail.mockResolvedValue({
    id: 3,
    tripType: "driver_post",
    fromText: "上海",
    toText: "杭州",
    departureDate: "2026-04-30",
    departureTime: "09:00",
    seatCount: 2,
    priceAmount: 68,
    isPriceNegotiable: false,
    contactWechat: "mingye",
    contactPhone: "13800138000",
    remark: "旧备注",
    status: "active",
  })

  renderRoute("/trips/3/edit")
  await userEvent.clear(screen.getByLabelText("起点"))
  await userEvent.type(screen.getByLabelText("起点"), "苏州")
  await userEvent.click(screen.getByRole("button", { name: /保\\s*存/ }))
  await userEvent.click(await screen.findByRole("button", { name: "确认" }))

  expect(mockUpdateAdminTrip).toHaveBeenCalledWith("3", expect.objectContaining({
    fromText: "苏州",
    remark: "旧备注",
    status: "active",
  }))
})
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-admin && npm run test -- trip-edit.test.tsx`

预期：FAIL，报错包含 `Unable to find label text 起点` 或 `mockUpdateAdminTrip not called`

- [ ] **步骤 3：编写最少实现代码**

```tsx
<Form form={form} layout="vertical">
  <Form.Item label="起点" name="fromText" rules={[{ required: true, message: "请输入起点" }]}>
    <Input />
  </Form.Item>
  <Form.Item label="终点" name="toText" rules={[{ required: true, message: "请输入终点" }]}>
    <Input />
  </Form.Item>
  <Form.Item label="出发日期" name="departureDate" rules={[{ required: true }]}>
    <Input placeholder="YYYY-MM-DD" />
  </Form.Item>
  <Form.Item label="出发时间" name="departureTime" rules={[{ required: true }]}>
    <Input placeholder="HH:mm" />
  </Form.Item>
  <Form.Item label="人数" name="seatCount" rules={[{ required: true }]}>
    <InputNumber min={1} max={6} style={{ width: "100%" }} />
  </Form.Item>
  <Form.Item label="备注" name="remark">
    <Input.TextArea rows={4} />
  </Form.Item>
  <ConfirmSubmitButton confirmTitle="确认保存修改吗？" onConfirm={() => void handleConfirm()}>
    保存
  </ConfirmSubmitButton>
</Form>
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-admin && npm run test -- trip-edit.test.tsx`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-admin/src/features/trips/api.ts mn-frontend-admin/src/features/trips/TripEditPage.tsx mn-frontend-admin/src/features/trips/TripDetailPage.tsx mn-frontend-admin/src/test/trip-edit.test.tsx
git commit -m "feat: 升级后台行程完整编辑页"
```

### 任务 6：补文档并执行全链路验证

**文件：**
- 修改：`README.md`
- 修改：`mn-backend/README.md`
- 修改：`docs/technical/api-checklist.md`

- [ ] **步骤 1：补充 SQL 导入与 P1 联调文档**

```md
## MySQL 初始化

在启动后端前，先导入初始化 SQL：

```bash
mysql -uroot -p moonick < mn-backend/docs/sql/001_init.sql
```

## P1 联调路径

1. H5 注册用户
2. H5 发布行程
3. Admin 编辑同一行程的起终点、时间、人数、联系方式、备注、状态
4. H5 回看详情与我的发布，确认字段同步
```

- [ ] **步骤 2：运行后端全量测试**

运行：`cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./...`

预期：PASS

- [ ] **步骤 3：运行 H5 全量测试与构建**

运行：`cd mn-frontend-h5 && npm run test && npm run build`

预期：PASS

- [ ] **步骤 4：运行 Admin 全量测试与构建**

运行：`cd mn-frontend-admin && npm run test && npm run build`

预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add README.md mn-backend/README.md docs/technical/api-checklist.md
git commit -m "docs: 补充 MySQL 初始化和 P1 联调说明"
```

## 自检

### 规格覆盖度

- 规格第 4 节“数据库与初始化 SQL”对应任务 1。
- 规格第 5 节“Repository 落库方案”对应任务 2、任务 3。
- 规格第 6 节“后台完整行程编辑”对应任务 4、任务 5。
- 规格第 7-9 节“实施顺序、风险控制、验证策略”对应任务 6。

### 占位符扫描

- 计划中未使用 `TODO`、`待定`、`后续实现`、`类似任务 N` 等占位写法。
- 每个任务都给出了精确文件路径、命令、预期结果和最小代码片段。

### 类型一致性

- 行程状态统一为 `active / full / closed / expired`。
- 行程类型统一为 `driver_post / passenger_post`。
- 后台完整编辑统一使用 `PUT /api/admin/v1/trips/:id`。
- MySQL 初始化脚本统一放在 `mn-backend/docs/sql/001_init.sql`。
