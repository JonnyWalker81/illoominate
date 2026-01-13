-- ============================================
-- SDK_USERS (Identified users from SDK)
-- ============================================
-- Stores users identified via the SDK's identify() call.
-- Similar to Canny's approach: auto-create on first feedback,
-- update on subsequent identify calls.

CREATE TABLE sdk_users (
    -- Internal ID (our UUID)
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Project scope (users are per-project)
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- External ID (the app's user ID from identify() call)
    external_id VARCHAR(255) NOT NULL,

    -- User info (from identify call)
    email VARCHAR(255),
    name VARCHAR(255),
    avatar_url TEXT,

    -- Custom traits/properties (flexible key-value storage)
    traits JSONB DEFAULT '{}',

    -- Optional link to authenticated user account (for future SSO/login)
    linked_user_id UUID,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint: external_id must be unique per project
    CONSTRAINT uq_sdk_users_project_external_id UNIQUE (project_id, external_id)
);

-- Indexes for sdk_users
CREATE INDEX idx_sdk_users_project ON sdk_users(project_id);
CREATE INDEX idx_sdk_users_external_id ON sdk_users(project_id, external_id);
CREATE INDEX idx_sdk_users_email ON sdk_users(project_id, email) WHERE email IS NOT NULL;
CREATE INDEX idx_sdk_users_last_seen ON sdk_users(project_id, last_seen_at DESC);

-- ============================================
-- Add sdk_user_id to feedback table
-- ============================================
ALTER TABLE feedback
    ADD COLUMN sdk_user_id UUID REFERENCES sdk_users(id) ON DELETE SET NULL;

CREATE INDEX idx_feedback_sdk_user ON feedback(sdk_user_id) WHERE sdk_user_id IS NOT NULL;

-- ============================================
-- Update constraint to include sdk_user_id
-- ============================================
-- Drop the old constraint
ALTER TABLE feedback DROP CONSTRAINT chk_feedback_author_or_identifier;

-- Add new constraint that includes sdk_user_id
ALTER TABLE feedback ADD CONSTRAINT chk_feedback_author_or_identifier
    CHECK (
        author_id IS NOT NULL
        OR submitter_email IS NOT NULL
        OR submitter_identifier IS NOT NULL
        OR sdk_user_id IS NOT NULL
    );

-- ============================================
-- Function to update updated_at timestamp
-- ============================================
CREATE OR REPLACE FUNCTION update_sdk_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_sdk_users_updated_at
    BEFORE UPDATE ON sdk_users
    FOR EACH ROW
    EXECUTE FUNCTION update_sdk_users_updated_at();
