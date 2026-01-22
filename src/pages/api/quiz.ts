export const prerender = false;

import type { APIRoute } from 'astro';
import { drizzle } from 'drizzle-orm/d1';
import { quizResponses } from '../../db/schema';
import { quizSchema } from '../../lib/validation';

interface Env {
  DB: D1Database;
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
      platform: (formData.get('platform') as string) || undefined,
      teamSize: (formData.get('teamSize') as string) || undefined,
      painPoints: (formData.get('painPoints') as string) || undefined,
    };

    const parsed = quizSchema.safeParse(data);
    if (!parsed.success) {
      return new Response(JSON.stringify({ error: parsed.error.errors[0].message }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    await db.insert(quizResponses).values({
      waitlistId: parsed.data.waitlistId,
      platform: parsed.data.platform || null,
      teamSize: parsed.data.teamSize || null,
      painPoints: parsed.data.painPoints || null,
      createdAt: new Date().toISOString(),
    });

    return new Response(JSON.stringify({ success: true }), {
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
