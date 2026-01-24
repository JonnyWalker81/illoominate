# Roadmap: Illoominate

## Overview

Illoominate delivers a native-first user feedback platform for indie developers and startups. The roadmap begins with a landing page to validate market interest, builds a solid multi-tenant foundation with authentication and workspace management, then progressively layers the core feedback loop (submission, organization, voting), delivers iOS SDK as the native differentiation before web, adds public-facing features (roadmap, notifications), and concludes with the Web SDK for broader reach. Each phase delivers a coherent, verifiable capability.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Landing Page** - Public-facing marketing site with waitlist collection
- [ ] **Phase 1.1: Landing Page CRO** - Conversion optimization for waitlist signups (INSERTED)
- [ ] **Phase 2: Foundation** - Database schema, Supabase setup, multi-tenant RLS
- [ ] **Phase 3: Authentication** - User accounts with email/password and OAuth
- [ ] **Phase 4: Workspace & Apps** - Organization and app management
- [ ] **Phase 5: Core API** - Go backend with feedback endpoints
- [ ] **Phase 6: Dashboard MVP** - Astro frontend for viewing and managing feedback
- [ ] **Phase 7: Feedback Submission** - Complete submission flow with attachments
- [ ] **Phase 8: Feedback Organization** - Boards, statuses, filtering, search
- [ ] **Phase 9: Voting & Engagement** - Upvotes, comments, subscriptions
- [ ] **Phase 10: iOS SDK** - Native SwiftUI SDK for iOS apps
- [ ] **Phase 11: Public Roadmap** - Public-facing roadmap and changelog
- [ ] **Phase 12: Notifications** - Email and in-app notification system
- [ ] **Phase 13: Web SDK** - JavaScript SDK for web applications

## Phase Details

### Phase 1: Landing Page
**Goal**: Visitors understand Illoominate's value proposition and can join the waitlist
**Depends on**: Nothing (first phase)
**Requirements**: LAND-01, LAND-02, LAND-03, LAND-04
**Success Criteria** (what must be TRUE):
  1. Visitor landing on illoominate.com sees clear value proposition explaining native SDKs and startup-friendly pricing
  2. Visitor can enter email and submit to join waitlist
  3. Visitor receives confirmation email after signup
  4. Admin can access dashboard to view and export waitlist entries
**Plans**: 5 plans (includes 1 gap closure)

Plans:
- [x] 01-01-PLAN.md — Infrastructure setup (Astro + Workers + D1 + Drizzle)
- [x] 01-02-PLAN.md — Landing page frontend (Hero, Features, Waitlist form)
- [x] 01-03-PLAN.md — Waitlist API and email confirmation
- [x] 01-04-PLAN.md — Quiz modal and admin dashboard
- [x] 01-05-PLAN.md — Quiz flow redesign (gap closure)

### Phase 1.1: Landing Page CRO (INSERTED)
**Goal**: Maximize waitlist signups and quiz completion through conversion optimization
**Depends on**: Phase 1 (enhances completed landing page)
**Requirements**: None (optimization work)
**Success Criteria** (what must be TRUE):
  1. Hero section uses outcome-oriented messaging with product teaser
  2. Trust block displays framework logos and momentum marker
  3. CTA copy optimized ("Claim Early Access" or similar)
  4. Quiz framed as incentive for position boost
  5. Mobile experience includes sticky CTA
**Plans**: 2 plans

Plans:
- [ ] 01.1-01-PLAN.md — Hero messaging + Trust block (above-the-fold optimization)
- [ ] 01.1-02-PLAN.md — Quiz incentive + Sticky CTA (engagement optimization)

### Phase 2: Foundation
**Goal**: Multi-tenant database schema with proper isolation exists
**Depends on**: Phase 1 (conceptually independent, but sequenced for focus)
**Requirements**: None directly (infrastructure enabling WORK-*, AUTH-*)
**Success Criteria** (what must be TRUE):
  1. Supabase project exists with PostgreSQL database
  2. Database schema includes workspaces, apps, users, feedback, and all supporting tables
  3. Row Level Security policies enforce workspace isolation (cross-tenant queries return empty)
  4. Database migrations are version-controlled and reproducible
**Plans**: 2 plans

Plans:
- [ ] 02-01-PLAN.md — Supabase CLI initialization and project structure
- [ ] 02-02-PLAN.md — Core schema with 10 tables and RLS policies

