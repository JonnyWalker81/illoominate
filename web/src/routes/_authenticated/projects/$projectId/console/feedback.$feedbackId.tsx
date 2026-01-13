import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/projects/$projectId/console/feedback/$feedbackId')({
  component: FeedbackDetailPage,
})

function FeedbackDetailPage() {
  const { projectId, feedbackId } = Route.useParams()

  return (
    <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <Link
        to="/projects/$projectId/console"
        params={{ projectId }}
        className="text-gray-500 hover:text-gray-700 mb-4 inline-block"
      >
        ← Back to Inbox
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 bg-white shadow rounded-lg p-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">
            Feedback Detail
          </h1>
          <p className="text-gray-500">
            Loading feedback {feedbackId}...
          </p>
        </div>

        {/* Triage panel */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Triage</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option>New</option>
                <option>Under Review</option>
                <option>Planned</option>
                <option>In Progress</option>
                <option>Completed</option>
                <option>Declined</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Visibility
              </label>
              <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option>Team Only</option>
                <option>Community</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Assignee
              </label>
              <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option>Unassigned</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
