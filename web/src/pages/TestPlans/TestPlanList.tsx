import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { testPlanService } from '../../services/testPlanService'
import { Link, useNavigate } from 'react-router-dom'
import Card from '../../components/ui/Card'
import Button from '../../components/ui/Button'
import LoadingSpinner from '../../components/ui/LoadingSpinner'
import { Plus, Play, Trash2 } from 'lucide-react'

export default function TestPlanList() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: testPlans, isLoading } = useQuery({
    queryKey: ['test-plans'],
    queryFn: () => testPlanService.getAll(),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => testPlanService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['test-plans'] })
    },
  })

  const startTestMutation = useMutation({
    mutationFn: (planId: string) => testPlanService.startTest(planId),
    onSuccess: (data) => {
      navigate(`/test-runs/${data.run_id}/live`)
    },
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Test Plans
        </h1>
        <Link to="/test-plans/new">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create New Plan
          </Button>
        </Link>
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 dark:border-gray-700">
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Name</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Target URL</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Method</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Users</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Duration</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Actions</th>
              </tr>
            </thead>
            <tbody>
              {testPlans?.map((plan) => (
                <tr key={plan.id} className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td className="py-3 px-4">
                    <Link
                      to={`/test-plans/${plan.id}`}
                      className="text-primary-600 hover:text-primary-700 font-medium"
                    >
                      {plan.name}
                    </Link>
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400 max-w-xs truncate">
                    {plan.target_url}
                  </td>
                  <td className="py-3 px-4">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                      {plan.method}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-900 dark:text-white">
                    {plan.concurrent_users}
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-900 dark:text-white">
                    {plan.duration_sec}s
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => startTestMutation.mutate(plan.id)}
                        disabled={startTestMutation.isPending}
                      >
                        <Play className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          if (window.confirm('Are you sure you want to delete this test plan?')) {
                            deleteMutation.mutate(plan.id)
                          }
                        }}
                        disabled={deleteMutation.isPending}
                      >
                        <Trash2 className="w-4 h-4 text-red-600" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}
