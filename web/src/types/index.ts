// ============================================
// Role types
// ============================================

export type Role = 'community' | 'viewer' | 'member' | 'admin' | 'owner'

export const ROLE_HIERARCHY: Record<Role, number> = {
  community: 0,
  viewer: 1,
  member: 2,
  admin: 3,
  owner: 4,
}

export const isTeamRole = (role: Role): boolean =>
  ['viewer', 'member', 'admin', 'owner'].includes(role)

export const canModify = (role: Role): boolean =>
  ['member', 'admin', 'owner'].includes(role)

export const isAdminOrOwner = (role: Role): boolean =>
  ['admin', 'owner'].includes(role)

// ============================================
// Feedback types
// ============================================

export type FeedbackType = 'bug' | 'feature' | 'general'

export type Visibility = 'TEAM_ONLY' | 'COMMUNITY'

export type FeedbackStatus =
  | 'new'
  | 'under_review'
  | 'planned'
  | 'in_progress'
  | 'completed'
  | 'declined'
  | 'duplicate'

export type Severity = 'low' | 'medium' | 'high' | 'critical'

export interface Feedback {
  id: string
  project_id: string
  author_id?: string
  assigned_to?: string
  canonical_id?: string
  title: string
  description: string
  type: FeedbackType
  status: FeedbackStatus
  severity?: Severity
  visibility: Visibility
  vote_count: number
  comment_count: number
  submitter_email?: string
  submitter_name?: string
  submitter_identifier?: string
  source: string
  source_metadata?: {
    user_traits?: Record<string, string>
    device_model?: string
    os_version?: string
    app_version?: string
    [key: string]: unknown
  }
  sdk_user?: SDKUser
  created_at: string
  updated_at: string
  resolved_at?: string
  tags?: Tag[]
  author?: User
  assignee?: User
  has_voted?: boolean
  can_delete?: boolean
}

// ============================================
// Project types
// ============================================

export interface Project {
  id: string
  name: string
  slug: string
  project_key: string
  settings: ProjectSettings
  logo_url?: string
  primary_color: string
  created_at: string
  updated_at: string
}

export interface ProjectSettings {
  default_visibility: {
    bug: Visibility
    feature: Visibility
    general: Visibility
  }
  allow_anonymous_feedback: boolean
  require_email_for_anonymous: boolean
  voting_enabled: boolean
  community_comments_enabled: boolean
  auto_close_duplicates: boolean
  notification_preferences: {
    new_feedback: boolean
    status_changes: boolean
    new_comments: boolean
  }
}

// ============================================
// Membership types
// ============================================

export interface Membership {
  id: string
  project_id: string
  user_id: string
  role: Role
  display_name?: string
  created_at: string
  updated_at: string
  project?: Project
  user?: User
}

// ============================================
// User types
// ============================================

export interface User {
  id: string
  email: string
  name?: string
  avatar_url?: string
}

// ============================================
// Identified User types (SDK users)
// ============================================

export interface SDKUser {
  id: string
  external_id: string
  email?: string
  name?: string
  traits?: Record<string, unknown>
}

export interface IdentifiedUser {
  id: string
  identifier: string
  email?: string
  name?: string
  traits?: Record<string, unknown>
  feedback_count: number
  first_seen: string
  last_seen: string
}

// ============================================
// Comment types
// ============================================

export interface Comment {
  id: string
  feedback_id: string
  author_id: string
  parent_id?: string
  body: string
  visibility: Visibility
  is_edited: boolean
  created_at: string
  updated_at: string
  author?: User
  replies?: Comment[]
}

// ============================================
// Tag types
// ============================================

export interface Tag {
  id: string
  project_id: string
  name: string
  slug: string
  color: string
  description?: string
  feedback_count?: number
}

// ============================================
// Invite types
// ============================================

export type InviteStatus = 'pending' | 'accepted' | 'expired' | 'revoked'

export interface Invite {
  id: string
  project_id: string
  invited_by: string
  email: string
  role: Role
  status: InviteStatus
  expires_at: string
  created_at: string
  accepted_at?: string
}

// ============================================
// Vote types
// ============================================

export interface Vote {
  id: string
  feedback_id: string
  user_id: string
  created_at: string
}

export interface VoteResult {
  feedback_id: string
  vote_count: number
  has_voted: boolean
}

// ============================================
// Pagination types
// ============================================

export interface PaginationMeta {
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: PaginationMeta
}

// ============================================
// API Response types
// ============================================

export interface ApiError {
  error: string
  code: string
  fields?: Record<string, string>
}

// ============================================
// SDK Token types
// ============================================

export interface SDKToken {
  id: string
  name: string
  token?: string // Only returned on creation
  token_prefix: string
  allowed_origins: string[]
  rate_limit: number
  is_active: boolean
  last_used_at?: string
  created_at: string
}

export interface CreateSDKTokenRequest {
  name: string
  allowed_origins?: string[]
  rate_limit_per_minute?: number
}

// ============================================
// Portal types (for feedback users)
// ============================================

export interface PortalNotificationPreferences {
  status_changes: boolean
  new_comments_on_my_feedback: boolean
  new_comments_on_voted_feedback?: boolean
  weekly_digest?: boolean
}

export interface PortalUserProfile {
  id: string
  user_id: string
  project_id: string
  notification_preferences: PortalNotificationPreferences
  created_at: string
  updated_at: string
}

export interface PortalFeedbackSummary {
  id: string
  title: string
  description: string
  type: string
  status: string
  vote_count: number
  comment_count: number
  has_voted: boolean
  created_at: string
  updated_at: string
}
