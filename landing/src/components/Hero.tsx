export default function Hero() {
  return (
    <section className="pt-24 pb-16 md:pt-32 md:pb-24 bg-gradient-hero">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left Column - Text */}
          <div className="text-center lg:text-left animate-fade-in">
            <div className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-900/50 text-indigo-300 rounded-full text-sm font-medium mb-6">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
              </span>
              Coming Soon - Join the Waitlist
            </div>

            <h1 className="text-4xl md:text-5xl lg:text-6xl font-extrabold text-white leading-tight">
              Turn User Feedback Into Your{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-purple-400">
                Product Roadmap
              </span>
            </h1>

            <p className="mt-6 text-lg md:text-xl text-slate-300 max-w-xl mx-auto lg:mx-0">
              Stop losing valuable feedback across scattered channels. Collect, organize, and prioritize feature requests in one place - with AI that actually helps.
            </p>

            <div className="mt-8 flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
              <a href="#waitlist" className="btn-primary text-lg px-8 py-4">
                Get Early Access
              </a>
            </div>

            <div className="mt-8 flex items-center gap-6 justify-center lg:justify-start text-sm text-slate-400">
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-emerald-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                Free tier available
              </div>
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-emerald-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                No credit card
              </div>
            </div>
          </div>

          {/* Right Column - Visual */}
          <div className="relative animate-slide-in-right">
            <div className="relative bg-slate-800 rounded-2xl shadow-2xl border border-slate-700 overflow-hidden">
              {/* Mockup Header */}
              <div className="px-4 py-3 bg-slate-700 border-b border-slate-600 flex items-center gap-2">
                <div className="flex gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-red-400"></div>
                  <div className="w-3 h-3 rounded-full bg-amber-400"></div>
                  <div className="w-3 h-3 rounded-full bg-emerald-400"></div>
                </div>
                <div className="ml-4 flex-1 h-6 bg-slate-600 rounded-md"></div>
              </div>

              {/* Mockup Content */}
              <div className="p-6">
                {/* Feature Request Card */}
                <div className="bg-slate-700/50 rounded-lg p-4 mb-4">
                  <div className="flex items-start gap-4">
                    <div className="flex flex-col items-center">
                      <button className="p-1 text-indigo-400 hover:bg-indigo-900/50 rounded">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
                        </svg>
                      </button>
                      <span className="text-lg font-bold text-indigo-400">47</span>
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-white">Dark mode support</h3>
                      <p className="text-sm text-slate-400 mt-1">Add dark mode theme option for better...</p>
                      <div className="flex gap-2 mt-3">
                        <span className="px-2 py-1 bg-purple-900/50 text-purple-300 text-xs rounded-full">UI/UX</span>
                        <span className="px-2 py-1 bg-amber-900/50 text-amber-300 text-xs rounded-full">In Progress</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Another Feature Request */}
                <div className="bg-slate-700/50 rounded-lg p-4 mb-4 opacity-70">
                  <div className="flex items-start gap-4">
                    <div className="flex flex-col items-center">
                      <button className="p-1 text-slate-500 hover:bg-slate-600 rounded">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
                        </svg>
                      </button>
                      <span className="text-lg font-bold text-slate-400">32</span>
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-white">API webhooks</h3>
                      <p className="text-sm text-slate-400 mt-1">Integrate with external services via...</p>
                      <div className="flex gap-2 mt-3">
                        <span className="px-2 py-1 bg-blue-900/50 text-blue-300 text-xs rounded-full">Integration</span>
                        <span className="px-2 py-1 bg-emerald-900/50 text-emerald-300 text-xs rounded-full">Planned</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Stats Bar */}
                <div className="flex items-center justify-between pt-4 border-t border-slate-700">
                  <div className="text-sm text-slate-400">23 feature requests</div>
                  <div className="flex -space-x-2">
                    <div className="w-8 h-8 rounded-full bg-indigo-500 border-2 border-slate-800 flex items-center justify-center text-xs text-white font-medium">JD</div>
                    <div className="w-8 h-8 rounded-full bg-emerald-500 border-2 border-slate-800 flex items-center justify-center text-xs text-white font-medium">SK</div>
                    <div className="w-8 h-8 rounded-full bg-amber-500 border-2 border-slate-800 flex items-center justify-center text-xs text-white font-medium">+5</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Floating elements */}
            <div className="absolute -top-4 -right-4 bg-slate-800 rounded-lg shadow-lg border border-slate-700 p-3 animate-bounce-subtle">
              <div className="flex items-center gap-2">
                <div className="w-10 h-10 rounded-full bg-emerald-900/50 flex items-center justify-center">
                  <svg className="w-5 h-5 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div>
                  <p className="text-xs font-medium text-white">Feature shipped!</p>
                  <p className="text-xs text-slate-400">Dark mode is live</p>
                </div>
              </div>
            </div>

            <div className="absolute -bottom-4 -left-4 bg-slate-800 rounded-lg shadow-lg border border-slate-700 p-3 animate-bounce-subtle animation-delay-300">
              <div className="flex items-center gap-2">
                <div className="w-10 h-10 rounded-full bg-indigo-900/50 flex items-center justify-center">
                  <svg className="w-5 h-5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                </div>
                <div>
                  <p className="text-xs font-medium text-white">+12 votes today</p>
                  <p className="text-xs text-slate-400">API webhooks</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
