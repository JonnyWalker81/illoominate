export const prerender = false;

import type { APIRoute } from 'astro';
import { drizzle } from 'drizzle-orm/d1';
import { eq, sql, desc, asc } from 'drizzle-orm';
import { Resend } from 'resend';
import { waitlist } from '../../db/schema';
import { waitlistSchema } from '../../lib/validation';
import { generateReferralCode } from '../../lib/referral';
import WelcomeEmail from '../../emails/WelcomeEmail';

interface Env {
  DB: D1Database;
  RESEND_API_KEY: string;
}

export const POST: APIRoute = async ({ request, locals }) => {
  const env = (locals as { runtime?: { env: Env } }).runtime?.env;
  
  if (!env?.DB) {
    return new Response(JSON.stringify({ error: 'Database not configured' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const db = drizzle(env.DB);

  try {
    // Parse form data
    const formData = await request.formData();
    const data = {
      email: formData.get('email') as string,
      name: (formData.get('name') as string) || undefined,
      source: (formData.get('source') as string) || undefined,
      referredBy: (formData.get('referredBy') as string) || undefined,
    };

    // Validate input
    const parsed = waitlistSchema.safeParse(data);
    if (!parsed.success) {
      return new Response(JSON.stringify({ error: parsed.error.errors[0].message }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const { email, name, source, referredBy } = parsed.data;

    // Check if email already exists
    const existing = await db
      .select()
      .from(waitlist)
      .where(eq(waitlist.email, email))
      .limit(1);

    if (existing.length > 0) {
      // Return existing entry's position and code
      const position = await getWaitlistPosition(db, existing[0].id);
      return new Response(JSON.stringify({
        success: true,
        position,
        referralCode: existing[0].referralCode,
        waitlistId: existing[0].id,
        message: 'You are already on the waitlist!',
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    // Generate unique referral code
    let referralCode = generateReferralCode();
    let attempts = 0;
    while (attempts < 10) {
      const existingCode = await db
        .select()
        .from(waitlist)
        .where(eq(waitlist.referralCode, referralCode))
        .limit(1);
      if (existingCode.length === 0) break;
      referralCode = generateReferralCode();
      attempts++;
    }

    // Insert new waitlist entry
    const [entry] = await db
      .insert(waitlist)
      .values({
        email,
        name: name || null,
        source: source || null,
        referralCode,
        referredBy: referredBy || null,
        referralCount: 0,
        createdAt: new Date().toISOString(),
      })
      .returning();

    // If referred, increment referrer's count
    if (referredBy) {
      await db
        .update(waitlist)
        .set({ referralCount: sql`${waitlist.referralCount} + 1` })
        .where(eq(waitlist.referralCode, referredBy));
    }

    // Calculate position
    const position = await getWaitlistPosition(db, entry.id);

    // Send confirmation email
    if (env.RESEND_API_KEY && env.RESEND_API_KEY !== 're_xxxxxxxxxxxx') {
      try {
        const resend = new Resend(env.RESEND_API_KEY);
        await resend.emails.send({
          from: 'Illoominate <noreply@illoominate.app>',
          to: email,
          subject: `You're #${position} on the Illoominate waitlist!`,
          react: WelcomeEmail({ name: name || undefined, position, referralCode }),
        });
      } catch (emailError) {
        console.error('Email send failed:', emailError);
      }
    }

    return new Response(JSON.stringify({
      success: true,
      position,
      referralCode,
      waitlistId: entry.id,
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    console.error('Waitlist signup error:', error);
    return new Response(JSON.stringify({ error: 'Something went wrong' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};

/**
 * Calculate waitlist position.
 * Ordered by: referral_count DESC, created_at ASC
 * Higher referrals = better position, earlier signup breaks ties.
 */
async function getWaitlistPosition(
  db: ReturnType<typeof drizzle>,
  entryId: number
): Promise<number> {
  // Get all entries ordered by position algorithm
  const entries = await db
    .select({ id: waitlist.id })
    .from(waitlist)
    .orderBy(desc(waitlist.referralCount), asc(waitlist.createdAt));

  const position = entries.findIndex((e) => e.id === entryId) + 1;
  return position || entries.length;
}
