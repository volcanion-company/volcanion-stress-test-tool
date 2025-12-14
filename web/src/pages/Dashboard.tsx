import { useQuery } from '@tanstack/react-query'
import { testRunService } from '../services/testRunService'
import { testPlanService } from '../services/testPlanService'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import LoadingSpinner from '../components/ui/LoadingSpinner'
import { Activity, FileText, CheckCircle, TrendingUp } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function Dashboard() {
  const { data: testRuns, isLoading: runsLoading } = useQuery({
    queryKey: ['test-runs'],
    queryFn: () => testRunService.getAll(),
  })

  const { data: testPlans, isLoading: plansLoading } = useQuery({
    queryKey: ['test-plans'],
    queryFn: () => testPlanService.getAll(),
  })

  if (runsLoading || plansLoading) {
    return <LoadingSpinner />
  }

  const activeTests = testRuns?.filter(run => run.status === 'running').length || 0
  const totalPlans = testPlans?.length || 0
  const recentRuns = testRuns?.slice(0, 5) || []
  const completedToday = testRuns?.filter(run => {
    const today = new Date().toDateString()
    return new Date(run.start_at).toDateString() === today && run.status === 'completed'
  }).length || 0

  // Calculate completion rate from test runs
  const calculateCompletionRate = (): string => {
    const totalRuns = testRuns?.length || 0
    if (totalRuns === 0) return 'N/A'
    
    const completedRuns = testRuns?.filter(run => run.status === 'completed').length || 0
    const rate = (completedRuns / totalRuns) * 100
    return rate.toFixed(1) + '%'
  }
  
  const completionRate = calculateCompletionRate()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Dashboard
        </h1>
        <Link to="/test-plans/new">
          <Button>Create New Test</Button>
        </Link>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Active Tests</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                {activeTests}
              </p>
            </div>
            <Activity className="w-12 h-12 text-primary-600" />
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Total Plans</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                {totalPlans}
              </p>
            </div>
            <FileText className="w-12 h-12 text-blue-600" />
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Completed Today</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                {completedToday}
              </p>
            </div>
            <CheckCircle className="w-12 h-12 text-green-600" />
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Completion Rate</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
                {completionRate}
              </p>
            </div>
            <TrendingUp className="w-12 h-12 text-purple-600" />
          </div>
        </Card>
      </div>

      {/* Recent Test Runs */}
      <Card title="Recent Test Runs" action={
        <Link to="/test-runs">
          <Button variant="ghost" size="sm">View All</Button>
        </Link>
      }>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 dark:border-gray-700">
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">ID</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Status</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Started</th>
                <th className="text-left py-3 px-4 font-medium text-gray-900 dark:text-white">Actions</th>
              </tr>
            </thead>
            <tbody>
              {recentRuns.map((run) => (
                <tr key={run.id} className="border-b border-gray-200 dark:border-gray-700">
                  <td className="py-3 px-4 text-sm text-gray-900 dark:text-white font-mono">
                    {run.id.substring(0, 8)}...
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
                    {new Date(run.start_at).toLocaleString()}
                  </td>
                  <td className="py-3 px-4">
                    <Link to={`/test-runs/${run.id}`}>
                      <Button variant="ghost" size="sm">View</Button>
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
