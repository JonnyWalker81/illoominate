import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';

export const waitlist = sqliteTable('waitlist', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  email: text('email').notNull().unique(),
  name: text('name'),
  source: text('source'),
  referralCode: text('referral_code').notNull().unique(),
  referredBy: text('referred_by'),
  referralCount: integer('referral_count').default(0),
  createdAt: text('created_at').notNull(),
});

export const quizResponses = sqliteTable('quiz_responses', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  waitlistId: integer('waitlist_id').references(() => waitlist.id),
  platform: text('platform'),
  teamSize: text('team_size'),
  painPoints: text('pain_points'),
  createdAt: text('created_at').notNull(),
});
