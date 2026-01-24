---
created: 2026-01-23T20:33
title: Landing page CRO overhaul for waitlist conversion
area: ui
files:
  - src/pages/index.astro
  - src/components/Hero.astro
  - src/components/Features.astro
  - src/components/Waitlist.astro
---

## Problem

The current landing page needs conversion rate optimization to maximize waitlist signups and post-signup quiz completion. Key gaps identified:

1. **Hero lacks outcome-oriented messaging** - Current "Native feedback, not webviews" is feature-focused, not benefit-focused
2. **No visual proof** - Abstract glow instead of product teaser showing the 60 FPS native experience
3. **Credibility gap** - Zero testimonials or social proof for a new product
4. **CTA could be stronger** - "Join the Waitlist" is generic vs. "Claim Early Access"
5. **Quiz positioning** - Not framed as an incentive for position boost

Target audience: Technical indie developers who value performance, transparency, and native quality.

## Solution

**Hero Section:**
- Headline variants: "Beautiful native feedback for apps that care about performance" or "Fix bugs 50x faster with native feedback"
- Sub-headline emphasizing economics: "Built for indie developers with pricing that doesn't punish success"
- Product teaser GIF showing shake-to-report at 60 FPS
- CTA: "Claim Early Access" or "Unlock Founding Pricing"

**Trust Block (borrowed social proof):**
- Framework logos: iOS, Android, React Native, Flutter
- Momentum marker: "Join X developers already on the list"
- Trust signals: "GDPR & CCPA Compliant", "Zero main thread impact"

**Features Block:**
- Chess layout (alternating text/imagery)
- Highlight 71 automatic data points
- Pricing comparison card vs. legacy per-seat competitors

**Waitlist Funnel:**
- Keep email-only entry
- Frame quiz as incentive: "Answer 3 questions to jump 50 spots"
- Multi-step conversational UI
- Referral loop on thank-you page

**Technical:**
- Maintain dark mode purple/blue aesthetic
- WebP assets, <2s load time
- Sticky mobile CTA

**Deliverable:** Updated HTML/Tailwind CSS + post-signup email sequence content map
