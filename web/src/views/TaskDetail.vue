<template>
  <div class="task-detail">
    <el-card v-loading="taskStore.loading">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <h2>任务详情</h2>
            <el-tag
              :type="connected ? 'success' : 'info'"
              size="small"
              style="margin-left: 10px"
            >
              {{ connected ? '实时更新已连接' : '实时更新未连接' }}
            </el-tag>
          </div>
          <div>
            <el-button
              v-if="task && (task.status === 'pending' || task.status === 'running')"
              type="danger"
              @click="handleCancel"
            >
              取消任务
            </el-button>
            <el-button @click="handleBack">返回</el-button>
          </div>
        </div>
      </template>

      <div v-if="task">
        <!-- Basic Information -->
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="任务ID">
            {{ task.id }}
          </el-descriptions-item>
          <el-descriptions-item label="任务名称">
            {{ task.name }}
          </el-descriptions-item>
          <el-descriptions-item label="任务描述" :span="2">
            {{ task.description }}
          </el-descriptions-item>
          <el-descriptions-item label="执行模式">
            <el-tag :type="getExecutionModeType(task.executionMode)">
              {{ getExecutionModeLabel(task.executionMode) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="任务状态">
            <el-tag :type="getStatusType(task.status)">
              {{ getStatusLabel(task.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="重试次数">
            {{ task.retryCount }}
          </el-descriptions-item>
          <el-descriptions-item label="执行节点">
            {{ task.nodeId || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(task.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatDate(task.updatedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ formatDate(task.startedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="完成时间">
            {{ formatDate(task.completedAt) }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- Schedule Configuration -->
        <el-descriptions
          v-if="task.scheduleConfig"
          title="调度配置"
          :column="2"
          border
          style="margin-top: 20px"
        >
          <el-descriptions-item
            v-if="task.scheduleConfig.scheduledTime"
            label="执行时间"
          >
            {{ formatDate(task.scheduleConfig.scheduledTime) }}
          </el-descriptions-item>
          <el-descriptions-item
            v-if="task.scheduleConfig.interval"
            label="间隔时长"
          >
            {{ task.scheduleConfig.interval }} 秒
          </el-descriptions-item>
          <el-descriptions-item
            v-if="task.scheduleConfig.cronExpr"
            label="Cron表达式"
          >
            {{ task.scheduleConfig.cronExpr }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- Callback Configuration -->
        <el-descriptions
          v-if="task.callbackConfig"
          title="回调配置"
          :column="2"
          border
          style="margin-top: 20px"
        >
          <el-descriptions-item label="回调协议">
            {{ task.callbackConfig.protocol.toUpperCase() }}
          </el-descriptions-item>
          <el-descriptions-item label="回调地址">
            {{ task.callbackConfig.url }}
          </el-descriptions-item>
          <el-descriptions-item
            v-if="task.callbackConfig.method"
            label="HTTP方法"
          >
            {{ task.callbackConfig.method }}
          </el-descriptions-item>
          <el-descriptions-item label="超时时间">
            {{ task.callbackConfig.timeout }} 秒
          </el-descriptions-item>
          <el-descriptions-item label="异步回调">
            {{ task.callbackConfig.isAsync ? '是' : '否' }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- Retry Policy -->
        <el-descriptions
          v-if="task.retryPolicy"
          title="重试策略"
          :column="2"
          border
          style="margin-top: 20px"
        >
          <el-descriptions-item label="最大重试次数">
            {{ task.retryPolicy.maxRetries }}
          </el-descriptions-item>
          <el-descriptions-item label="重试间隔">
            {{ task.retryPolicy.retryInterval }} 秒
          </el-descriptions-item>
          <el-descriptions-item label="退避因子">
            {{ task.retryPolicy.backoffFactor }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- Concurrency Limit -->
        <el-descriptions
          v-if="task.concurrencyLimit"
          title="并发控制"
          :column="2"
          border
          style="margin-top: 20px"
        >
          <el-descriptions-item label="子任务并发限制">
            {{ task.concurrencyLimit || '不限制' }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- Alert Policy -->
        <el-descriptions
          v-if="task.alertPolicy"
          title="报警策略"
          :column="2"
          border
          style="margin-top: 20px"
        >
          <el-descriptions-item label="启用失败报警">
            {{ task.alertPolicy.enableFailureAlert ? '是' : '否' }}
          </el-descriptions-item>
          <el-descriptions-item label="重试阈值">
            {{ task.alertPolicy.retryThreshold }}
          </el-descriptions-item>
          <el-descriptions-item label="超时阈值">
            {{ task.alertPolicy.timeoutThreshold }} 秒
          </el-descriptions-item>
          <el-descriptions-item label="通知渠道">
            <el-tag
              v-for="channel in task.alertPolicy.channels"
              :key="channel.type"
              style="margin-right: 5px"
            >
              {{ getChannelLabel(channel.type) }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <!-- Sub Tasks -->
        <div v-if="task.parentId === undefined" style="margin-top: 30px">
          <el-divider content-position="left">
            <h3>子任务列表</h3>
          </el-divider>
          
          <el-table
            :data="subTasks"
            :loading="loadingSubTasks"
            style="width: 100%"
          >
            <el-table-column prop="id" label="子任务ID" width="200" />
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="description" label="描述" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="retryCount" label="重试次数" width="100" />
            <el-table-column prop="createdAt" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDate(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button
                  type="primary"
                  size="small"
                  @click="handleViewSubTask(row.id)"
                >
                  查看
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- Execution History -->
        <div style="margin-top: 30px">
          <el-divider content-position="left">
            <h3>执行历史</h3>
          </el-divider>
          
          <el-timeline>
            <el-timeline-item
              v-if="task.createdAt"
              :timestamp="formatDate(task.createdAt)"
              placement="top"
            >
              <el-card>
                <h4>任务创建</h4>
                <p>任务已创建，等待执行</p>
              </el-card>
            </el-timeline-item>
            
            <el-timeline-item
              v-if="task.startedAt"
              :timestamp="formatDate(task.startedAt)"
              placement="top"
            >
              <el-card>
                <h4>开始执行</h4>
                <p>任务开始执行</p>
              </el-card>
            </el-timeline-item>
            
            <el-timeline-item
              v-if="task.retryCount > 0"
              placement="top"
            >
              <el-card>
                <h4>重试执行</h4>
                <p>任务已重试 {{ task.retryCount }} 次</p>
              </el-card>
            </el-timeline-item>
            
            <el-timeline-item
              v-if="task.completedAt"
              :timestamp="formatDate(task.completedAt)"
              :type="task.status === 'success' ? 'success' : 'danger'"
              placement="top"
            >
              <el-card>
                <h4>执行完成</h4>
                <p>
                  任务执行{{ task.status === 'success' ? '成功' : '失败' }}
                </p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>

      <el-empty v-else description="任务不存在" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTaskStore } from '@/stores/task'
import { useWebSocket } from '@/composables/useWebSocket'
import { taskApi } from '@/api/task'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Task, TaskStatus, ExecutionMode, ChannelType } from '@/types/task'

const route = useRoute()
const router = useRouter()
const taskStore = useTaskStore()
const { connected, subscribe } = useWebSocket()

const task = computed(() => taskStore.currentTask)
const subTasks = ref<Task[]>([])
const loadingSubTasks = ref(false)

onMounted(async () => {
  const taskId = route.params.id as string
  if (taskId) {
    await loadTask(taskId)
    await loadSubTasks(taskId)
    
    // Subscribe to WebSocket events for real-time updates
    const unsubscribe = subscribe('*', (message) => {
      // Reload task if it's the current task or a sub-task
      if (message.data.id === taskId || message.data.parentId === taskId) {
        loadTask(taskId)
        loadSubTasks(taskId)
      }
    })
    
    // Store unsubscribe function for cleanup
    onUnmounted(() => {
      unsubscribe()
    })
  }
})

const loadTask = async (id: string) => {
  try {
    await taskStore.fetchTask(id)
  } catch (error) {
    ElMessage.error('加载任务详情失败')
  }
}

const loadSubTasks = async (parentId: string) => {
  try {
    loadingSubTasks.value = true
    subTasks.value = await taskApi.getSubTasks(parentId)
  } catch (error) {
    console.error('Failed to load sub tasks:', error)
  } finally {
    loadingSubTasks.value = false
  }
}

const handleCancel = async () => {
  if (!task.value) return

  try {
    await ElMessageBox.confirm('确定要取消这个任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await taskStore.cancelTask(task.value.id)
    ElMessage.success('任务已取消')
    await loadTask(task.value.id)
  } catch (error) {
    // User cancelled
  }
}

const handleBack = () => {
  router.push('/tasks')
}

const handleViewSubTask = (id: string) => {
  router.push(`/tasks/${id}`)
}

const getStatusLabel = (status: TaskStatus): string => {
  const labels: Record<TaskStatus, string> = {
    pending: '等待中',
    running: '执行中',
    success: '成功',
    failed: '失败',
    cancelled: '已取消'
  }
  return labels[status] || status
}

const getStatusType = (status: TaskStatus): string => {
  const types: Record<TaskStatus, string> = {
    pending: 'info',
    running: 'warning',
    success: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return types[status] || 'info'
}

const getExecutionModeLabel = (mode: ExecutionMode): string => {
  const labels: Record<ExecutionMode, string> = {
    immediate: '立即执行',
    scheduled: '定时执行',
    interval: '固定间隔',
    cron: 'Cron'
  }
  return labels[mode] || mode
}

const getExecutionModeType = (mode: ExecutionMode): string => {
  const types: Record<ExecutionMode, string> = {
    immediate: 'success',
    scheduled: 'warning',
    interval: 'primary',
    cron: 'info'
  }
  return types[mode] || 'info'
}

const getChannelLabel = (type: ChannelType): string => {
  const labels: Record<ChannelType, string> = {
    email: '邮件',
    webhook: 'Webhook',
    sms: '短信'
  }
  return labels[type] || type
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}
</script>

<style scoped>
.task-detail {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.card-header h2 {
  margin: 0;
}

.el-descriptions {
  margin-top: 20px;
}

.el-divider h3 {
  margin: 0;
}
</style>
