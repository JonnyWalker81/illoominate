import { createFileRoute, useParams } from '@tanstack/react-router'
import { ProjectHeader } from '../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../components/ConsoleNav'

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const { projectId } = useParams({ from: '/_authenticated/projects/$projectId/console/settings' })

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />
      <ConsoleNav projectId={projectId} />

      <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-8">Project Settings</h1>

      <div className="bg-white shadow rounded-lg divide-y">
        {/* General Settings */}
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">General</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Project Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                placeholder="Project name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Project Key
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                disabled
                placeholder="proj_xxxxx"
              />
            </div>
          </div>
        </div>

        {/* Visibility Settings */}
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Default Visibility</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-700">Bug Reports</span>
              <select className="px-3 py-2 border border-gray-300 rounded-md">
                <option>Team Only</option>
                <option>Community</option>
              </select>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-700">Feature Requests</span>
              <select className="px-3 py-2 border border-gray-300 rounded-md">
                <option>Team Only</option>
                <option>Community</option>
              </select>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-700">General Feedback</span>
              <select className="px-3 py-2 border border-gray-300 rounded-md">
                <option>Team Only</option>
                <option>Community</option>
              </select>
            </div>
          </div>
        </div>

        {/* Community Settings */}
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Community</h2>
          <div className="space-y-4">
            <label className="flex items-center space-x-3">
              <input type="checkbox" className="rounded" defaultChecked />
              <span className="text-sm text-gray-700">Enable voting</span>
            </label>
            <label className="flex items-center space-x-3">
              <input type="checkbox" className="rounded" defaultChecked />
              <span className="text-sm text-gray-700">Enable community comments</span>
            </label>
            <label className="flex items-center space-x-3">
              <input type="checkbox" className="rounded" />
              <span className="text-sm text-gray-700">Allow anonymous feedback</span>
            </label>
          </div>
        </div>

        {/* Save Button */}
        <div className="p-6 bg-gray-50">
          <button className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90">
            Save Changes
          </button>
        </div>
      </div>
      </div>
    </div>
  )
}
