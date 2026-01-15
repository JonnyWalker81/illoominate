import type { APIRoute } from 'astro';

interface Env {
  DB: D1Database;
}

interface QuizSubmitRequest {
  session_id: string;
  responses: {
    question_index: number;
    question_text: string;
    selected_option: string;
    option_value: number;
  }[];
  email?: string;
}

type QuizResultType =
  | 'feedback_firefighter'
  | 'feedback_gatherer'
  | 'feedback_pro'
  | 'feedback_champion';

interface QuizResult {
  type: QuizResultType;
  title: string;
  emoji: string;
  description: string;
  recommendations: string[];
}

const quizResults: Record<QuizResultType, QuizResult> = {
  feedback_firefighter: {
    type: 'feedback_firefighter',
    title: 'Feedback Firefighter',
    emoji: '🔥',
    description:
      "You're putting out fires instead of preventing them. Illoominate will help you get organized and stop losing valuable user insights.",
    recommendations: [
      'Centralize all feedback in one place',
      'Set up automated feedback collection',
      'Create a weekly feedback review process',
    ],
  },
  feedback_gatherer: {
    type: 'feedback_gatherer',
    title: 'Feedback Gatherer',
    emoji: '📥',
    description:
      "You're collecting feedback but struggling to act on it effectively. Illoominate will help you prioritize and close the loop with users.",
    recommendations: [
      'Implement a voting system for feature requests',
      'Tag and categorize feedback by theme',
      'Start sharing a public roadmap',
    ],
  },
  feedback_pro: {
    type: 'feedback_pro',
    title: 'Feedback Pro',
    emoji: '⭐',
    description:
      "You have decent systems but there's room for improvement. Illoominate will streamline your workflow and add powerful automation.",
    recommendations: [
      'Add real-time status updates for features',
      'Enable user notifications when requests ship',
      'Integrate feedback into your product analytics',
    ],
  },
  feedback_champion: {
    type: 'feedback_champion',
    title: 'Feedback Champion',
    emoji: '🏆',
    description:
      "Impressive! You've got great practices. Illoominate can still help you scale and add AI-powered features.",
    recommendations: [
      'Explore AI-powered duplicate detection',
      'Set up advanced analytics and reporting',
      'Consider enterprise features for larger teams',
    ],
  },
};

function calculateResult(scores: number[]): QuizResultType {
  const total = scores.reduce((sum, score) => sum + score, 0);

  if (total <= 8) {
    return 'feedback_firefighter';
  }
  if (total <= 12) {
    return 'feedback_gatherer';
  }
  if (total <= 16) {
    return 'feedback_pro';
  }
  return 'feedback_champion';
}

export const POST: APIRoute = async ({ request, locals }) => {
  const runtime = (locals as any).runtime;
  const env = runtime.env as Env;

  try {
    const body = (await request.json()) as QuizSubmitRequest;
    const { session_id, responses, email } = body;

    // Validate request
    if (!session_id || !responses || !Array.isArray(responses)) {
      return jsonResponse({ error: 'Invalid request format.' }, 400);
    }

    if (responses.length !== 5) {
      return jsonResponse({ error: 'Quiz requires exactly 5 responses.' }, 400);
    }

    // Calculate score
    const scores = responses.map((r) => r.option_value);
    const totalScore = scores.reduce((sum, score) => sum + score, 0);
    const resultType = calculateResult(scores);
    const result = quizResults[resultType];

    // Find waitlist entry if email provided
    let waitlistId: number | null = null;
    if (email) {
      const entry = await env.DB.prepare(
        'SELECT id FROM waitlist WHERE email = ?'
      )
        .bind(email.toLowerCase())
        .first<{ id: number }>();

      if (entry) {
        waitlistId = entry.id;

        // Update waitlist with quiz results
        await env.DB.prepare(
          `UPDATE waitlist SET
            quiz_completed = true,
            quiz_score = ?,
            quiz_result_type = ?,
            quiz_responses = ?,
            quiz_session_id = ?
          WHERE id = ?`
        )
          .bind(totalScore, resultType, JSON.stringify(responses), session_id, waitlistId)
          .run();
      }
    }

    // Store individual quiz responses
    const stmt = env.DB.prepare(
      `INSERT INTO quiz_responses (
        waitlist_id, session_id, question_index, question_text,
        selected_option, option_value
      ) VALUES (?, ?, ?, ?, ?, ?)`
    );

    const batch = responses.map((response) =>
      stmt.bind(
        waitlistId,
        session_id,
        response.question_index,
        response.question_text,
        response.selected_option,
        response.option_value
      )
    );

    await env.DB.batch(batch);

    return jsonResponse({
      success: true,
      score: totalScore,
      result_type: resultType,
      title: result.title,
      emoji: result.emoji,
      description: result.description,
      recommendations: result.recommendations,
    });
  } catch (error) {
    console.error('Quiz submit error:', error);
    return jsonResponse(
      { error: 'Something went wrong. Please try again.' },
      500
    );
  }
};

function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
    },
  });
}
