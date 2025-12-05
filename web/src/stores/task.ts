import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { taskApi } from '@/api/task'
import type { Task, TaskFilter, CreateTaskRequest } from '@/types/task'

export const useTaskStore = defineStore('task', () => {
  const tasks = ref<Task[]>([])
  const currentTask = ref<Task | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const filter = ref<TaskFilter>({
    page: 1,
    pageSize: 20
  })

  const fetchTasks = async () => {
    loading.value = true
    try {
      const response = await taskApi.listTasks(filter.value)
      tasks.value = response.tasks
      total.value = response.total
    } finally {
      loading.value = false
    }
  }

  const fetchTask = async (id: string) => {
    loading.value = true
    try {
      currentTask.value = await taskApi.getTask(id)
      return currentTask.value
    } finally {
      loading.value = false
    }
  }

  const createTask = async (data: CreateTaskRequest) => {
    loading.value = true
    try {
      const task = await taskApi.createTask(data)
      await fetchTasks()
      return task
    } finally {
      loading.value = false
    }
  }

  const cancelTask = async (id: string) => {
    loading.value = true
    try {
      await taskApi.cancelTask(id)
      await fetchTasks()
    } finally {
      loading.value = false
    }
  }

  const updateFilter = (newFilter: Partial<TaskFilter>) => {
    filter.value = { ...filter.value, ...newFilter }
  }

  const resetFilter = () => {
    filter.value = {
      page: 1,
      pageSize: 20
    }
  }

  return {
    tasks,
    currentTask,
    total,
    loading,
    filter,
    fetchTasks,
    fetchTask,
    createTask,
    cancelTask,
    updateFilter,
    resetFilter
  }
})