### Phase 3: Authentication
**Goal**: Users can securely create accounts and sign in
**Depends on**: Phase 2
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05
**Success Criteria** (what must be TRUE):
  1. User can create account with email and password
  2. User can sign in with Google OAuth
  3. User can sign in with GitHub OAuth
  4. User can reset forgotten password via email link
  5. User session persists across browser refresh and new tabs
**Plans**: TBD

Plans:
- [ ] 03-01: Supabase Auth configuration
- [ ] 03-02: OAuth provider setup (Google, GitHub)
- [ ] 03-03: Password reset and session handling

### Phase 4: Workspace & Apps
**Goal**: Developers can create workspaces and register apps for SDK integration
**Depends on**: Phase 3
**Requirements**: WORK-01, WORK-02, WORK-03, WORK-04
**Success Criteria** (what must be TRUE):
  1. Authenticated user can create a workspace (organization)
  2. Workspace owner can create multiple apps within the workspace
  3. Each app displays unique API keys for SDK authentication
  4. Workspace owner can invite team members and remove them
**Plans**: TBD

Plans:
- [ ] 04-01: Workspace CRUD and management
- [ ] 04-02: App registration and API key generation
- [ ] 04-03: Team member invitation and management

### Phase 5: Core API
**Goal**: Go backend serves authenticated API requests for all platform operations
**Depends on**: Phase 4
**Requirements**: None directly (infrastructure enabling SUBM-*, ORG-*, VOTE-*)
**Success Criteria** (what must be TRUE):
  1. Go service runs on Cloud Run with health checks passing
  2. API authenticates SDK requests via API key header
  3. API authenticates dashboard requests via Supabase JWT
  4. Feedback CRUD endpoints return appropriate responses
  5. API versioning (/v1/) is in place for SDK compatibility
**Plans**: TBD

Plans:
- [ ] 05-01: Go project setup with Chi router and sqlc
- [ ] 05-02: Authentication middleware (API key + JWT)
- [ ] 05-03: Core feedback endpoints
- [ ] 05-04: Cloud Run deployment configuration

### Phase 6: Dashboard MVP
**Goal**: Admins can view feedback submitted to their apps
**Depends on**: Phase 5
**Requirements**: None directly (UI for SUBM-*, ORG-*, VOTE-*)
**Success Criteria** (what must be TRUE):
  1. Authenticated user sees dashboard with workspace/app selector
  2. Dashboard displays list of feedback items for selected app
  3. User can navigate between workspace settings, apps, and feedback
  4. Dashboard uses Astro with React islands for interactivity
**Plans**: TBD

Plans:
- [ ] 06-01: Astro project setup with CloudFlare Pages
- [ ] 06-02: Authentication flow and layout
- [ ] 06-03: Feedback list view with basic filtering

### Phase 7: Feedback Submission
**Goal**: Users can submit feedback with full metadata through the API
**Depends on**: Phase 5, Phase 6
**Requirements**: SUBM-01, SUBM-02, SUBM-03, SUBM-04
**Success Criteria** (what must be TRUE):
  1. API accepts feedback submission with type (Bug, Feature Request, UI/UX)
  2. Feedback includes title and description fields
  3. User can attach images/screenshots to feedback (stored in Supabase Storage)
  4. Submission automatically captures device and app metadata from request
**Plans**: TBD

Plans:
- [ ] 07-01: Feedback submission API with types
- [ ] 07-02: Image attachment handling with Supabase Storage
- [ ] 07-03: Metadata capture and storage

### Phase 8: Feedback Organization
**Goal**: Admins can organize, filter, and manage feedback effectively
**Depends on**: Phase 7
**Requirements**: ORG-01, ORG-02, ORG-03, ORG-04, ORG-05
**Success Criteria** (what must be TRUE):
  1. Admin can create boards to organize feedback (e.g., "Features", "Bugs")
  2. Admin can change feedback status through workflow (Open, Under Review, Planned, In Progress, Complete)
  3. Admin can filter feedback by status, type, votes, and date
  4. Admin can search feedback by text content
  5. Admin can add tags/labels to feedback items
**Plans**: TBD

Plans:
- [ ] 08-01: Boards CRUD and assignment
- [ ] 08-02: Status workflow implementation
- [ ] 08-03: Filtering, search, and tags

### Phase 9: Voting & Engagement
**Goal**: Users can vote on and engage with feedback items
**Depends on**: Phase 8
**Requirements**: VOTE-01, VOTE-02, VOTE-03, VOTE-04
**Success Criteria** (what must be TRUE):
  1. User can upvote feedback items (one vote per user per item)
  2. Vote counts are visible on feedback items in real-time
  3. User can comment on feedback items
  4. User can subscribe to feedback items for status updates
