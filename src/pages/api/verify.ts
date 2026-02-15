export const prerender = false;

import type { APIRoute } from 'astro';
import { drizzle } from 'drizzle-orm/d1';
import { eq, sql, desc, asc } from 'drizzle-orm';
import { Resend } from 'resend';
import { waitlist } from '../../db/schema';
import WelcomeEmail from '../../emails/WelcomeEmail';

interface Env {
  DB: D1Database;
  RESEND_API_KEY: string;
}

export const GET: APIRoute = async ({ request, locals, redirect }) => {
  const env = (locals as { runtime?: { env: Env } }).runtime?.env;

  if (!env?.DB) {
    return redirect('/verified?error=config');
  }

  const db = drizzle(env.DB);
  const url = new URL(request.url);
  const token = url.searchParams.get('token');

  if (!token) {
    return redirect('/verified?error=missing_token');
  }

  try {
    // Find entry by verification token
    const existing = await db
      .select()
      .from(waitlist)
      .where(eq(waitlist.verificationToken, token))
      .limit(1);

    if (existing.length === 0) {
      return redirect('/verified?error=invalid_token');
    }

    const entry = existing[0];

    // Check if already verified
    if (entry.verified) {
      const position = await getWaitlistPosition(db, entry.id);
      return redirect(`/verified?success=true&position=${position}&code=${entry.referralCode}&wid=${entry.id}`);
    }

    // Mark as verified
    await db
      .update(waitlist)
      .set({
        verified: 1,
        verificationToken: null,
        verifiedAt: new Date().toISOString(),
      })
      .where(eq(waitlist.id, entry.id));

    // Increment referrer's count if this entry was referred
    if (entry.referredBy) {
      await db
        .update(waitlist)
        .set({ referralCount: sql`${waitlist.referralCount} + 1` })
        .where(eq(waitlist.referralCode, entry.referredBy));
    }

    // Calculate position (now that they're verified)
    const position = await getWaitlistPosition(db, entry.id);

    // Send welcome email with position
    if (env.RESEND_API_KEY && env.RESEND_API_KEY !== 're_xxxxxxxxxxxx') {
      try {
        const resend = new Resend(env.RESEND_API_KEY);
        await resend.emails.send({
          from: 'Illoominate <noreply@illoominate.app>',
          to: entry.email,
          subject: `You're #${position} on the Illoominate waitlist!`,
          react: WelcomeEmail({
            name: entry.name || undefined,
            position,
            referralCode: entry.referralCode,
          }),
        });
      } catch (emailError) {
        console.error('Welcome email send failed:', emailError);
      }
    }

    return redirect(`/verified?success=true&position=${position}&code=${entry.referralCode}&wid=${entry.id}`);

  } catch (error) {
    console.error('Verification error:', error);
    return redirect('/verified?error=server');
  }
};

/**
 * Calculate waitlist position (only counts verified entries).
 */
async function getWaitlistPosition(
  db: ReturnType<typeof drizzle>,
  entryId: number
): Promise<number> {
  const entries = await db
    .select({ id: waitlist.id })
    .from(waitlist)
    .where(eq(waitlist.verified, 1))
    .orderBy(desc(waitlist.referralCount), asc(waitlist.createdAt));

  const position = entries.findIndex((e) => e.id === entryId) + 1;
  return position || entries.length;
}
