import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { testPlanService } from '../../services/testPlanService'
import Card from '../../components/ui/Card'
import Button from '../../components/ui/Button'
import type { CreateTestPlanRequest } from '../../types/api'

export default function TestPlanWizard() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState<CreateTestPlanRequest>({
    name: '',
    target_url: '',
    method: 'GET',
    concurrent_users: 10,
    duration_sec: 60,
    timeout_ms: 5000,
    rate_pattern: 'fixed',
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTestPlanRequest) => testPlanService.create(data),
    onSuccess: (data) => {
      navigate(`/test-plans/${data.id}`)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createMutation.mutate(formData)
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        Create Test Plan
      </h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card title="Basic Configuration">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Test Name
              </label>
              <input
                type="text"
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Target URL
              </label>
              <input
                type="url"
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                value={formData.target_url}
                onChange={(e) => setFormData({ ...formData, target_url: e.target.value })}
                placeholder="https://api.example.com/endpoint"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                HTTP Method
              </label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                value={formData.method}
                onChange={(e) => setFormData({ ...formData, method: e.target.value })}
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
                <option value="PATCH">PATCH</option>
              </select>
            </div>
          </div>
        </Card>

        <Card title="Load Configuration">
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Concurrent Users
                </label>
                <input
                  type="number"
                  min="1"
                  max="10000"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  value={formData.concurrent_users}
                  onChange={(e) => setFormData({ ...formData, concurrent_users: parseInt(e.target.value) })}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Duration (seconds)
                </label>
                <input
                  type="number"
                  min="1"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  value={formData.duration_sec}
                  onChange={(e) => setFormData({ ...formData, duration_sec: parseInt(e.target.value) })}
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Rate Pattern
              </label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                value={formData.rate_pattern}
                onChange={(e) => setFormData({ ...formData, rate_pattern: e.target.value })}
              >
                <option value="fixed">Fixed RPS</option>
                <option value="step">Step Pattern</option>
                <option value="ramp">Ramp Pattern</option>
                <option value="spike">Spike Pattern</option>
              </select>
            </div>

            {formData.rate_pattern === 'fixed' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Target RPS
                </label>
                <input
                  type="number"
                  min="1"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  value={formData.target_rps || ''}
                  onChange={(e) => setFormData({ ...formData, target_rps: parseInt(e.target.value) || undefined })}
                  placeholder="Leave empty for unlimited"
                />
              </div>
            )}
          </div>
        </Card>

        <div className="flex justify-end gap-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => navigate('/test-plans')}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={createMutation.isPending}>
            {createMutation.isPending ? 'Creating...' : 'Create Test Plan'}
          </Button>
        </div>
      </form>
    </div>
  )
}
