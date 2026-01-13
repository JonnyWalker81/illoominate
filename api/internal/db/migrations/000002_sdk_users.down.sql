-- Revert sdk_users migration

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_sdk_users_updated_at ON sdk_users;
DROP FUNCTION IF EXISTS update_sdk_users_updated_at();

-- Restore original constraint
ALTER TABLE feedback DROP CONSTRAINT IF EXISTS chk_feedback_author_or_identifier;
ALTER TABLE feedback ADD CONSTRAINT chk_feedback_author_or_identifier
    CHECK (author_id IS NOT NULL OR submitter_email IS NOT NULL OR submitter_identifier IS NOT NULL);

-- Remove sdk_user_id from feedback
DROP INDEX IF EXISTS idx_feedback_sdk_user;
ALTER TABLE feedback DROP COLUMN IF EXISTS sdk_user_id;

-- Drop sdk_users table
DROP TABLE IF EXISTS sdk_users;
