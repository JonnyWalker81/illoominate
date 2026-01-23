export const prerender = false;

import type { APIRoute } from 'astro';
import { drizzle } from 'drizzle-orm/d1';
import { desc, asc } from 'drizzle-orm';
import { waitlist, quizResponses } from '../../../db/schema';

interface Env {
  DB: D1Database;
  ADMIN_TOKEN: string;
}

export const GET: APIRoute = async ({ request, locals }) => {
  const env = (locals as { runtime?: { env: Env } }).runtime?.env;

  if (!env?.DB || !env?.ADMIN_TOKEN) {
    return new Response(JSON.stringify({ error: 'Server not configured' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Check admin token
  const authHeader = request.headers.get('Authorization');
  const token = authHeader?.replace('Bearer ', '');

  if (!token || token !== env.ADMIN_TOKEN) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const db = drizzle(env.DB);
  const url = new URL(request.url);
  const format = url.searchParams.get('format') || 'json';

  try {
    // Get all waitlist entries
    const entries = await db
      .select()
      .from(waitlist)
      .orderBy(desc(waitlist.referralCount), asc(waitlist.createdAt));

    // Get quiz responses separately
    const quizzes = await db.select().from(quizResponses);
    const quizMap = new Map(quizzes.map(q => [q.waitlistId, q]));

    // Build response data with positions (only count verified for position)
    const verifiedEntries = entries.filter(e => e.verified);
    const data = entries.map((entry) => {
      const quiz = quizMap.get(entry.id);
      // Position is only for verified entries
      const position = entry.verified
        ? verifiedEntries.findIndex(e => e.id === entry.id) + 1
        : null;
      return {
        position,
        email: entry.email,
        name: entry.name,
        source: entry.source,
        referralCode: entry.referralCode,
        referredBy: entry.referredBy,
        referralCount: entry.referralCount,
        verified: entry.verified === 1,
        verifiedAt: entry.verifiedAt,
        createdAt: entry.createdAt,
        quiz: quiz ? {
          platform: quiz.platform,
          teamSize: quiz.teamSize,
          painPoints: quiz.painPoints,
        } : null,
      };
    });

    if (format === 'csv') {
      const csv = generateCSV(data);
      return new Response(csv, {
        headers: {
          'Content-Type': 'text/csv',
          'Content-Disposition': `attachment; filename="waitlist-${new Date().toISOString().split('T')[0]}.csv"`,
        },
      });
    }

    return new Response(JSON.stringify({ entries: data, total: data.length }), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error('Admin API error:', error);
    return new Response(JSON.stringify({ error: 'Something went wrong' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};

interface WaitlistEntry {
  position: number | null;
  email: string;
  name: string | null;
  source: string | null;
  referralCode: string;
  referredBy: string | null;
  referralCount: number | null;
  verified: boolean;
  verifiedAt: string | null;
  createdAt: string;
  quiz: {
    platform: string | null;
    teamSize: string | null;
    painPoints: string | null;
  } | null;
}

function generateCSV(data: WaitlistEntry[]): string {
  const headers = [
    'Position', 'Email', 'Name', 'Source', 'Referral Code',
    'Referred By', 'Referral Count', 'Verified', 'Verified At', 'Signed Up',
    'Platform', 'Team Size', 'Pain Points'
  ];

  const rows = data.map(d => [
    d.position ?? '-',
    d.email,
    d.name || '',
    d.source || '',
    d.referralCode,
    d.referredBy || '',
    d.referralCount ?? 0,
    d.verified ? 'Yes' : 'No',
    d.verifiedAt || '',
    d.createdAt,
    d.quiz?.platform || '',
    d.quiz?.teamSize || '',
    (d.quiz?.painPoints || '').replace(/"/g, '""'),
  ].map(v => `"${v}"`).join(','));

  return [headers.join(','), ...rows].join('\n');
}
