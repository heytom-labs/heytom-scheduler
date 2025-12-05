<template>
  <div class="task-create">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>创建任务</h2>
          <el-button @click="handleBack">返回</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="140px"
      >
        <!-- Basic Information -->
        <el-divider content-position="left">基本信息</el-divider>
        
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>

        <el-form-item label="任务描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入任务描述"
          />
        </el-form-item>

        <!-- Execution Mode -->
        <el-divider content-position="left">执行模式</el-divider>

        <el-form-item label="执行模式" prop="executionMode">
          <el-radio-group v-model="form.executionMode">
            <el-radio label="immediate">立即执行</el-radio>
            <el-radio label="scheduled">定时执行</el-radio>
            <el-radio label="interval">固定间隔</el-radio>
            <el-radio label="cron">Cron表达式</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          v-if="form.executionMode === 'scheduled'"
          label="执行时间"
          prop="scheduleConfig.scheduledTime"
        >
          <el-date-picker
            v-model="form.scheduleConfig.scheduledTime"
            type="datetime"
            placeholder="选择执行时间"
            format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>

        <el-form-item
          v-if="form.executionMode === 'interval'"
          label="间隔时长(秒)"
          prop="scheduleConfig.interval"
        >
          <el-input-number
            v-model="form.scheduleConfig.interval"
            :min="1"
            placeholder="请输入间隔秒数"
          />
        </el-form-item>

        <el-form-item
          v-if="form.executionMode === 'cron'"
          label="Cron表达式"
          prop="scheduleConfig.cronExpr"
        >
          <el-input
            v-model="form.scheduleConfig.cronExpr"
            placeholder="例如: 0 0 * * * (每小时执行)"
          />
        </el-form-item>

        <!-- Callback Configuration -->
        <el-divider content-position="left">回调配置</el-divider>

        <el-form-item label="回调协议">
          <el-radio-group v-model="form.callbackConfig.protocol">
            <el-radio label="http">HTTP</el-radio>
            <el-radio label="grpc">gRPC</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="回调URL" prop="callbackConfig.url">
          <el-input
            v-model="form.callbackConfig.url"
            placeholder="请输入回调地址"
          />
        </el-form-item>

        <el-form-item v-if="form.callbackConfig.protocol === 'http'" label="HTTP方法">
          <el-select v-model="form.callbackConfig.method">
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>

        <el-form-item label="超时时间(秒)">
          <el-input-number
            v-model="form.callbackConfig.timeout"
            :min="1"
            :max="300"
          />
        </el-form-item>

        <el-form-item label="异步回调">
          <el-switch v-model="form.callbackConfig.isAsync" />
        </el-form-item>

        <!-- Retry Policy -->
        <el-divider content-position="left">重试策略</el-divider>

        <el-form-item label="最大重试次数">
          <el-input-number
            v-model="form.retryPolicy.maxRetries"
            :min="0"
            :max="10"
          />
        </el-form-item>

        <el-form-item label="重试间隔(秒)">
          <el-input-number
            v-model="form.retryPolicy.retryInterval"
            :min="1"
            :max="3600"
          />
        </el-form-item>

        <el-form-item label="退避因子">
          <el-input-number
            v-model="form.retryPolicy.backoffFactor"
            :min="1"
            :max="10"
            :step="0.1"
          />
        </el-form-item>

        <!-- Concurrency Limit -->
        <el-divider content-position="left">并发控制</el-divider>

        <el-form-item label="子任务并发限制">
          <el-input-number
            v-model="form.concurrencyLimit"
            :min="0"
            placeholder="0表示不限制"
          />
          <span class="form-tip">设置为0表示不限制并发数</span>
        </el-form-item>

        <!-- Alert Policy -->
        <el-divider content-position="left">报警策略</el-divider>

        <el-form-item label="启用失败报警">
          <el-switch v-model="form.alertPolicy.enableFailureAlert" />
        </el-form-item>

        <el-form-item label="重试阈值">
          <el-input-number
            v-model="form.alertPolicy.retryThreshold"
            :min="0"
          />
          <span class="form-tip">重试次数超过此值时触发报警</span>
        </el-form-item>

        <el-form-item label="超时阈值(秒)">
          <el-input-number
            v-model="form.alertPolicy.timeoutThreshold"
            :min="0"
          />
        </el-form-item>

        <el-form-item label="通知渠道">
          <el-checkbox-group v-model="selectedChannels">
            <el-checkbox label="email">邮件</el-checkbox>
            <el-checkbox label="webhook">Webhook</el-checkbox>
            <el-checkbox label="sms">短信</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <!-- Sub Tasks -->
        <el-divider content-position="left">子任务配置</el-divider>

        <el-form-item label="子任务">
          <el-button type="primary" @click="handleAddSubTask">添加子任务</el-button>
        </el-form-item>

        <div v-for="(subTask, index) in form.subTasks" :key="index" class="sub-task-item">
          <el-card>
            <template #header>
              <div class="sub-task-header">
                <span>子任务 {{ index + 1 }}</span>
                <el-button
                  type="danger"
                  size="small"
                  @click="handleRemoveSubTask(index)"
                >
                  删除
                </el-button>
              </div>
            </template>
            <el-form-item label="名称">
              <el-input v-model="subTask.name" placeholder="子任务名称" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input
                v-model="subTask.description"
                type="textarea"
                :rows="2"
                placeholder="子任务描述"
              />
            </el-form-item>
          </el-card>
        </div>

        <!-- Submit Buttons -->
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit">
            创建任务
          </el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button @click="handleBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useTaskStore } from '@/stores/task'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import type { CreateTaskRequest, ExecutionMode } from '@/types/task'

