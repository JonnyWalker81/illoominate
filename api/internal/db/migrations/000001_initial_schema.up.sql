-- ============================================
-- FullDisclosure Initial Schema
-- Multi-tenant Feedback + Feature Voting Platform
-- ============================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ENUMS
-- ============================================

-- User roles within a project (ordered by permission level)
CREATE TYPE membership_role AS ENUM (
    'community',   -- Beta testers, can see COMMUNITY visibility items, vote, comment
    'viewer',      -- Can view team-only content, cannot modify
    'member',      -- Can create/edit feedback, triage
    'admin',       -- Full access except project deletion
    'owner'        -- Full access including project deletion
);

-- Types of feedback
CREATE TYPE feedback_type AS ENUM (
    'bug',
    'feature',
    'general'
);

-- Visibility levels for feedback and comments
CREATE TYPE visibility_level AS ENUM (
    'TEAM_ONLY',   -- Only team members (viewer, member, admin, owner)
    'COMMUNITY'    -- Visible to community members as well
);

-- Status workflow for feedback items
CREATE TYPE feedback_status AS ENUM (
    'new',
    'under_review',
    'planned',
    'in_progress',
    'completed',
    'declined',
    'duplicate'
);

-- Severity levels (primarily for bugs)
CREATE TYPE severity_level AS ENUM (
    'low',
    'medium',
    'high',
    'critical'
);

-- Invite status
CREATE TYPE invite_status AS ENUM (
    'pending',
    'accepted',
    'expired',
    'revoked'
);

-- Attachment upload status
CREATE TYPE attachment_status AS ENUM (
    'pending',     -- Upload initiated, awaiting completion
    'uploaded',    -- Successfully uploaded to GCS
    'failed',      -- Upload failed
    'deleted'      -- Soft deleted
);

-- Activity log action types
CREATE TYPE activity_action AS ENUM (
    'created',
    'updated',
    'status_changed',
    'visibility_changed',
    'merged',
    'commented',
    'voted',
    'unvoted',
    'tagged',
    'untagged',
    'assigned',
    'unassigned',
    'attachment_added',
    'attachment_removed'
);

-- ============================================
-- PROJECTS
-- ============================================
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Identifiers
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    project_key VARCHAR(32) NOT NULL UNIQUE,

    -- Settings (JSONB for flexibility)
    settings JSONB NOT NULL DEFAULT '{
        "default_visibility": {
            "bug": "TEAM_ONLY",
            "feature": "COMMUNITY",
            "general": "TEAM_ONLY"
        },
        "allow_anonymous_feedback": true,
        "require_email_for_anonymous": false,
        "voting_enabled": true,
        "community_comments_enabled": true,
        "auto_close_duplicates": true,
        "notification_preferences": {
            "new_feedback": true,
            "status_changes": true,
            "new_comments": true
        }
    }'::jsonb,

    -- Branding
    logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#6366F1',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ
);

-- Indexes for projects
CREATE INDEX idx_projects_slug ON projects(slug) WHERE archived_at IS NULL;
CREATE INDEX idx_projects_project_key ON projects(project_key) WHERE archived_at IS NULL;
CREATE INDEX idx_projects_created_at ON projects(created_at);

-- ============================================
-- MEMBERSHIPS (Project-User-Role junction)
-- ============================================
CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,  -- References auth.users (Supabase)

    -- Role
    role membership_role NOT NULL DEFAULT 'member',

    -- Metadata
    display_name VARCHAR(100),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint: one membership per user per project
    CONSTRAINT uq_memberships_project_user UNIQUE (project_id, user_id)
);

-- Indexes for memberships
CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_project_id ON memberships(project_id);
CREATE INDEX idx_memberships_role ON memberships(project_id, role);

-- ============================================
-- INVITES (Email-based project invitations)
-- ============================================
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL,

    -- Invite details
    email VARCHAR(255) NOT NULL,
    role membership_role NOT NULL DEFAULT 'member',
    token VARCHAR(64) NOT NULL UNIQUE,

    -- Status tracking
    status invite_status NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMPTZ
);

-- Partial unique index for pending invites
CREATE UNIQUE INDEX idx_invites_pending_unique
    ON invites(project_id, email)
    WHERE status = 'pending';

CREATE INDEX idx_invites_token ON invites(token) WHERE status = 'pending';
CREATE INDEX idx_invites_email ON invites(email);
CREATE INDEX idx_invites_expires ON invites(expires_at) WHERE status = 'pending';

-- ============================================
-- SDK TOKENS (For anonymous SDK authentication)
-- ============================================
CREATE TABLE sdk_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign key
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Token details
    name VARCHAR(100) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    token_prefix VARCHAR(12) NOT NULL,

    -- Permissions
    allowed_origins TEXT[],
    rate_limit_per_minute INTEGER DEFAULT 60,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sdk_tokens_project ON sdk_tokens(project_id) WHERE is_active = true;
