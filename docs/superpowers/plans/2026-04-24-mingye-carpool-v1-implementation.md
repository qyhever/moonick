# 明叶同行 v1 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 在现有 `mn-backend` 骨架上完成明叶同行 v1 的后端、H5 业务端和 PC 管理端首版可用闭环。

**架构：** 后端继续使用单体分层 `Go + Gin`，新增用户、管理员、行程、收藏、文件上传和定时任务模块；H5 与 Admin 分别作为独立的 `React + Vite` 应用接入同一后端，通过 `/api/v1` 与 `/api/admin/v1` 两套路由和双令牌鉴权隔离。开发顺序按“后端领域能力优先、前端骨架跟进、核心页面闭环、后台运营能力、联调验收”推进。

**技术栈：** Go、Gin、MySQL、JWT、Zap、Viper、React、Vite、TypeScript、React Router、Axios、Zustand、Ant Design、Vitest、Testing Library

---

## 文件结构

### 后端 `mn-backend`

- 创建：`mn-backend/internal/model/entity/user.go`
  责任：定义 `User`、`Admin`、`Trip`、`TripFavorite`、`FileAsset` 核心实体。
- 创建：`mn-backend/internal/model/request/auth.go`
  责任：定义用户端与管理端登录、注册、刷新、资料更新、行程查询与编辑请求体。
- 创建：`mn-backend/internal/model/response/auth.go`
  责任：定义 token、用户资料、行程详情、看板汇总等对外响应结构。
- 创建：`mn-backend/internal/pkg/jwt/jwt.go`
  责任：统一签发和解析 `accessToken` / `refreshToken`。
- 创建：`mn-backend/internal/pkg/password/password.go`
  责任：封装密码哈希与校验。
- 创建：`mn-backend/internal/pkg/pagination/pagination.go`
  责任：统一 `pageNum` / `pageSize` 分页计算。
- 创建：`mn-backend/internal/pkg/timeutil/departure.go`
  责任：处理 `departure_date` + `departure_time` 组装和过期判断。
- 创建：`mn-backend/internal/pkg/storage/r2.go`
  责任：封装 Cloudflare R2 上传。
- 创建：`mn-backend/internal/middleware/auth_user.go`
  责任：校验用户 access token。
- 创建：`mn-backend/internal/middleware/auth_admin.go`
  责任：校验管理员 access token。
- 创建：`mn-backend/internal/middleware/request_id.go`
  责任：为请求注入 request id 并透传日志。
- 创建：`mn-backend/internal/controller/auth_controller.go`
  责任：用户注册、登录、刷新、当前登录态接口。
- 创建：`mn-backend/internal/controller/user_controller.go`
  责任：用户资料、联系方式、头像接口。
- 创建：`mn-backend/internal/controller/trip_controller.go`
  责任：H5 行程列表、详情、创建、编辑、状态更新接口。
- 创建：`mn-backend/internal/controller/favorite_controller.go`
  责任：收藏 toggle 和我的收藏接口。
- 创建：`mn-backend/internal/controller/admin_auth_controller.go`
  责任：管理员登录、刷新、当前登录态接口。
- 创建：`mn-backend/internal/controller/admin_trip_controller.go`
  责任：后台行程列表、详情、编辑接口。
- 创建：`mn-backend/internal/controller/admin_user_controller.go`
  责任：后台用户列表、详情、用户行程列表接口。
- 创建：`mn-backend/internal/controller/file_controller.go`
  责任：头像上传接口。
- 创建：`mn-backend/internal/service/auth_service.go`
  责任：用户/管理员认证流程。
- 创建：`mn-backend/internal/service/user_service.go`
  责任：用户资料与联系方式更新。
- 创建：`mn-backend/internal/service/trip_service.go`
  责任：H5 行程相关业务规则。
- 创建：`mn-backend/internal/service/favorite_service.go`
  责任：收藏 toggle 与收藏列表。
- 创建：`mn-backend/internal/service/admin_service.go`
  责任：看板、后台行程、后台用户查询与编辑。
- 创建：`mn-backend/internal/service/file_service.go`
  责任：文件校验、上传、资产入库。
- 创建：`mn-backend/internal/task/trip_expire_task.go`
  责任：过期行程扫描任务。
- 创建：`mn-backend/internal/repository/mysql/user_repository.go`
  责任：用户读写与资料更新。
- 创建：`mn-backend/internal/repository/mysql/admin_repository.go`
  责任：管理员查询与登录更新。
- 创建：`mn-backend/internal/repository/mysql/trip_repository.go`
  责任：行程列表、详情、创建、更新、过期扫描。
- 创建：`mn-backend/internal/repository/mysql/favorite_repository.go`
  责任：收藏关系增删查。
- 修改：`mn-backend/internal/api/router.go`
  责任：注册 `/api/v1` 与 `/api/admin/v1` 路由、鉴权中间件和上传接口。
- 修改：`mn-backend/internal/config/config.go`
  责任：追加 MySQL、JWT、R2、管理员账号等配置读取。
