import { createFileRoute, Outlet, Navigate, Link, useParams } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'

export const Route = createFileRoute('/portal/$projectId')({
  component: PortalLayout,
})

function PortalLayout() {
  const { user, isLoading } = useAuth()
  const { projectId } = useParams({ from: '/portal/$projectId' })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/auth/login" />
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex h-14 items-center justify-between">
            <div className="flex items-center space-x-8">
              <span className="text-lg font-semibold text-gray-900">
                Feedback Portal
              </span>
              <div className="flex space-x-1">
                <Link
                  to={`/portal/${projectId}/my-feedback` as string}
                  className="px-3 py-2 text-sm font-medium rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 [&.active]:bg-gray-100 [&.active]:text-gray-900"
                  activeProps={{ className: 'bg-gray-100 text-gray-900' }}
                >
                  My Feedback
                </Link>
                <Link
                  to={`/portal/${projectId}/features` as string}
                  className="px-3 py-2 text-sm font-medium rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 [&.active]:bg-gray-100 [&.active]:text-gray-900"
                  activeProps={{ className: 'bg-gray-100 text-gray-900' }}
                >
                  Feature Requests
                </Link>
                <Link
                  to={`/portal/${projectId}/settings` as string}
                  className="px-3 py-2 text-sm font-medium rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 [&.active]:bg-gray-100 [&.active]:text-gray-900"
                  activeProps={{ className: 'bg-gray-100 text-gray-900' }}
                >
                  Settings
                </Link>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <span className="text-sm text-gray-500">{user.email}</span>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-5xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  )
}
