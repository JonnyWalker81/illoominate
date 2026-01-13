import { describe, it, expect } from 'vitest';
import {
  calculateResult,
  getQuizResult,
  getResultFromScores,
  quizQuestions,
  quizResults,
  type QuizResultType,
} from './quiz-logic';

describe('quizQuestions', () => {
  it('should have exactly 5 questions', () => {
    expect(quizQuestions).toHaveLength(5);
  });

  it('should have sequential IDs starting from 1', () => {
    quizQuestions.forEach((question, index) => {
      expect(question.id).toBe(index + 1);
    });
  });

  it('should have 4 options per question', () => {
    quizQuestions.forEach((question) => {
      expect(question.options).toHaveLength(4);
    });
  });

  it('should have non-empty question text', () => {
    quizQuestions.forEach((question) => {
      expect(question.question).toBeTruthy();
      expect(question.question.length).toBeGreaterThan(10);
    });
  });

  it('should have valid option values (1-4)', () => {
    quizQuestions.forEach((question) => {
      question.options.forEach((option) => {
        expect(option.value).toBeGreaterThanOrEqual(1);
        expect(option.value).toBeLessThanOrEqual(4);
      });
    });
  });

  it('should have non-empty option text', () => {
    quizQuestions.forEach((question) => {
      question.options.forEach((option) => {
        expect(option.text).toBeTruthy();
        expect(option.text.length).toBeGreaterThan(0);
      });
    });
  });
});

describe('quizResults', () => {
  const resultTypes: QuizResultType[] = [
    'feedback_firefighter',
    'feedback_gatherer',
    'feedback_pro',
    'feedback_champion',
  ];

  it('should have all 4 result types', () => {
    expect(Object.keys(quizResults)).toHaveLength(4);
    resultTypes.forEach((type) => {
      expect(quizResults[type]).toBeDefined();
    });
  });

  it('should have matching type field in each result', () => {
    resultTypes.forEach((type) => {
      expect(quizResults[type].type).toBe(type);
    });
  });

  it('should have non-empty title for each result', () => {
    resultTypes.forEach((type) => {
      expect(quizResults[type].title).toBeTruthy();
      expect(quizResults[type].title.length).toBeGreaterThan(0);
    });
  });

  it('should have emoji for each result', () => {
    resultTypes.forEach((type) => {
      expect(quizResults[type].emoji).toBeTruthy();
    });
  });

  it('should have non-empty description for each result', () => {
    resultTypes.forEach((type) => {
      expect(quizResults[type].description).toBeTruthy();
      expect(quizResults[type].description.length).toBeGreaterThan(20);
    });
  });

  it('should have at least 2 recommendations for each result', () => {
    resultTypes.forEach((type) => {
      expect(quizResults[type].recommendations.length).toBeGreaterThanOrEqual(2);
      quizResults[type].recommendations.forEach((rec) => {
        expect(rec).toBeTruthy();
        expect(rec.length).toBeGreaterThan(0);
      });
    });
  });
});