- 测试：`mn-backend/internal/service/auth_service_test.go`
  责任：覆盖注册、登录、刷新、管理员登录。
- 测试：`mn-backend/internal/service/user_service_test.go`
  责任：覆盖资料更新、联系方式校验、头像上传失败回退。
- 测试：`mn-backend/internal/service/trip_service_test.go`
  责任：覆盖行程创建、编辑、状态更新、列表筛选。
- 测试：`mn-backend/internal/service/favorite_service_test.go`
  责任：覆盖收藏 toggle 与收藏列表排序。
- 测试：`mn-backend/internal/service/admin_service_test.go`
  责任：覆盖看板聚合、后台行程编辑、用户详情查询。
- 测试：`mn-backend/internal/api/router_test.go`
  责任：覆盖关键接口路由与鉴权。

### H5 前端 `mn-frontend-h5`

- 创建：`mn-frontend-h5/package.json`
  责任：定义 Vite、React、Vitest、Testing Library 依赖和脚本。
- 创建：`mn-frontend-h5/src/main.tsx`
  责任：应用入口。
- 创建：`mn-frontend-h5/src/router/index.tsx`
  责任：定义公共页、受保护页、登录回跳。
- 创建：`mn-frontend-h5/src/store/auth.ts`
  责任：保存 token、当前用户、登出与恢复登录态。
- 创建：`mn-frontend-h5/src/lib/http.ts`
  责任：Axios 实例、自动 refresh、401 重放。
- 创建：`mn-frontend-h5/src/features/trips/api.ts`
  责任：封装行程相关请求。
- 创建：`mn-frontend-h5/src/features/trips/components/TripCard.tsx`
  责任：首页、我的发布、我的收藏复用卡片。
- 创建：`mn-frontend-h5/src/features/trips/pages/TripDetailPage.tsx`
  责任：行程详情与底部操作栏。
- 创建：`mn-frontend-h5/src/features/trips/pages/PublishPage.tsx`
  责任：发布行程页。
- 创建：`mn-frontend-h5/src/features/trips/pages/EditTripPage.tsx`
  责任：编辑本人行程页。
- 创建：`mn-frontend-h5/src/features/trips/pages/MyTripsPage.tsx`
  责任：我的发布列表。
- 创建：`mn-frontend-h5/src/features/trips/pages/MyFavoritesPage.tsx`
  责任：我的收藏列表。
- 创建：`mn-frontend-h5/src/features/profile/pages/ProfilePage.tsx`
  责任：个人中心主页。
- 创建：`mn-frontend-h5/src/features/profile/components/AvatarUploader.tsx`
  责任：头像上传与失败回退。
- 创建：`mn-frontend-h5/src/pages/LoginPage.tsx`
  责任：登录页。
- 创建：`mn-frontend-h5/src/pages/RegisterPage.tsx`
  责任：注册页。
- 创建：`mn-frontend-h5/src/test/auth-routing.test.tsx`
  责任：覆盖受保护路由和登录回跳。
- 创建：`mn-frontend-h5/src/test/publish-form.test.tsx`
  责任：覆盖发布表单校验。
- 创建：`mn-frontend-h5/src/test/favorite-toggle.test.tsx`
  责任：覆盖收藏 toggle。
- 创建：`mn-frontend-h5/src/test/avatar-upload.test.tsx`
  责任：覆盖头像失败回退。

### Admin 前端 `mn-frontend-admin`

- 创建：`mn-frontend-admin/package.json`
  责任：定义 Vite、React、Ant Design、Vitest、Testing Library 依赖和脚本。
- 创建：`mn-frontend-admin/src/main.tsx`
  责任：后台入口。
- 创建：`mn-frontend-admin/src/router/index.tsx`
  责任：登录态路由守卫与后台页面组织。
- 创建：`mn-frontend-admin/src/layout/AdminLayout.tsx`
  责任：侧栏、顶栏、内容区布局。
- 创建：`mn-frontend-admin/src/lib/http.ts`
  责任：Admin token 自动携带与 refresh。
- 创建：`mn-frontend-admin/src/features/auth/store.ts`
  责任：管理员 token 与当前会话存储。
- 创建：`mn-frontend-admin/src/features/auth/LoginPage.tsx`
  责任：管理员登录页。
- 创建：`mn-frontend-admin/src/features/dashboard/DashboardPage.tsx`
  责任：轻量看板。
- 创建：`mn-frontend-admin/src/features/trips/TripListPage.tsx`
  责任：后台行程列表页。
- 创建：`mn-frontend-admin/src/features/trips/TripDetailPage.tsx`
  责任：后台行程详情页。
- 创建：`mn-frontend-admin/src/features/trips/TripEditPage.tsx`
  责任：后台行程编辑页。
- 创建：`mn-frontend-admin/src/features/trips/components/TripSearchForm.tsx`
  责任：后台行程查询表单。
- 创建：`mn-frontend-admin/src/features/users/UserListPage.tsx`
  责任：后台用户列表页。
