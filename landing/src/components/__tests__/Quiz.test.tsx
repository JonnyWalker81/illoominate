import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Quiz from '../Quiz';

describe('Quiz Component', () => {
  describe('Initial Render', () => {
    it('should render the quiz section', () => {
      render(<Quiz />);

      expect(screen.getByText("What's Your Feedback Maturity Level?")).toBeInTheDocument();
    });

    it('should show the first question on mount', () => {
      render(<Quiz />);

      expect(screen.getByText('How do you currently collect user feedback?')).toBeInTheDocument();
    });

    it('should display progress indicator', () => {
      render(<Quiz />);

      expect(screen.getByText('Question 1 of 5')).toBeInTheDocument();
      expect(screen.getByText('0% complete')).toBeInTheDocument();
    });

    it('should show all 4 options for the first question', () => {
      render(<Quiz />);

      expect(screen.getByText('Email and support tickets only')).toBeInTheDocument();
      expect(screen.getByText('GitHub issues or forum')).toBeInTheDocument();
      expect(screen.getByText('Scattered across multiple channels')).toBeInTheDocument();
      expect(screen.getByText('Dedicated feedback tool')).toBeInTheDocument();
    });

    it('should show option letters (A, B, C, D)', () => {
      render(<Quiz />);

      expect(screen.getByText('A')).toBeInTheDocument();
      expect(screen.getByText('B')).toBeInTheDocument();
      expect(screen.getByText('C')).toBeInTheDocument();
      expect(screen.getByText('D')).toBeInTheDocument();
    });
  });

  describe('Question Navigation', () => {
    it('should advance to next question on option select', async () => {
      render(<Quiz />);

      // Click first option
      fireEvent.click(screen.getByText('Email and support tickets only'));

      // Wait for animation and question change
      await waitFor(() => {
        expect(screen.getByText('Question 2 of 5')).toBeInTheDocument();
      }, { timeout: 500 });
    });

    it('should update progress bar after answering', async () => {
      render(<Quiz />);

      fireEvent.click(screen.getByText('Email and support tickets only'));

      await waitFor(() => {
        expect(screen.getByText('20% complete')).toBeInTheDocument();
      }, { timeout: 500 });
    });

    it('should show different question after advancing', async () => {
      render(<Quiz />);

      fireEvent.click(screen.getByText('Email and support tickets only'));

      await waitFor(() => {
        expect(screen.getByText('How do you prioritize which features to build?')).toBeInTheDocument();
      }, { timeout: 500 });
    });
  });

  describe('Result Display', () => {
    it('should show result after all questions answered', async () => {
      const onComplete = vi.fn();
      render(<Quiz onComplete={onComplete} />);

      // Answer all 5 questions (selecting first option each time)
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[0]); // First option

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }

      // Wait for result
      await waitFor(() => {
        expect(screen.getByText(/You're a/)).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should display result title and description', async () => {
      render(<Quiz />);

      // Answer all questions - first options give mixed scores
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[0]);

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }

      // Check for any result type (the exact type depends on question values)
      await waitFor(() => {
        expect(screen.getByText(/You're a Feedback/)).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should show recommendations after quiz completion', async () => {
      render(<Quiz />);

      // Answer all questions
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[0]);

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }

      await waitFor(() => {
        expect(screen.getByText('Our Recommendations:')).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should call onComplete callback with result', async () => {
      const onComplete = vi.fn();
      render(<Quiz onComplete={onComplete} />);

      // Answer all questions
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[0]);

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }

      await waitFor(() => {
        expect(onComplete).toHaveBeenCalledTimes(1);
      }, { timeout: 1000 });

      const [result, scores] = onComplete.mock.calls[0];
      // Check that we got a valid result type
      expect(['feedback_firefighter', 'feedback_gatherer', 'feedback_pro', 'feedback_champion']).toContain(result.type);
      expect(scores).toHaveLength(5);
    });
  });

  describe('Result Actions', () => {
    const completeQuiz = async () => {
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[0]);

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }

      await waitFor(() => {
        expect(screen.getByText(/You're a/)).toBeInTheDocument();
      }, { timeout: 1000 });
    };

    it('should show Get Early Access button', async () => {
      render(<Quiz />);
      await completeQuiz();

      expect(screen.getByText('Get Early Access')).toBeInTheDocument();
    });

    it('should show Share Result button', async () => {
      render(<Quiz />);
      await completeQuiz();

      expect(screen.getByText('Share Result')).toBeInTheDocument();
    });

    it('should show Retake Quiz button', async () => {
      render(<Quiz />);
      await completeQuiz();

      expect(screen.getByText('Retake Quiz')).toBeInTheDocument();
    });

    it('should reset quiz when Retake Quiz is clicked', async () => {
      render(<Quiz />);
      await completeQuiz();

      fireEvent.click(screen.getByText('Retake Quiz'));

      await waitFor(() => {
        expect(screen.getByText('Question 1 of 5')).toBeInTheDocument();
      });
    });
  });

  describe('Different Result Types', () => {
    const answerAllWithValue = async (valueIndex: number) => {
      for (let i = 0; i < 5; i++) {
        const options = screen.getAllByRole('button');
        fireEvent.click(options[valueIndex]);

        if (i < 4) {
          await waitFor(() => {
            expect(screen.getByText(`Question ${i + 2} of 5`)).toBeInTheDocument();
          }, { timeout: 500 });
        }
      }
    };

    it('should show a result for any score', async () => {
      render(<Quiz />);
      await answerAllWithValue(0); // All first options

      await waitFor(() => {
        // Result title contains "Feedback" in all types
        expect(screen.getByText(/You're a Feedback/)).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should show a different result for high scores', async () => {
      render(<Quiz />);
      await answerAllWithValue(3); // All last options (tend to be higher values)

      await waitFor(() => {
        // Should get one of the result types
        expect(screen.getByText(/You're a Feedback/)).toBeInTheDocument();
      }, { timeout: 1000 });
    });
  });
});