describe('calculateResult', () => {
  describe('score range: 5-8 (Feedback Firefighter)', () => {
    it('should return feedback_firefighter for score 5 (minimum possible)', () => {
      expect(calculateResult([1, 1, 1, 1, 1])).toBe('feedback_firefighter');
    });

    it('should return feedback_firefighter for score 6', () => {
      expect(calculateResult([1, 1, 1, 1, 2])).toBe('feedback_firefighter');
    });

    it('should return feedback_firefighter for score 7', () => {
      expect(calculateResult([1, 1, 1, 2, 2])).toBe('feedback_firefighter');
    });

    it('should return feedback_firefighter for score 8', () => {
      expect(calculateResult([1, 1, 2, 2, 2])).toBe('feedback_firefighter');
    });
  });

  describe('score range: 9-12 (Feedback Gatherer)', () => {
    it('should return feedback_gatherer for score 9', () => {
      expect(calculateResult([1, 2, 2, 2, 2])).toBe('feedback_gatherer');
    });

    it('should return feedback_gatherer for score 10', () => {
      expect(calculateResult([2, 2, 2, 2, 2])).toBe('feedback_gatherer');
    });

    it('should return feedback_gatherer for score 11', () => {
      expect(calculateResult([2, 2, 2, 2, 3])).toBe('feedback_gatherer');
    });

    it('should return feedback_gatherer for score 12', () => {
      expect(calculateResult([2, 2, 2, 3, 3])).toBe('feedback_gatherer');
    });
  });

  describe('score range: 13-16 (Feedback Pro)', () => {
    it('should return feedback_pro for score 13', () => {
      expect(calculateResult([2, 2, 3, 3, 3])).toBe('feedback_pro');
    });

    it('should return feedback_pro for score 14', () => {
      expect(calculateResult([2, 3, 3, 3, 3])).toBe('feedback_pro');
    });

    it('should return feedback_pro for score 15', () => {
      expect(calculateResult([3, 3, 3, 3, 3])).toBe('feedback_pro');
    });

    it('should return feedback_pro for score 16', () => {
      expect(calculateResult([3, 3, 3, 3, 4])).toBe('feedback_pro');
    });
  });

  describe('score range: 17-20 (Feedback Champion)', () => {
    it('should return feedback_champion for score 17', () => {
      expect(calculateResult([3, 3, 3, 4, 4])).toBe('feedback_champion');
    });

    it('should return feedback_champion for score 18', () => {
      expect(calculateResult([3, 3, 4, 4, 4])).toBe('feedback_champion');
    });

    it('should return feedback_champion for score 19', () => {
      expect(calculateResult([3, 4, 4, 4, 4])).toBe('feedback_champion');
    });

    it('should return feedback_champion for score 20 (maximum possible)', () => {
      expect(calculateResult([4, 4, 4, 4, 4])).toBe('feedback_champion');
    });
  });

  it('should handle empty array (edge case)', () => {
    expect(calculateResult([])).toBe('feedback_firefighter');
  });
});

describe('getQuizResult', () => {
  it('should return correct result object for each type', () => {
    const types: QuizResultType[] = [
      'feedback_firefighter',
      'feedback_gatherer',
      'feedback_pro',
      'feedback_champion',
    ];

    types.forEach((type) => {
      const result = getQuizResult(type);
      expect(result).toEqual(quizResults[type]);
    });
  });

  it('should return result with all required fields', () => {
    const result = getQuizResult('feedback_firefighter');
    expect(result).toHaveProperty('type');
    expect(result).toHaveProperty('title');
    expect(result).toHaveProperty('emoji');
    expect(result).toHaveProperty('description');
    expect(result).toHaveProperty('recommendations');
  });
});

describe('getResultFromScores', () => {
  it('should return full result object for low scores', () => {
    const result = getResultFromScores([1, 1, 1, 1, 1]);
    expect(result.type).toBe('feedback_firefighter');
    expect(result.title).toBe('Feedback Firefighter');
    expect(result.emoji).toBe('🔥');
  });

  it('should return full result object for medium-low scores', () => {
    const result = getResultFromScores([2, 2, 2, 2, 2]);
    expect(result.type).toBe('feedback_gatherer');
    expect(result.title).toBe('Feedback Gatherer');
    expect(result.emoji).toBe('📥');
  });

  it('should return full result object for medium-high scores', () => {
    const result = getResultFromScores([3, 3, 3, 3, 3]);
    expect(result.type).toBe('feedback_pro');
    expect(result.title).toBe('Feedback Pro');
    expect(result.emoji).toBe('⭐');
  });

  it('should return full result object for high scores', () => {
    const result = getResultFromScores([4, 4, 4, 4, 4]);
    expect(result.type).toBe('feedback_champion');
    expect(result.title).toBe('Feedback Champion');
    expect(result.emoji).toBe('🏆');
  });

  it('should include recommendations array', () => {
    const result = getResultFromScores([1, 1, 1, 1, 1]);
    expect(Array.isArray(result.recommendations)).toBe(true);
    expect(result.recommendations.length).toBeGreaterThan(0);
  });
});
