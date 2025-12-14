import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import { ErrorBoundary } from './components/ui/ErrorBoundary'
import { OfflineIndicator } from './hooks/useNetworkStatus'
import { DashboardSkeleton, ListSkeleton, DetailSkeleton, FormSkeleton } from './components/ui/Skeleton'
import AppLayout from './components/layout/AppLayout'

// Lazy load pages for code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'))
const TestPlanList = lazy(() => import('./pages/TestPlans/TestPlanList'))
const TestPlanDetail = lazy(() => import('./pages/TestPlans/TestPlanDetail'))
const TestPlanWizard = lazy(() => import('./pages/TestPlans/TestPlanWizard'))
const TestRunList = lazy(() => import('./pages/TestRuns/TestRunList'))
const TestRunDetail = lazy(() => import('./pages/TestRuns/TestRunDetail'))
const TestRunLive = lazy(() => import('./pages/TestRuns/TestRunLive'))
const Login = lazy(() => import('./pages/Auth/Login'))

// Page loading wrapper with appropriate skeleton
function PageSuspense({ children, skeleton }: { children: React.ReactNode; skeleton: React.ReactNode }) {
  return (
    <Suspense fallback={skeleton}>
      {children}
    </Suspense>
  )
}

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <Routes>
          {/* Public routes */}
          <Route 
            path="/login" 
            element={
              <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div></div>}>
                <Login />
              </Suspense>
            } 
          />
          
          {/* Protected routes */}
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <ErrorBoundary>
                    <Routes>
                      <Route 
                        path="/" 
                        element={
                          <PageSuspense skeleton={<DashboardSkeleton />}>
                            <Dashboard />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-plans" 
                        element={
                          <PageSuspense skeleton={<ListSkeleton />}>
                            <TestPlanList />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-plans/new" 
                        element={
                          <PageSuspense skeleton={<FormSkeleton fields={6} />}>
                            <TestPlanWizard />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-plans/:id" 
                        element={
                          <PageSuspense skeleton={<DetailSkeleton />}>
                            <TestPlanDetail />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-runs" 
                        element={
                          <PageSuspense skeleton={<ListSkeleton />}>
                            <TestRunList />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-runs/:id" 
                        element={
                          <PageSuspense skeleton={<DetailSkeleton />}>
                            <TestRunDetail />
                          </PageSuspense>
                        } 
                      />
                      <Route 
                        path="/test-runs/:id/live" 
                        element={
                          <PageSuspense skeleton={<DetailSkeleton />}>
                            <TestRunLive />
                          </PageSuspense>
                        } 
                      />
                      <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                  </ErrorBoundary>
                </AppLayout>
              </ProtectedRoute>
            }
          />
        </Routes>
        <OfflineIndicator />
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App
