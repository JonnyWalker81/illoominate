-- ============================================
-- FullDisclosure Initial Schema - Rollback
-- ============================================

-- Drop helper functions
DROP FUNCTION IF EXISTS get_user_project_role(UUID, UUID);
DROP FUNCTION IF EXISTS is_admin_or_owner(UUID, UUID);
DROP FUNCTION IF EXISTS can_modify(UUID, UUID);
DROP FUNCTION IF EXISTS is_team_member(UUID, UUID);
DROP FUNCTION IF EXISTS generate_invite_token();
DROP FUNCTION IF EXISTS generate_project_key();

-- Drop triggers
DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS trg_feedback_updated_at ON feedback;
DROP TRIGGER IF EXISTS trg_memberships_updated_at ON memberships;
DROP TRIGGER IF EXISTS trg_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS trg_comments_count ON comments;
DROP TRIGGER IF EXISTS trg_votes_count ON votes;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS update_feedback_comment_count();
DROP FUNCTION IF EXISTS update_feedback_vote_count();

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS feedback_tags;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS sdk_tokens;
DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS projects;

-- Drop enums
DROP TYPE IF EXISTS activity_action;
DROP TYPE IF EXISTS attachment_status;
DROP TYPE IF EXISTS invite_status;
DROP TYPE IF EXISTS severity_level;
DROP TYPE IF EXISTS feedback_status;
DROP TYPE IF EXISTS visibility_level;
DROP TYPE IF EXISTS feedback_type;
DROP TYPE IF EXISTS membership_role;
