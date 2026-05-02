# Admin 固定壳布局设计

## 背景

`mn-frontend-admin` 当前使用 Ant Design `Layout` 作为后台主壳。现状是左侧菜单、顶部 Header 和页面内容处于同一滚动上下文，导致当页面内容较长时，侧边菜单和顶部 Header 会随页面一起滚动。

本次目标是将管理端调整为典型后台壳布局：

- 左侧侧边菜单固定在视口左侧
- 顶部 Header 固定在右侧主区域顶部
- 只有右侧内容区内部滚动

## 范围

本次只调整 `mn-frontend-admin/src/layout/AdminLayout.tsx` 的壳布局样式，不修改路由结构、业务页面组件和菜单项定义。

不在本次范围内：

- 响应式折叠侧边栏
- Header 吸顶阴影等视觉增强
- 页面内锚点滚动优化

## 方案对比

### 方案 A：使用 fixed 固定 Sider 和 Header

做法：

- `Sider` 使用 `position: fixed`，固定在左侧并占满 `100vh`
- `Header` 使用 `position: fixed`，固定在右侧主区域顶部
- 右侧 `Layout` 通过 `margin-left` 让出侧边栏宽度
- `Content` 通过 `margin-top` 和固定高度形成独立滚动容器

优点：

- 改动范围小
- 与当前组件结构兼容
- 风险低，便于快速验证

缺点：

- 需要手工维护侧边栏宽度和 Header 高度常量

### 方案 B：纯 flex 滚动容器

做法：

- 外层壳容器固定为 `100vh` 且 `overflow: hidden`
- 右侧主区域使用 `flex` 划分 Header 和 Content
- 只给 Content 设置滚动

优点：

- 结构更整洁，不依赖 `fixed`

缺点：

- 需要更细致处理 Ant Design `Layout` 的高度与拉伸行为
- 当前代码上回归风险更高

## 选型

采用方案 A。

理由：当前 `AdminLayout` 结构集中且简单，使用 `fixed` 能以最小改动满足需求，不引入页面结构重排。

## 详细设计

### 布局常量

在 `AdminLayout.tsx` 中提取两个布局常量：

- 侧边栏宽度：`232`
- Header 高度：`64`

常量用于统一控制 `Sider` 宽度、主内容左偏移、Header 定位和 Content 高度计算，避免多处硬编码不一致。

### 左侧固定菜单

`Sider` 改为固定定位，核心行为：

- `position: fixed`
- `left: 0`
- `top: 0`
- `bottom: 0`
- `height: 100vh`
- `overflow: auto`

这样在菜单项变多时，菜单自身仍可滚动，但不会带动页面整体滚动。

### 顶部固定 Header

`Header` 改为固定定位，核心行为：

- `position: fixed`
- `top: 0`
- `left: siderWidth`
- `right: 0`
- `height: headerHeight`
- 较高的 `z-index`

这样 Header 始终停留在顶部，右侧操作区不会因内容滚动而移走。

### 右侧主内容滚动

右侧主区域 `Layout` 通过 `margin-left: siderWidth` 与左侧固定菜单避让。

`Content` 改为独立滚动容器，核心行为：

- `margin-top: headerHeight`
- `height: calc(100vh - headerHeight)`
- `overflow-y: auto`

结果是业务页面长内容只在 Content 内部滚动，外层窗口不再承担主滚动职责。

## 风险与兼容性

- 如果某些页面依赖 `window` 级滚动监听，本次调整后行为会变化。根据当前代码检索，尚未发现管理端存在该依赖。
- Header 固定后，页面首屏如果未给 Content 预留顶部空间，会发生内容遮挡。本次通过 `margin-top` 显式规避。
- `Sider` 固定后，右侧主区域必须始终保留左边距，否则会覆盖菜单。本次通过统一常量控制。

## 测试策略

### 自动化

新增或补充 `AdminLayout` 相关测试，验证：

- 菜单项仍可正常跳转
- Header 中管理员信息和退出按钮正常渲染
- 固定布局关键样式存在，包括 `Sider` 固定、`Header` 固定、`Content` 独立滚动

### 手工验证

重点检查以下页面：

- 看板页
- 行程列表页
- 行程详情/编辑页
- 用户列表页

验证标准：

- 页面滚动时左侧菜单不动
- 页面滚动时顶部 Header 不动
- 右侧内容区能独立滚动到底部
- 顶部退出按钮可正常点击

## 成功标准

满足以下条件视为完成：

- Admin 侧边菜单固定在左侧，不随内容滚动
- Admin 顶部 Header 固定在顶部，不随内容滚动
- 右侧内容区独立滚动，页面无首屏遮挡
- 现有测试通过，构建通过