- 创建：`mn-frontend-admin/src/features/users/UserDetailPage.tsx`
  责任：后台用户详情页。
- 创建：`mn-frontend-admin/src/test/login-guard.test.tsx`
  责任：覆盖登录拦截。
- 创建：`mn-frontend-admin/src/test/dashboard-page.test.tsx`
  责任：覆盖看板展示。
- 创建：`mn-frontend-admin/src/test/trip-edit.test.tsx`
  责任：覆盖行程编辑确认。
- 创建：`mn-frontend-admin/src/test/user-readonly.test.tsx`
  责任：覆盖用户只读展示。

## 任务 1：扩展后端配置、路由和基础设施

**文件：**
- 修改：`mn-backend/internal/config/config.go`
- 修改：`mn-backend/internal/api/router.go`
- 创建：`mn-backend/internal/middleware/request_id.go`
- 创建：`mn-backend/internal/pkg/jwt/jwt.go`
- 创建：`mn-backend/internal/pkg/password/password.go`
- 创建：`mn-backend/internal/pkg/pagination/pagination.go`
- 创建：`mn-backend/internal/pkg/timeutil/departure.go`
- 测试：`mn-backend/internal/api/router_test.go`

- [ ] **步骤 1：编写失败的路由与基础设施测试**

```go
func TestSetupRouter_RegistersProtectedDomains(t *testing.T) {
	r := SetupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	adminReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/auth/me", nil)
	adminRec := httptest.NewRecorder()
	r.ServeHTTP(adminRec, adminReq)
	assert.Equal(t, http.StatusUnauthorized, adminRec.Code)
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && go test ./internal/api -run TestSetupRouter_RegistersProtectedDomains -v`
预期：FAIL，报错包含 `404` 或 `undefined: SetupRouter protected routes`

- [ ] **步骤 3：编写最少实现代码**

```go
v1 := r.Group("/api/v1")
adminV1 := r.Group("/api/admin/v1")

authController := controller.NewAuthController(authService)
adminAuthController := controller.NewAdminAuthController(authService)

v1.GET("/auth/me", middleware.RequireUserAuth(jwtManager), authController.Me)
adminV1.GET("/auth/me", middleware.RequireAdminAuth(jwtManager), adminAuthController.Me)
```

```go
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && go test ./internal/api -run TestSetupRouter_RegistersProtectedDomains -v`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/config/config.go mn-backend/internal/api/router.go mn-backend/internal/middleware/request_id.go mn-backend/internal/pkg/jwt/jwt.go mn-backend/internal/pkg/password/password.go mn-backend/internal/pkg/pagination/pagination.go mn-backend/internal/pkg/timeutil/departure.go mn-backend/internal/api/router_test.go
git commit -m "feat: 初始化明叶同行后端基础设施"
```

## 任务 2：实现用户端与管理员端认证、资料和头像上传后端能力

**文件：**
- 创建：`mn-backend/internal/model/entity/user.go`
- 创建：`mn-backend/internal/model/request/auth.go`
- 创建：`mn-backend/internal/model/response/auth.go`
- 创建：`mn-backend/internal/repository/mysql/user_repository.go`
- 创建：`mn-backend/internal/repository/mysql/admin_repository.go`
- 创建：`mn-backend/internal/service/auth_service.go`
- 创建：`mn-backend/internal/service/user_service.go`
- 创建：`mn-backend/internal/service/file_service.go`
- 创建：`mn-backend/internal/controller/auth_controller.go`
- 创建：`mn-backend/internal/controller/user_controller.go`
- 创建：`mn-backend/internal/controller/admin_auth_controller.go`
- 创建：`mn-backend/internal/controller/file_controller.go`
- 创建：`mn-backend/internal/pkg/storage/r2.go`
- 测试：`mn-backend/internal/service/auth_service_test.go`
- 测试：`mn-backend/internal/service/user_service_test.go`

- [ ] **步骤 1：编写失败的认证与资料测试**

```go
func TestAuthService_RegisterAndLogin(t *testing.T) {
	svc := newAuthServiceForTest()

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Phone: "13800138000",
		Password: "secret123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.AccessToken)

	loginResp, err := svc.Login(context.Background(), request.LoginRequest{
		Phone: "13800138000",
		Password: "secret123",
	})
	require.NoError(t, err)
	assert.Equal(t, "13800138000", loginResp.User.Phone)
}
```

```go
func TestUserService_UpdateAvatarRollbackOnUploadError(t *testing.T) {
	svc := newUserServiceForTest(uploadStub{err: errors.New("r2 down")})
	err := svc.UpdateAvatar(context.Background(), 1001, fakeFileHeader("avatar.png"))
	require.ErrorContains(t, err, "r2 down")
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && go test ./internal/service -run 'TestAuthService_RegisterAndLogin|TestUserService_UpdateAvatarRollbackOnUploadError' -v`
预期：FAIL，报错包含 `undefined: Register`、`undefined: UpdateAvatar`

- [ ] **步骤 3：编写最少实现代码**

```go
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (s *AuthService) Register(ctx context.Context, req request.RegisterRequest) (*response.AuthPayload, error) {
	if _, err := s.userRepo.FindByPhone(ctx, req.Phone); err == nil {
		return nil, errors.New("该手机号已注册，请直接登录")
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, entity.User{
		Phone:        req.Phone,
		PasswordHash: hash,
		Nickname:     "用户" + req.Phone[7:],
		Status:       "active",
	})
	if err != nil {
		return nil, err
	}

	return s.buildUserAuthPayload(user)
}
```

```go
func (s *UserService) UpdateContact(ctx context.Context, userID int64, req request.UpdateContactRequest) error {
	if req.DefaultWechat == "" && req.DefaultPhone == "" {
		return errors.New("请填写至少一种联系方式")
	}
	return s.userRepo.UpdateContact(ctx, userID, req.DefaultWechat, req.DefaultPhone)
}
```

```go
func (c *FileController) UploadAvatar(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请选择头像文件")
		return
	}
	url, err := c.fileService.UploadAvatar(ctx, mustUserID(ctx), file)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, gin.H{"url": url})
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && go test ./internal/service -run 'TestAuthService_RegisterAndLogin|TestUserService_UpdateAvatarRollbackOnUploadError' -v`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/model/entity/user.go mn-backend/internal/model/request/auth.go mn-backend/internal/model/response/auth.go mn-backend/internal/repository/mysql/user_repository.go mn-backend/internal/repository/mysql/admin_repository.go mn-backend/internal/service/auth_service.go mn-backend/internal/service/user_service.go mn-backend/internal/service/file_service.go mn-backend/internal/controller/auth_controller.go mn-backend/internal/controller/user_controller.go mn-backend/internal/controller/admin_auth_controller.go mn-backend/internal/controller/file_controller.go mn-backend/internal/pkg/storage/r2.go mn-backend/internal/service/auth_service_test.go mn-backend/internal/service/user_service_test.go
git commit -m "feat: 实现用户和管理员认证能力"
```

