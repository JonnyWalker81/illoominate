import { useState, useEffect } from 'react';
import Quiz from './Quiz';
import { type QuizResult } from '../lib/quiz-logic';

interface WaitlistProps {
  referralCode?: string;
}

interface SuccessState {
  inviteCode: string;
  message: string;
}

export default function Waitlist({ referralCode: initialReferralCode }: WaitlistProps) {
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [referralSource, setReferralSource] = useState('');
  const [referralCode, setReferralCode] = useState(initialReferralCode || '');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<SuccessState | null>(null);
  const [copied, setCopied] = useState(false);
  const [quizCompleted, setQuizCompleted] = useState(false);
  const [quizSkipped, setQuizSkipped] = useState(false);
  const [quizResult, setQuizResult] = useState<QuizResult | null>(null);

  // Extract referral code from URL on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const ref = params.get('ref');
    if (ref && !referralCode) {
      setReferralCode(ref);
    }
  }, [referralCode]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const response = await fetch('/api/waitlist/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email,
          name: name || undefined,
          referral_source: referralSource || undefined,
          referral_code: referralCode || undefined,
        }),
      });

      const data = await response.json() as {
        error?: string;
        invite_code?: string;
        message?: string;
      };

      if (!response.ok) {
        throw new Error(data.error || 'Something went wrong');
      }

      setSuccess({
        inviteCode: data.invite_code || '',
        message: data.message || 'Success',
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setIsSubmitting(false);
    }
  };

  const copyReferralLink = () => {
    if (success?.inviteCode) {
      const link = `${window.location.origin}?ref=${success.inviteCode}`;
      navigator.clipboard.writeText(link);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleQuizComplete = (result: QuizResult) => {
    setQuizResult(result);
    setQuizCompleted(true);
  };

  if (success) {
    const referralLink = `${window.location.origin}?ref=${success.inviteCode}`;

    return (
      <section id="waitlist" className="section-padding bg-slate-900">
        <div className="max-w-2xl mx-auto text-center">
          <div className="bg-slate-800 rounded-2xl p-8 md:p-12 animate-fade-in">
            <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-emerald-500/20 flex items-center justify-center">
              <svg className="w-8 h-8 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>

            <h2 className="text-2xl md:text-3xl font-bold text-white mb-4">
              You're on the list!
            </h2>
            <p className="text-slate-300 mb-8">
              Check your email to verify and lock in your spot. Share your referral link to move up the waitlist!
            </p>

            <div className="bg-slate-700/50 rounded-xl p-4 mb-6">
              <p className="text-sm text-slate-400 mb-2">Your referral link:</p>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  readOnly
                  value={referralLink}
                  className="flex-1 px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white text-sm truncate"
                />
                <button
                  onClick={copyReferralLink}
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition-colors text-sm font-medium whitespace-nowrap"
                >
                  {copied ? 'Copied!' : 'Copy'}
                </button>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center mb-8">
              <a
                href={`https://twitter.com/intent/tweet?text=${encodeURIComponent(`I just joined the @illoominate waitlist! Join me and help shape the future of feedback management: ${referralLink}`)}`}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 px-6 py-3 bg-[#1DA1F2] hover:bg-[#1a8cd8] text-white rounded-lg transition-colors font-medium"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M23.953 4.57a10 10 0 01-2.825.775 4.958 4.958 0 002.163-2.723c-.951.555-2.005.959-3.127 1.184a4.92 4.92 0 00-8.384 4.482C7.69 8.095 4.067 6.13 1.64 3.162a4.822 4.822 0 00-.666 2.475c0 1.71.87 3.213 2.188 4.096a4.904 4.904 0 01-2.228-.616v.06a4.923 4.923 0 003.946 4.827 4.996 4.996 0 01-2.212.085 4.936 4.936 0 004.604 3.417 9.867 9.867 0 01-6.102 2.105c-.39 0-.779-.023-1.17-.067a13.995 13.995 0 007.557 2.209c9.053 0 13.998-7.496 13.998-13.985 0-.21 0-.42-.015-.63A9.935 9.935 0 0024 4.59z"/>
                </svg>
                Share on Twitter
              </a>
            </div>

            {/* Optional Quiz */}
            {!quizCompleted && !quizSkipped && (
              <div className="border-t border-slate-700 pt-8">
                <h3 className="text-lg font-semibold text-white mb-2">
                  While you wait...
                </h3>
                <p className="text-slate-400 mb-6">
                  Discover your feedback maturity level with our quick quiz!
                </p>
                <div className="bg-slate-700/30 rounded-xl p-6">
                  <Quiz onComplete={handleQuizComplete} />
                </div>
                <button
                  onClick={() => setQuizSkipped(true)}
                  className="text-slate-400 hover:text-slate-300 text-sm mt-4 transition-colors"
                >
                  Skip quiz →
                </button>
              </div>
            )}

            {/* Quiz Result */}
            {quizCompleted && quizResult && (
              <div className="border-t border-slate-700 pt-8">
                <div className="text-4xl mb-4">{quizResult.emoji}</div>
                <h3 className="text-xl font-bold text-white mb-2">
                  You're a {quizResult.title}!
                </h3>
                <p className="text-slate-300 text-sm">
                  {quizResult.description}
                </p>
              </div>
            )}
          </div>
        </div>
      </section>
    );
  }

  return (
    <section id="waitlist" className="section-padding bg-slate-900">
      <div className="max-w-2xl mx-auto">
        <div className="text-center mb-8">
          <span className="inline-block px-4 py-1 bg-indigo-500/20 text-indigo-300 rounded-full text-sm font-medium mb-4">
            Join the Waitlist
          </span>
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            Be First in Line
          </h2>
          <p className="text-lg text-slate-300">
            Get early access, exclusive features, and help shape the product. Limited spots available.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="bg-slate-800 rounded-2xl p-8 md:p-10">
          {error && (
            <div className="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-lg text-red-300 text-sm">
              {error}
            </div>
          )}

          <div className="space-y-5">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-slate-300 mb-2">
                Email address <span className="text-red-400">*</span>
              </label>
              <input
                type="email"
                id="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@company.com"
                className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              />
            </div>

            <div>
              <label htmlFor="name" className="block text-sm font-medium text-slate-300 mb-2">
                Name <span className="text-slate-500">(optional)</span>
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Jane Doe"
                className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              />
            </div>

            <div>
              <label htmlFor="referralSource" className="block text-sm font-medium text-slate-300 mb-2">
                How did you hear about us? <span className="text-slate-500">(optional)</span>
              </label>
              <select
                id="referralSource"
                value={referralSource}
                onChange={(e) => setReferralSource(e.target.value)}
                className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              >
                <option value="">Select...</option>
                <option value="reddit">Reddit</option>
                <option value="search">Search Engine (Google, etc.)</option>
                <option value="social">Social Media</option>
                <option value="friend">Friend or Colleague</option>
                <option value="blog">Blog or Article</option>
                <option value="other">Other</option>
              </select>
            </div>

            {referralCode && (
              <div>
                <label htmlFor="referralCode" className="block text-sm font-medium text-slate-300 mb-2">
                  Referral code
                </label>
                <input
                  type="text"
                  id="referralCode"
                  value={referralCode}
                  onChange={(e) => setReferralCode(e.target.value)}
                  className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600 rounded-lg text-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                  readOnly={!!initialReferralCode}
                />
              </div>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full mt-6 px-6 py-4 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-600/50 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-colors text-lg"
          >
            {isSubmitting ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Joining...
              </span>
            ) : (
              'Join the Waitlist'
            )}
          </button>

          <p className="mt-4 text-center text-sm text-slate-400">
            We'll never spam you. Unsubscribe anytime.
          </p>
        </form>
      </div>
    </section>
  );
}
