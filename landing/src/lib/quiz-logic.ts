export interface QuizQuestion {
  id: number;
  question: string;
  options: QuizOption[];
}

export interface QuizOption {
  text: string;
  value: number;
}

export type QuizResultType =
  | 'feedback_firefighter'
  | 'feedback_gatherer'
  | 'feedback_pro'
  | 'feedback_champion';

export interface QuizResult {
  type: QuizResultType;
  title: string;
  emoji: string;
  description: string;
  recommendations: string[];
}

export const quizQuestions: QuizQuestion[] = [
  {
    id: 1,
    question: 'How do you currently collect user feedback?',
    options: [
      { text: 'Email and support tickets only', value: 1 },
      { text: 'GitHub issues or forum', value: 2 },
      { text: 'Scattered across multiple channels', value: 1 },
      { text: 'Dedicated feedback tool', value: 3 },
    ],
  },
  {
    id: 2,
    question: 'How do you prioritize which features to build?',
    options: [
      { text: 'Gut feeling / loudest customers', value: 1 },
      { text: 'Internal team decisions only', value: 1 },
      { text: 'Some user input, mostly internal', value: 2 },
      { text: 'Data-driven with user voting', value: 4 },
    ],
  },
  {
    id: 3,
    question: 'How long does it take to find if a feature has been requested before?',
    options: [
      { text: 'Minutes - we have good organization', value: 4 },
      { text: '15-30 minutes of searching', value: 2 },
      { text: 'Could take hours across platforms', value: 1 },
      { text: "No idea - we'd just build it", value: 1 },
    ],
  },
  {
    id: 4,
    question: 'How do users know their feedback was received?',
    options: [
      { text: 'Automated acknowledgment + status updates', value: 4 },
      { text: 'Manual email reply', value: 2 },
      { text: 'They can check a public board', value: 3 },
      { text: "They don't - it's a black hole", value: 1 },
    ],
  },
  {
    id: 5,
    question: 'How do you communicate when requested features ship?',
    options: [
      { text: 'Changelog + direct notifications', value: 4 },
      { text: 'Blog post or release notes', value: 2 },
      { text: 'Social media announcement', value: 1 },
      { text: "We don't specifically notify requesters", value: 1 },
    ],
  },
];

export const quizResults: Record<QuizResultType, QuizResult> = {
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

/**
 * Calculate the quiz result type based on the scores from each question.
 * @param scores - Array of scores (one per question)
 * @returns The result type based on total score
 */
export function calculateResult(scores: number[]): QuizResultType {
  const total = scores.reduce((sum, score) => sum + score, 0);

  // Score ranges:
  // 5-8: Feedback Firefighter
  // 9-12: Feedback Gatherer
  // 13-16: Feedback Pro
  // 17-20: Feedback Champion
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

/**
 * Get the full result object based on the result type.
 * @param type - The quiz result type
 * @returns The full quiz result object
 */
export function getQuizResult(type: QuizResultType): QuizResult {
  return quizResults[type];
}

/**
 * Calculate result and return the full result object.
 * @param scores - Array of scores (one per question)
 * @returns The full quiz result object
 */
export function getResultFromScores(scores: number[]): QuizResult {
  const type = calculateResult(scores);
  return getQuizResult(type);
}
