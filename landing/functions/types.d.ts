/// <reference types="@cloudflare/workers-types" />

export interface Env {
  // D1 Database
  DB: D1Database;

  // Environment variables
  TURNSTILE_SECRET_KEY: string;
  RESEND_API_KEY: string;
  RESEND_AUDIENCE_ID: string;
  RESEND_WEBHOOK_SECRET: string;
  VERIFICATION_BASE_URL: string;
  ADMIN_EMAIL: string;
  FROM_EMAIL: string;
}

export interface WaitlistEntry {
  id: number;
  email: string;
  name: string | null;
  email_status: 'pending' | 'verified' | 'bounced' | 'invalid' | 'unsubscribed';
  verification_token: string | null;
  verification_sent_at: string | null;
  verified_at: string | null;
  verification_attempts: number;
  consent_given_at: string | null;
  consent_ip_address: string | null;
  consent_user_agent: string | null;
  privacy_policy_version: string;
  marketing_consent: boolean;
  unsubscribed_at: string | null;
  unsubscribe_reason: string | null;
  unsubscribe_token: string | null;
  score: number;
  invite_code: string | null;
  referral_code: string | null;
  referrals_count: number;
  tier: 'standard' | 'early_access' | 'vip' | 'beta_tester';
  is_vip: boolean;
  quiz_completed: boolean;
  quiz_score: number | null;
  quiz_result_type: string | null;
  quiz_responses: string | null;
  referral_source: string | null;
  ip_address: string | null;
  user_agent: string | null;
  tags: string | null;
  custom_metadata: string | null;
  resend_contact_id: string | null;
  resend_audience_id: string | null;
  last_synced_to_resend_at: string | null;
  resend_sync_status: 'pending' | 'synced' | 'failed';
  engagement_score: number;
  last_email_opened_at: string | null;
  last_email_clicked_at: string | null;
  total_emails_sent: number;
  total_emails_opened: number;
  total_emails_clicked: number;
  created_at: string;
  updated_at: string;
}

export interface RateLimit {
  ip_address: string;
  attempt_count: number;
  first_attempt_at: string;
  last_attempt_at: string;
}

export interface WaitlistJoinRequest {
  email: string;
  name?: string;
  referral_code?: string;
  referral_source?: 'reddit' | 'search' | 'social' | 'friend' | 'blog' | 'other';
  turnstile_token?: string;
}

export interface QuizSubmitRequest {
  session_id: string;
  responses: {
    question_index: number;
    question_text: string;
    selected_option: string;
    option_value: number;
  }[];
  email?: string;
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
}
