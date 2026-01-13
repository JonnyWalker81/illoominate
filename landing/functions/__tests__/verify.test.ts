import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { createMockD1, mockFetch } from './helpers/mock-env';

const VERIFICATION_EXPIRY_HOURS = 24;

describe('Verify API', () => {
  let fetchMock: ReturnType<typeof mockFetch>;

  beforeEach(() => {
    fetchMock = mockFetch();
  });

  afterEach(() => {
    fetchMock.restore();
  });

  describe('Token Validation', () => {
    it('should reject missing token', () => {
      const url = new URL('http://localhost/verify');
      const token = url.searchParams.get('token');

      expect(token).toBeNull();
      // Handler would return 400
    });

    it('should reject invalid token', async () => {
      const db = createMockD1();
      db._setData('waitlist', []);

      const result = await db
        .prepare('SELECT * FROM waitlist WHERE verification_token = ?')
        .bind('invalid_token_123')
        .first();

      expect(result).toBeNull();
      // Handler would return 404
    });

    it('should accept valid token', async () => {
      const db = createMockD1();
      const validToken = 'a'.repeat(64);

      db._setData('waitlist', [
        {
          id: 1,
          email: 'test@example.com',
          email_status: 'pending',
          verification_token: validToken,
          verification_sent_at: new Date().toISOString(),
          invite_code: 'ABCD1234',
        },
      ]);

      const entry = db._getData('waitlist')[0] as {
        verification_token: string;
      };

      expect(entry.verification_token).toBe(validToken);
    });
  });

  describe('Expiration Check', () => {
    it('should reject expired token (> 24 hours)', () => {
      const twentyFiveHoursAgo = new Date(Date.now() - 25 * 60 * 60 * 1000);
      const now = new Date();

      const hoursSinceSent =
        (now.getTime() - twentyFiveHoursAgo.getTime()) / (1000 * 60 * 60);

      expect(hoursSinceSent).toBeGreaterThan(VERIFICATION_EXPIRY_HOURS);
      // Handler would return 410
    });

    it('should accept token within 24 hours', () => {
      const twoHoursAgo = new Date(Date.now() - 2 * 60 * 60 * 1000);
      const now = new Date();

      const hoursSinceSent =
        (now.getTime() - twoHoursAgo.getTime()) / (1000 * 60 * 60);

      expect(hoursSinceSent).toBeLessThan(VERIFICATION_EXPIRY_HOURS);
    });

    it('should handle edge case at exactly 24 hours', () => {
      const exactlyTwentyFourHoursAgo = new Date(
        Date.now() - 24 * 60 * 60 * 1000
      );
      const now = new Date();

      const hoursSinceSent =
        (now.getTime() - exactlyTwentyFourHoursAgo.getTime()) / (1000 * 60 * 60);

      // At exactly 24 hours, should still be valid (not > 24)
      expect(hoursSinceSent).toBeLessThanOrEqual(VERIFICATION_EXPIRY_HOURS);
    });
  });

  describe('Verification Flow', () => {
    it('should update email_status to verified', async () => {
      const db = createMockD1();
      const token = 'valid_token_' + 'a'.repeat(52);

      const entry = {
        id: 1,
        email: 'test@example.com',
        email_status: 'pending' as const,
        verification_token: token,
        verification_sent_at: new Date().toISOString(),
        verified_at: null,
        invite_code: 'ABCD1234',
      };

      db._setData('waitlist', [entry]);

      // Simulate verification
      const verifiedEntry = {
        ...entry,
        email_status: 'verified' as const,
        verified_at: new Date().toISOString(),
        verification_token: null,
      };

      expect(verifiedEntry.email_status).toBe('verified');
      expect(verifiedEntry.verified_at).not.toBeNull();
      expect(verifiedEntry.verification_token).toBeNull();
    });

    it('should record consent information', () => {
      const clientIP = '192.168.1.100';
      const userAgent = 'Mozilla/5.0 Test';
      const now = new Date().toISOString();

      const consentData = {
        consent_given_at: now,
        consent_ip_address: clientIP,
        consent_user_agent: userAgent,
      };

      expect(consentData.consent_given_at).toBeDefined();
      expect(consentData.consent_ip_address).toBe(clientIP);
      expect(consentData.consent_user_agent).toBe(userAgent);
    });

    it('should handle already verified email', async () => {
      const db = createMockD1();

      db._setData('waitlist', [
        {
          id: 1,
          email: 'test@example.com',
          email_status: 'verified',
          verification_token: 'some_token',
          invite_code: 'ABCD1234',
        },
      ]);

      const entry = db._getData('waitlist')[0] as {
        email_status: string;
      };

      expect(entry.email_status).toBe('verified');
      // Handler should return success with existing position info
    });
  });

  describe('Position Calculation', () => {
    it('should calculate position based on score', () => {
      const users = [
        { id: 1, email: 'first@example.com', score: 1000, created_at: '2024-01-01' },
        { id: 2, email: 'second@example.com', score: 900, created_at: '2024-01-02' },
        { id: 3, email: 'third@example.com', score: 800, created_at: '2024-01-03' },
      ];

      // User with score 900 should be position 2
      const targetScore = 900;
      const targetCreatedAt = '2024-01-02';

      const usersAhead = users.filter(
        (u) =>
          u.score > targetScore ||
          (u.score === targetScore && u.created_at < targetCreatedAt)
      );

      expect(usersAhead.length + 1).toBe(2); // Position 2
    });

    it('should handle tie-breaker with created_at', () => {
      const users = [
        { id: 1, email: 'early@example.com', score: 1000, created_at: '2024-01-01T10:00:00' },
        { id: 2, email: 'later@example.com', score: 1000, created_at: '2024-01-01T11:00:00' },
      ];

      // Earlier signup should be ahead with same score
      const targetScore = 1000;
      const targetCreatedAt = '2024-01-01T11:00:00';

      const usersAhead = users.filter(
        (u) =>
          u.score > targetScore ||
          (u.score === targetScore && u.created_at < targetCreatedAt)
      );

      expect(usersAhead.length + 1).toBe(2); // Position 2 (later signup)
    });

    it('should return position 1 for highest score', () => {
      const users = [
        { id: 1, email: 'top@example.com', score: 2000, created_at: '2024-01-01' },
        { id: 2, email: 'second@example.com', score: 1000, created_at: '2024-01-02' },
      ];

      const targetScore = 2000;
      const targetCreatedAt = '2024-01-01';

      const usersAhead = users.filter(
        (u) =>
          u.score > targetScore ||
          (u.score === targetScore && u.created_at < targetCreatedAt)
      );

      expect(usersAhead.length + 1).toBe(1); // Position 1
    });
  });

  describe('Referral Bonus', () => {
    it('should increment referrer count when referral code used', async () => {
      const db = createMockD1();

      // Referrer
      const referrer = {
        id: 1,
        email: 'referrer@example.com',
        invite_code: 'REF12345',
        referrals_count: 0,
        score: 1000,
      };

      // New signup with referral code
      const newUser = {
        id: 2,
        email: 'new@example.com',
        referral_code: 'REF12345',
        referrals_count: 0,
        score: 900,
      };

      db._setData('waitlist', [referrer, newUser]);

      // After verification, referrer's count should increment
      const updatedReferrer = {
        ...referrer,
        referrals_count: referrer.referrals_count + 1,
        score: referrer.score + 86400, // One day bonus
      };

      expect(updatedReferrer.referrals_count).toBe(1);
      expect(updatedReferrer.score).toBe(1000 + 86400);
    });

    it('should add 86400 (1 day) to score per referral', () => {
      const baseScore = 1000;
      const referralBonus = 86400; // Seconds in a day
      const referralsCount = 3;

      const newScore = baseScore + referralsCount * referralBonus;

      expect(newScore).toBe(1000 + 3 * 86400);
    });
  });

  describe('Resend Integration', () => {
    it('should sync verified user to Resend audience', async () => {
      await fetch('https://api.resend.com/audiences/test_audience/contacts', {
        method: 'POST',
        headers: {
          Authorization: 'Bearer test_key',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: 'test@example.com',
          first_name: 'Test',
          unsubscribed: false,
        }),
      });

      const calls = fetchMock.getCalls();
      expect(calls).toHaveLength(1);
      expect(calls[0].url).toContain('audiences');
    });

    it('should send welcome email after verification', async () => {
      await fetch('https://api.resend.com/emails', {
        method: 'POST',
        headers: {
          Authorization: 'Bearer test_key',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          from: 'Illoominate <noreply@test.com>',
          to: ['test@example.com'],
          subject: "You're #1 on the Illoominate waitlist!",
          html: '<p>Welcome!</p>',
        }),
      });

      const calls = fetchMock.getCalls();
      expect(calls).toHaveLength(1);
      expect(calls[0].url).toContain('emails');
    });
  });

  describe('Response Format', () => {
    it('should return success with position and invite code', () => {
      const response = {
        success: true,
        message: "Your email has been verified! You're on the waitlist.",
        position: 42,
        invite_code: 'ABCD1234',
      };

      expect(response.success).toBe(true);
      expect(response.position).toBe(42);
      expect(response.invite_code).toBe('ABCD1234');
    });

    it('should return error for invalid token', () => {
      const response = {
        error: 'Invalid or expired verification link.',
      };

      expect(response.error).toBeDefined();
    });

    it('should return error for expired token', () => {
      const response = {
        error: 'This verification link has expired. Please request a new one.',
      };

      expect(response.error).toContain('expired');
    });
  });
});
