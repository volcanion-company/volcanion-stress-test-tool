import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { testRunService } from '../../services/testRunService'
import Card from '../../components/ui/Card'
import LoadingSpinner from '../../components/ui/LoadingSpinner'
import Button from '../../components/ui/Button'
import { Activity, Wifi, WifiOff } from 'lucide-react'
import { useWebSocket } from '../../hooks/useWebSocket'
import { useState } from 'react'
import type { TestMetrics } from '../../types/api'

export default function TestRunLive() {
  const { id } = useParams<{ id: string }>()
  const [liveMetrics, setLiveMetrics] = useState<TestMetrics | null>(null)

  const { data: testRun } = useQuery({
    queryKey: ['test-run', id],
    queryFn: () => testRunService.getById(id!),
    enabled: !!id,
  })

  // WebSocket connection for live metrics
  const wsUrl = id ? `http://localhost:8080/api/test-runs/${id}/ws/metrics` : null
  const { data: wsData, isConnected } = useWebSocket<TestMetrics>(wsUrl, {
    onMessage: (data) => {
      setLiveMetrics(data)
    },
  })

  // Fallback to HTTP polling if WebSocket fails
  const { data: polledMetrics, isLoading } = useQuery({
    queryKey: ['test-run-live-metrics', id],
    queryFn: () => testRunService.getLiveMetrics(id!),
    enabled: !!id && !isConnected,
    refetchInterval: isConnected ? false : 2000,
  })

  const metrics = liveMetrics || wsData || polledMetrics

  const handleStopTest = async () => {
    if (!id) return
    try {
      await testRunService.stop(id)
    } catch (error) {
      console.error('Failed to stop test:', error)
    }
  }

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (!metrics) {
    return <div>Unable to load metrics</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <Activity className="w-8 h-8 text-primary-600 animate-pulse" />
          Live Test Monitor
        </h1>
        <div className="flex items-center gap-4">
          {/* Connection Status */}
          <div className="flex items-center gap-2">
            {isConnected ? (
              <>
                <Wifi className="w-4 h-4 text-green-600" />
                <span className="text-sm text-green-600">Live</span>
              </>
            ) : (
              <>
                <WifiOff className="w-4 h-4 text-orange-600" />
                <span className="text-sm text-orange-600">Polling</span>
              </>
            )}
          </div>
          <span className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-100 text-blue-800 animate-pulse">
            {testRun?.status || 'running'}
          </span>
          {testRun?.status === 'running' && (
            <Button variant="danger" onClick={handleStopTest}>
              Stop Test
            </Button>
          )}
        </div>
      </div>

      {/* Live Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-6">
        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.total_requests.toLocaleString()}
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Success Rate</p>
          <p className="text-3xl font-bold text-green-600 mt-1">
            {metrics.total_requests > 0 
              ? ((metrics.success_requests / metrics.total_requests) * 100).toFixed(1)
              : '0'}%
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Avg Response</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.avg_latency_ms.toFixed(0)}ms
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Current RPS</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.current_rps.toFixed(0)}
          </p>
        </Card>

        <Card className="p-6">
          <p className="text-sm text-gray-600 dark:text-gray-400">Active Workers</p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">
            {metrics.active_workers}
          </p>
        </Card>
      </div>

      {/* Live Percentiles */}
      <Card title="Response Time Percentiles (Live)">
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
                    style={{ width: `${((count as number) / metrics.total_requests) * 100}%` }}
                  />
                </div>
                <span className="text-sm text-gray-600 dark:text-gray-400 w-20 text-right">
                  {(count as number).toLocaleString()}
                </span>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
