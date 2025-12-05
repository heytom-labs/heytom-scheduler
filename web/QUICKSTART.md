# 快速启动指南

## 前置要求

- Node.js 18+ 
- npm 或 yarn
- 后端服务运行在 `http://localhost:8080`

## 安装步骤

### 1. 安装依赖

```bash
cd web
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

服务器将在 http://localhost:3000 启动

### 3. 访问应用

在浏览器中打开 http://localhost:3000

## 功能演示

### 创建任务

1. 点击"创建任务"按钮
2. 填写任务信息：
   - 任务名称：例如 "测试任务"
   - 任务描述：例如 "这是一个测试任务"
3. 选择执行模式：
   - 立即执行：任务创建后立即执行
   - 定时执行：选择未来的某个时间点
   - 固定间隔：设置间隔秒数（如60秒）
   - Cron：输入Cron表达式（如 `0 * * * *` 每小时执行）
4. 配置回调：
   - 回调协议：HTTP 或 gRPC
   - 回调地址：例如 `http://localhost:9000/callback`
   - 超时时间：默认30秒
5. 配置重试策略（可选）：
   - 最大重试次数：默认3次
   - 重试间隔：默认60秒
   - 退避因子：默认2（指数退避）
6. 配置并发限制（可选）：
   - 子任务并发数：0表示不限制
7. 配置报警策略（可选）：
   - 启用失败报警
   - 设置重试阈值
   - 选择通知渠道
8. 添加子任务（可选）：
   - 点击"添加子任务"
   - 填写子任务名称和描述
9. 点击"创建任务"

### 查看任务列表

1. 在任务列表页面可以看到所有任务
2. 使用筛选功能：
   - 搜索框：输入任务名称或描述
   - 状态筛选：选择特定状态
   - 执行模式筛选：选择特定执行模式
3. 点击"搜索"应用筛选
4. 点击"重置"清除筛选条件
5. 使用分页控件浏览更多任务

### 查看任务详情

1. 在任务列表中点击任务行或"查看"按钮
2. 查看任务的完整信息：
   - 基本信息
   - 调度配置
   - 回调配置
   - 重试策略
   - 并发控制
   - 报警策略
   - 子任务列表
   - 执行历史
3. 如果任务正在执行，可以点击"取消任务"

### 实时更新

- 页面右上角显示WebSocket连接状态
- 绿色标签表示"实时更新已连接"
- 灰色标签表示"实时更新未连接"
- 当任务状态变化时，页面会自动刷新

## 常见问题

### Q: 无法连接到后端API

**A:** 确保后端服务正在运行：
```bash
# 在项目根目录
go run cmd/scheduler/main.go
```

### Q: WebSocket连接失败

**A:** 检查后端是否支持WebSocket端点 `/ws`

### Q: 页面显示错误

**A:** 打开浏览器开发者工具（F12）查看控制台错误信息

### Q: 任务创建失败

**A:** 检查：
1. 回调地址是否正确
2. Cron表达式格式是否正确
3. 所有必填字段是否已填写

## 开发模式

### 热重载

开发服务器支持热重载，修改代码后会自动刷新页面。

### 调试

使用Vue DevTools浏览器扩展进行调试：
- 查看组件树
- 检查Pinia状态
- 追踪事件

### API代理

开发服务器配置了API代理：
- `/api/*` 请求会被代理到 `http://localhost:8080`

## 生产构建

### 构建

```bash
npm run build
```

构建产物在 `dist/` 目录

### 预览

```bash
npm run preview
```

在 http://localhost:4173 预览生产构建

### 部署

将 `dist/` 目录部署到Web服务器（如Nginx、Apache）

**Nginx配置示例：**

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 技术支持

如有问题，请查看：
- [完整实现文档](./IMPLEMENTATION.md)
- [API文档](./src/api/README.md)
- [项目README](./README.md)