## 任务 3：实现行程、收藏、后台看板与过期任务后端能力

**文件：**
- 创建：`mn-backend/internal/model/entity/trip.go`
- 创建：`mn-backend/internal/model/request/trip.go`
- 创建：`mn-backend/internal/model/response/trip.go`
- 创建：`mn-backend/internal/repository/mysql/trip_repository.go`
- 创建：`mn-backend/internal/repository/mysql/favorite_repository.go`
- 创建：`mn-backend/internal/service/trip_service.go`
- 创建：`mn-backend/internal/service/favorite_service.go`
- 创建：`mn-backend/internal/service/admin_service.go`
- 创建：`mn-backend/internal/controller/trip_controller.go`
- 创建：`mn-backend/internal/controller/favorite_controller.go`
- 创建：`mn-backend/internal/controller/admin_trip_controller.go`
- 创建：`mn-backend/internal/controller/admin_user_controller.go`
- 创建：`mn-backend/internal/task/trip_expire_task.go`
- 测试：`mn-backend/internal/service/trip_service_test.go`
- 测试：`mn-backend/internal/service/favorite_service_test.go`
- 测试：`mn-backend/internal/service/admin_service_test.go`

- [ ] **步骤 1：编写失败的行程与收藏测试**

```go
func TestTripService_CreateTripRejectsSameRoute(t *testing.T) {
	svc := newTripServiceForTest()
	_, err := svc.CreateTrip(context.Background(), 1001, request.UpsertTripRequest{
		TripType: "driver_post",
		FromText: "上海",
		ToText: "上海",
		DepartureDate: "2026-04-25",
		DepartureTime: "10:00",
		SeatCount: 3,
		IsPriceNegotiable: true,
		ContactPhone: "13800138000",
	})
	require.ErrorContains(t, err, "起点和终点不能相同")
}
```

```go
func TestFavoriteService_ToggleFavorite(t *testing.T) {
	svc := newFavoriteServiceForTest()

	first, err := svc.Toggle(context.Background(), 1001, 2001)
	require.NoError(t, err)
	assert.True(t, first.Favorited)

	second, err := svc.Toggle(context.Background(), 1001, 2001)
	require.NoError(t, err)
	assert.False(t, second.Favorited)
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-backend && go test ./internal/service -run 'TestTripService_CreateTripRejectsSameRoute|TestFavoriteService_ToggleFavorite' -v`
预期：FAIL，报错包含 `undefined: CreateTrip`、`undefined: Toggle`

- [ ] **步骤 3：编写最少实现代码**

