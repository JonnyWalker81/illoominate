# Project Research Summary

**Project:** Illoominate
**Domain:** User feedback collection and management platform (Canny/UserVoice competitor)
**Researched:** 2026-01-21
**Confidence:** HIGH

## Executive Summary

Illoominate is a user feedback platform targeting indie developers and startups with native SDKs as its primary differentiator. The research validates the pre-defined stack (Go + Supabase + Astro + CloudFlare) as well-aligned with 2025/2026 best practices for this domain. The table stakes features are well-established across competitors (feedback boards, voting, status workflow, roadmap), meaning differentiation must come from execution quality, developer experience, and native SDK integration rather than feature novelty.

The recommended approach prioritizes multi-tenant data isolation from day one (RLS policies with workspace_id), builds core feedback loop validation before investing in SDKs, and focuses on simplicity over enterprise feature parity. The architecture follows clean separation: collection layer (SDKs) → API layer (Go/Cloud Run) → data layer (Supabase PostgreSQL) → presentation layer (Astro islands). This enables independent scaling and clear boundaries.

Key risks center on multi-tenant isolation failures (critical security issue), voting system bias creating misleading priorities, and underpricing destroying unit economics. Mitigation requires RLS policies from day one with automated cross-tenant tests, voting UX that reduces herd behavior (randomized order, vote limits), and value-based pricing above $50/mo minimum. The biggest strategic risk is building native SDKs before validating the core feedback loop—research strongly recommends web-first validation before iOS investment.

## Key Findings

### Recommended Stack

The pre-selected stack is strongly validated by current best practices. Go with Chi router, sqlc, and slog provides a lightweight, stdlib-aligned backend that avoids ORM complexity. Supabase PostgreSQL accessed directly via pgx/v5 delivers better performance than REST API for the backend, while Supabase Auth and Realtime handle authentication and live updates. Astro 5's Server Islands and ClientRouter enable SPA-like dashboard experiences with partial hydration, and Nanostores solves cross-island state management where React Context breaks.

**Core technologies:**
- **Go 1.24+ with Chi v5**: HTTP router — 100% net/http compatible, lightweight, no framework lock-in
- **sqlc v1.30.0**: Type-safe SQL generation — compiles SQL to Go, avoids ORM overhead and N+1 queries
- **Supabase PostgreSQL**: Multi-tenant database — RLS for workspace isolation, realtime for voting updates
- **Astro 5.16+**: Dashboard meta-framework — partial hydration, View Transitions, React islands for interactivity
- **Nanostores 1.1.0**: Cross-island state — 286 bytes, framework-agnostic, no provider wrappers
- **CloudFlare Workers + D1**: Landing page — edge-first, globally distributed, isolated from product database
- **Swift 6.0+ (SwiftUI)**: iOS SDK — native UI, async/await, SwiftUI-first with UIKit bridge
- **Rollup + esbuild**: Web SDK bundling — optimal library output, tree-shakeable, multiple formats

**Critical version/pattern notes:**
- RLS must be enabled on all Supabase tables (83% of exposed databases involve RLS misconfigurations - CVE-2025-48757)
- Use service key for SDK writes (bypass RLS, enforce in API layer)
- Denormalize vote_count column to avoid expensive COUNT queries at scale
- Set Cloud Run minimum instances > 0 to avoid 300ms-1s cold start latency

### Expected Features

The user feedback platform market has clear feature tiers. Table stakes are established—without these, users won't consider the product. Differentiation opportunities lie in native SDK quality, multi-app workspace simplicity, and transparent pricing, not in adding more features than competitors.

**Must have (table stakes):**
- Feedback submission (text + optional image) — core value proposition
- Voting/upvoting — industry standard since UserVoice, simple one-per-user is sufficient
- Status workflow (Under Review, Planned, In Progress, Shipped, Closed) — 4-6 statuses cover all needs
- Public feedback board — filterable, searchable list view
- Public roadmap — shows commitment, Kanban or timeline view
- Changelog/announcements — close the feedback loop, notify voters when features ship
- Email notifications — status changes, comments, shipped features
- Basic admin dashboard — list, filter, change status, respond
- User authentication — email/password + OAuth (Google/GitHub)
- Search and filter — by status, category, votes, date
- Custom branding — logo, colors, custom domain
- Embeddable widget — web widget at minimum for in-app collection

