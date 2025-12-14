import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { testPlanService } from '../../services/testPlanService'
import Card from '../../components/ui/Card'
import LoadingSpinner from '../../components/ui/LoadingSpinner'

export default function TestPlanDetail() {
  const { id } = useParams<{ id: string }>()

  const { data: plan, isLoading } = useQuery({
    queryKey: ['test-plan', id],
    queryFn: () => testPlanService.getById(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (!plan) {
    return <div>Test plan not found</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        {plan.name}
      </h1>

      <Card title="Configuration">
        <dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Target URL</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white break-all">{plan.target_url}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Method</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white">{plan.method}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Concurrent Users</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white">{plan.concurrent_users}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Duration</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white">{plan.duration_sec} seconds</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Rate Pattern</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white">{plan.rate_pattern}</dd>
          </div>
          {plan.target_rps && (
            <div>
              <dt className="text-sm font-medium text-gray-600 dark:text-gray-400">Target RPS</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white">{plan.target_rps}</dd>
            </div>
          )}
        </dl>
      </Card>

      {Object.keys(plan.headers).length > 0 && (
        <Card title="Headers">
          <pre className="text-sm text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-900 p-4 rounded-lg overflow-auto">
            {JSON.stringify(plan.headers, null, 2)}
          </pre>
        </Card>
      )}

      {plan.body && (
        <Card title="Request Body">
          <pre className="text-sm text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-900 p-4 rounded-lg overflow-auto">
            {plan.body}
          </pre>
        </Card>
      )}
    </div>
  )
}
