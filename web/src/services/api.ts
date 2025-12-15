import axios from 'axios'

const API_BASE_URL = process.env.VITE_API_URL || 'http://localhost:8080/api/v1'

// Auth API
export const authApi = {
  login: (username: string, password: string) => {
    console.log(API_BASE_URL);
    return apiClient.post('/auth/login', { username, password })
  },
  // register: (email: string, password: string) => apiClient.post('/auth/register', { email, password }),
};

// Test Plan API
export const testPlanApi = {
  getAll: () => apiClient.get('/test-plans').then(res => res.data),
  getById: (id: string) => apiClient.get(`/test-plans/${id}`).then(res => res.data),
  create: (data: any) => apiClient.post('/test-plans', data).then(res => res.data),
  delete: (id: string) => apiClient.delete(`/test-plans/${id}`),
  startTest: (planId: string) => apiClient.post(`/test-plans/${planId}/start`).then(res => res.data),
};

// Test Run API
export const testRunApi = {
  getAll: () => apiClient.get('/test-runs').then(res => res.data),
  getById: (id: string) => apiClient.get(`/test-runs/${id}`).then(res => res.data),
  stop: (id: string) => apiClient.post(`/test-runs/${id}/stop`),
  getMetrics: (id: string) => apiClient.get(`/test-runs/${id}/metrics`).then(res => res.data),
  getLiveMetrics: (id: string) => apiClient.get(`/test-runs/${id}/metrics/live`).then(res => res.data),
};

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add request interceptor for auth token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Add response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
