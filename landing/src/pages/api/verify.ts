import type { APIRoute } from 'astro';

interface Env {
  DB: D1Database;
  RESEND_API_KEY: string;
  RESEND_AUDIENCE_ID: string;
  VERIFICATION_BASE_URL: string;
  FROM_EMAIL: string;
}

interface WaitlistEntry {
  id: number;
  email: string;
  name: string | null;
  email_status: string;
  verification_sent_at: string | null;
  invite_code: string | null;
  referral_code: string | null;
  score: number;
  created_at: string;
}

const VERIFICATION_EXPIRY_HOURS = 24;

export const GET: APIRoute = async ({ request, locals }) => {
  try {
    // Access Cloudflare runtime
    const runtime = (locals as any).runtime;
    if (!runtime) {
      console.error('Runtime not available');
      return jsonResponse({ error: 'Runtime not available' }, 500);
    }

    const env = runtime.env as Env;
    if (!env?.DB) {
      console.error('DB binding not available');
      return jsonResponse({ error: 'Database not configured' }, 500);
    }

    const url = new URL(request.url);
    const token = url.searchParams.get('token');

    if (!token) {
      return jsonResponse({ error: 'Verification token is required.' }, 400);
    }

    // Find entry by verification token
    const entry = await env.DB.prepare(
      `SELECT id, email, name, email_status, verification_sent_at, invite_code, referral_code
       FROM waitlist WHERE verification_token = ?`
    )
      .bind(token)
      .first<Pick<WaitlistEntry, 'id' | 'email' | 'name' | 'email_status' | 'verification_sent_at' | 'invite_code' | 'referral_code'>>();

    if (!entry) {
      return jsonResponse(
        { error: 'Invalid or expired verification link.' },
        404
      );
    }

    if (entry.email_status === 'verified') {
      // Already verified - return position info
      const position = await calculatePosition(env.DB, entry.id);
      return jsonResponse({
        success: true,
        message: 'Your email is already verified.',
        position,
        invite_code: entry.invite_code,
      });
    }

    // Check if token is expired (24 hours)
    if (entry.verification_sent_at) {
      const sentAt = new Date(entry.verification_sent_at);
      const now = new Date();
      const hoursSinceSent = (now.getTime() - sentAt.getTime()) / (1000 * 60 * 60);

      if (hoursSinceSent > VERIFICATION_EXPIRY_HOURS) {
        return jsonResponse(
          { error: 'This verification link has expired. Please request a new one.' },
          410
        );
      }
    }

    // Get client info for consent tracking
    const clientIP = request.headers.get('CF-Connecting-IP') || 'unknown';
    const userAgent = request.headers.get('User-Agent') || '';

    // Mark as verified
    await env.DB.prepare(
      `UPDATE waitlist SET
        email_status = 'verified',
        verified_at = datetime('now'),
        consent_given_at = datetime('now'),
        consent_ip_address = ?,
        consent_user_agent = ?,
        verification_token = NULL
      WHERE id = ?`
    )
      .bind(clientIP, userAgent, entry.id)
      .run();

    // Calculate position
    const position = await calculatePosition(env.DB, entry.id);

    // Sync to Resend audience
    await syncToResend(env, entry.email, entry.name);

    // Send welcome email
    await sendWelcomeEmail(env, entry.email, entry.name, position, entry.invite_code!);

    return jsonResponse({
      success: true,
      message: "Your email has been verified! You're on the waitlist.",
      position,
      invite_code: entry.invite_code,
    });
  } catch (error) {
    console.error('Verify error:', error);
    return jsonResponse(
      { error: 'Something went wrong. Please try again.' },
      500
    );
  }
};

async function calculatePosition(db: D1Database, userId: number): Promise<number> {
  // Get user's score
  const user = await db
    .prepare('SELECT score, created_at FROM waitlist WHERE id = ?')
    .bind(userId)
    .first<Pick<WaitlistEntry, 'score' | 'created_at'>>();

  if (!user) {
    return 0;
  }

  // Count users with higher score, or same score but earlier signup
  const result = await db
    .prepare(
      `SELECT COUNT(*) as count FROM waitlist
       WHERE email_status = 'verified'
       AND (score > ? OR (score = ? AND created_at < ?))`
    )
    .bind(user.score, user.score, user.created_at)
    .first<{ count: number }>();

  return (result?.count || 0) + 1;
}

async function syncToResend(
  env: Env,
  email: string,
  name: string | null
): Promise<void> {
  try {
    const response = await fetch(
      `https://api.resend.com/audiences/${env.RESEND_AUDIENCE_ID}/contacts`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${env.RESEND_API_KEY}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          first_name: name || undefined,
          unsubscribed: false,
        }),
      }
    );

    if (!response.ok) {
      const error = await response.text();
      console.error('Resend sync error:', error);
    }
  } catch (error) {
    console.error('Failed to sync to Resend:', error);
  }
}

async function sendWelcomeEmail(
  env: Env,
  email: string,
  name: string | null,
  position: number,
  inviteCode: string
): Promise<void> {
  const referralLink = `${env.VERIFICATION_BASE_URL}?ref=${inviteCode}`;
  const greeting = name ? `Hi ${name}` : 'Hi there';

  try {
    await fetch('https://api.resend.com/emails', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${env.RESEND_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        from: `Illoominate <${env.FROM_EMAIL}>`,
        to: [email],
        subject: `You're #${position} on the Illoominate waitlist!`,
        html: `
          <div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
            <h1 style="color: #4F46E5;">${greeting}, you're in!</h1>
            <p>Your spot on the Illoominate waitlist is confirmed. You're currently <strong>#${position}</strong> in line.</p>

            <div style="background-color: #F8FAFC; border-radius: 12px; padding: 24px; margin: 24px 0;">
              <h3 style="margin-top: 0; color: #1E293B;">Want to move up?</h3>
              <p style="margin-bottom: 16px;">Share your personal invite link. Each verified signup bumps you up!</p>
              <p style="background-color: #E2E8F0; padding: 12px; border-radius: 8px; font-family: monospace; font-size: 14px; word-break: break-all;">${referralLink}</p>
            </div>

            <p style="color: #64748B;">Your invite code: <strong>${inviteCode}</strong></p>

            <h3 style="color: #1E293B;">What's next?</h3>
            <ul style="color: #475569;">
              <li>We'll notify you when it's your turn</li>
              <li>Early access users get exclusive features</li>
              <li>Top referrers get VIP treatment</li>
            </ul>

            <p>Questions? Just reply to this email.</p>

            <hr style="border: none; border-top: 1px solid #E2E8F0; margin: 32px 0;">
            <p style="color: #94A3B8; font-size: 12px;">Illoominate - Turn user feedback into your product roadmap.</p>
          </div>
        `,
      }),
    });
  } catch (error) {
    console.error('Failed to send welcome email:', error);
  }
}

function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
    },
  });
}
