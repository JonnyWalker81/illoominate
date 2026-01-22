import { z } from 'zod';

export const waitlistSchema = z.object({
  email: z.string().email('Invalid email address'),
  name: z.string().max(100).optional(),
  source: z.string().max(200).optional(),
  referredBy: z.string().length(6).optional().transform(val => val?.toUpperCase()),
});

export type WaitlistInput = z.infer<typeof waitlistSchema>;

export const quizSchema = z.object({
  waitlistId: z.number().int().positive(),
  role: z.enum(['developer', 'founder', 'pm', 'designer', 'other']).optional(),
  platform: z.enum(['ios', 'android', 'web', 'multiple']).optional(),
  teamSize: z.enum(['solo', 'small', 'medium', 'large']).optional(),
  disappointmentLevel: z.enum(['very', 'somewhat', 'not']).optional(),
  painPoints: z.string().max(1000).optional(),
});

export type QuizInput = z.infer<typeof quizSchema>;
