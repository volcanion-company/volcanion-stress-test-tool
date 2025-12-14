import { apiClient } from './api'
import type { TestRun, Metrics } from '../types/api'

export const testRunService = {
  getAll: async (): Promise<TestRun[]> => {
    const response = await apiClient.get('/test-runs')
    return response.data
  },

  getById: async (id: string): Promise<TestRun> => {
    const response = await apiClient.get(`/test-runs/${id}`)
    return response.data
  },

  stop: async (id: string): Promise<void> => {
    await apiClient.post(`/test-runs/${id}/stop`)
  },

  getMetrics: async (id: string): Promise<Metrics> => {
    const response = await apiClient.get(`/test-runs/${id}/metrics`)
    return response.data
  },

  getLiveMetrics: async (id: string): Promise<Metrics> => {
    const response = await apiClient.get(`/test-runs/${id}/metrics/live`)
    return response.data
  },
}
