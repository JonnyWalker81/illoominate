import { supabase } from './supabase'

const API_URL = import.meta.env.VITE_API_URL || '/api'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  params?: Record<string, string | number | boolean | undefined>
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, params } = options

  // Build URL with query params
  const url = new URL(`${API_URL}${endpoint}`, window.location.origin)
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    })
  }

  // Get auth token
  const { data: { session } } = await supabase.auth.getSession()
  const token = session?.access_token

  const response = await fetch(url.toString(), {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...(body ? { body: JSON.stringify(body) } : {}),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(error.message || error.error || 'Request failed')
  }

  // Handle 204 No Content responses
  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

// Helper methods
const get = <T>(endpoint: string, params?: RequestOptions['params']) =>
  request<T>(endpoint, { method: 'GET', params })

const post = <T>(endpoint: string, body?: unknown) =>
  request<T>(endpoint, { method: 'POST', body })

const patch = <T>(endpoint: string, body?: unknown) =>
  request<T>(endpoint, { method: 'PATCH', body })

const del = <T>(endpoint: string) =>
  request<T>(endpoint, { method: 'DELETE' })

// API client
export const api = {
  // Community endpoints
  community: {
    listFeatures: (projectId: string, params?: { sort?: string; status?: string; page?: number }) =>
      get(`/community/projects/${projectId}/feature-requests`, params),
    getFeature: (projectId: string, feedbackId: string) =>
      get(`/community/projects/${projectId}/feature-requests/${feedbackId}`),
    vote: (projectId: string, feedbackId: string) =>
      post(`/community/projects/${projectId}/feature-requests/${feedbackId}/vote`),
    unvote: (projectId: string, feedbackId: string) =>
      del(`/community/projects/${projectId}/feature-requests/${feedbackId}/vote`),
    listComments: (projectId: string, feedbackId: string) =>
      get(`/community/projects/${projectId}/feature-requests/${feedbackId}/comments`),
    createComment: (projectId: string, feedbackId: string, body: string) =>
      post(`/community/projects/${projectId}/feature-requests/${feedbackId}/comments`, { body }),
  },

  // Creator endpoints
  creator: {
    listFeedback: (projectId: string, params?: Record<string, unknown>) =>
      get(`/creator/projects/${projectId}/feedback`, params as RequestOptions['params']),
    getFeedback: (projectId: string, feedbackId: string) =>
      get(`/creator/projects/${projectId}/feedback/${feedbackId}`),
    createFeedback: (projectId: string, data: unknown) =>
      post(`/creator/projects/${projectId}/feedback`, data),
    updateFeedback: (projectId: string, feedbackId: string, data: unknown) =>
      patch(`/creator/projects/${projectId}/feedback/${feedbackId}`, data),
    mergeFeedback: (projectId: string, feedbackId: string, canonicalId: string) =>
      post(`/creator/projects/${projectId}/feedback/${feedbackId}/merge`, { canonical_id: canonicalId }),

    // Tags
    listTags: (projectId: string) =>
      get(`/creator/projects/${projectId}/tags`),
    createTag: (projectId: string, data: { name: string; color?: string }) =>
      post(`/creator/projects/${projectId}/tags`, data),
    updateTag: (projectId: string, tagId: string, data: { name?: string; color?: string }) =>
      patch(`/creator/projects/${projectId}/tags/${tagId}`, data),
    deleteTag: (projectId: string, tagId: string) =>
      del(`/creator/projects/${projectId}/tags/${tagId}`),

    // Members
    listMembers: (projectId: string) =>
      get(`/creator/projects/${projectId}/members`),
    addMember: (projectId: string, data: { user_id: string; role: string }) =>
      post(`/creator/projects/${projectId}/members`, data),
    updateMember: (projectId: string, memberId: string, data: { role: string }) =>
      patch(`/creator/projects/${projectId}/members/${memberId}`, data),
    removeMember: (projectId: string, memberId: string) =>
      del(`/creator/projects/${projectId}/members/${memberId}`),
    inviteMember: (projectId: string, data: { email: string; role: string }) =>
      post(`/creator/projects/${projectId}/members/invite`, data),

    // Settings
    getSettings: (projectId: string) =>
      get(`/creator/projects/${projectId}/settings`),
    updateSettings: (projectId: string, data: unknown) =>
      patch(`/creator/projects/${projectId}/settings`, data),

    // SDK Tokens
    listSDKTokens: (projectId: string) =>
      get(`/creator/projects/${projectId}/sdk-tokens`),
    createSDKToken: (projectId: string, data: { name: string; allowed_origins?: string[]; rate_limit_per_minute?: number }) =>
      post(`/creator/projects/${projectId}/sdk-tokens`, data),
    revokeSDKToken: (projectId: string, tokenId: string) =>
      del(`/creator/projects/${projectId}/sdk-tokens/${tokenId}`),

    // Users (identified SDK users)
    listUsers: (projectId: string, params?: { search?: string }) =>
      get(`/creator/projects/${projectId}/users`, params),
    getUserFeedback: (projectId: string, userId: string) =>
      get(`/creator/projects/${projectId}/users/${encodeURIComponent(userId)}/feedback`),
  },

  // Invites
  acceptInvite: (token: string) =>
    post(`/invites/${token}/accept`),

  // User
  getCurrentUser: () =>
    get('/me'),

  // Projects
  projects: {
    list: () =>
      get('/creator/projects'),
    get: (projectId: string) =>
      get(`/creator/projects/${projectId}`),
    create: (data: { name: string }) =>
      post('/creator/projects', data),
    delete: (projectId: string) =>
      del(`/creator/projects/${projectId}`),
  },

  // Feature Requests (Community)
  featureRequests: {
    list: (projectId: string, params?: { sort?: string; status?: string }) =>
      get(`/community/projects/${projectId}/feature-requests`, params),
    create: (projectId: string, data: { title: string; description: string }) =>
      post(`/community/projects/${projectId}/feature-requests`, data),
    delete: (projectId: string, feedbackId: string) =>
      del(`/community/projects/${projectId}/feature-requests/${feedbackId}`),
  },

  // Portal endpoints (for feedback users)
  portal: {
    // Profile
    getProfile: (projectId: string) =>
      get(`/portal/${projectId}/me`),
    updateNotifications: (projectId: string, prefs: {
      status_changes?: boolean
      new_comments_on_my_feedback?: boolean
      new_comments_on_voted_feedback?: boolean
      weekly_digest?: boolean
    }) =>
      patch(`/portal/${projectId}/me/notifications`, prefs),

    // Feedback
    listMyFeedback: (projectId: string) =>
      get(`/portal/${projectId}/my-feedback`),
    listFeatures: (projectId: string, params?: { limit?: number; offset?: number }) =>
      get(`/portal/${projectId}/feature-requests`, params),

    // Voting
    vote: (projectId: string, feedbackId: string) =>
      post(`/portal/${projectId}/feature-requests/${feedbackId}/vote`),
    unvote: (projectId: string, feedbackId: string) =>
      del(`/portal/${projectId}/feature-requests/${feedbackId}/vote`),
  },
}
