export const prerender = false;

import type { APIRoute } from 'astro';
import { drizzle } from 'drizzle-orm/d1';
import { eq, sql } from 'drizzle-orm';
import { quizResponses, waitlist } from '../../db/schema';
import { quizSchema } from '../../lib/validation';

interface Env {
  DB: D1Database;
}

// Position boost for completing quiz (equivalent to 5 referrals)
const QUIZ_COMPLETION_BOOST = 5;

/**
 * Generate personalized type based on quiz answers
 * Uses role + platform + team size to create a memorable label
 */
function generatePersonalizedType(
  role?: string,
  platform?: string,
  teamSize?: string
): string {
  let prefix = '';
  if (teamSize === 'solo') {
    prefix = 'Solo ';
  } else if (teamSize === 'large') {
    prefix = 'Enterprise ';
  }

  let suffix = '';
  switch (role) {
    case 'founder':
      suffix = 'Founder';
      break;
    case 'developer':
      suffix = 'Developer';
      break;
    case 'pm':
      suffix = 'Product Lead';
      break;
    case 'designer':
      suffix = 'Designer';
      break;
    default:
      suffix = 'Builder';
  }

  let platformName = '';
  switch (platform) {
    case 'ios':
      platformName = 'iOS';
      break;
    case 'android':
      platformName = 'Android';
      break;
    case 'web':
      platformName = 'Web';
      break;
    case 'multiple':
      return `${prefix}Multi-Platform ${suffix}`.trim();
    default:
      platformName = 'Digital';
  }

  return `${prefix}${platformName} ${suffix}`.trim();
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
    const formData = await request.formData();
    const data = {
      waitlistId: Number(formData.get('waitlistId')),
      role: (formData.get('role') as string) || undefined,
      platform: (formData.get('platform') as string) || undefined,
      teamSize: (formData.get('teamSize') as string) || undefined,
      disappointmentLevel: (formData.get('disappointmentLevel') as string) || undefined,
      painPoints: (formData.get('painPoints') as string) || undefined,
    };

    const parsed = quizSchema.safeParse(data);
    if (!parsed.success) {
      return new Response(JSON.stringify({ error: parsed.error.errors[0].message }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    // Check if quiz was meaningfully completed (at least PMF question answered)
    const quizCompleted = parsed.data.disappointmentLevel ? 1 : 0;

    // Insert quiz response
    await db.insert(quizResponses).values({
      waitlistId: parsed.data.waitlistId,
      role: parsed.data.role || null,
      platform: parsed.data.platform || null,
      teamSize: parsed.data.teamSize || null,
      disappointmentLevel: parsed.data.disappointmentLevel || null,
      painPoints: parsed.data.painPoints || null,
      quizCompleted,
      createdAt: new Date().toISOString(),
    });

    // Generate personalized type
    const personalizedType = generatePersonalizedType(
      parsed.data.role,
      parsed.data.platform,
      parsed.data.teamSize
    );

    // Apply position boost if quiz completed
    // Boost position by incrementing referral count (position sorted by referralCount desc)
    let positionBoost = 0;
    if (quizCompleted) {
      await db
        .update(waitlist)
        .set({ 
          referralCount: sql`${waitlist.referralCount} + ${QUIZ_COMPLETION_BOOST}`
        })
        .where(eq(waitlist.id, parsed.data.waitlistId));
      
      positionBoost = QUIZ_COMPLETION_BOOST;
    }

    return new Response(JSON.stringify({ 
      success: true,
      type: personalizedType,
      positionBoost,
      quizCompleted: Boolean(quizCompleted),
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error('Quiz submission error:', error);
    return new Response(JSON.stringify({ error: 'Something went wrong' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
