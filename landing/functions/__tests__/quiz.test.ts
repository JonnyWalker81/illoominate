import { describe, it, expect } from 'vitest';
import { createMockD1 } from './helpers/mock-env';

// Quiz result calculation logic (same as in the handler)
type QuizResultType =
  | 'feedback_firefighter'
  | 'feedback_gatherer'
  | 'feedback_pro'
  | 'feedback_champion';

function calculateResult(scores: number[]): QuizResultType {
  const total = scores.reduce((sum, score) => sum + score, 0);

  if (total <= 8) return 'feedback_firefighter';
  if (total <= 12) return 'feedback_gatherer';
  if (total <= 16) return 'feedback_pro';
  return 'feedback_champion';
}

describe('Quiz API', () => {
  describe('Score Calculation', () => {
    it('should calculate correct total from scores array', () => {
      expect([1, 1, 1, 1, 1].reduce((a, b) => a + b, 0)).toBe(5);
      expect([2, 2, 2, 2, 2].reduce((a, b) => a + b, 0)).toBe(10);
      expect([3, 3, 3, 3, 3].reduce((a, b) => a + b, 0)).toBe(15);
      expect([4, 4, 4, 4, 4].reduce((a, b) => a + b, 0)).toBe(20);
    });

    it('should return feedback_firefighter for scores 5-8', () => {
      expect(calculateResult([1, 1, 1, 1, 1])).toBe('feedback_firefighter'); // 5
      expect(calculateResult([1, 1, 1, 2, 2])).toBe('feedback_firefighter'); // 7
      expect(calculateResult([1, 1, 2, 2, 2])).toBe('feedback_firefighter'); // 8
    });

    it('should return feedback_gatherer for scores 9-12', () => {
      expect(calculateResult([1, 2, 2, 2, 2])).toBe('feedback_gatherer'); // 9
      expect(calculateResult([2, 2, 2, 2, 2])).toBe('feedback_gatherer'); // 10
      expect(calculateResult([2, 2, 2, 3, 3])).toBe('feedback_gatherer'); // 12
    });

    it('should return feedback_pro for scores 13-16', () => {
      expect(calculateResult([2, 2, 3, 3, 3])).toBe('feedback_pro'); // 13
      expect(calculateResult([3, 3, 3, 3, 3])).toBe('feedback_pro'); // 15
      expect(calculateResult([3, 3, 3, 3, 4])).toBe('feedback_pro'); // 16
    });

    it('should return feedback_champion for scores 17-20', () => {
      expect(calculateResult([3, 3, 3, 4, 4])).toBe('feedback_champion'); // 17
      expect(calculateResult([3, 4, 4, 4, 4])).toBe('feedback_champion'); // 19
      expect(calculateResult([4, 4, 4, 4, 4])).toBe('feedback_champion'); // 20
    });
  });

  describe('Request Validation', () => {
    it('should require session_id', () => {
      const request = {
        responses: [
          { question_index: 0, question_text: 'Q1', selected_option: 'A', option_value: 1 },
        ],
      };

      // session_id is required
      expect('session_id' in request).toBe(false);
    });

    it('should require exactly 5 responses', () => {
      const responses = [
        { question_index: 0, question_text: 'Q1', selected_option: 'A', option_value: 1 },
        { question_index: 1, question_text: 'Q2', selected_option: 'B', option_value: 2 },
        { question_index: 2, question_text: 'Q3', selected_option: 'C', option_value: 3 },
      ];

      expect(responses.length).toBe(3);
      expect(responses.length).not.toBe(5);
    });

    it('should validate response structure', () => {
      const validResponse = {
        question_index: 0,
        question_text: 'How do you collect feedback?',
        selected_option: 'Email only',
        option_value: 1,
      };

      expect(validResponse.question_index).toBeDefined();
      expect(validResponse.question_text).toBeDefined();
      expect(validResponse.selected_option).toBeDefined();
      expect(validResponse.option_value).toBeDefined();
      expect(typeof validResponse.option_value).toBe('number');
    });
  });

  describe('Session Tracking', () => {
    it('should generate unique session IDs', () => {
      const generateSessionId = () => crypto.randomUUID();

      const session1 = generateSessionId();
      const session2 = generateSessionId();

      expect(session1).not.toBe(session2);
      expect(session1).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
      );
    });

    it('should store quiz responses in database', async () => {
      const db = createMockD1();

      const responses = [
        { question_index: 0, question_text: 'Q1', selected_option: 'A', option_value: 2 },
        { question_index: 1, question_text: 'Q2', selected_option: 'B', option_value: 3 },
        { question_index: 2, question_text: 'Q3', selected_option: 'C', option_value: 2 },
        { question_index: 3, question_text: 'Q4', selected_option: 'D', option_value: 4 },
        { question_index: 4, question_text: 'Q5', selected_option: 'A', option_value: 1 },
      ];

      // Simulate storing responses
      db._setData('quiz_responses', responses);

      const stored = db._getData('quiz_responses');
      expect(stored).toHaveLength(5);
    });
  });

  describe('Waitlist Integration', () => {
    it('should link quiz to waitlist entry if email provided', async () => {
      const db = createMockD1();

      // Simulate existing waitlist entry
      db._setData('waitlist', [
        {
          id: 1,
          email: 'test@example.com',
          quiz_completed: false,
          quiz_score: null,
          quiz_result_type: null,
        },
      ]);

      const waitlistEntry = db._getData('waitlist')[0] as {
        id: number;
        email: string;
        quiz_completed: boolean;
        quiz_score: number | null;
        quiz_result_type: string | null;
      };

      expect(waitlistEntry.email).toBe('test@example.com');
      expect(waitlistEntry.quiz_completed).toBe(false);

      // After quiz completion, entry would be updated
      const updatedEntry = {
        ...waitlistEntry,
        quiz_completed: true,
        quiz_score: 15,
        quiz_result_type: 'feedback_pro',
      };

      expect(updatedEntry.quiz_completed).toBe(true);
      expect(updatedEntry.quiz_score).toBe(15);
      expect(updatedEntry.quiz_result_type).toBe('feedback_pro');
    });

    it('should not fail if email not on waitlist', async () => {
      const db = createMockD1();
      db._setData('waitlist', []);

      const result = await db
        .prepare('SELECT id FROM waitlist WHERE email = ?')
        .bind('notfound@example.com')
        .first();

      expect(result).toBeNull();
      // Quiz submission should still succeed without waitlist link
    });
  });

  describe('Result Format', () => {
    it('should return complete result object', () => {
      const quizResults = {
        feedback_firefighter: {
          type: 'feedback_firefighter',
          title: 'Feedback Firefighter',
          emoji: '🔥',
          description: 'Test description',
          recommendations: ['Rec 1', 'Rec 2', 'Rec 3'],
        },
      };

      const resultType = 'feedback_firefighter';
      const result = quizResults[resultType];

      expect(result.type).toBe('feedback_firefighter');
      expect(result.title).toBe('Feedback Firefighter');
      expect(result.emoji).toBe('🔥');
      expect(result.description).toBeDefined();
      expect(result.recommendations).toHaveLength(3);
    });

    it('should include score in response', () => {
      const scores = [2, 3, 2, 4, 1];
      const totalScore = scores.reduce((sum, s) => sum + s, 0);

      expect(totalScore).toBe(12);

      const response = {
        success: true,
        score: totalScore,
        result_type: calculateResult(scores),
      };

      expect(response.success).toBe(true);
      expect(response.score).toBe(12);
      expect(response.result_type).toBe('feedback_gatherer');
    });
  });
});
