# Phase 1: Landing Page - Context

**Gathered:** 2026-01-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Public-facing marketing site where visitors understand Illoominate's value proposition and can join the waitlist. Visitors can enter email to sign up, receive confirmation email, and optionally complete a survey. Admin can view and export waitlist entries.

</domain>

<decisions>
## Implementation Decisions

### Visual Identity
- Dark mode default — dark backgrounds, light text
- Cool color palette (blue/purple) for primary accent colors
- Geometric/modern typography — clean sans-serif like Inter or Geist
- Visual elements: combination of minimal/abstract backgrounds, product screenshots, icons and diagrams
- Design inspiration: Linear and Vercel aesthetics

### Page Structure
- Sections: Hero → Features → Waitlist (minimal, focused)
- Hero: Headline + tagline + CTA — straight to the point, no product visuals in hero
- Features: Icon cards grid — 3-4 feature cards with icons and brief descriptions
- No social proof section for now — skip until testimonials exist

### Waitlist Experience
- Form fields: Email (required), name (optional), "how did you hear about us" (optional)
- After submission: Show success confirmation, then present optional quiz
- Post-signup quiz:
  - Platform focus (iOS, Android, Web)
  - Team size/stage
  - Current feedback collection pain points
  - Goal: identify differentiators and high-priority features
  - Quiz is entirely skippable — lowest friction
- Confirmation email includes referral CTA
- Referral system: shareable referral code (not link)
- Referral reward: higher position on beta access list
- Show exact waitlist position to users ("You are #47")

### Messaging & Copy
- Tone: Friendly professional — warm but competent, approachable startup vibe
- Key features to highlight:
  - Native SDKs (iOS, Web) — the native experience differentiation
  - Dashboard + organization (boards, statuses, voting)
  - Public roadmap and transparency features
- No specific pricing mentioned — just "startup-friendly pricing" without details
- Main competitor differentiation: Canny
  - Canny is web-embedded; Illoominate offers native mobile experience
  - Canny pricing isn't indie-friendly; Illoominate is more affordable

### Claude's Discretion
- Primary value proposition angle — research what pain points resonate most
- Technical depth of feature descriptions — research what's most effective
- Specific copy and headlines
- Exact quiz questions
- Loading states and micro-interactions
- Specific icon choices

</decisions>

<specifics>
## Specific Ideas

- Design should feel like Linear meets Vercel — clean, dark, developer-focused but approachable
- Position as the native-first, indie-friendly alternative to Canny
- Referral system to encourage viral growth from waitlist

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-landing-page*
*Context gathered: 2026-01-21*
