-- Illoominate Waitlist Database Schema
-- D1 SQLite Database for storing waitlist signups with full email management

-- ============================================================================
-- MAIN WAITLIST TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS waitlist (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  -- Basic contact information
  email TEXT NOT NULL UNIQUE,
  name TEXT,

  -- Email verification & status
  email_status TEXT NOT NULL DEFAULT 'pending'
    CHECK(email_status IN ('pending', 'verified', 'bounced', 'invalid', 'unsubscribed')),
  verification_token TEXT UNIQUE,
  verification_sent_at DATETIME,
  verified_at DATETIME,
  verification_attempts INTEGER DEFAULT 0,

  -- Consent & compliance (GDPR)
  consent_given_at DATETIME,
  consent_ip_address TEXT,
  consent_user_agent TEXT,
  privacy_policy_version TEXT DEFAULT 'v1.0',
  marketing_consent BOOLEAN DEFAULT true,

  -- Unsubscribe handling
  unsubscribed_at DATETIME,
  unsubscribe_reason TEXT,
  unsubscribe_token TEXT UNIQUE,

  -- Waitlist position & referrals (score-based ranking)
  score REAL DEFAULT 0,
  invite_code TEXT UNIQUE,
  referral_code TEXT,           -- Code they used to sign up (who referred them)
  referrals_count INTEGER DEFAULT 0,
  tier TEXT DEFAULT 'standard'
    CHECK(tier IN ('standard', 'early_access', 'vip', 'beta_tester')),
  is_vip BOOLEAN DEFAULT false,

  -- Quiz results
  quiz_completed BOOLEAN DEFAULT false,
  quiz_score INTEGER,
  quiz_result_type TEXT CHECK(quiz_result_type IN (
    'feedback_firefighter', 'feedback_gatherer',
    'feedback_pro', 'feedback_champion'
  )),
  quiz_responses TEXT,          -- JSON array of responses
  quiz_session_id TEXT,         -- Links to quiz_responses.session_id

  -- Attribution tracking (how did you hear about us)
  referral_source TEXT CHECK(referral_source IN (
    'reddit', 'search', 'social', 'friend', 'blog', 'other'
  )),
  ip_address TEXT,
  user_agent TEXT,

  -- Custom data
  tags TEXT,                    -- JSON array for flexible tagging
  custom_metadata TEXT,         -- JSON object for additional data

  -- Resend integration
  resend_contact_id TEXT UNIQUE,
  resend_audience_id TEXT,
  last_synced_to_resend_at DATETIME,
  resend_sync_status TEXT DEFAULT 'pending'
    CHECK(resend_sync_status IN ('pending', 'synced', 'failed')),

  -- Engagement tracking
  engagement_score INTEGER DEFAULT 0,
  last_email_opened_at DATETIME,
  last_email_clicked_at DATETIME,
  total_emails_sent INTEGER DEFAULT 0,
  total_emails_opened INTEGER DEFAULT 0,
  total_emails_clicked INTEGER DEFAULT 0,

  -- Timestamps
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- QUIZ RESPONSES TABLE (detailed tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS quiz_responses (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  waitlist_id INTEGER,
  session_id TEXT NOT NULL,
  question_index INTEGER NOT NULL,
  question_text TEXT NOT NULL,
  selected_option TEXT NOT NULL,
  option_value INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (waitlist_id) REFERENCES waitlist(id) ON DELETE SET NULL
);

-- ============================================================================
-- RATE LIMITING TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS rate_limits (
  ip_address TEXT PRIMARY KEY,
  attempt_count INTEGER DEFAULT 1,
  first_attempt_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  last_attempt_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- EMAIL EVENTS TABLE (from Resend webhooks)
-- ============================================================================

CREATE TABLE IF NOT EXISTS email_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  waitlist_id INTEGER NOT NULL,
  campaign_id INTEGER,
  event_type TEXT NOT NULL
    CHECK(event_type IN ('sent', 'delivered', 'opened', 'clicked',
                          'bounced', 'complained', 'unsubscribed', 'delivery_delayed')),
  resend_email_id TEXT,
  link_url TEXT,                -- For click events
  user_agent TEXT,
  ip_address TEXT,
  event_timestamp DATETIME NOT NULL,
  raw_webhook_data TEXT,        -- Full JSON payload
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (waitlist_id) REFERENCES waitlist(id) ON DELETE CASCADE
);

-- ============================================================================
-- WAITLIST TABLE INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_waitlist_email ON waitlist(email);
CREATE INDEX IF NOT EXISTS idx_waitlist_email_status ON waitlist(email_status);
CREATE INDEX IF NOT EXISTS idx_waitlist_verification_token ON waitlist(verification_token);
CREATE INDEX IF NOT EXISTS idx_waitlist_unsubscribe_token ON waitlist(unsubscribe_token);
CREATE INDEX IF NOT EXISTS idx_waitlist_score ON waitlist(score DESC);
CREATE INDEX IF NOT EXISTS idx_waitlist_invite_code ON waitlist(invite_code);
CREATE INDEX IF NOT EXISTS idx_waitlist_referral_code ON waitlist(referral_code);
CREATE INDEX IF NOT EXISTS idx_waitlist_tier ON waitlist(tier);
CREATE INDEX IF NOT EXISTS idx_waitlist_quiz_result ON waitlist(quiz_result_type);
CREATE INDEX IF NOT EXISTS idx_waitlist_quiz_session_id ON waitlist(quiz_session_id);
CREATE INDEX IF NOT EXISTS idx_waitlist_resend_contact_id ON waitlist(resend_contact_id);
CREATE INDEX IF NOT EXISTS idx_waitlist_resend_sync_status ON waitlist(resend_sync_status);
CREATE INDEX IF NOT EXISTS idx_waitlist_engagement_score ON waitlist(engagement_score DESC);
CREATE INDEX IF NOT EXISTS idx_waitlist_created_at ON waitlist(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_waitlist_referral_source ON waitlist(referral_source);
CREATE INDEX IF NOT EXISTS idx_waitlist_marketing_consent ON waitlist(marketing_consent);
CREATE INDEX IF NOT EXISTS idx_waitlist_is_vip ON waitlist(is_vip);

-- ============================================================================
-- OTHER TABLE INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_quiz_responses_session ON quiz_responses(session_id);
CREATE INDEX IF NOT EXISTS idx_quiz_responses_waitlist ON quiz_responses(waitlist_id);
CREATE INDEX IF NOT EXISTS idx_rate_limits_last_attempt ON rate_limits(last_attempt_at);
CREATE INDEX IF NOT EXISTS idx_email_events_waitlist_id ON email_events(waitlist_id);
CREATE INDEX IF NOT EXISTS idx_email_events_type ON email_events(event_type);
CREATE INDEX IF NOT EXISTS idx_email_events_timestamp ON email_events(event_timestamp DESC);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update timestamp on any row update
CREATE TRIGGER IF NOT EXISTS update_waitlist_timestamp
AFTER UPDATE ON waitlist
BEGIN
  UPDATE waitlist SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Calculate initial score on insert
-- Score = (CONSTANT - signup_timestamp), higher = earlier signup
-- CONSTANT = 2524608000 (Unix timestamp for Jan 1, 2050)
CREATE TRIGGER IF NOT EXISTS calculate_initial_score
AFTER INSERT ON waitlist
BEGIN
  UPDATE waitlist
  SET score = 2524608000 - CAST(strftime('%s', NEW.created_at) AS INTEGER)
  WHERE id = NEW.id;
END;

-- Auto-generate invite code on insert (8 uppercase hex characters)
CREATE TRIGGER IF NOT EXISTS generate_invite_code
AFTER INSERT ON waitlist
BEGIN
  UPDATE waitlist
  SET invite_code = UPPER(SUBSTR(HEX(RANDOMBLOB(4)), 1, 8))
  WHERE id = NEW.id AND invite_code IS NULL;
END;

-- Recalculate score when email is verified (adds referral bonus)
-- Each referral adds 86400 seconds (1 day) worth of score
CREATE TRIGGER IF NOT EXISTS recalculate_score_on_verify
AFTER UPDATE OF email_status ON waitlist
WHEN NEW.email_status = 'verified' AND OLD.email_status != 'verified'
BEGIN
  UPDATE waitlist
  SET score = (2524608000 - CAST(strftime('%s', NEW.created_at) AS INTEGER))
            + (NEW.referrals_count * 86400)
  WHERE id = NEW.id;
END;

-- Increment referral count when someone uses a referral code
CREATE TRIGGER IF NOT EXISTS increment_referral_count
AFTER UPDATE OF email_status ON waitlist
WHEN NEW.email_status = 'verified' AND OLD.email_status != 'verified'
     AND NEW.referral_code IS NOT NULL
BEGIN
  UPDATE waitlist
  SET referrals_count = referrals_count + 1,
      score = (2524608000 - CAST(strftime('%s', created_at) AS INTEGER))
            + ((referrals_count + 1) * 86400)
  WHERE invite_code = NEW.referral_code;
END;

-- ============================================================================
-- USEFUL QUERIES (for reference)
-- ============================================================================

-- Get waitlist with dynamic positions (score-based ranking)
-- SELECT
--   email, name, score, referrals_count,
--   ROW_NUMBER() OVER (ORDER BY score DESC, created_at ASC) as position
-- FROM waitlist
-- WHERE email_status = 'verified'
-- ORDER BY score DESC, created_at ASC;

-- Get a single user's position
-- SELECT
--   email, score,
--   (SELECT COUNT(*) FROM waitlist w2
--    WHERE w2.email_status = 'verified'
--    AND (w2.score > w1.score OR (w2.score = w1.score AND w2.created_at < w1.created_at))) + 1 as position
-- FROM waitlist w1
-- WHERE email = 'user@example.com' AND email_status = 'verified';

-- Top referrers
-- SELECT email, name, invite_code, referrals_count
-- FROM waitlist
-- WHERE referrals_count > 0
-- ORDER BY referrals_count DESC
-- LIMIT 10;

-- Cleanup old rate limits (run hourly)
-- DELETE FROM rate_limits WHERE last_attempt_at < datetime('now', '-1 hour');
