import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { testRunService } from '../../services/testRunService'
import Card from '../../components/ui/Card'
import LoadingSpinner from '../../components/ui/LoadingSpinner'
import { ResponseTimeChart } from '../../components/charts/ResponseTimeChart'
import { ThroughputChart } from '../../components/charts/ThroughputChart'
import { StatusCodeChart } from '../../components/charts/StatusCodeChart'
import { VirtualUsersChart } from '../../components/charts/VirtualUsersChart'
import { useMemo } from 'react'

export default function TestRunDetail() {
  const { id } = useParams<{ id: string }>()

  const { data: testRun, isLoading: runLoading } = useQuery({
    queryKey: ['test-run', id],
    queryFn: () => testRunService.getById(id!),
    enabled: !!id,
  })

  const { data: metrics, isLoading: metricsLoading } = useQuery({
    queryKey: ['test-run-metrics', id],
    queryFn: () => testRunService.getMetrics(id!),
    enabled: !!id,
  })

  // Generate mock time-series data for charts (in real app, fetch from backend)
  const chartData = useMemo(() => {
    if (!metrics) return { responseTime: [], throughput: [], virtualUsers: [] };
    
    // Generate 20 data points spread over the test duration
    const points = 20;
    const data = Array.from({ length: points }, (_, i) => {
      const timestamp = new Date(Date.now() - (points - i) * 60000).toISOString();
      return {
        timestamp,
        responseTime: {
          timestamp,
          avg: metrics.avg_latency_ms + (Math.random() - 0.5) * 20,
          p50: metrics.p50_latency_ms + (Math.random() - 0.5) * 15,
          p95: metrics.p95_latency_ms + (Math.random() - 0.5) * 30,
          p99: metrics.p99_latency_ms + (Math.random() - 0.5) * 40,
        },
        throughput: {
          timestamp,
          rps: metrics.current_rps + (Math.random() - 0.5) * metrics.current_rps * 0.2,
          successRate: (metrics.success_requests / metrics.total_requests) * 100,
        },
        virtualUsers: {
          timestamp,
          active: Math.floor(metrics.active_workers + (Math.random() - 0.5) * 5),
          total: metrics.active_workers,
        },
      };
    });

    return {
      responseTime: data.map(d => d.responseTime),
      throughput: data.map(d => d.throughput),
      virtualUsers: data.map(d => d.virtualUsers),
    };
  }, [metrics]);

  if (runLoading || metricsLoading) {
    return <LoadingSpinner />
  }

  if (!testRun || !metrics) {
    return <div>Test run not found</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Test Run Details
        </h1>
        <span className={`px-4 py-2 rounded-lg text-sm font-medium ${
          testRun.status === 'completed' ? 'bg-green-100 text-green-800' :
          testRun.status === 'failed' ? 'bg-red-100 text-red-800' :
          'bg-gray-100 text-gray-800'
        }`}>
          {testRun.status}
        </span>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.total_requests.toLocaleString()}
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Success Rate</p>
          <p className="text-3xl font-bold text-green-600 mt-1">
            {((metrics.success_requests / metrics.total_requests) * 100).toFixed(1)}%
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Avg Response Time</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.avg_latency_ms.toFixed(0)}ms
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Throughput</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.requests_per_sec.toFixed(0)} RPS
          </p>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card title="Response Time Over Time">
          <ResponseTimeChart data={chartData.responseTime} />
        </Card>

        <Card title="Throughput Over Time">
          <ThroughputChart data={chartData.throughput} />
        </Card>

        <Card title="Status Code Distribution">
          <StatusCodeChart data={metrics.status_codes} />
        </Card>

        <Card title="Virtual Users Over Time">
          <VirtualUsersChart data={chartData.virtualUsers} />
        </Card>
      </div>

      {/* Percentiles */}
      <Card title="Response Time Percentiles">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <p className="text-sm text-gray-600 dark:text-gray-400">P50</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">
              {metrics.p50_latency_ms.toFixed(0)}ms
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-600 dark:text-gray-400">P75</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">
              {metrics.p75_latency_ms.toFixed(0)}ms
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-600 dark:text-gray-400">P95</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">
              {metrics.p95_latency_ms.toFixed(0)}ms
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-600 dark:text-gray-400">P99</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">
              {metrics.p99_latency_ms.toFixed(0)}ms
            </p>
          </div>
        </div>
      </Card>

      {/* Status Codes */}
      <Card title="Status Code Distribution">
        <div className="space-y-2">
          {Object.entries(metrics.status_codes).map(([code, count]) => (
            <div key={code} className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900 dark:text-white">
                {code}
              </span>
              <div className="flex items-center gap-4">
                <div className="w-64 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      code.startsWith('2') ? 'bg-green-600' :
                      code.startsWith('4') ? 'bg-yellow-600' :
                      'bg-red-600'
                    }`}
                    style={{ width: `${(count / metrics.total_requests) * 100}%` }}
                  />
                </div>
                <span className="text-sm text-gray-600 dark:text-gray-400 w-20 text-right">
                  {count.toLocaleString()} ({((count / metrics.total_requests) * 100).toFixed(1)}%)
                </span>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
