import { createFileRoute, useParams, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ProjectHeader } from '../../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../../components/ConsoleNav'
import { api } from '../../../../../../lib/api'
import type { IdentifiedUser } from '../../../../../../types'

interface FeedbackItem {
  id: string
  title: string
  description: string
  type: string
  status: string
  vote_count: number
  comment_count: number
  source: string
  submitter_email?: string
  submitter_name?: string
  submitter_identifier?: string
  source_metadata?: {
    user_traits?: Record<string, string>
    [key: string]: unknown
  }
  created_at: string
  updated_at: string
}

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/users/$userId')({
  component: UserDetailPage,
})

function UserDetailPage() {
  const { projectId, userId } = useParams({ from: '/_authenticated/projects/$projectId/console/users/$userId' })
  const decodedUserId = decodeURIComponent(userId)

  const { data: users = [] } = useQuery<IdentifiedUser[]>({
    queryKey: ['users', projectId],
    queryFn: () => api.creator.listUsers(projectId) as Promise<IdentifiedUser[]>,
  })

  const user = users.find(u => u.identifier === decodedUserId)

  const { data: feedback = [], isLoading, error } = useQuery<FeedbackItem[]>({
    queryKey: ['user-feedback', projectId, decodedUserId],
    queryFn: () => api.creator.getUserFeedback(projectId, decodedUserId) as Promise<FeedbackItem[]>,
  })

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'bug': return 'bg-red-100 text-red-800'
      case 'feature': return 'bg-blue-100 text-blue-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'new': return 'bg-yellow-100 text-yellow-800'
      case 'under_review': return 'bg-purple-100 text-purple-800'
      case 'planned': return 'bg-blue-100 text-blue-800'
      case 'in_progress': return 'bg-indigo-100 text-indigo-800'
      case 'completed': return 'bg-green-100 text-green-800'
      case 'closed': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />
      <ConsoleNav projectId={projectId} />

      <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        {/* Breadcrumb */}
        <div className="mb-4">
          <Link
            to={`/projects/${projectId}/console/users` as string}
            className="text-blue-600 hover:underline text-sm"
          >
            &larr; Back to Users
          </Link>
        </div>

        {/* User Info Card */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{decodedUserId}</h1>
              <div className="mt-2 space-y-1 text-sm text-gray-500">
                {user?.email && <p>Email: {user.email}</p>}
                {user?.name && <p>Name: {user.name}</p>}
                {user && (
                  <>
                    <p>First seen: {formatDate(user.first_seen)}</p>
                    <p>Last seen: {formatDate(user.last_seen)}</p>
                    <p>Total feedback: {user.feedback_count}</p>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* User Traits */}
          {user?.traits && Object.keys(user.traits).length > 0 && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <h2 className="text-sm font-medium text-gray-700 mb-3">User Traits</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                {Object.entries(user.traits).map(([key, value]) => (
                  <div key={key} className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-500 uppercase tracking-wide">{key}</div>
                    <div className="text-sm font-medium text-gray-900 mt-1">{String(value)}</div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Feedback List */}
        <h2 className="text-xl font-bold text-gray-900 mb-4">Feedback History</h2>
        <div className="bg-white shadow rounded-lg">
          {isLoading ? (
            <div className="p-6 text-center text-gray-500">Loading...</div>
          ) : error ? (
            <div className="p-6 text-center text-red-500">Failed to load feedback</div>
          ) : feedback.length === 0 ? (
            <div className="p-6 text-center text-gray-500">No feedback from this user yet.</div>
          ) : (
            <ul className="divide-y divide-gray-200">
              {feedback.map((item) => (
                <li key={item.id} className="p-4 hover:bg-gray-50">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2 mb-1">
                        <span className={`px-2 py-0.5 text-xs font-medium rounded ${getTypeColor(item.type)}`}>
                          {item.type}
                        </span>
                        <span className={`px-2 py-0.5 text-xs font-medium rounded ${getStatusColor(item.status)}`}>
                          {item.status.replace('_', ' ')}
                        </span>
                        {item.source && item.source !== 'web' && (
                          <span className="px-2 py-0.5 text-xs font-medium rounded bg-indigo-100 text-indigo-800">
                            {item.source}
                          </span>
                        )}
                      </div>
                      <h3 className="text-sm font-medium text-gray-900 truncate">{item.title}</h3>
                      <p className="text-sm text-gray-500 truncate">{item.description}</p>
                      <div className="mt-1 text-xs text-gray-400">
                        {formatDate(item.created_at)}
                      </div>
                    </div>
                    <div className="ml-4 flex items-center space-x-4 text-sm text-gray-500">
                      <span title="Votes">{item.vote_count} votes</span>
                      <span title="Comments">{item.comment_count} comments</span>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}
