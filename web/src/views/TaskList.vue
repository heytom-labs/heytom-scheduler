<template>
  <div class="task-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <h2>任务列表</h2>
            <el-tag
              :type="connected ? 'success' : 'info'"
              size="small"
              style="margin-left: 10px"
            >
              {{ connected ? '实时更新已连接' : '实时更新未连接' }}
            </el-tag>
          </div>
          <el-button type="primary" @click="handleCreate">创建任务</el-button>
        </div>
      </template>

      <!-- Filter Section -->
      <div class="filter-section">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="搜索">
            <el-input
              v-model="filterForm.search"
              placeholder="搜索任务名称或描述"
              clearable
              @clear="handleSearch"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select
              v-model="filterForm.status"
              placeholder="选择状态"
              clearable
              @change="handleSearch"
            >
              <el-option label="等待中" value="pending" />
              <el-option label="执行中" value="running" />
              <el-option label="成功" value="success" />
              <el-option label="失败" value="failed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="执行模式">
            <el-select
              v-model="filterForm.executionMode"
              placeholder="选择执行模式"
              clearable
              @change="handleSearch"
            >
              <el-option label="立即执行" value="immediate" />
              <el-option label="定时执行" value="scheduled" />
              <el-option label="固定间隔" value="interval" />
              <el-option label="Cron" value="cron" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- Task Table -->
      <el-table
        :data="taskStore.tasks"
        :loading="taskStore.loading"
        style="width: 100%"
        @row-click="handleRowClick"
      >
        <el-table-column prop="id" label="任务ID" width="200" />
        <el-table-column prop="name" label="任务名称" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="executionMode" label="执行模式" width="120">
          <template #default="{ row }">
            <el-tag :type="getExecutionModeType(row.executionMode)">
              {{ getExecutionModeLabel(row.executionMode) }}
            </el-tag>
          </template>
        </el-table-column>
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
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click.stop="handleView(row.id)"
            >
              查看
            </el-button>
            <el-button
              v-if="row.status === 'pending' || row.status === 'running'"
              type="danger"
              size="small"
              @click.stop="handleCancel(row.id)"
            >
              取消
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="taskStore.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useTaskStore } from '@/stores/task'
import { useWebSocket } from '@/composables/useWebSocket'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { TaskStatus, ExecutionMode } from '@/types/task'

const router = useRouter()
const taskStore = useTaskStore()
const { connected, subscribe } = useWebSocket()

const currentPage = ref(1)
const pageSize = ref(20)

const filterForm = reactive({
  search: '',
  status: undefined as TaskStatus | undefined,
  executionMode: undefined as ExecutionMode | undefined
})

onMounted(() => {
  loadTasks()
  
  // Subscribe to WebSocket events for real-time updates
  const unsubscribe = subscribe('*', (message) => {
    console.log('Received WebSocket message:', message)
    
    // Refresh task list when any task event occurs
    if (['task_update', 'task_status_change', 'task_created', 'task_cancelled'].includes(message.type)) {
      loadTasks()
    }
  })
  
  // Store unsubscribe function for cleanup
  onUnmounted(() => {
    unsubscribe()
  })
})

const loadTasks = async () => {
  await taskStore.fetchTasks()
}

const handleSearch = () => {
  currentPage.value = 1
  taskStore.updateFilter({
    search: filterForm.search || undefined,
    status: filterForm.status,
    executionMode: filterForm.executionMode,
    page: currentPage.value,
    pageSize: pageSize.value
  })
  loadTasks()
}

const handleReset = () => {
  filterForm.search = ''
  filterForm.status = undefined
  filterForm.executionMode = undefined
  currentPage.value = 1
  taskStore.resetFilter()
  loadTasks()
}

const handlePageChange = (page: number) => {
  taskStore.updateFilter({ page })
  loadTasks()
}

const handleSizeChange = (size: number) => {
  currentPage.value = 1
  taskStore.updateFilter({ page: 1, pageSize: size })
  loadTasks()
}

const handleCreate = () => {
  router.push('/tasks/create')
}

const handleView = (id: string) => {
  router.push(`/tasks/${id}`)
}

const handleRowClick = (row: any) => {
  handleView(row.id)
}

const handleCancel = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要取消这个任务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await taskStore.cancelTask(id)
    ElMessage.success('任务已取消')
  } catch (error) {
    // User cancelled
  }
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

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}
</script>

<style scoped>
.task-list {
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

.filter-section {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.el-table {
  cursor: pointer;
}
</style>
