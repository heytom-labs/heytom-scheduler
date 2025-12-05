# API 模块

## 概述

API模块提供了与后端服务通信的接口，包括HTTP REST API和WebSocket实时通信。

## 文件说明

### client.ts

Axios HTTP客户端配置，提供统一的请求/响应拦截器。

**特性:**
- 自动添加请求头
- 统一错误处理
- 自动显示错误消息

### task.ts

任务相关的API接口。

**接口:**
- `listTasks(filter)` - 获取任务列表
- `getTask(id)` - 获取任务详情
- `createTask(data)` - 创建任务
- `cancelTask(id)` - 取消任务
- `getTaskStatus(id)` - 获取任务状态
- `getSubTasks(parentId)` - 获取子任务列表

### websocket.ts

WebSocket服务，提供实时通信功能。

**特性:**
- 自动重连机制
- 事件订阅/取消订阅
- 连接状态管理

**使用示例:**

```typescript
import { wsService } from '@/api/websocket'

// 连接WebSocket
wsService.connect('ws://localhost:8080/ws')

// 订阅事件
wsService.on('task_update', (message) => {
  console.log('Task updated:', message.data)
})

// 发送消息
wsService.send({ type: 'subscribe', taskId: '123' })

// 断开连接
wsService.disconnect()
```

**消息类型:**

- `task_update` - 任务更新
- `task_status_change` - 任务状态变化
- `task_created` - 任务创建
- `task_cancelled` - 任务取消

## Composables

### useWebSocket

Vue组合式函数，简化WebSocket的使用。

**使用示例:**

```vue
<script setup lang="ts">
import { useWebSocket } from '@/composables/useWebSocket'

const { connected, subscribe } = useWebSocket()

// 订阅事件
const unsubscribe = subscribe('task_update', (message) => {
  console.log('Task updated:', message.data)
})

// 组件卸载时自动清理
onUnmounted(() => {
  unsubscribe()
})
</script>

<template>
  <div>
    <el-tag :type="connected ? 'success' : 'info'">
      {{ connected ? '已连接' : '未连接' }}
    </el-tag>
  </div>
</template>
```

## 配置

### 后端地址配置

在 `vite.config.ts` 中配置代理：

```typescript
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

### WebSocket地址

默认WebSocket地址为 `ws://localhost:8080/ws`，可以在连接时指定：

```typescript
wsService.connect('ws://your-backend-url/ws')
```

## 错误处理

所有API请求的错误都会通过Element Plus的消息组件自动显示。

WebSocket连接失败会自动尝试重连，最多重试5次。