const router = useRouter()
const taskStore = useTaskStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const selectedChannels = ref<string[]>([])

const form = reactive<CreateTaskRequest>({
  name: '',
  description: '',
  executionMode: 'immediate' as ExecutionMode,
  scheduleConfig: {
    scheduledTime: undefined,
    interval: undefined,
    cronExpr: undefined
  },
  callbackConfig: {
    protocol: 'http',
    url: '',
    method: 'POST',
    timeout: 30,
    isAsync: false
  },
  retryPolicy: {
    maxRetries: 3,
    retryInterval: 60,
    backoffFactor: 2
  },
  concurrencyLimit: 0,
  alertPolicy: {
    enableFailureAlert: true,
    retryThreshold: 3,
    timeoutThreshold: 300,
    channels: []
  },
  subTasks: []
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { min: 1, max: 100, message: '长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入任务描述', trigger: 'blur' }
  ],
  executionMode: [
    { required: true, message: '请选择执行模式', trigger: 'change' }
  ],
  'callbackConfig.url': [
    { required: true, message: '请输入回调地址', trigger: 'blur' }
  ]
}

watch(selectedChannels, (channels) => {
  form.alertPolicy!.channels = channels.map(type => ({
    type: type as any,
    config: {}
  }))
})

const handleAddSubTask = () => {
  form.subTasks = form.subTasks || []
  form.subTasks.push({
    name: '',
    description: '',
    executionMode: 'immediate' as ExecutionMode,
    callbackConfig: {
      protocol: 'http',
      url: form.callbackConfig?.url || '',
      method: 'POST',
      timeout: 30,
      isAsync: false
    }
  })
}

const handleRemoveSubTask = (index: number) => {
  form.subTasks?.splice(index, 1)
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    // Prepare the request data
    const requestData: CreateTaskRequest = {
      ...form,
      scheduleConfig: form.executionMode === 'immediate' ? undefined : form.scheduleConfig
    }

    // Convert scheduledTime to ISO string if present
    if (requestData.scheduleConfig?.scheduledTime) {
      requestData.scheduleConfig.scheduledTime = new Date(
        requestData.scheduleConfig.scheduledTime
      ).toISOString()
    }

    await taskStore.createTask(requestData)
    ElMessage.success('任务创建成功')
    router.push('/tasks')
  } catch (error) {
    console.error('Failed to create task:', error)
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  formRef.value?.resetFields()
  form.subTasks = []
  selectedChannels.value = []
}

const handleBack = () => {
  router.back()
}
</script>

<style scoped>
.task-create {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
}

.form-tip {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.sub-task-item {
  margin-bottom: 20px;
}

.sub-task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.el-divider {
  margin: 30px 0 20px 0;
}
</style>
