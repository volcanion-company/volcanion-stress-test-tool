import { useQuery } from '@tanstack/react-query'
import { testRunService } from '../../services/testRunService'
import { Link } from 'react-router-dom'
import Card from '../../components/ui/Card'
import LoadingSpinner from '../../components/ui/LoadingSpinner'
import { formatDistanceToNow } from 'date-fns'

export default function TestRunList() {
  const { data: testRuns, isLoading } = useQuery({
    queryKey: ['test-runs'],
    queryFn: () => testRunService.getAll(),
  })

  if (isLoading) {
    return <LoadingSpinner />
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        Test Runs
      </h1>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 dark:border-gray-700">
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Run ID</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Status</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Started</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Duration</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Actions</th>
              </tr>
            </thead>
            <tbody>
              {testRuns?.map((run) => (
                <tr key={run.id} className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td className="py-3 px-4 text-sm font-mono text-gray-900 dark:text-white">
                    {run.id.substring(0, 12)}...
                  </td>
                  <td className="py-3 px-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      run.status === 'running' ? 'bg-blue-100 text-blue-800' :
                      run.status === 'completed' ? 'bg-green-100 text-green-800' :
                      run.status === 'failed' ? 'bg-red-100 text-red-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {run.status}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400">
                    {formatDistanceToNow(new Date(run.start_at), { addSuffix: true })}
                  </td>
                  <td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400">
                    {run.end_at ? (
                      `${Math.round((new Date(run.end_at).getTime() - new Date(run.start_at).getTime()) / 1000)}s`
                    ) : (
                      'In progress'
                    )}
                  </td>
                  <td className="py-3 px-4">
                    <Link
                      to={run.status === 'running' ? `/test-runs/${run.id}/live` : `/test-runs/${run.id}`}
                      className="text-primary-600 hover:text-primary-700 text-sm font-medium"
                    >
                      {run.status === 'running' ? 'Monitor' : 'View Details'}
                    </Link>
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