CREATE INDEX idx_sdk_tokens_hash ON sdk_tokens(token_hash) WHERE is_active = true;
CREATE INDEX idx_sdk_tokens_prefix ON sdk_tokens(token_prefix) WHERE is_active = true;

-- ============================================
-- TAGS (Project-scoped labels)
-- ============================================
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign key
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Tag details
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6B7280',
    description TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique tag names per project
    CONSTRAINT uq_tags_project_slug UNIQUE (project_id, slug)
);

CREATE INDEX idx_tags_project ON tags(project_id);

-- ============================================
-- FEEDBACK (Core feedback items)
-- ============================================
CREATE TABLE feedback (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_id UUID,
    assigned_to UUID,

    -- Merge tracking (for duplicates)
    canonical_id UUID REFERENCES feedback(id) ON DELETE SET NULL,

    -- Content
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,

    -- Classification
    type feedback_type NOT NULL,
    status feedback_status NOT NULL DEFAULT 'new',
    severity severity_level,
    visibility visibility_level NOT NULL,

    -- Denormalized counts (for performance)
    vote_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Anonymous submitter info (when author_id is NULL)
    submitter_email VARCHAR(255),
    submitter_name VARCHAR(100),
    submitter_identifier VARCHAR(255),

    -- Source tracking
    source VARCHAR(50) DEFAULT 'web',
    source_url TEXT,
    source_metadata JSONB,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT chk_feedback_author_or_identifier
        CHECK (author_id IS NOT NULL OR submitter_email IS NOT NULL OR submitter_identifier IS NOT NULL),
    CONSTRAINT chk_feedback_not_self_canonical
        CHECK (canonical_id IS NULL OR canonical_id != id)
);

-- Indexes for feedback
CREATE INDEX idx_feedback_project ON feedback(project_id);
CREATE INDEX idx_feedback_project_status ON feedback(project_id, status);
CREATE INDEX idx_feedback_project_type ON feedback(project_id, type);
CREATE INDEX idx_feedback_project_visibility ON feedback(project_id, visibility);
CREATE INDEX idx_feedback_canonical ON feedback(canonical_id) WHERE canonical_id IS NOT NULL;
CREATE INDEX idx_feedback_author ON feedback(author_id) WHERE author_id IS NOT NULL;
CREATE INDEX idx_feedback_assigned ON feedback(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_feedback_vote_count ON feedback(project_id, vote_count DESC);
CREATE INDEX idx_feedback_created_at ON feedback(project_id, created_at DESC);

-- Full-text search index
CREATE INDEX idx_feedback_search ON feedback
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- ============================================
-- FEEDBACK_TAGS (Junction table)
-- ============================================
CREATE TABLE feedback_tags (
    feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,

    PRIMARY KEY (feedback_id, tag_id)
);

CREATE INDEX idx_feedback_tags_tag ON feedback_tags(tag_id);

-- ============================================
-- VOTES (Unique per feedback + user)
-- ============================================
CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint: one vote per user per feedback
    CONSTRAINT uq_votes_feedback_user UNIQUE (feedback_id, user_id)
);

CREATE INDEX idx_votes_user ON votes(user_id);
CREATE INDEX idx_votes_feedback ON votes(feedback_id);

-- ============================================
-- COMMENTS
-- ============================================
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,

    -- Content
    body TEXT NOT NULL,
    visibility visibility_level NOT NULL DEFAULT 'COMMUNITY',

    -- Edit tracking
    is_edited BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_feedback ON comments(feedback_id);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_visibility ON comments(feedback_id, visibility);

-- ============================================
-- ATTACHMENTS
-- ============================================
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    feedback_id UUID REFERENCES feedback(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    uploaded_by UUID,

    -- File info
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,

    -- Storage
    gcs_bucket VARCHAR(100) NOT NULL,
    gcs_path TEXT NOT NULL,

    -- Upload tracking
    status attachment_status NOT NULL DEFAULT 'pending',
    upload_expires_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uploaded_at TIMESTAMPTZ,

    -- Must belong to either feedback or comment (or neither during upload)
    CONSTRAINT chk_attachment_parent
        CHECK (
            (feedback_id IS NULL AND comment_id IS NULL) OR
            (feedback_id IS NOT NULL AND comment_id IS NULL) OR
            (feedback_id IS NULL AND comment_id IS NOT NULL)
        )
);

CREATE INDEX idx_attachments_feedback ON attachments(feedback_id) WHERE feedback_id IS NOT NULL;
CREATE INDEX idx_attachments_comment ON attachments(comment_id) WHERE comment_id IS NOT NULL;
CREATE INDEX idx_attachments_status ON attachments(status);
CREATE INDEX idx_attachments_pending ON attachments(upload_expires_at)
    WHERE status = 'pending';

