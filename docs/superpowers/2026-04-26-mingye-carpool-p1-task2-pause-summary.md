# 明叶同行 P1 任务 2 暂停前进度摘要

生成时间：2026-04-26
分支：`feature`
阶段：P1 任务 2（用户与管理员仓储切换到真实 MySQL）暂停中

## 当前结论

- 任务 1 已完成
- 任务 2 已完成一轮实现
- 任务 2 已通过规格审查
- 任务 2 尚未完成代码质量修复闭环，因此当前不能宣称任务 2 已完成

这意味着：

- 当前工作区里的任务 2 代码已经不再是最初子代理交付版本
- 用户仓储、管理员仓储和 `router` 注入链已经基本切到真实 MySQL 路径
- 我已经亲自完成过任务 2 当前版本的关键验证
- 但代码质量复审抓到了 2 条必须修复的真实 bug 和 1 条高风险测试偏差
- 在这些问题修完并回归通过之前，任务 2 仍处于“已实现、未闭环”状态

## 本轮已完成内容

### 1. 用户仓储支持真实 MySQL 路径

已修改：

- `mn-backend/internal/repository/mysql/user_repository.go`
- `mn-backend/internal/repository/mysql/user_repository_test.go`

当前已具备：

- `Create`
- `FindByPhone`
- `FindByID`
- `UpdateProfile`
- `UpdateContact`
- `UpdateAvatarURL`
- `List`
- `Count`

当前实现边界：

- 传入非空 `db` 时，`UserRepository` 走真实 MySQL 查询和写入
- 显式传入 `nil` 时，强制回退到内存实现
- 无参数调用时，仍保留旧兼容行为，允许继续使用共享 DB 或内存回退
- 重复手机号仍映射为 `mysql.ErrUserPhoneAlreadyExists`
- 查无数据仍保持 `nil, nil`

### 2. 管理员仓储支持真实 MySQL 查询与 seed upsert

已修改：

- `mn-backend/internal/repository/mysql/admin_repository.go`
- `mn-backend/internal/repository/mysql/admin_repository_test.go`

当前已具备：

- `FindByUsername`
- `FindByID`
- `Upsert`

当前实现边界：

- 新增 `NewAdminRepositoryWithDB(db)`，显式按传入 `db` 决定是否走真实 MySQL
- `NewAdminRepository()` 保留为纯内存构造器
- 管理员 seed 在有库时通过 `repo.Upsert` 落库
- `db == nil` 时，仍保留内存回退路径

### 3. Router 注入链已改成显式数据库注入

已修改：

- `mn-backend/internal/api/router.go`

当前行为：

- `initMySQLDB(cfg)` 负责初始化 DB 并设置共享 DB
- `userRepo := mysql.NewUserRepository(db)`
- `adminRepo := newAdminRepositoryFromConfig(cfg, db)`
- `newAdminRepositoryFromConfig(cfg, db)` 在有库时通过 `mysql.NewAdminRepositoryWithDB(db)` 构造管理员仓储
- 管理员 seed 不再依赖“内存 seed 仓储”作为登录数据源

### 4. 测试辅助已补齐到当前阶段

已新增：

- `mn-backend/internal/repository/mysql/repository_db_test_helper_test.go`

用途：

- 用自定义 `database/sql` 假驱动覆盖任务 2 的仓储 DB 路径
- 支撑用户与管理员仓储的数据库路径测试

## 我已实际验证通过的命令

以下命令已由我亲自执行并通过：

```bash
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestUserRepository_|TestAdminRepository_' -v
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/service -run 'TestAuthService|TestAdminService' -v
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/api -run 'Test.*Router|TestSetupRouter' -v
```

当前可以确认：

- `repository/mysql` 当前任务 2 覆盖范围内的测试通过
- `service` 层认证与后台摘要等受影响逻辑未被打坏
- `router` 与管理员登录相关路径在当前实现下通过

