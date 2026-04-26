# 明叶同行联调与验收检查清单

## 文档目的

本清单用于记录当前仓库已经落地并可验证的联调路径，不包含尚未实现的需求点。

---

## 一、启动前检查

- [ ] 已准备 MySQL，并按后端配置可正常连接
- [ ] 已检查 `mn-backend/internal/config/dev.yml`
- [ ] 已配置 JWT `secret`
- [ ] 已配置 R2 上传参数
- [ ] 已配置管理员账号种子：
  - [ ] `auth.admin.username`
  - [ ] `auth.admin.password`
  - [ ] `auth.admin.name`

---

## 二、本地启动顺序

- [ ] 启动后端

```bash
cd mn-backend
MOONICK_ENV=dev make dev
```

- [ ] 启动 H5

```bash
cd mn-frontend-h5
npm install
npm run dev
```

- [ ] 启动 Admin

```bash
cd mn-frontend-admin
npm install
npm run dev
```

---

## 三、自动化验证

### 3.1 后端

- [x] 已执行：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

### 3.2 H5

- [x] 已执行：

```bash
cd mn-frontend-h5
npm run test
npm run build
```

### 3.3 Admin

- [x] 已执行：

```bash
cd mn-frontend-admin
npm run test
npm run build
```

说明：

- `mn-frontend-admin` 构建目前存在大包体积 warning，但构建成功

---

## 四、H5 人工回归清单

### 4.1 游客路径

- [ ] 游客访问首页正常
- [ ] 游客访问行程详情页正常
- [ ] 游客点击发布页会跳转登录页

### 4.2 登录与注册

- [ ] 用户可注册
- [ ] 用户可登录
- [ ] 未登录访问受保护页面时，登录后能按 `redirect` 回跳

### 4.3 行程

- [ ] 发布页能校验：
  - [ ] 起点和终点不能相同
  - [ ] 出发时间不能早于当前时间
  - [ ] 至少填写一种联系方式
  - [ ] 人数范围必须在 `1 ~ 6`
- [ ] 发布成功后跳转详情页
- [ ] 本人可编辑自己的行程
- [ ] 本人可把自己的行程设为满员或关闭
- [ ] 他人行程可收藏
- [ ] 已满或已关闭行程不会允许继续收藏

### 4.4 个人中心

- [ ] 头像上传成功时能更新展示
- [ ] 上传失败时会回退原头像
- [ ] 昵称更新正常
- [ ] 默认手机号更新正常
- [ ] 默认微信号更新正常
- [ ] 我的发布数量与我的收藏数量能展示

---

## 五、Admin 人工回归清单

### 5.1 登录与守卫

- [ ] 未登录访问 `/dashboard` 会跳转 `/login`
- [ ] 管理员登录成功后进入后台首页

### 5.2 看板

- [ ] 首页展示：
  - [ ] 行程总数
  - [ ] 用户总数
  - [ ] 当前有效行程数
  - [ ] 过期行程数
  - [ ] 收藏总数

### 5.3 行程管理

- [ ] 行程列表可打开
- [ ] 可按关键字和状态筛选
- [ ] 行程详情可查看
- [ ] 行程编辑页可打开
- [ ] 点击保存前会出现二次确认
- [ ] 当前可修改行程状态：
  - [ ] `active`
  - [ ] `full`
  - [ ] `closed`

### 5.4 用户管理

- [ ] 用户列表可打开
- [ ] 用户详情可查看
- [ ] 用户详情页不出现封禁、删除等越界操作

---

## 六、当前已知边界

- H5 `refresh` 接口尚未接入，当前是占位逻辑
- H5 行程表单当前未包含价格、备注等未落地字段
- Admin 行程编辑当前是状态编辑，不是完整字段编辑
- Admin 构建仍存在 chunk size warning，后续可单独做拆包优化