-- ============================================
-- ACTIVITY LOG (Timeline of actions)
-- ============================================
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign keys
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    feedback_id UUID REFERENCES feedback(id) ON DELETE CASCADE,
    actor_id UUID,

    -- Action details
    action activity_action NOT NULL,

    -- Change tracking (before/after values)
    changes JSONB,

    -- Additional context
    metadata JSONB,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for activity_log
CREATE INDEX idx_activity_project ON activity_log(project_id, created_at DESC);
CREATE INDEX idx_activity_feedback ON activity_log(feedback_id, created_at DESC)
    WHERE feedback_id IS NOT NULL;
CREATE INDEX idx_activity_actor ON activity_log(actor_id) WHERE actor_id IS NOT NULL;
CREATE INDEX idx_activity_action ON activity_log(project_id, action);

-- ============================================
-- TRIGGERS FOR DENORMALIZED COUNTS
-- ============================================

-- Update vote_count on feedback when votes change
CREATE OR REPLACE FUNCTION update_feedback_vote_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE feedback SET vote_count = vote_count + 1, updated_at = NOW()
        WHERE id = NEW.feedback_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE feedback SET vote_count = vote_count - 1, updated_at = NOW()
        WHERE id = OLD.feedback_id;
        RETURN OLD;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_votes_count
AFTER INSERT OR DELETE ON votes
FOR EACH ROW EXECUTE FUNCTION update_feedback_vote_count();

-- Update comment_count on feedback when comments change
CREATE OR REPLACE FUNCTION update_feedback_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE feedback SET comment_count = comment_count + 1, updated_at = NOW()
        WHERE id = NEW.feedback_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE feedback SET comment_count = comment_count - 1, updated_at = NOW()
        WHERE id = OLD.feedback_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' AND OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        -- Soft delete
        UPDATE feedback SET comment_count = comment_count - 1, updated_at = NOW()
        WHERE id = NEW.feedback_id;
        RETURN NEW;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comments_count
AFTER INSERT OR DELETE OR UPDATE OF deleted_at ON comments
FOR EACH ROW EXECUTE FUNCTION update_feedback_comment_count();

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_projects_updated_at
BEFORE UPDATE ON projects
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_memberships_updated_at
BEFORE UPDATE ON memberships
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_feedback_updated_at
BEFORE UPDATE ON feedback
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_comments_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Generate a random project key
CREATE OR REPLACE FUNCTION generate_project_key()
RETURNS VARCHAR(32) AS $$
DECLARE
    chars TEXT := 'abcdefghijklmnopqrstuvwxyz0123456789';
    result VARCHAR(32) := 'proj_';
    i INTEGER;
BEGIN
    FOR i IN 1..27 LOOP
        result := result || substr(chars, floor(random() * length(chars) + 1)::integer, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Generate a random invite token
CREATE OR REPLACE FUNCTION generate_invite_token()
RETURNS VARCHAR(64) AS $$
BEGIN
    RETURN encode(gen_random_bytes(32), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Check if a user is a team member (not community)
CREATE OR REPLACE FUNCTION is_team_member(p_project_id UUID, p_user_id UUID)
RETURNS BOOLEAN AS $$
    SELECT EXISTS(
        SELECT 1 FROM memberships
        WHERE project_id = p_project_id
        AND user_id = p_user_id
        AND role IN ('viewer', 'member', 'admin', 'owner')
    );
$$ LANGUAGE sql STABLE;

-- Check if user can modify (member, admin, owner)
CREATE OR REPLACE FUNCTION can_modify(p_project_id UUID, p_user_id UUID)
RETURNS BOOLEAN AS $$
    SELECT EXISTS(
        SELECT 1 FROM memberships
        WHERE project_id = p_project_id
        AND user_id = p_user_id
        AND role IN ('member', 'admin', 'owner')
    );
$$ LANGUAGE sql STABLE;

-- Check if user is admin or owner
CREATE OR REPLACE FUNCTION is_admin_or_owner(p_project_id UUID, p_user_id UUID)
RETURNS BOOLEAN AS $$
    SELECT EXISTS(
        SELECT 1 FROM memberships
        WHERE project_id = p_project_id
        AND user_id = p_user_id
        AND role IN ('admin', 'owner')
    );
$$ LANGUAGE sql STABLE;

-- Get user's role in a project
CREATE OR REPLACE FUNCTION get_user_project_role(p_project_id UUID, p_user_id UUID)
RETURNS membership_role AS $$
    SELECT role FROM memberships
    WHERE project_id = p_project_id AND user_id = p_user_id;
$$ LANGUAGE sql STABLE;
