import { useState } from 'react'
import { createFileRoute, useParams, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ProjectHeader } from '../../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../../components/ConsoleNav'
import { api } from '../../../../../../lib/api'
import type { IdentifiedUser } from '../../../../../../types'

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/users/')({
  component: UsersPage,
})

function UsersPage() {
  const { projectId } = useParams({ from: '/_authenticated/projects/$projectId/console/users/' })
  const [search, setSearch] = useState('')

  const { data: users = [], isLoading, error } = useQuery<IdentifiedUser[]>({
    queryKey: ['users', projectId, search],
    queryFn: () => api.creator.listUsers(projectId, search ? { search } : undefined) as Promise<IdentifiedUser[]>,
  })

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />
      <ConsoleNav projectId={projectId} />

      <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Identified Users</h1>
          <div className="flex space-x-2">
            <input
              type="text"
              placeholder="Search users..."
              className="px-3 py-2 border border-gray-300 rounded-md text-sm w-64"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>

        <div className="bg-white shadow rounded-lg overflow-hidden">
          {isLoading ? (
            <div className="p-6 text-center text-gray-500">Loading...</div>
          ) : error ? (
            <div className="p-6 text-center text-red-500">Failed to load users</div>
          ) : users.length === 0 ? (
            <div className="p-6 text-center text-gray-500">
              {search ? 'No users found matching your search.' : 'No identified users yet.'}
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    User ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Traits
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Feedback
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Seen
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {users.map((user) => (
                  <tr key={user.identifier} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Link
                        to={`/projects/${projectId}/console/users/${encodeURIComponent(user.identifier)}` as string}
                        className="text-blue-600 hover:underline font-medium"
                      >
                        {user.name || user.email || user.identifier}
                      </Link>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {user.email || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {user.name || '-'}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {user.traits ? (
                          Object.entries(user.traits).slice(0, 3).map(([key, value]) => (
                            <span
                              key={key}
                              className="px-2 py-1 text-xs bg-gray-100 rounded text-gray-600"
                            >
                              {key}: {String(value)}
                            </span>
                          ))
                        ) : (
                          <span className="text-gray-400 text-sm">-</span>
                        )}
                        {user.traits && Object.keys(user.traits).length > 3 && (
                          <span className="px-2 py-1 text-xs text-gray-400">
                            +{Object.keys(user.traits).length - 3} more
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {user.feedback_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(user.last_seen)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  )
}
