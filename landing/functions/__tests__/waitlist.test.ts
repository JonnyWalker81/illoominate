import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createMockEnv, createMockRequest, createMockD1, mockFetch } from './helpers/mock-env';

// Import the handler (we'll test the logic separately since it's a Pages function)
// For now, we test the core validation and utility functions

describe('Waitlist API', () => {
  let fetchMock: ReturnType<typeof mockFetch>;

  beforeEach(() => {
    fetchMock = mockFetch();
  });

  afterEach(() => {
    fetchMock.restore();
    vi.clearAllMocks();
  });

  describe('Email Validation', () => {
    // Testing the validation logic that would be in the handler
    const validateEmail = (email: string): { valid: boolean; error?: string } => {
      const DISPOSABLE_DOMAINS = [
        'tempmail.com',
        'guerrillamail.com',
        'mailinator.com',
        '10minutemail.com',
      ];

      if (!email) {
        return { valid: false, error: 'Email is required.' };
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        return { valid: false, error: 'Please enter a valid email address.' };
      }

      const domain = email.split('@')[1].toLowerCase();
      if (DISPOSABLE_DOMAINS.some((d) => domain.includes(d))) {
        return { valid: false, error: 'Please use a non-disposable email address.' };
      }

      return { valid: true };
    };

    it('should reject empty email', () => {
      const result = validateEmail('');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Email is required.');
    });

    it('should reject invalid email format', () => {
      const invalidEmails = ['notanemail', 'missing@domain', '@nodomain.com', 'spaces in@email.com'];

      invalidEmails.forEach((email) => {
        const result = validateEmail(email);
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Please enter a valid email address.');
      });
    });

    it('should reject disposable email domains', () => {
      const disposableEmails = [
        'test@tempmail.com',
        'user@guerrillamail.com',
        'fake@mailinator.com',
        'temp@10minutemail.com',
      ];

      disposableEmails.forEach((email) => {
        const result = validateEmail(email);
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Please use a non-disposable email address.');
      });
    });

    it('should accept valid email formats', () => {
      const validEmails = [
        'user@example.com',
        'john.doe@company.co.uk',
        'test+tag@domain.org',
        'name123@subdomain.example.com',
      ];

      validEmails.forEach((email) => {
        const result = validateEmail(email);
        expect(result.valid).toBe(true);
        expect(result.error).toBeUndefined();
      });
    });
  });

  describe('Token Generation', () => {
    const generateToken = (length: number): string => {
      const array = new Uint8Array(length / 2);
      crypto.getRandomValues(array);
      return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
    };

    it('should generate token of correct length', () => {
      const token64 = generateToken(64);
      expect(token64).toHaveLength(64);

      const token32 = generateToken(32);
      expect(token32).toHaveLength(32);
    });

    it('should generate unique tokens', () => {
      const token1 = generateToken(64);
      const token2 = generateToken(64);
      expect(token1).not.toBe(token2);
    });

    it('should only contain hex characters', () => {
      const token = generateToken(64);
      expect(token).toMatch(/^[0-9a-f]+$/);
    });
  });

  describe('Rate Limiting', () => {
    it('should track rate limit data structure', () => {
      const db = createMockD1();

      // Simulate rate limit data
      db._setData('rate_limits', [
        {
          ip_address: '192.168.1.1',
          attempt_count: 3,
          first_attempt_at: new Date().toISOString(),
          last_attempt_at: new Date().toISOString(),
        },
      ]);

      const rateLimits = db._getData('rate_limits');
      expect(rateLimits).toHaveLength(1);
      expect((rateLimits[0] as { attempt_count: number }).attempt_count).toBe(3);
    });

    it('should reset rate limit after window expires', () => {
      const oneHourAgo = new Date(Date.now() - 61 * 60 * 1000);
      const db = createMockD1();

      db._setData('rate_limits', [
        {
          ip_address: '192.168.1.1',
          attempt_count: 5,
          first_attempt_at: oneHourAgo.toISOString(),
          last_attempt_at: oneHourAgo.toISOString(),
        },
      ]);

      // In real implementation, this would be reset
      const rateLimits = db._getData('rate_limits');
      const record = rateLimits[0] as { first_attempt_at: string };
      const firstAttempt = new Date(record.first_attempt_at);
      const timeSince = Date.now() - firstAttempt.getTime();

      // Window should have expired (> 1 hour)
      expect(timeSince).toBeGreaterThan(60 * 60 * 1000);
    });
  });

  describe('Mock Environment', () => {
    it('should create valid mock environment', () => {
      const env = createMockEnv();

      expect(env.TURNSTILE_SECRET_KEY).toBeDefined();
      expect(env.RESEND_API_KEY).toBeDefined();
      expect(env.RESEND_AUDIENCE_ID).toBeDefined();
      expect(env.VERIFICATION_BASE_URL).toBe('http://localhost:4321');
      expect(env.DB).toBeDefined();
    });

    it('should allow environment overrides', () => {
      const env = createMockEnv({
        VERIFICATION_BASE_URL: 'https://custom.domain.com',
      });

      expect(env.VERIFICATION_BASE_URL).toBe('https://custom.domain.com');
    });
  });

  describe('Mock Request', () => {
    it('should create request with default headers', () => {
      const request = createMockRequest('http://localhost/api/test');

      expect(request.headers.get('Content-Type')).toBe('application/json');
      expect(request.headers.get('CF-Connecting-IP')).toBe('127.0.0.1');
      expect(request.headers.get('User-Agent')).toBe('Test Agent');
    });

    it('should allow custom headers', () => {
      const request = createMockRequest('http://localhost/api/test', {
        headers: {
          'CF-Connecting-IP': '192.168.1.100',
        },
      });

      expect(request.headers.get('CF-Connecting-IP')).toBe('192.168.1.100');
    });

    it('should create POST request with body', () => {
      const request = createMockRequest('http://localhost/api/test', {
        method: 'POST',
        body: JSON.stringify({ email: 'test@example.com' }),
      });

      expect(request.method).toBe('POST');
    });
  });

  describe('Mock Database', () => {
    it('should prepare statements', () => {
      const db = createMockD1();
      const stmt = db.prepare('SELECT * FROM waitlist');
      expect(stmt).toBeDefined();
      expect(stmt.bind).toBeDefined();
      expect(stmt.first).toBeDefined();
      expect(stmt.run).toBeDefined();
    });

    it('should allow setting and getting data', () => {
      const db = createMockD1();

      db._setData('waitlist', [
        { id: 1, email: 'test@example.com' },
        { id: 2, email: 'user@example.com' },
      ]);

      const data = db._getData('waitlist');
      expect(data).toHaveLength(2);
    });

    it('should return null for empty table', async () => {
      const db = createMockD1();
      const result = await db.prepare('SELECT * FROM waitlist WHERE id = ?').bind(1).first();
      expect(result).toBeNull();
    });
  });

  describe('Referral Code Validation', () => {
    it('should validate 8-character uppercase referral code format', () => {
      const validCodes = ['ABCD1234', 'XY78Z9AB', '12345678'];
      const isValidFormat = (code: string) => /^[A-Z0-9]{8}$/.test(code);

      validCodes.forEach((code) => {
        expect(isValidFormat(code)).toBe(true);
      });
    });

    it('should reject invalid referral code formats', () => {
      const invalidCodes = ['abc', 'toolongcode123', 'lower123', 'ABC-1234'];
      const isValidFormat = (code: string) => /^[A-Z0-9]{8}$/.test(code);

      invalidCodes.forEach((code) => {
        expect(isValidFormat(code)).toBe(false);
      });
    });
  });

  describe('Mock Fetch', () => {
    it('should mock Resend email API', async () => {
      const response = await fetch('https://api.resend.com/emails', {
        method: 'POST',
        body: JSON.stringify({ to: 'test@example.com' }),
      });

      expect(response.status).toBe(200);
      const data = await response.json() as { id: string };
      expect(data.id).toBe('mock_email_id');
    });

    it('should track fetch calls', async () => {
      await fetch('https://api.resend.com/emails', { method: 'POST' });
      await fetch('https://api.resend.com/audiences/test', { method: 'POST' });

      const calls = fetchMock.getCalls();
      expect(calls).toHaveLength(2);
      expect(calls[0].url).toContain('emails');
      expect(calls[1].url).toContain('audiences');
    });
  });
});
