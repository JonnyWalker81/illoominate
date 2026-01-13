-- Rollback: Feedback Portal

-- Drop index on sdk_users
DROP INDEX IF EXISTS idx_sdk_users_linked_user;

-- Drop notification queue
DROP TABLE IF EXISTS notification_queue;

-- Drop portal votes
DROP TABLE IF EXISTS portal_votes;

-- Drop portal user profiles
DROP TABLE IF EXISTS portal_user_profiles;
