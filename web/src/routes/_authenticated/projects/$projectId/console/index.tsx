import { useState } from 'react'
import { createFileRoute, useParams, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ProjectHeader } from '../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../components/ConsoleNav'
import { api } from '../../../../../lib/api'
import type { IdentifiedUser, SDKUser } from '../../../../../types'

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
  sdk_user?: SDKUser
  created_at: string
  updated_at: string
}

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/')({
  component: ConsoleInboxPage,
})

function ConsoleInboxPage() {
  const { projectId } = useParams({ from: '/_authenticated/projects/$projectId/console/' })
  const [userFilter, setUserFilter] = useState('')

  const { data: feedback = [], isLoading, error } = useQuery<FeedbackItem[]>({
    queryKey: ['feedback', projectId, userFilter],
    queryFn: () => api.creator.listFeedback(projectId, userFilter ? { user: userFilter } : undefined) as Promise<FeedbackItem[]>,
  })

  const { data: users = [] } = useQuery<IdentifiedUser[]>({
    queryKey: ['users', projectId],
    queryFn: () => api.creator.listUsers(projectId) as Promise<IdentifiedUser[]>,
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
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Feedback Inbox</h1>
          <div className="flex space-x-2">
            <select className="px-3 py-2 border border-gray-300 rounded-md text-sm">
              <option>All Types</option>
              <option>Bug</option>
              <option>Feature</option>
              <option>General</option>
            </select>
            <select className="px-3 py-2 border border-gray-300 rounded-md text-sm">
              <option>All Status</option>
              <option>New</option>
              <option>Under Review</option>
              <option>Planned</option>
              <option>In Progress</option>
              <option>Completed</option>
            </select>
            <select
              className="px-3 py-2 border border-gray-300 rounded-md text-sm"
              value={userFilter}
              onChange={(e) => setUserFilter(e.target.value)}
            >
              <option value="">All Users</option>
              {users.map((user) => (
                <option key={user.identifier} value={user.identifier}>
                  {user.identifier} {user.email ? `(${user.email})` : ''}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg">
          {isLoading ? (
            <div className="p-6 text-center text-gray-500">Loading...</div>
          ) : error ? (
            <div className="p-6 text-center text-red-500">Failed to load feedback</div>
          ) : feedback.length === 0 ? (
            <div className="p-6 text-center text-gray-500">No feedback items yet.</div>
          ) : (
            <ul className="divide-y divide-gray-200">
              {feedback.map((item) => (
                <li key={item.id} className="p-4 hover:bg-gray-50 cursor-pointer">
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
                      <div className="mt-1 flex items-center flex-wrap gap-x-4 gap-y-1 text-xs text-gray-400">
                        {/* Display SDK user info (preferred) or fall back to legacy fields */}
                        {item.sdk_user ? (
                          <>
                            <Link
                              to={`/projects/${projectId}/console/users/${encodeURIComponent(item.sdk_user.external_id)}` as string}
                              className="text-blue-600 hover:underline font-medium"
                            >
                              {item.sdk_user.name || item.sdk_user.email || item.sdk_user.external_id}
                            </Link>
                            {/* Show email separately only if name is displayed */}
                            {item.sdk_user.name && item.sdk_user.email && <span>{item.sdk_user.email}</span>}
                            {item.sdk_user.traits && Object.keys(item.sdk_user.traits).length > 0 && (
                              <span className="inline-flex flex-wrap gap-1">
                                {Object.entries(item.sdk_user.traits).slice(0, 3).map(([key, value]) => (
                                  <span key={key} className="px-1.5 py-0.5 bg-gray-100 rounded text-gray-600">
                                    {key}: {String(value)}
                                  </span>
                                ))}
                                {Object.keys(item.sdk_user.traits).length > 3 && (
                                  <span className="px-1.5 py-0.5 text-gray-400">
                                    +{Object.keys(item.sdk_user.traits).length - 3} more
                                  </span>
                                )}
                              </span>
                            )}
                          </>
                        ) : (
                          <>
                            {item.submitter_identifier && (
                              <Link
                                to={`/projects/${projectId}/console/users/${encodeURIComponent(item.submitter_identifier)}` as string}
                                className="text-blue-600 hover:underline font-medium"
                              >
                                {item.submitter_identifier}
                              </Link>
                            )}
                            {item.submitter_email && <span>{item.submitter_email}</span>}
                            {item.source_metadata?.user_traits && (
                              <span className="inline-flex flex-wrap gap-1">
                                {Object.entries(item.source_metadata.user_traits).slice(0, 3).map(([key, value]) => (
                                  <span key={key} className="px-1.5 py-0.5 bg-gray-100 rounded text-gray-600">
                                    {key}: {value}
                                  </span>
                                ))}
                                {Object.keys(item.source_metadata.user_traits).length > 3 && (
                                  <span className="px-1.5 py-0.5 text-gray-400">
                                    +{Object.keys(item.source_metadata.user_traits).length - 3} more
                                  </span>
                                )}
                              </span>
                            )}
                          </>
                        )}
                        <span>{formatDate(item.created_at)}</span>
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
