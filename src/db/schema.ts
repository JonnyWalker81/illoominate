import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';

export const waitlist = sqliteTable('waitlist', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  email: text('email').notNull().unique(),
  name: text('name'),
  source: text('source'),
  referralCode: text('referral_code').notNull().unique(),
  referredBy: text('referred_by'),
  referralCount: integer('referral_count').default(0),
  verified: integer('verified').default(0),
  verificationToken: text('verification_token'),
  verifiedAt: text('verified_at'),
  createdAt: text('created_at').notNull(),
});

export const quizResponses = sqliteTable('quiz_responses', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  waitlistId: integer('waitlist_id').references(() => waitlist.id),
  role: text('role'), // developer, founder, pm, designer, other
  platform: text('platform'),
  teamSize: text('team_size'),
  disappointmentLevel: text('disappointment_level'), // very, somewhat, not (PMF metric)
  painPoints: text('pain_points'),
  quizCompleted: integer('quiz_completed').default(0), // 1 if completed, 0 if skipped
  createdAt: text('created_at').notNull(),
});