**Plans**: TBD

Plans:
- [ ] 09-01: Voting API with denormalized counts
- [ ] 09-02: Comments system
- [ ] 09-03: Subscription and real-time updates

### Phase 10: iOS SDK
**Goal**: iOS developers can integrate native feedback submission in their apps
**Depends on**: Phase 7 (submission API must exist)
**Requirements**: IOS-01, IOS-02, IOS-03, IOS-04
**Success Criteria** (what must be TRUE):
  1. iOS SDK provides native SwiftUI interface for submitting feedback
  2. Host app can pass user identity (id, email, name) to SDK
  3. SDK queues feedback when offline and syncs when connection restored
  4. SDK automatically captures device metadata (iOS version, device model, app version)
**Plans**: TBD

Plans:
- [ ] 10-01: Swift Package setup and project structure
- [ ] 10-02: SwiftUI feedback submission UI
- [ ] 10-03: User identity and metadata capture
- [ ] 10-04: Offline queueing and sync

### Phase 11: Public Roadmap
**Goal**: Users can view public roadmap and admins can publish updates
**Depends on**: Phase 8, Phase 9
**Requirements**: ROAD-01, ROAD-02, ROAD-03, ROAD-04
**Success Criteria** (what must be TRUE):
  1. Public URL displays roadmap with planned/in-progress/shipped items
  2. Admin has private roadmap layer for internal planning (not visible to public)
  3. Admin can publish changelog entries when features ship
  4. Roadmap widget can be embedded on external sites via iframe/script
**Plans**: TBD

Plans:
- [ ] 11-01: Public roadmap view
- [ ] 11-02: Private roadmap layer for admins
- [ ] 11-03: Changelog and embeddable widget

### Phase 12: Notifications
**Goal**: Users receive timely updates about feedback they care about
**Depends on**: Phase 9, Phase 11
**Requirements**: NOTF-01, NOTF-02, NOTF-03, NOTF-04
**Success Criteria** (what must be TRUE):
  1. User receives email when features they voted for ship
  2. User has in-app notification center showing recent activity
  3. User can configure notification preferences (email on/off, types)
  4. User receives email when someone replies to their comment
**Plans**: TBD

Plans:
- [ ] 12-01: Email notification system with Resend
- [ ] 12-02: In-app notification center
- [ ] 12-03: Notification preferences

### Phase 13: Web SDK
**Goal**: Web developers can integrate feedback submission in their websites
**Depends on**: Phase 7 (submission API), Phase 10 (SDK patterns proven)
**Requirements**: WEB-01, WEB-02, WEB-03, WEB-04
**Success Criteria** (what must be TRUE):
  1. Web SDK provides JavaScript API for submitting feedback programmatically
  2. Host app can pass user identity to SDK
  3. Optional embeddable widget provides UI without custom code
  4. SDK automatically captures browser metadata (browser, OS, viewport)
**Plans**: TBD

Plans:
- [ ] 13-01: TypeScript SDK setup with Rollup bundling
- [ ] 13-02: Core API and user identity
- [ ] 13-03: Embeddable widget component
- [ ] 13-04: Browser metadata capture

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 1.1 -> 2 -> 3 -> ... -> 13

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Landing Page | 5/5 | Complete | 2026-01-22 |
| 1.1 Landing Page CRO | 0/2 | Not started | - |
| 2. Foundation | 0/2 | Not started | - |
| 3. Authentication | 0/3 | Not started | - |
| 4. Workspace & Apps | 0/3 | Not started | - |
| 5. Core API | 0/4 | Not started | - |
| 6. Dashboard MVP | 0/3 | Not started | - |
| 7. Feedback Submission | 0/3 | Not started | - |
| 8. Feedback Organization | 0/3 | Not started | - |
| 9. Voting & Engagement | 0/3 | Not started | - |
| 10. iOS SDK | 0/4 | Not started | - |
| 11. Public Roadmap | 0/3 | Not started | - |
| 12. Notifications | 0/3 | Not started | - |
| 13. Web SDK | 0/4 | Not started | - |

---
*Roadmap created: 2026-01-21*
*Phase 1 planned: 2026-01-21*
*Phase 1 completed: 2026-01-22*
*Phase 1.1 inserted: 2026-01-23*
*Phase 1.1 planned: 2026-01-23*
*Total phases: 14 | Total plans: 45+ (estimated)*