## 当前未闭环问题

代码质量复审给出的结论是：任务 2 还不能结束，至少需要补这 3 项，其中前 2 项必须修。

### 1. 用户更新“值未变化”时可能被误判为不存在

涉及文件：

- `mn-backend/internal/repository/mysql/user_repository.go`

问题说明：

- `updateUserRowsAffected` 目前把 `RowsAffected() == 0` 直接映射为 `ErrUserNotFound`
- 在 MySQL 中，这并不等价于“记录不存在”
- 当用户提交未变化的昵称、联系方式、头像，或者 `updated_at` 同秒未变化时，也可能返回 `0`

真实风险：

- 现有用户更新资料时可能被错误返回“用户不存在”

### 2. 内存版管理员 Upsert 会留下旧用户名索引

涉及文件：

- `mn-backend/internal/repository/mysql/admin_repository.go`

问题说明：

- 内存版 `AdminRepository.Upsert` 更新同一 `ID` 的管理员用户名时，只写入新用户名索引
- 旧用户名索引没有删除

真实风险：

- 同一管理员改过用户名后，旧用户名仍可能查到这条管理员
- 无 MySQL 配置时，当前 `router` 仍可能走到这条内存路径

### 3. 仓储测试替身对真实行为模拟偏乐观

涉及文件：

- `mn-backend/internal/repository/mysql/repository_db_test_helper_test.go`
- `mn-backend/internal/repository/mysql/user_repository_test.go`
- `mn-backend/internal/repository/mysql/admin_repository_test.go`

问题说明：

- update 相关替身现在只要记录存在就固定返回 `rowsAffected = 1`
- 这样测不出第 1 条“值未变化却误判不存在”的真实问题
- 列表查询辅助对部分 `limit / offset` 场景也偏乐观

真实风险：

- 测试会掩盖真实 MySQL 行为差异
- 后续补边界测试时容易出现“假驱动通过，真实 MySQL 出错”的情况

## 当前涉及文件

- `mn-backend/internal/api/router.go`
- `mn-backend/internal/repository/mysql/user_repository.go`
- `mn-backend/internal/repository/mysql/admin_repository.go`
- `mn-backend/internal/repository/mysql/user_repository_test.go`
- `mn-backend/internal/repository/mysql/admin_repository_test.go`
- `mn-backend/internal/repository/mysql/repository_db_test_helper_test.go`

## 下次恢复建议

建议按以下顺序继续：

1. 先完成任务 2 的代码质量修复：

- 修正 `updateUserRowsAffected` 对 `RowsAffected() == 0` 的错误语义判断
- 修正内存版 `AdminRepository.Upsert` 的旧用户名索引污染
- 调整仓储测试替身，使其能暴露上述问题

2. 再补对应测试：

- “更新相同值不应误判用户不存在”
- “管理员改用户名后旧用户名失效、新用户名生效”

3. 修完后重新执行：

```bash
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/repository/mysql -run 'TestUserRepository_|TestAdminRepository_' -v
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/service -run 'TestAuthService|TestAdminService' -v
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./internal/api -run 'Test.*Router|TestSetupRouter' -v
```

4. 验证通过后，再重新做任务 2 的代码质量复审闭环
5. 只有代码质量复审通过后，才能把任务 2 标记为完成并进入任务 3

## 备注

- 当前环境仍然不能稳定写入 `.git/index.lock`，无法代为提交 commit
- `.codex/config.toml` 仍是工作区里的无关改动，不属于本轮 P1
- 本文档只记录 P1 任务 2 在暂停时的局部状态，不替代以下全局文档：
  - `docs/superpowers/specs/2026-04-26-mingye-carpool-p1-mysql-design.md`
  - `docs/superpowers/plans/2026-04-26-mingye-carpool-p1-mysql-implementation.md`
  - `docs/superpowers/2026-04-26-mingye-carpool-final-summary.md`