```go
func (s *TripService) validateUpsert(req request.UpsertTripRequest, now time.Time) error {
	if strings.TrimSpace(req.FromText) == strings.TrimSpace(req.ToText) {
		return errors.New("起点和终点不能相同")
	}
	departureAt, err := timeutil.CombineDeparture(req.DepartureDate, req.DepartureTime, now.Location())
	if err != nil {
		return err
	}
	if departureAt.Before(now) {
		return errors.New("出发时间不能早于当前时间")
	}
	if req.ContactWechat == "" && req.ContactPhone == "" {
		return errors.New("请填写至少一种联系方式")
	}
	return nil
}
```

```go
func (s *FavoriteService) Toggle(ctx context.Context, userID, tripID int64) (*response.ToggleFavoriteResponse, error) {
	exists, err := s.favoriteRepo.Exists(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.favoriteRepo.Delete(ctx, userID, tripID); err != nil {
			return nil, err
		}
		return &response.ToggleFavoriteResponse{Favorited: false}, nil
	}
	if err := s.favoriteRepo.Create(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return &response.ToggleFavoriteResponse{Favorited: true}, nil
}
```

```go
func (t *TripExpireTask) Run(ctx context.Context) error {
	return t.tripRepo.ExpireTripsBefore(ctx, time.Now())
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-backend && go test ./internal/service -run 'TestTripService_CreateTripRejectsSameRoute|TestFavoriteService_ToggleFavorite' -v`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-backend/internal/model/entity/trip.go mn-backend/internal/model/request/trip.go mn-backend/internal/model/response/trip.go mn-backend/internal/repository/mysql/trip_repository.go mn-backend/internal/repository/mysql/favorite_repository.go mn-backend/internal/service/trip_service.go mn-backend/internal/service/favorite_service.go mn-backend/internal/service/admin_service.go mn-backend/internal/controller/trip_controller.go mn-backend/internal/controller/favorite_controller.go mn-backend/internal/controller/admin_trip_controller.go mn-backend/internal/controller/admin_user_controller.go mn-backend/internal/task/trip_expire_task.go mn-backend/internal/service/trip_service_test.go mn-backend/internal/service/favorite_service_test.go mn-backend/internal/service/admin_service_test.go
git commit -m "feat: 实现行程收藏和后台查询能力"
```

## 任务 4：搭建 H5 应用骨架、鉴权与登录注册闭环

**文件：**
- 创建：`mn-frontend-h5/package.json`
- 创建：`mn-frontend-h5/tsconfig.json`
- 创建：`mn-frontend-h5/vite.config.ts`
- 创建：`mn-frontend-h5/src/main.tsx`
- 创建：`mn-frontend-h5/src/App.tsx`
- 创建：`mn-frontend-h5/src/router/index.tsx`
- 创建：`mn-frontend-h5/src/lib/http.ts`
- 创建：`mn-frontend-h5/src/store/auth.ts`
- 创建：`mn-frontend-h5/src/pages/LoginPage.tsx`
- 创建：`mn-frontend-h5/src/pages/RegisterPage.tsx`
- 创建：`mn-frontend-h5/src/pages/HomePage.tsx`
- 测试：`mn-frontend-h5/src/test/auth-routing.test.tsx`

- [ ] **步骤 1：编写失败的 H5 路由与回跳测试**

```tsx
it("redirects guest to login and jumps back after login", async () => {
  renderWithRouter("/publish");

  expect(await screen.findByText("登录")).toBeInTheDocument();

  await userEvent.type(screen.getByLabelText("手机号"), "13800138000");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByText("发布行程")).toBeInTheDocument();
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-h5 && npm install && npm run test -- auth-routing.test.tsx`
预期：FAIL，报错包含 `Cannot find module '../router'` 或 `Unable to find text 发布行程`

- [ ] **步骤 3：编写最少实现代码**

```tsx
const router = createBrowserRouter([
  { path: "/", element: <HomePage /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  {
    path: "/publish",
    element: (
      <RequireAuth>
        <PublishPage />
      </RequireAuth>
    ),
  },
]);
```

```ts
api.interceptors.response.use(undefined, async (error) => {
  if (error.response?.status !== 401 || error.config.__retried) {
    return Promise.reject(error);
  }

  error.config.__retried = true;
  await authStore.getState().refresh();
  return api.request(error.config);
});
```

```ts
login: async (payload) => {
  const res = await api.post("/api/v1/auth/login", payload);
  set({
    accessToken: res.data.data.accessToken,
    refreshToken: res.data.data.refreshToken,
    user: res.data.data.user,
  });
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-h5 && npm run test -- auth-routing.test.tsx`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-h5/package.json mn-frontend-h5/tsconfig.json mn-frontend-h5/vite.config.ts mn-frontend-h5/src/main.tsx mn-frontend-h5/src/App.tsx mn-frontend-h5/src/router/index.tsx mn-frontend-h5/src/lib/http.ts mn-frontend-h5/src/store/auth.ts mn-frontend-h5/src/pages/LoginPage.tsx mn-frontend-h5/src/pages/RegisterPage.tsx mn-frontend-h5/src/pages/HomePage.tsx mn-frontend-h5/src/test/auth-routing.test.tsx
git commit -m "feat: 初始化 H5 应用骨架和认证流程"
```

## 任务 5：实现 H5 首页、详情、发布/编辑、个人中心与收藏闭环

**文件：**
- 创建：`mn-frontend-h5/src/features/trips/api.ts`
- 创建：`mn-frontend-h5/src/features/trips/components/TripCard.tsx`
- 创建：`mn-frontend-h5/src/features/trips/pages/TripDetailPage.tsx`
- 创建：`mn-frontend-h5/src/features/trips/pages/PublishPage.tsx`
- 创建：`mn-frontend-h5/src/features/trips/pages/EditTripPage.tsx`
- 创建：`mn-frontend-h5/src/features/trips/pages/MyTripsPage.tsx`
- 创建：`mn-frontend-h5/src/features/trips/pages/MyFavoritesPage.tsx`
- 创建：`mn-frontend-h5/src/features/profile/pages/ProfilePage.tsx`
- 创建：`mn-frontend-h5/src/features/profile/components/AvatarUploader.tsx`
- 测试：`mn-frontend-h5/src/test/publish-form.test.tsx`
- 测试：`mn-frontend-h5/src/test/favorite-toggle.test.tsx`
- 测试：`mn-frontend-h5/src/test/avatar-upload.test.tsx`

- [ ] **步骤 1：编写失败的表单、收藏与头像测试**

```tsx
it("blocks publish when departure is earlier than now", async () => {
  render(<PublishPage />);

  await userEvent.type(screen.getByLabelText("起点"), "上海");
  await userEvent.type(screen.getByLabelText("终点"), "杭州");
  await userEvent.type(screen.getByLabelText("手机号"), "13800138000");
  await userEvent.click(screen.getByRole("button", { name: "发布" }));

  expect(await screen.findByText("出发时间不能早于当前时间")).toBeInTheDocument();
});
```

```tsx
it("toggles favorite state without toast", async () => {
  render(<TripDetailPage />);
  const button = await screen.findByRole("button", { name: "收藏" });
  await userEvent.click(button);
  expect(button).toHaveAttribute("data-favorited", "true");
});
```

```tsx
it("restores previous avatar when upload fails", async () => {
  render(<AvatarUploader initialUrl="https://cdn.example.com/a.png" />);
  await userEvent.upload(screen.getByLabelText("上传头像"), makeFile("b.png"));
  expect(await screen.findByAltText("当前头像")).toHaveAttribute("src", "https://cdn.example.com/a.png");
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-h5 && npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx`
预期：FAIL，报错包含 `PublishPage is not defined`、`Unable to find role button 收藏`

- [ ] **步骤 3：编写最少实现代码**

```tsx
const handleSubmit = async () => {
  if (form.fromText.trim() === form.toText.trim()) {
    toast.error("起点和终点不能相同");
    return;
  }
  if (new Date(`${form.departureDate} ${form.departureTime}`).getTime() < Date.now()) {
    toast.error("出发时间不能早于当前时间");
    return;
  }
  if (!form.contactWechat && !form.contactPhone) {
    toast.error("请填写至少一种联系方式");
    return;
  }
  await createTrip(form);
  navigate(`/trips/${created.id}`);
};
```

```tsx
<button
  type="button"
  data-favorited={favorited}
  disabled={trip.status !== "active"}
  onClick={handleToggleFavorite}
>
  收藏
</button>
```

```tsx
const onUpload = async (file: File) => {
  const previous = avatarUrl;
  const localPreview = URL.createObjectURL(file);
  setAvatarUrl(localPreview);
  try {
    const nextUrl = await uploadAvatar(file);
    setAvatarUrl(nextUrl);
  } catch {
    setAvatarUrl(previous);
    toast.error("服务器异常，请稍后再试");
  }
};
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-h5 && npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-h5/src/features/trips/api.ts mn-frontend-h5/src/features/trips/components/TripCard.tsx mn-frontend-h5/src/features/trips/pages/TripDetailPage.tsx mn-frontend-h5/src/features/trips/pages/PublishPage.tsx mn-frontend-h5/src/features/trips/pages/EditTripPage.tsx mn-frontend-h5/src/features/trips/pages/MyTripsPage.tsx mn-frontend-h5/src/features/trips/pages/MyFavoritesPage.tsx mn-frontend-h5/src/features/profile/pages/ProfilePage.tsx mn-frontend-h5/src/features/profile/components/AvatarUploader.tsx mn-frontend-h5/src/test/publish-form.test.tsx mn-frontend-h5/src/test/favorite-toggle.test.tsx mn-frontend-h5/src/test/avatar-upload.test.tsx
git commit -m "feat: 完成 H5 核心业务页面"
```

## 任务 6：搭建 Admin 应用骨架、登录态和首页看板

**文件：**
- 创建：`mn-frontend-admin/package.json`
- 创建：`mn-frontend-admin/tsconfig.json`
- 创建：`mn-frontend-admin/vite.config.ts`
- 创建：`mn-frontend-admin/src/main.tsx`
- 创建：`mn-frontend-admin/src/router/index.tsx`
- 创建：`mn-frontend-admin/src/layout/AdminLayout.tsx`
- 创建：`mn-frontend-admin/src/lib/http.ts`
- 创建：`mn-frontend-admin/src/features/auth/store.ts`
- 创建：`mn-frontend-admin/src/features/auth/LoginPage.tsx`
- 创建：`mn-frontend-admin/src/features/dashboard/DashboardPage.tsx`
- 测试：`mn-frontend-admin/src/test/login-guard.test.tsx`
- 测试：`mn-frontend-admin/src/test/dashboard-page.test.tsx`

- [ ] **步骤 1：编写失败的后台登录与看板测试**

```tsx
it("redirects anonymous admin user to login", async () => {
  renderWithRouter("/dashboard");
  expect(await screen.findByRole("button", { name: "登录" })).toBeInTheDocument();
});
```

```tsx
it("renders four summary cards", async () => {
  render(<DashboardPage />);
  expect(await screen.findByText("行程总数")).toBeInTheDocument();
  expect(await screen.findByText("今日新增行程数")).toBeInTheDocument();
  expect(await screen.findByText("用户总数")).toBeInTheDocument();
  expect(await screen.findByText("当前有效行程数")).toBeInTheDocument();
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-admin && npm install && npm run test -- login-guard.test.tsx dashboard-page.test.tsx`
预期：FAIL，报错包含 `Cannot find module './layout/AdminLayout'` 或 `Unable to find text 行程总数`

- [ ] **步骤 3：编写最少实现代码**

```tsx
const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: (
      <RequireAdminAuth>
        <AdminLayout />
      </RequireAdminAuth>
    ),
    children: [{ index: true, element: <DashboardPage /> }],
  },
]);
```

```tsx
export function DashboardPage() {
  const { data } = useDashboardSummary();
  return (
    <Row gutter={16}>
      <Col span={6}><Card title="行程总数">{data.totalTrips}</Card></Col>
      <Col span={6}><Card title="今日新增行程数">{data.todayTrips}</Card></Col>
      <Col span={6}><Card title="用户总数">{data.totalUsers}</Card></Col>
      <Col span={6}><Card title="当前有效行程数">{data.activeTrips}</Card></Col>
    </Row>
  );
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-admin && npm run test -- login-guard.test.tsx dashboard-page.test.tsx`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-admin/package.json mn-frontend-admin/tsconfig.json mn-frontend-admin/vite.config.ts mn-frontend-admin/src/main.tsx mn-frontend-admin/src/router/index.tsx mn-frontend-admin/src/layout/AdminLayout.tsx mn-frontend-admin/src/lib/http.ts mn-frontend-admin/src/features/auth/store.ts mn-frontend-admin/src/features/auth/LoginPage.tsx mn-frontend-admin/src/features/dashboard/DashboardPage.tsx mn-frontend-admin/src/test/login-guard.test.tsx mn-frontend-admin/src/test/dashboard-page.test.tsx
git commit -m "feat: 初始化管理后台骨架和看板"
```

## 任务 7：实现 Admin 行程管理与用户查询页面

**文件：**
- 创建：`mn-frontend-admin/src/features/trips/TripListPage.tsx`
- 创建：`mn-frontend-admin/src/features/trips/TripDetailPage.tsx`
- 创建：`mn-frontend-admin/src/features/trips/TripEditPage.tsx`
- 创建：`mn-frontend-admin/src/features/trips/components/TripSearchForm.tsx`
- 创建：`mn-frontend-admin/src/features/users/UserListPage.tsx`
- 创建：`mn-frontend-admin/src/features/users/UserDetailPage.tsx`
- 创建：`mn-frontend-admin/src/components/ConfirmSubmitButton.tsx`
- 测试：`mn-frontend-admin/src/test/trip-edit.test.tsx`
- 测试：`mn-frontend-admin/src/test/user-readonly.test.tsx`

- [ ] **步骤 1：编写失败的后台编辑与只读测试**

```tsx
it("requires confirmation before saving trip edit", async () => {
  render(<TripEditPage />);
  await userEvent.click(await screen.findByRole("button", { name: "保存" }));
  expect(await screen.findByText("确认保存修改吗？")).toBeInTheDocument();
});
```

```tsx
it("does not render destructive actions on user detail page", async () => {
  render(<UserDetailPage />);
  expect(await screen.findByText("基本资料")).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "封禁" })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "删除" })).not.toBeInTheDocument();
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-admin && npm run test -- trip-edit.test.tsx user-readonly.test.tsx`
预期：FAIL，报错包含 `TripEditPage is not defined` 或 `Unable to find text 基本资料`

- [ ] **步骤 3：编写最少实现代码**

```tsx
<Form form={form} layout="vertical" onFinish={openConfirm}>
  <Form.Item name="fromText" label="起点" rules={[{ required: true }]}>
    <Input />
  </Form.Item>
  <ConfirmSubmitButton
    confirmTitle="确认保存修改吗？"
    onConfirm={() => form.submit()}
  >
    保存
  </ConfirmSubmitButton>
