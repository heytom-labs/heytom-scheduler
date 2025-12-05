# 任务调度系统 - 管理后台

基于 Vue 3 + TypeScript + Vite + Element Plus 的任务调度系统管理后台。

## 技术栈

- **Vue 3**: 渐进式 JavaScript 框架
- **TypeScript**: JavaScript 的超集，提供类型安全
- **Vite**: 下一代前端构建工具
- **Element Plus**: Vue 3 UI 组件库
- **Vue Router**: Vue.js 官方路由管理器
- **Pinia**: Vue 3 状态管理库
- **Axios**: HTTP 客户端

## 开发

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
web/
├── src/
│   ├── api/          # API 接口
│   ├── components/   # 公共组件
│   ├── router/       # 路由配置
│   ├── stores/       # Pinia 状态管理
│   ├── types/        # TypeScript 类型定义
│   ├── views/        # 页面组件
│   ├── App.vue       # 根组件
│   └── main.ts       # 入口文件
├── index.html        # HTML 模板
├── package.json      # 项目配置
├── tsconfig.json     # TypeScript 配置
└── vite.config.ts    # Vite 配置
```

## 功能特性

- ✅ 任务列表展示
- ✅ 任务创建
- ✅ 任务详情查看
- ✅ 任务状态实时更新
- ✅ 任务筛选和搜索
- ✅ 子任务管理
- ✅ 重试策略配置
- ✅ 报警策略配置
