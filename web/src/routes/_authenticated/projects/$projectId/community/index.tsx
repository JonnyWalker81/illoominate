import { useState, useEffect } from 'react'
import { createFileRoute, useParams } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X, ChevronUp, MessageSquare, ChevronDown, Send, SlidersHorizontal, Trash2 } from 'lucide-react'
import { ProjectHeader } from '../../../../../components/ProjectHeader'
import { api } from '../../../../../lib/api'
import type { Feedback } from '../../../../../types'

interface Comment {
  id: string
  feedback_id: string
  author_id: string
  body: string
  is_edited: boolean
  created_at: string
  updated_at: string
}

export const Route = createFileRoute('/_authenticated/projects/$projectId/community/')({
  component: CommunityPage,
})

const submitRequestSchema = z.object({
  title: z
    .string()
    .min(1, 'Title is required')
    .max(200, 'Title must be 200 characters or less'),
  description: z
    .string()
    .min(1, 'Description is required')
    .max(5000, 'Description must be 5000 characters or less'),
})

type SubmitRequestForm = z.infer<typeof submitRequestSchema>

function CommunityPage() {
  const { projectId } = useParams({ from: '/_authenticated/projects/$projectId/community/' })
  const [featureRequests, setFeatureRequests] = useState<Feedback[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [votedIds, setVotedIds] = useState<Set<string>>(new Set())
  const [votingIds, setVotingIds] = useState<Set<string>>(new Set())
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [comments, setComments] = useState<Record<string, Comment[]>>({})
  const [loadingComments, setLoadingComments] = useState<Set<string>>(new Set())
  const [newComment, setNewComment] = useState('')
  const [submittingComment, setSubmittingComment] = useState(false)
  const [sortBy, setSortBy] = useState<string>('votes')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const fetchFeatureRequests = async (sort?: string, status?: string) => {
    try {
      const params: { sort?: string; status?: string } = {}
      if (sort) params.sort = sort
      if (status && status !== 'all') params.status = status

      const data = (await api.featureRequests.list(projectId, params)) as Feedback[]
      const requests = Array.isArray(data) ? data : []
      setFeatureRequests(requests)

      // Initialize voted state from API response
      const voted = new Set<string>()
      requests.forEach((r) => {
        if (r.has_voted) voted.add(r.id)
      })
      setVotedIds(voted)
    } catch (err) {
      console.error('Failed to fetch feature requests:', err)
      setFeatureRequests([])
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchFeatureRequests(sortBy, filterStatus)
  }, [projectId, sortBy, filterStatus])

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SubmitRequestForm>({
    resolver: zodResolver(submitRequestSchema),
  })

  const onSubmit = async (data: SubmitRequestForm) => {
    setIsSubmitting(true)
    setError(null)
    try {
      await api.featureRequests.create(projectId, data)
      setIsModalOpen(false)
      reset()
      fetchFeatureRequests()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit request')
    } finally {
      setIsSubmitting(false)
    }
  }

  const closeModal = () => {
    setIsModalOpen(false)
    reset()
    setError(null)
  }

  const fetchComments = async (feedbackId: string) => {
    if (loadingComments.has(feedbackId)) return

    setLoadingComments((prev) => new Set(prev).add(feedbackId))
    try {
      const data = (await api.community.listComments(projectId, feedbackId)) as Comment[]
      setComments((prev) => ({ ...prev, [feedbackId]: data }))
    } catch (err) {
      console.error('Failed to fetch comments:', err)
    } finally {
      setLoadingComments((prev) => {
        const next = new Set(prev)
        next.delete(feedbackId)
        return next
      })
    }
  }

  const toggleExpanded = (feedbackId: string) => {
    if (expandedId === feedbackId) {
      setExpandedId(null)
      setNewComment('')
    } else {
      setExpandedId(feedbackId)
      if (!comments[feedbackId]) {
        fetchComments(feedbackId)
      }
    }
  }

  const handleSubmitComment = async (feedbackId: string) => {
    if (!newComment.trim() || submittingComment) return

    setSubmittingComment(true)
    try {
      const comment = (await api.community.createComment(projectId, feedbackId, newComment.trim())) as Comment
      setComments((prev) => ({
        ...prev,
        [feedbackId]: [...(prev[feedbackId] || []), comment],
      }))
      setNewComment('')
      // Update comment count
      setFeatureRequests((prev) =>
        prev.map((r) =>
          r.id === feedbackId ? { ...r, comment_count: r.comment_count + 1 } : r
        )
      )
    } catch (err) {
      console.error('Failed to create comment:', err)
    } finally {
      setSubmittingComment(false)
    }
  }

  const handleDelete = async (feedbackId: string) => {
    if (deletingId) return
    if (!confirm('Are you sure you want to delete this feature request? This action cannot be undone.')) {
      return
    }

    setDeletingId(feedbackId)
    try {
      await api.featureRequests.delete(projectId, feedbackId)
      setFeatureRequests((prev) => prev.filter((r) => r.id !== feedbackId))
      if (expandedId === feedbackId) {
        setExpandedId(null)
      }
    } catch (err) {
      console.error('Failed to delete feature request:', err)
      alert('Failed to delete feature request')
    } finally {
      setDeletingId(null)
    }
  }

  const handleVote = async (feedbackId: string) => {
    if (votingIds.has(feedbackId)) return

    setVotingIds((prev) => new Set(prev).add(feedbackId))
    const hasVoted = votedIds.has(feedbackId)

    try {
      if (hasVoted) {
        await api.community.unvote(projectId, feedbackId)
        setVotedIds((prev) => {
          const next = new Set(prev)
          next.delete(feedbackId)
          return next
        })
        setFeatureRequests((prev) =>
          prev.map((r) =>
            r.id === feedbackId ? { ...r, vote_count: r.vote_count - 1 } : r
          )
        )
      } else {
        await api.community.vote(projectId, feedbackId)
        setVotedIds((prev) => new Set(prev).add(feedbackId))
        setFeatureRequests((prev) =>
          prev.map((r) =>
            r.id === feedbackId ? { ...r, vote_count: r.vote_count + 1 } : r
          )
        )
      }
    } catch (err) {
      console.error('Failed to vote:', err)
    } finally {
      setVotingIds((prev) => {
        const next = new Set(prev)
        next.delete(feedbackId)
        return next
      })
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />

      <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Feature Requests</h1>
          <button
            onClick={() => setIsModalOpen(true)}
            className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90"
          >
            Submit Request
          </button>
        </div>

        {/* Sorting and Filtering */}
        <div className="flex flex-wrap items-center gap-4 mb-6 bg-white rounded-lg shadow p-4">
          <div className="flex items-center gap-2">
            <SlidersHorizontal className="w-4 h-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Sort:</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
            >
              <option value="votes">Most Votes</option>
              <option value="newest">Newest</option>
              <option value="oldest">Oldest</option>
              <option value="comments">Most Comments</option>
            </select>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-gray-700">Status:</span>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
            >
              <option value="all">All</option>
              <option value="new">New</option>
              <option value="planned">Planned</option>
              <option value="in_progress">In Progress</option>
              <option value="completed">Completed</option>
            </select>
          </div>

          {(sortBy !== 'votes' || filterStatus !== 'all') && (
            <button
              onClick={() => {
                setSortBy('votes')
                setFilterStatus('all')
              }}
              className="text-sm text-gray-500 hover:text-gray-700 underline"
            >
              Reset filters
            </button>
          )}
        </div>

        {isLoading ? (
          <div className="bg-white shadow rounded-lg p-6">
            <p className="text-gray-500 text-center py-8">Loading...</p>
          </div>
        ) : featureRequests.length === 0 ? (
          <div className="bg-white shadow rounded-lg p-6">
            <p className="text-gray-500 text-center py-8">
              No feature requests yet. Be the first to submit one!
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {featureRequests.map((request) => (
              <div
                key={request.id}
                className="bg-white shadow rounded-lg overflow-hidden"
              >
                <div className="p-6 hover:bg-gray-50 transition-colors">
                  <div className="flex gap-4">
                    <div className="flex flex-col items-center">
                      <button
                        onClick={() => handleVote(request.id)}
                        disabled={votingIds.has(request.id)}
                        className={`p-1 transition-colors ${
                          votedIds.has(request.id)
                            ? 'text-primary'
                            : 'text-gray-400 hover:text-primary'
                        } disabled:opacity-50`}
                      >
                        <ChevronUp className="w-6 h-6" />
                      </button>
                      <span className={`text-sm font-semibold ${
                        votedIds.has(request.id) ? 'text-primary' : 'text-gray-700'
                      }`}>
                        {request.vote_count}
                      </span>
                    </div>
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-gray-900 mb-1">
                        {request.title}
                      </h3>
                      <p className={`text-gray-600 text-sm mb-3 ${expandedId === request.id ? '' : 'line-clamp-2'}`}>
                        {request.description}
                      </p>
                      <div className="flex items-center gap-4 text-sm text-gray-500">
                        <button
                          onClick={() => toggleExpanded(request.id)}
                          className="flex items-center gap-1 hover:text-primary transition-colors"
                        >
                          <MessageSquare className="w-4 h-4" />
                          {request.comment_count}
                          {expandedId === request.id ? (
                            <ChevronUp className="w-4 h-4" />
                          ) : (
                            <ChevronDown className="w-4 h-4" />
                          )}
                        </button>
                        <span
                          className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                            request.status === 'planned'
                              ? 'bg-blue-100 text-blue-800'
                              : request.status === 'in_progress'
                                ? 'bg-yellow-100 text-yellow-800'
                                : request.status === 'completed'
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {request.status.replace('_', ' ')}
                        </span>
                        <button
                          onClick={() => handleDelete(request.id)}
                          disabled={deletingId === request.id || !request.can_delete}
                          className={`flex items-center gap-1 transition-colors ${
                            request.can_delete
                              ? 'text-gray-400 hover:text-red-600'
                              : 'text-gray-200 cursor-not-allowed'
                          } disabled:opacity-50`}
                          title={request.can_delete ? "Delete feature request" : "You don't have permission to delete"}
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Comments Section */}
                {expandedId === request.id && (
                  <div className="border-t bg-gray-50 p-6">
                    {loadingComments.has(request.id) ? (
                      <p className="text-gray-500 text-sm text-center py-4">Loading comments...</p>
                    ) : (
                      <>
                        {(comments[request.id] || []).length === 0 ? (
                          <p className="text-gray-500 text-sm text-center py-4">
                            No comments yet. Be the first to comment!
                          </p>
                        ) : (
                          <div className="space-y-4 mb-4">
                            {(comments[request.id] || []).map((comment) => (
                              <div key={comment.id} className="bg-white rounded-lg p-4 shadow-sm">
                                <p className="text-gray-700 text-sm">{comment.body}</p>
                                <p className="text-gray-400 text-xs mt-2">
                                  {new Date(comment.created_at).toLocaleDateString()}
                                  {comment.is_edited && ' (edited)'}
                                </p>
                              </div>
                            ))}
                          </div>
                        )}

                        {/* Add Comment Form */}
                        <div className="flex gap-2">
                          <input
                            type="text"
                            value={newComment}
                            onChange={(e) => setNewComment(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter' && !e.shiftKey) {
                                e.preventDefault()
                                handleSubmitComment(request.id)
                              }
                            }}
                            placeholder="Write a comment..."
                            className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-sm"
                          />
                          <button
                            onClick={() => handleSubmitComment(request.id)}
                            disabled={!newComment.trim() || submittingComment}
                            className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            <Send className="w-4 h-4" />
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Submit Request Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/50 transition-opacity"
              onClick={closeModal}
            />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">
                  Submit Feature Request
                </h2>
                <button
                  onClick={closeModal}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              <form onSubmit={handleSubmit(onSubmit)}>
                <div className="mb-4">
                  <label
                    htmlFor="title"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Title
                  </label>
                  <input
                    type="text"
                    id="title"
                    {...register('title')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
                    placeholder="Brief summary of your request"
                  />
                  {errors.title && (
                    <p className="mt-1 text-sm text-red-600">
                      {errors.title.message}
                    </p>
                  )}
                </div>

                <div className="mb-4">
                  <label
                    htmlFor="description"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Description
                  </label>
                  <textarea
                    id="description"
                    rows={5}
                    {...register('description')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
                    placeholder="Describe your feature request in detail..."
                  />
                  {errors.description && (
                    <p className="mt-1 text-sm text-red-600">
                      {errors.description.message}
                    </p>
                  )}
                </div>

                {error && (
                  <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                    <p className="text-sm text-red-600">{error}</p>
                  </div>
                )}

                <div className="flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={closeModal}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isSubmitting ? 'Submitting...' : 'Submit Request'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
