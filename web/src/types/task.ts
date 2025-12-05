export type ExecutionMode = 'immediate' | 'scheduled' | 'interval' | 'cron'
export type TaskStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
export type CallbackProtocol = 'http' | 'grpc'
export type ServiceDiscoveryType = 'static' | 'consul' | 'etcd' | 'kubernetes'
export type ChannelType = 'email' | 'webhook' | 'sms'

export interface ScheduleConfig {
  scheduledTime?: string
  interval?: number
  cronExpr?: string
}

export interface RetryPolicy {
  maxRetries: number
  retryInterval: number
  backoffFactor: number
}

export interface CallbackConfig {
  protocol: CallbackProtocol
  url: string
  method?: string
  grpcService?: string
  grpcMethod?: string
  headers?: Record<string, string>
  timeout: number
  isAsync: boolean
  serviceName?: string
  discoveryType?: ServiceDiscoveryType
}

export interface NotificationChannel {
  type: ChannelType
  config: Record<string, string>
}

export interface AlertPolicy {
  enableFailureAlert: boolean
  retryThreshold: number
  timeoutThreshold: number
  channels: NotificationChannel[]
}

export interface Task {
  id: string
  name: string
  description: string
  parentId?: string
  executionMode: ExecutionMode
  scheduleConfig?: ScheduleConfig
  callbackConfig?: CallbackConfig
  retryPolicy?: RetryPolicy
  concurrencyLimit?: number
  alertPolicy?: AlertPolicy
  status: TaskStatus
  retryCount: number
  nodeId?: string
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
  metadata?: Record<string, string>
}

export interface CreateTaskRequest {
  name: string
  description: string
  parentId?: string
  executionMode: ExecutionMode
  scheduleConfig?: ScheduleConfig
  callbackConfig?: CallbackConfig
  retryPolicy?: RetryPolicy
  concurrencyLimit?: number
  alertPolicy?: AlertPolicy
  metadata?: Record<string, string>
  subTasks?: CreateTaskRequest[]
}

export interface TaskFilter {
  status?: TaskStatus
  executionMode?: ExecutionMode
  parentId?: string
  search?: string
  page?: number
  pageSize?: number
}

export interface TaskListResponse {
  tasks: Task[]
  total: number
  page: number
  pageSize: number
}
