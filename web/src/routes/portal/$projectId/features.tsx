import { createFileRoute, useParams } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { PortalFeedbackSummary, PaginatedResponse } from '@/types'

export const Route = createFileRoute('/portal/$projectId/features')({
  component: FeaturesPage,
})

function FeaturesPage() {
  const { projectId } = useParams({ from: '/portal/$projectId/features' })
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery<PaginatedResponse<PortalFeedbackSummary>>({
    queryKey: ['portal', 'features', projectId],
    queryFn: () => api.portal.listFeatures(projectId) as Promise<PaginatedResponse<PortalFeedbackSummary>>,
  })

  const voteMutation = useMutation({
    mutationFn: (feedbackId: string) => api.portal.vote(projectId, feedbackId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portal', 'features', projectId] })
    },
  })

  const unvoteMutation = useMutation({
    mutationFn: (feedbackId: string) => api.portal.unvote(projectId, feedbackId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portal', 'features', projectId] })
    },
  })

  const handleVote = (feedbackId: string, hasVoted: boolean) => {
    if (hasVoted) {
      unvoteMutation.mutate(feedbackId)
    } else {
      voteMutation.mutate(feedbackId)
    }
  }

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

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto" />
        <p className="mt-4 text-gray-500">Loading feature requests...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">Failed to load feature requests. Please try again.</p>
      </div>
    )
  }

  const features = data?.data ?? []

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Feature Requests</h1>

      {features.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <p className="text-gray-500">No feature requests yet.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {features.map((feature) => (
            <div
              key={feature.id}
              className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:border-gray-300 transition-colors"
            >
              <div className="flex items-start gap-4">
                <button
                  onClick={() => handleVote(feature.id, feature.has_voted)}
                  disabled={voteMutation.isPending || unvoteMutation.isPending}
                  className={`
                    flex flex-col items-center justify-center min-w-[60px] py-2 px-3 rounded-lg border transition-colors
                    ${feature.has_voted
                      ? 'bg-primary/10 border-primary text-primary'
                      : 'bg-gray-50 border-gray-200 text-gray-500 hover:border-gray-300'
                    }
                    ${(voteMutation.isPending || unvoteMutation.isPending) ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                  `}
                  aria-label={feature.has_voted ? 'Remove vote' : 'Vote for this feature'}
                >
                  <svg
                    className={`w-5 h-5 ${feature.has_voted ? 'text-primary' : 'text-gray-400'}`}
                    fill={feature.has_voted ? 'currentColor' : 'none'}
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <span className="text-sm font-semibold">{feature.vote_count}</span>
                </button>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${getStatusColor(feature.status)}`}>
                      {feature.status.replace('_', ' ')}
                    </span>
                  </div>
                  <h3 className="text-base font-medium text-gray-900">{feature.title}</h3>
                  <p className="text-sm text-gray-500 mt-1 line-clamp-2">{feature.description}</p>
                  <div className="mt-2 flex items-center gap-4 text-xs text-gray-400">
                    <span>{feature.comment_count} comments</span>
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