**Should have (competitive differentiators):**
- Native SDKs (iOS, Web) — Canny uses WebView for mobile; native is a gap
- Multi-app workspaces — indie devs have multiple products, Canny charges per workspace
- AI duplicate detection — becoming table stakes (UserJot, Featurebase have this)
- User segmentation — filter by plan/revenue/cohort to prioritize high-value customers
- Guest posting — users submit without account creation (lower friction)
- Webhooks — enables custom automation without building integrations

**Defer (v2+):**
- Jira/Linear/GitHub sync — HIGH complexity, each integration is ongoing maintenance burden
- Slack/Discord integrations — popular but not critical path
- Analytics dashboard — simple counts sufficient initially, sophisticated analytics after data volume
- Import from Canny/others — valuable for switching but defer until demand exists
- Custom fields — power user feature, enable when requested

**Anti-features (deliberately avoid):**
- Complex voting systems — weighted votes, limits, resets overcomplicate simple signal
- Gamification/points — badges/leaderboards attract wrong behavior, distract from core value
- Forums/discussions — scope creep into community platforms, different product
- Full project management — sprints, assignments, time tracking is Linear/Jira, not feedback tool
- Enterprise SSO/SCIM — months of work for market that doesn't need it (indie/startup target)

### Architecture Approach

Multi-tier architecture with clear separation enables independent scaling and maintenance. Collection layer (native SDKs) captures feedback at source. API layer (Go on Cloud Run) handles authentication, business logic, and orchestration. Data layer (Supabase PostgreSQL with RLS) provides multi-tenant isolation. Presentation layer (Astro islands) delivers interactive dashboard with partial hydration. Marketing layer (CloudFlare Workers + D1) keeps landing page separate from product database.

**Major components:**
1. **Go Backend (Clean Architecture)** — Stateless Cloud Run containers with domain/usecase/repository/infrastructure layers. Direct PostgreSQL via pgx/v5 + sqlc for performance, supabase-go only for Auth/Storage/Realtime. API key authentication for SDKs, JWT validation for dashboard.
2. **Supabase PostgreSQL (Multi-tenant RLS)** — Shared database with workspace_id column pattern. RLS policies enforce workspace membership checks via workspace_members table. Denormalized vote_count with triggers for performance. Service role for SDK writes bypasses RLS (enforce in API).
3. **Astro Dashboard (Islands Architecture)** — Static shell (layout/nav) with React islands for interactive components. Hydration strategy: client:load for critical (FeedbackTable), client:visible for repeated (VoteButton), client:idle for deferred (FilterPanel). Nanostores for cross-island state.
4. **Native SDKs** — iOS: SwiftUI-first with async/await, XCFramework distribution via SPM. Web: TypeScript source, Rollup + esbuild for ESM/CJS/IIFE outputs, tree-shakeable, <5KB core. Both include optional UI widgets and offline queueing.
5. **CloudFlare Workers + D1** — Edge-first landing page with globally distributed D1 for waitlist. Isolated from product database for simplicity, cost, and latency. Resend integration for email notifications.

**Key patterns:**
- Multi-tenant isolation via RLS with workspace_id column (not schema-per-tenant)
- SDK authentication via API keys (X-API-Key header), dashboard via Supabase Auth JWT
- Realtime voting updates via Postgres Changes subscriptions filtered by workspace_id
- Vote count denormalization with triggers to avoid expensive COUNT queries
- API versioning from day one (/v1/feedback) for SDK backwards compatibility

### Critical Pitfalls

Research identified 17 pitfalls across security, product, technical, and business domains. The top 5 could kill the project or require major rewrites:

1. **Multi-tenant data isolation failures** — RLS misconfiguration allows workspace data leaks. Prevention: Design RLS policies with workspace_id checks from day one, store authorization in app_metadata (not user-modifiable user_metadata), implement automated cross-tenant access tests, never use service_role keys in client code. Index workspace_id and user_id for RLS performance.

