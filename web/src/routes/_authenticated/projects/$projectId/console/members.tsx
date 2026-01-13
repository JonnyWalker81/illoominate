import { createFileRoute, useParams } from '@tanstack/react-router'
import { useState } from 'react'
import { ProjectHeader } from '../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../components/ConsoleNav'

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/members')({
  component: MembersPage,
})

function MembersPage() {
  const { projectId } = useParams({ from: '/_authenticated/projects/$projectId/console/members' })
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('member')

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />
      <ConsoleNav projectId={projectId} />

      <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-8">Team Members</h1>

      {/* Invite Form */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Invite Member</h2>
        <div className="flex space-x-4">
          <input
            type="email"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
            placeholder="email@example.com"
          />
          <select
            value={inviteRole}
            onChange={(e) => setInviteRole(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md"
          >
            <option value="viewer">Viewer</option>
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
          <button className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90">
            Send Invite
          </button>
        </div>
      </div>

      {/* Members List */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b">
          <h2 className="text-lg font-semibold text-gray-900">Current Members</h2>
        </div>
        <ul className="divide-y">
          <li className="px-6 py-4 flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center">
                <span className="text-gray-500 text-sm font-medium">?</span>
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">You</p>
                <p className="text-sm text-gray-500">owner</p>
              </div>
            </div>
            <span className="px-2 py-1 text-xs font-medium bg-purple-100 text-purple-800 rounded">
              Owner
            </span>
          </li>
        </ul>
      </div>

      {/* Pending Invites */}
      <div className="bg-white shadow rounded-lg mt-6">
        <div className="px-6 py-4 border-b">
          <h2 className="text-lg font-semibold text-gray-900">Pending Invites</h2>
        </div>
        <div className="p-6 text-center text-gray-500">
          No pending invites.
        </div>
      </div>
      </div>
    </div>
  )
}
