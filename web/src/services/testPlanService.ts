import { apiClient } from './api'
import type { TestPlan, CreateTestPlanRequest } from '../types/api'

export const testPlanService = {
  getAll: async (): Promise<TestPlan[]> => {
    const response = await apiClient.get('/test-plans')
    return response.data
  },

  getById: async (id: string): Promise<TestPlan> => {
    const response = await apiClient.get(`/test-plans/${id}`)
    return response.data
  },

  create: async (data: CreateTestPlanRequest): Promise<TestPlan> => {
    const response = await apiClient.post('/test-plans', data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/test-plans/${id}`)
  },

  startTest: async (planId: string): Promise<{ run_id: string }> => {
    const response = await apiClient.post(`/test-runs/start`, { plan_id: planId })
    return response.data
  },
}