</Form>
```

```tsx
export function UserDetailPage() {
  const { data } = useUserDetail();
  return (
    <>
      <Descriptions title="基本资料" bordered>
        <Descriptions.Item label="昵称">{data.nickname}</Descriptions.Item>
        <Descriptions.Item label="手机号">{data.phone}</Descriptions.Item>
      </Descriptions>
      <Card title="发布行程列表">
        <Table rowKey="id" columns={tripColumns} dataSource={data.trips} pagination={false} />
      </Card>
    </>
  );
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-admin && npm run test -- trip-edit.test.tsx user-readonly.test.tsx`
预期：PASS

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-admin/src/features/trips/TripListPage.tsx mn-frontend-admin/src/features/trips/TripDetailPage.tsx mn-frontend-admin/src/features/trips/TripEditPage.tsx mn-frontend-admin/src/features/trips/components/TripSearchForm.tsx mn-frontend-admin/src/features/users/UserListPage.tsx mn-frontend-admin/src/features/users/UserDetailPage.tsx mn-frontend-admin/src/components/ConfirmSubmitButton.tsx mn-frontend-admin/src/test/trip-edit.test.tsx mn-frontend-admin/src/test/user-readonly.test.tsx
git commit -m "feat: 完成后台行程和用户管理页面"
```

## 任务 8：联调、验收与开发文档补全

**文件：**
- 修改：`README.md`
- 修改：`mn-backend/README.md`
- 创建：`mn-frontend-h5/README.md`
- 创建：`mn-frontend-admin/README.md`
- 创建：`docs/technical/api-checklist.md`

- [ ] **步骤 1：编写失败的验收检查清单**

```md
- [ ] H5 游客访问首页、详情正常
- [ ] H5 未登录访问发布页跳转登录，登录后回跳发布页
- [ ] H5 发布成功后跳转详情页并提示“发布成功”
- [ ] H5 收藏 toggle 正常，未登录时跳转登录页
- [ ] Admin 未登录访问后台自动跳转登录页
- [ ] Admin 看板四张卡片展示正确
- [ ] Admin 可编辑行程并二次确认
```

- [ ] **步骤 2：运行全量测试与构建**

运行：`cd mn-backend && go test ./...`
预期：PASS

运行：`cd mn-frontend-h5 && npm run test && npm run build`
预期：PASS

运行：`cd mn-frontend-admin && npm run test && npm run build`
预期：PASS

- [ ] **步骤 3：补充启动与联调文档**

```md
## 本地启动顺序

1. 启动 MySQL，并导入 `docs/technical/backend.md` 中的表结构。
2. 在 `mn-backend/internal/config/dev.yml` 中配置 `mysql`、`jwt`、`r2`。
3. 执行 `cd mn-backend && make dev`。
4. 执行 `cd mn-frontend-h5 && npm run dev`。
5. 执行 `cd mn-frontend-admin && npm run dev`。
```

- [ ] **步骤 4：人工回归验证**

运行：`open http://localhost:5173 && open http://localhost:5174`
预期：H5 与 Admin 均可访问，核心路径按验收清单逐项通过

- [ ] **步骤 5：Commit**

```bash
git add README.md mn-backend/README.md mn-frontend-h5/README.md mn-frontend-admin/README.md docs/technical/api-checklist.md
git commit -m "docs: 补充明叶同行联调和验收说明"
```

## 自检

### 规格覆盖度

- `docs/requirements/overview.md` 中的双端范围、鉴权隔离、分页规范、状态枚举，分别覆盖在任务 1、任务 2、任务 3、任务 4、任务 6。
- `docs/requirements/h5.md` 中的首页、详情、发布、编辑、个人中心、登录注册、收藏与头像上传，覆盖在任务 4、任务 5。
- `docs/requirements/admin.md` 中的管理员登录、看板、行程管理、用户管理，覆盖在任务 2、任务 3、任务 6、任务 7。
- `docs/technical/backend.md` 中的数据模型、路由域、文件上传、过期任务、测试策略，覆盖在任务 1、任务 2、任务 3、任务 8。
- `docs/technical/h5.md` 与 `docs/technical/admin.md` 的页面结构、接口映射和测试重点，分别覆盖在任务 4、任务 5、任务 6、任务 7。

### 占位符扫描

- 已检查全文，未出现未落地说明、延后实现提示或跨任务引用式写法。
- 每个代码步骤均给出明确代码片段，每个验证步骤均给出明确命令与预期结果。

### 类型一致性

- 用户端路由统一使用 `/api/v1/...`，管理端统一使用 `/api/admin/v1/...`。
- 行程状态统一使用 `active / full / closed / expired`。
- 行程类型统一使用 `driver_post / passenger_post`。
- 分页参数统一使用 `pageNum`、`pageSize`。
- token 字段统一使用 `accessToken`、`refreshToken`。
