import client from './client'
import type { Task, CreateTaskRequest, TaskFilter, TaskListResponse } from '@/types/task'

export const taskApi = {
  // Get task list
  listTasks(filter: TaskFilter = {}): Promise<TaskListResponse> {
    return client.get('/tasks', { params: filter })
  },

  // Get task by ID
  getTask(id: string): Promise<Task> {
    return client.get(`/tasks/${id}`)
  },

  // Create task
  createTask(data: CreateTaskRequest): Promise<Task> {
    return client.post('/tasks', data)
  },

  // Cancel task
  cancelTask(id: string): Promise<void> {
    return client.post(`/tasks/${id}/cancel`)
  },

  // Get task status
  getTaskStatus(id: string): Promise<{ status: string; retryCount: number }> {
    return client.get(`/tasks/${id}/status`)
  },

  // Get sub-tasks
  getSubTasks(parentId: string): Promise<Task[]> {
    return client.get(`/tasks/${parentId}/subtasks`)
  }
}
