# Illoominate

## What This Is

Illoominate is a user feedback collection platform for developers. It lets app creators collect bug reports, feature requests, and UX feedback through native SDKs, while giving their users a voice through voting and public roadmaps. Think Canny, but with truly native mobile experiences and pricing that doesn't punish success.

## Core Value

Native-feeling feedback submission with transparent, startup-friendly pricing. The SDK experience must feel like a natural part of the host app, not an embedded webview.

## Requirements

### Validated

(None yet — ship to validate)

### Active

**Landing Page & Waitlist**
- [ ] Landing page explaining Illoominate value proposition
- [ ] Email waitlist signup with confirmation
- [ ] Waitlist data stored and manageable

**Workspace & Apps**
- [ ] Developers can create workspaces (organizations)
- [ ] Workspaces can contain multiple apps/products
- [ ] Apps have unique identifiers for SDK integration

**User Identity**
- [ ] SDK can pass user identity (id, email, name, avatar) from host app
- [ ] Users can later log in via email and link to existing feedback
- [ ] Anonymous feedback not supported (identity required)

**Feedback Submission**
- [ ] Users can submit feedback with type (Bug, Feature Request, UI/UX)
- [ ] Feedback includes title, description, and optional attachments
- [ ] Native SDK UI for submission (iOS Swift, Web JS initially)

**Feedback Organization**
- [ ] Admins can create boards to organize feedback (e.g., "Features", "Bugs")
- [ ] Feedback has status workflow (Open, Under Review, Planned, In Progress, Complete, Closed)
- [ ] Admins can view all feedback across all users

**Voting & Engagement**
- [ ] Users can vote on feature requests
- [ ] Users can see all feedback for apps they've submitted to
- [ ] Vote counts visible to users and admins

**Roadmap**
- [ ] Public roadmap showing planned/in-progress/shipped features
- [ ] Private roadmap layer for internal organization (admin only)
- [ ] Users can see public roadmap and vote on items

**Notifications**
- [ ] Users notified (email + in-app) when features they voted for ship
- [ ] Admin notifications for new feedback

**SDKs (v1)**
- [ ] iOS SDK (Swift) — native UI for submitting feedback
- [ ] Web SDK (JavaScript) — for websites

### Out of Scope

- **Android SDK** — v2, after iOS/Web proven
- **React Native SDK** — v2, cross-platform later
- **Browse/vote in native SDK** — v2, users go to web for this initially
- **Real-time chat/support** — Not a support tool, feedback only
- **Video attachments** — Storage/bandwidth costs, maybe v2
- **OAuth login (Google, GitHub)** — Email/password sufficient for v1
- **Mobile app for admins** — Web dashboard only for v1
- **AI features** — No auto-categorization or smart replies for v1
- **Integrations (Jira, Slack, etc.)** — v2, focus on core first
- **Self-hosted option** — SaaS only for v1

## Context

**Market Position**: Competing with Canny ($79/mo for Pro, per-user pricing) and UserVoice ($999+/mo enterprise). Targeting indie developers, solo developers, and startups who find existing tools too expensive or punishing as they grow.

**Pricing Strategy**:
| Tier | Users | Apps | Price |
|------|-------|------|-------|
| Free | 100 | 1 | $0 |
| Starter | 1,000 | 3 | $29/mo |
| Pro | Unlimited | 10 | $79/mo |
| Business | Unlimited | Unlimited | Custom |

**Key Competitor Pain Points**:
- Canny's "tracked user" model counts anyone who votes or comments — costs spike with success
- UserVoice is enterprise-only, no self-serve, requires sales calls
- Most tools use webviews for mobile, not native experiences

**Native SDK Rationale**: The submission experience is the first touchpoint. A native UI that matches the host app's design language builds trust and feels professional. Webviews feel like afterthoughts.

## Constraints

- **Tech Stack - Landing**: CloudFlare Pages (Astro) + Workers + D1 + Resend API — specified for waitlist handling
- **Tech Stack - Backend**: Go on Google Cloud Run — performance and simplicity
- **Tech Stack - Database**: Supabase PostgreSQL — auth and data
- **Tech Stack - Frontend**: Astro (exploring fit for both landing and web app)
- **Build Order**: Landing page with waitlist must be built first
- **SDK Platforms**: iOS (Swift) and Web (JS) are priority; Android/React Native deferred
- **No Enterprise Features**: SSO, advanced compliance, etc. are v2+

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Native SDKs over webviews | Differentiation, better UX, feels professional | — Pending |
| Supabase for auth + data | Managed PostgreSQL, built-in auth, good DX | — Pending |
| Go backend on Cloud Run | Performance, simplicity, scales well | — Pending |
| Landing page first | Build waitlist before product, validate interest | — Pending |
| Boards + statuses for org | Matches Canny mental model users expect | — Pending |
| Public + private roadmap | Transparency for users, control for teams | — Pending |
| Per-tier pricing (not per-user) | Differentiation from Canny's punishing model | — Pending |

---
*Last updated: 2026-01-21 after initialization*