2. **Voting system bias creating misleading priorities** — Feature voting naturally biases toward vocal minorities, early submissions, and momentum. Top-voted features get built but go unused. Prevention: Randomize feature display order (don't default-sort by votes), hide vote counts to reduce herd behavior, implement vote limits per user to force prioritization, weight votes by customer value (MRR/plan tier), capture context with votes (why they want it).

3. **Building SDK before validating core product** — Native SDKs are significant investment. Building before proving value proposition wastes engineering time. Prevention: Build web dashboard and API first, use web-based submission initially even for mobile users, validate with web-only flow, invest in native SDKs only after proving users want the product. Start with single SDK (Web) before iOS.

4. **RLS performance degradation at scale** — Complex RLS policies with subqueries/joins slow every query. Performance degrades as data grows. Prevention: Index all columns used in RLS policies (workspace_id, user_id), use custom JWT claims for tenant context to avoid subqueries, run EXPLAIN ANALYZE on queries with RLS enabled, load test with realistic data volumes early.

5. **Underpricing destroys unit economics** — Indie developer pricing ($0-79/mo) has thin margins. Underpricing attracts low-value users who churn fast and demand disproportionate support. Prevention: Price on value not cost-to-serve, free tier must have clear limitations, annual prepay option from launch, track support costs per pricing tier, be willing to raise prices if underpriced.

**Phase-specific warnings:**
- Phase 1 (Core Data Model): Multi-tenant isolation (#1), RLS performance (#4) must be correct before any data exists
- Voting implementation: Voting bias (#2) requires UX that reduces herd behavior
- SDK development: Building too early (#3), versioning strategy (#9) from day one
- Pricing launch: Underpricing (#5), clear "user" definitions, upsell path design

## Implications for Roadmap

Based on research, recommended phase structure follows dependency order and risk mitigation patterns:

### Phase 1: Foundation (Database + Auth)
**Rationale:** Everything depends on multi-tenant data structure being correct. RLS misconfiguration is the #1 critical pitfall—fixing after data exists is nearly impossible. Authentication establishes workspace membership which all RLS policies reference.

**Delivers:** Supabase project configured, database schema with RLS policies, workspace/app/user tables, workspace_members for authorization, Supabase Auth integrated with Google/GitHub OAuth.

**Addresses:** Multi-tenant isolation (PITFALLS.md #1), flexible data model (PITFALLS.md #10), auth vs custom identity separation (PITFALLS.md #11)

**Avoids:** Data isolation failures, migration-heavy schema changes later

**Research flag:** SKIP PHASE RESEARCH — RLS patterns are well-documented in Supabase official docs

### Phase 2: Core API (Go Backend)
**Rationale:** SDKs and dashboard both need API to function. Building this second enables testing multi-tenancy via direct API calls before UI complexity. Clean Architecture pattern from research provides clear layer separation for iterating.

**Delivers:** Go service on Cloud Run with Chi router, feedback CRUD endpoints, API key authentication for SDKs, JWT validation for dashboard, sqlc-generated type-safe queries, structured logging with slog.

**Uses:** Go 1.24+ (STACK.md), Chi v5, sqlc v1.30.0, pgx/v5 for direct PostgreSQL access

**Implements:** Backend API component (ARCHITECTURE.md), stateless request-scoped connections

**Avoids:** Cold start latency (PITFALLS.md #7) by configuring minimum instances from start

**Research flag:** SKIP PHASE RESEARCH — Go/Chi/sqlc patterns are standard, well-documented

### Phase 3: Dashboard MVP (Astro + React Islands)
**Rationale:** Need to visualize data and test end-to-end flow before investing in native SDKs. Validates core feedback loop works (submit → view → manage → update status). Enables dogfooding before external users.

**Delivers:** Astro dashboard on CloudFlare Pages, authentication flow, feedback list view, filtering by status/category, basic admin actions (change status, respond), workspace/app management UI.

**Uses:** Astro 5.16+ (STACK.md), React 18.3+ islands, Nanostores for cross-island state, Tailwind + shadcn/ui for components

**Implements:** Astro islands architecture (ARCHITECTURE.md), hydration strategy (client:load/visible/idle)

**Addresses:** UX simplicity (PITFALLS.md #6) — focus on indie dev needs, not enterprise feature parity

**Research flag:** SKIP PHASE RESEARCH — Astro 5 islands architecture is well-documented

### Phase 4: Voting + Realtime
**Rationale:** Voting is table stakes but requires foundation in place. Realtime updates enhance UX but are non-critical path. Grouping together since both touch same data (feedback.vote_count).

**Delivers:** Vote submission API, votes table with uniqueness constraint, denormalized vote_count with trigger, Supabase Realtime subscriptions for live updates, vote button in dashboard.

**Implements:** Voting flow (ARCHITECTURE.md), denormalized count pattern, Realtime Postgres Changes

**Addresses:** Voting bias (PITFALLS.md #2) — randomize display order, consider hiding vote counts initially

**Avoids:** Global COUNT queries (PITFALLS.md #5 anti-pattern) via denormalization

**Research flag:** SKIP PHASE RESEARCH — Voting patterns well-established, Realtime officially documented

### Phase 5: Web SDK
**Rationale:** Web SDK validates SDK design patterns before iOS investment (lower cost to iterate). Enables dogfooding with web apps. Proves SDK → API → Dashboard flow works end-to-end. Native iOS only after this validates.

**Delivers:** TypeScript SDK source, Rollup + esbuild build config, ESM/CJS/IIFE outputs, type declarations, basic feedback submission API, optional embeddable widget.

**Uses:** TypeScript 5.7+ (STACK.md), Rollup 4.x, esbuild 0.24+ for optimal library bundling

**Implements:** Web SDK architecture (ARCHITECTURE.md), tree-shakeable <5KB core

**Addresses:** SDK versioning (PITFALLS.md #9) — API versioning from day one, semantic versioning, deprecation policy

**Avoids:** Building SDK before validation (PITFALLS.md #3) — web dashboard already proves value

**Research flag:** SKIP PHASE RESEARCH — TypeScript library bundling patterns are standard

### Phase 6: Email Notifications
**Rationale:** Close the feedback loop (table stakes) but non-blocking. Users can manually check status. Async processing pattern established before scaling concerns.

**Delivers:** Resend integration, email templates for status changes/comments/shipped features, notification preferences per user, async processing (queue or Cloud Tasks).

**Uses:** Resend API (STACK.md), Go resend-go SDK

**Addresses:** Notification fatigue (PITFALLS.md #5) — conservative defaults, digest options, granular controls

**Avoids:** Synchronous email sending (ARCHITECTURE.md anti-pattern #3)

**Research flag:** SKIP PHASE RESEARCH — Resend integration is straightforward, well-documented

### Phase 7: Public Board + Roadmap
**Rationale:** User-facing transparency features. Depends on dashboard/voting existing. Enables external users to submit and vote without full dashboard access.

**Delivers:** Public-facing feedback board (read-only for non-members), public roadmap view (status-based Kanban), embeddable widgets for both, guest posting capability.

**Addresses:** Public roadmap commitment overload (PITFALLS.md #4) — Now/Next/Later format without dates, clear communication

**Implements:** Public board architecture, guest identity handling

**Research flag:** SKIP PHASE RESEARCH — Public board UX patterns well-established across competitors

### Phase 8: iOS SDK
**Rationale:** Platform-specific, highest complexity. Only after Web SDK validates patterns. Requires native expertise and longer iteration cycle. Native differentiation vs Canny's WebView approach.

**Delivers:** Swift Package with SwiftUI SDK, XCFramework binary distribution, async/await API, optional feedback widget UI, offline queueing and sync.

**Uses:** Swift 6.0+ (STACK.md), SwiftUI for iOS 16+, SPM + XCFramework distribution

**Implements:** iOS SDK architecture (ARCHITECTURE.md), SwiftUI-first with UIKit bridge

**Addresses:** SDK versioning (PITFALLS.md #9) — same API contract as Web SDK, consistent behavior

**Research flag:** POSSIBLE PHASE RESEARCH — If team lacks iOS experience, research SwiftUI patterns, XCFramework distribution, and SPM best practices. Otherwise skip.

### Phase 9: AI Features (Duplicate Detection + Categorization)
**Rationale:** Becoming table stakes based on competitor analysis but can defer until scale demands it. Manual duplicate merging works initially. Categorization can be user-selected first.

**Delivers:** AI duplicate detection on submission (show similar before creating), automatic categorization suggestions, merge duplicates workflow for admins.

**Addresses:** Search and discovery (PITFALLS.md #14) — full-text search + semantic similarity

**Research flag:** NEEDS PHASE RESEARCH — AI/ML integration patterns, embedding models, similarity search in PostgreSQL (pgvector), LLM API costs and latency

### Phase 10: Multi-App Workspaces
**Rationale:** Competitive differentiator but requires workspace foundation from Phase 1. Complex UI/UX for managing multiple apps. Defer until single-app experience is solid.

**Delivers:** App switcher UI, workspace-level settings, per-app API keys, cross-app analytics and filtering.

**Implements:** Multi-app workspace architecture (ARCHITECTURE.md), app-level isolation within workspace

**Research flag:** SKIP PHASE RESEARCH — Architecture already designed, UI/UX patterns are standard

### Phase 11: Landing Page + Waitlist
**Rationale:** Marketing separate from product. Can be built in parallel anytime. Zero dependencies on main system. Validates demand before features complete.

**Delivers:** CloudFlare Pages static site, Workers for waitlist API, D1 database for emails, Resend for confirmation emails, referral tracking.

**Uses:** CloudFlare Workers + D1 (STACK.md), Astro for static generation

**Implements:** Landing architecture (ARCHITECTURE.md), edge-first with isolated database

**Research flag:** SKIP PHASE RESEARCH — CloudFlare Workers/D1 patterns well-documented

### Phase Ordering Rationale

- **Foundation first (1-2):** Database and API must be correct before building on top. Multi-tenant isolation mistakes are nearly impossible to fix retroactively.
- **Validate before SDK investment (3-4):** Dashboard + voting proves core value proposition. Web SDK validates patterns before expensive iOS work.
- **User-facing features after dogfooding (7):** Public board/roadmap built after internal use reveals UX issues.
- **AI deferred to later (9):** Manual workflows sufficient at low scale. AI becomes valuable when volume demands automation.
- **Landing page parallel (11):** No dependencies, can build anytime or even before Phase 1.

**Dependency chain:**
```
Phase 1 (Foundation) → Phase 2 (API) → Phase 3 (Dashboard) → Phase 4 (Voting)
                                    → Phase 5 (Web SDK) → Phase 8 (iOS SDK)
                                    → Phase 6 (Email) → Phase 7 (Public Board)
                                    → Phase 9 (AI) → Phase 10 (Multi-App)

Phase 11 (Landing) — independent, can run in parallel
```

### Research Flags

**Needs deeper research during planning:**
- **Phase 9 (AI Features):** AI/ML integration patterns sparse, need research on embedding models, pgvector for similarity search, LLM API selection (OpenAI vs alternatives), cost/latency tradeoffs
- **Phase 8 (iOS SDK):** Only if team lacks iOS expertise — research SwiftUI patterns, XCFramework distribution, offline sync strategies

**Standard patterns (skip research-phase):**
- **Phases 1-7, 10-11:** Well-documented patterns in official docs (Supabase, Astro, Go, CloudFlare)
- Go/Chi/sqlc backend patterns are 2025 standard
- Astro islands architecture officially documented
- Supabase RLS multi-tenancy has comprehensive guides
- CloudFlare Workers/D1 have extensive official examples

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies validated via official docs and 2025 best practices. Chi/sqlc/slog are stdlib-aligned Go standards. Astro 5 is stable with official islands architecture. Supabase RLS patterns well-documented. |
| Features | HIGH | User feedback platform feature tiers are well-established across 10+ competitors analyzed. Table stakes clear. Differentiators validated against competitive gaps (native SDKs, multi-app simplicity). |
| Architecture | HIGH | Multi-tier separation follows proven patterns. Supabase multi-tenant RLS architecture has official guides and community validation. Astro islands architecture officially documented. Clean Architecture for Go is standard. |
| Pitfalls | MEDIUM-HIGH | RLS security issues verified via CVE and official warnings. Voting bias validated across multiple product management sources. Pricing pitfalls cross-referenced with SaaS research. Some pitfalls inferred from general patterns. |

**Overall confidence:** HIGH

### Gaps to Address

Areas requiring validation during implementation or user testing:

- **Anonymous feedback decision:** PROJECT.md specifies no anonymous feedback. Validate this with target users (indie devs). If demand exists, schema already supports it (nullable user_id). Have strong rationale ready for why identity matters.

- **Pricing structure validation:** Three tiers defined but need to validate with target market that limits/pricing resonate. Test "user" definition clarity (submitters only? voters too?) to avoid Canny's tracked-user confusion. Ensure upsell path exists between Pro ($79) and Business (custom).

- **AI feature timing:** Research says AI duplicate detection is "becoming table stakes" but manual workflows may be sufficient initially. Validate during Phase 3-4 whether manual duplicate merging is manageable or AI needed sooner.

- **CloudFlare D1 for landing vs Supabase:** Research confidence is MEDIUM for D1 (relatively new, GA but evolving). If D1 causes issues, fallback is using Supabase for waitlist (higher latency but more familiar).

- **iOS SDK scope:** XCFramework distribution is standard but offline sync patterns for feedback SDK are less documented. May need iteration on local queue + batch sync strategy.

- **Notification preferences granularity:** Research warns of notification fatigue but doesn't specify ideal preference controls. Start conservative (digest by default), iterate based on user feedback on what they want to control.

## Sources

### Primary (HIGH confidence)
- [Supabase Row Level Security Docs](https://supabase.com/docs/guides/database/postgres/row-level-security) — Multi-tenant isolation, RLS policy patterns
- [Supabase RLS Best Practices](https://supabase.com/docs/guides/troubleshooting/rls-performance-and-best-practices-Z5Jjwv) — Performance optimization, indexing
- [Supabase Realtime Postgres Changes](https://supabase.com/docs/guides/realtime/postgres-changes) — Live voting updates
- [Astro 5.0 Release](https://astro.build/blog/astro-5/) — Server Islands, ClientRouter
- [Astro Islands Architecture](https://docs.astro.build/en/concepts/islands/) — Partial hydration patterns
- [Astro View Transitions](https://docs.astro.build/en/guides/view-transitions/) — SPA-like navigation
- [Nanostores GitHub](https://github.com/nanostores/nanostores) — Cross-island state management
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog) — Standard library logging
- [sqlc GitHub](https://github.com/sqlc-dev/sqlc) — Type-safe SQL generation
- [Chi GitHub](https://github.com/go-chi/chi) — HTTP router
- [CloudFlare D1 Docs](https://developers.cloudflare.com/d1/) — Edge database patterns
- [Apple: Distributing Binary Frameworks](https://developer.apple.com/documentation/xcode/distributing-binary-frameworks-as-swift-packages) — XCFramework via SPM
- [Resend Go SDK](https://resend.com/docs/send-with-go) — Email integration

### Secondary (MEDIUM confidence)
- [Canny Official](https://canny.io/) — Feature comparison, pricing structure
- [Featurebase](https://www.featurebase.app/) — Competitive feature analysis
- [Fider GitHub](https://github.com/getfider/fider) — Open source reference implementation
- [LogRocket: Go Frameworks 2025](https://blog.logrocket.com/top-go-frameworks-2025/) — Framework selection rationale
- [Go Clean Architecture](https://dev.to/kittipat1413/structuring-a-go-project-with-clean-architecture-a-practical-example-3b3f) — Backend structure patterns
- [Supabase Multi-Tenant Architecture](https://dev.to/blackie360/-enforcing-row-level-security-in-supabase-a-deep-dive-into-lockins-multi-tenant-architecture-4hd2) — RLS tenant isolation
- [Supabase Best Practices](https://www.leanware.co/insights/supabase-best-practices) — Performance optimization
- [How Feature Voting Forums Failed](https://www.productboard.com/blog/how-feature-voting-forums-failed-us/) — Voting bias pitfalls
- [Feature Voting Pitfalls](https://www.savio.io/blog/feature-voting/) — Voting system design
- [Should You Share Your Roadmap Publicly](https://www.launchnotes.com/blog/should-you-share-your-product-roadmap-publicly) — Public roadmap warnings
- [SaaS Pricing Mistakes](https://mucker.com/blog/saas-startup-pricing-mistakes-and-how-to-fix-them-with-kyle-poyar/) — Pricing pitfalls
- [SwiftUI vs UIKit 2025](https://www.alimertgulec.com/en/blog/swiftui-vs-uikit-2025) — iOS SDK technology choice
- [This Dot: JS Build Tools 2025](https://www.thisdot.co/blog/the-2025-guide-to-js-build-tools) — Web SDK bundling

### Tertiary (LOW confidence - patterns inferred)
- [G2 Reviews for Canny](https://www.g2.com/products/canny/reviews) — User pain points
- [Canny Alternatives Analysis](https://www.featurebase.app/blog/canny-alternatives) — Competitive gaps
- Various blog posts on feedback tools, SaaS pricing, indie developer patterns — Aggregate trends

---
*Research completed: 2026-01-21*
*Ready for roadmap: yes*
