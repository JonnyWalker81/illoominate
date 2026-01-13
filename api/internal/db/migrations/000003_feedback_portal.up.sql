-- Migration: Feedback Portal
-- Creates tables for portal user profiles, votes, and notification queue
-- Adds index for SDK user linking

-- Portal user profiles: stores notification preferences for Supabase users per project
CREATE TABLE portal_user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,  -- References Supabase auth.users.id
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    notification_preferences JSONB NOT NULL DEFAULT '{"status_changes": true, "new_comments_on_my_feedback": true}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_portal_user_project UNIQUE (user_id, project_id)
);

CREATE INDEX idx_portal_user_profiles_user ON portal_user_profiles(user_id);
CREATE INDEX idx_portal_user_profiles_project ON portal_user_profiles(project_id);

-- Trigger to auto-update updated_at
CREATE TRIGGER trg_portal_user_profiles_updated_at
    BEFORE UPDATE ON portal_user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Portal votes: tracks votes from portal users (separate from team member votes)
CREATE TABLE portal_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,  -- References Supabase auth.users.id
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_portal_votes UNIQUE (feedback_id, user_id)
);

CREATE INDEX idx_portal_votes_feedback ON portal_votes(feedback_id);
CREATE INDEX idx_portal_votes_user ON portal_votes(user_id);
CREATE INDEX idx_portal_votes_project ON portal_votes(project_id);

-- Notification queue: async email notifications for portal users
CREATE TABLE notification_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,  -- References Supabase auth.users.id
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    sent_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_queue_pending ON notification_queue(created_at)
    WHERE sent_at IS NULL AND failed_at IS NULL;
CREATE INDEX idx_notification_queue_user ON notification_queue(user_id);

-- Add index for SDK user linking (speeds up queries for linked_user_id)
CREATE INDEX idx_sdk_users_linked_user ON sdk_users(linked_user_id)
    WHERE linked_user_id IS NOT NULL;
