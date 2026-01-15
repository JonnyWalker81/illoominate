import { useState } from 'react';
import {
  quizQuestions,
  getResultFromScores,
  type QuizResult,
} from '../lib/quiz-logic';

export interface QuizResponse {
  question_index: number;
  question_text: string;
  selected_option: string;
  option_value: number;
}

interface QuizProps {
  onComplete?: (result: QuizResult, scores: number[], responses: QuizResponse[]) => void;
}

export default function Quiz({ onComplete }: QuizProps) {
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [scores, setScores] = useState<number[]>([]);
  const [responses, setResponses] = useState<QuizResponse[]>([]);
  const [result, setResult] = useState<QuizResult | null>(null);
  const [isAnimating, setIsAnimating] = useState(false);

  const handleAnswer = (value: number, optionText: string) => {
    if (isAnimating) return;

    setIsAnimating(true);
    const newScores = [...scores, value];
    setScores(newScores);

    const question = quizQuestions[currentQuestion];
    const newResponse: QuizResponse = {
      question_index: currentQuestion,
      question_text: question.question,
      selected_option: optionText,
      option_value: value,
    };
    const newResponses = [...responses, newResponse];
    setResponses(newResponses);

    setTimeout(() => {
      if (currentQuestion < quizQuestions.length - 1) {
        setCurrentQuestion(currentQuestion + 1);
        setIsAnimating(false);
      } else {
        const quizResult = getResultFromScores(newScores);
        setResult(quizResult);
        onComplete?.(quizResult, newScores, newResponses);
        setIsAnimating(false);
      }
    }, 300);
  };

  const resetQuiz = () => {
    setCurrentQuestion(0);
    setScores([]);
    setResponses([]);
    setResult(null);
  };

  const shareResult = () => {
    const text = `I got "${result?.title}" on the Illoominate Feedback Maturity Quiz! ${result?.emoji} Take the quiz: ${window.location.origin}#quiz`;
    if (navigator.share) {
      navigator.share({ text }).catch(() => {
        navigator.clipboard.writeText(text);
      });
    } else {
      navigator.clipboard.writeText(text);
    }
  };

  const progress = ((currentQuestion + (result ? 1 : 0)) / quizQuestions.length) * 100;

  if (result) {
    return (
      <section id="quiz" className="section-padding bg-gradient-to-br from-slate-900 via-indigo-950 to-purple-950">
        <div className="max-w-3xl mx-auto">
          <div className="bg-slate-800 rounded-2xl shadow-xl p-8 md:p-12 text-center animate-fade-in border border-slate-700">
            <div className="text-6xl mb-4">{result.emoji}</div>
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              You're a {result.title}!
            </h2>
            <p className="text-lg text-slate-300 mb-8 max-w-xl mx-auto">
              {result.description}
            </p>

            <div className="bg-slate-700/50 rounded-xl p-6 mb-8 text-left">
              <h3 className="font-semibold text-white mb-4">Our Recommendations:</h3>
              <ul className="space-y-3">
                {result.recommendations.map((rec, index) => (
                  <li key={index} className="flex items-start gap-3">
                    <svg className="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    <span className="text-slate-300">{rec}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="#waitlist" className="btn-primary">
                Get Early Access
              </a>
              <button onClick={shareResult} className="btn-secondary">
                Share Result
              </button>
              <button onClick={resetQuiz} className="btn-ghost">
                Retake Quiz
              </button>
            </div>
          </div>
        </div>
      </section>
    );
  }

  const question = quizQuestions[currentQuestion];

  return (
    <section id="quiz" className="section-padding bg-gradient-to-br from-slate-900 via-indigo-950 to-purple-950">
      <div className="max-w-3xl mx-auto">
        <div className="text-center mb-8">
          <span className="inline-block px-4 py-1 bg-indigo-900/50 text-indigo-300 rounded-full text-sm font-medium mb-4">
            Interactive Quiz
          </span>
          <h2 className="section-title">What's Your Feedback Maturity Level?</h2>
          <p className="section-subtitle">
            Answer 5 quick questions to discover how you stack up and get personalized recommendations.
          </p>
        </div>

        <div className="bg-slate-800 rounded-2xl shadow-xl p-8 md:p-12 border border-slate-700">
          {/* Progress bar */}
          <div className="mb-8">
            <div className="flex justify-between text-sm text-slate-400 mb-2">
              <span>Question {currentQuestion + 1} of {quizQuestions.length}</span>
              <span>{Math.round(progress)}% complete</span>
            </div>
            <div className="w-full h-2 bg-slate-600 rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-indigo-500 to-purple-600 rounded-full transition-all duration-500 ease-out"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>

          {/* Question */}
          <div className={`transition-opacity duration-300 ${isAnimating ? 'opacity-0' : 'opacity-100'}`}>
            <h3 className="text-xl md:text-2xl font-semibold text-white mb-6">
              {question.question}
            </h3>

            <div className="space-y-3">
              {question.options.map((option, index) => (
                <button
                  key={index}
                  onClick={() => handleAnswer(option.value, option.text)}
                  className="w-full p-4 text-left border-2 border-slate-600 rounded-xl hover:border-indigo-500 hover:bg-indigo-900/30 transition-all duration-200 group"
                >
                  <div className="flex items-center gap-4">
                    <span className="w-8 h-8 rounded-full bg-slate-700 group-hover:bg-indigo-900/50 flex items-center justify-center text-sm font-medium text-slate-300 group-hover:text-indigo-300 transition-colors">
                      {String.fromCharCode(65 + index)}
                    </span>
                    <span className="text-slate-300 group-hover:text-white">
                      {option.text}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
