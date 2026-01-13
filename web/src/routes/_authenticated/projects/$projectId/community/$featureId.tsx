import { createFileRoute, Link } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/projects/$projectId/community/$featureId')({
  component: FeatureDetailPage,
})

function FeatureDetailPage() {
  const { projectId, featureId } = Route.useParams()

  return (
    <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <Link
        to="/projects/$projectId/community"
        params={{ projectId }}
        className="text-gray-500 hover:text-gray-700 mb-4 inline-block"
      >
        ← Back to Features
      </Link>

      <div className="bg-white shadow rounded-lg p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">
          Feature Request
        </h1>
        <p className="text-gray-500">
          Loading feature {featureId}...
        </p>
      </div>
    </div>
  )
}
