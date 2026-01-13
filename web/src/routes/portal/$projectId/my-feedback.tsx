import { createFileRoute, useParams } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { PortalFeedbackSummary } from '@/types'

export const Route = createFileRoute('/portal/$projectId/my-feedback')({
  component: MyFeedbackPage,
})

function MyFeedbackPage() {
  const { projectId } = useParams({ from: '/portal/$projectId/my-feedback' })

  const { data: feedback = [], isLoading, error } = useQuery<PortalFeedbackSummary[]>({
    queryKey: ['portal', 'my-feedback', projectId],
    queryFn: () => api.portal.listMyFeedback(projectId) as Promise<PortalFeedbackSummary[]>,
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'new': return 'bg-yellow-100 text-yellow-800'
      case 'under_review': return 'bg-purple-100 text-purple-800'
      case 'planned': return 'bg-blue-100 text-blue-800'
      case 'in_progress': return 'bg-indigo-100 text-indigo-800'
      case 'completed': return 'bg-green-100 text-green-800'
      case 'declined': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto" />
        <p className="mt-4 text-gray-500">Loading your feedback...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">Failed to load feedback. Please try again.</p>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">My Feedback</h1>

      {feedback.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <p className="text-gray-500">You haven&apos;t submitted any feedback yet.</p>
          <p className="text-sm text-gray-400 mt-2">
            Feedback you submit through the app will appear here.
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 divide-y divide-gray-200">
          {feedback.map((item) => (
            <div key={item.id} className="p-4 hover:bg-gray-50">
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${getStatusColor(item.status)}`}>
                      {item.status.replace('_', ' ')}
                    </span>
                    <span className="text-xs text-gray-400">
                      {item.type}
                    </span>
                  </div>
                  <h3 className="text-sm font-medium text-gray-900">{item.title}</h3>
                  <p className="text-sm text-gray-500 line-clamp-2 mt-1">{item.description}</p>
                  <div className="mt-2 flex items-center gap-4 text-xs text-gray-400">
                    <span>{formatDate(item.created_at)}</span>
                    <span>{item.vote_count} votes</span>
                    <span>{item.comment_count} comments</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
