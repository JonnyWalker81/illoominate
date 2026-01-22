# Pitfalls Research: Illoominate

**Domain:** User feedback collection platform
**Researched:** 2026-01-21
**Confidence:** MEDIUM-HIGH (multiple sources cross-referenced)

## Critical Pitfalls

Mistakes that could kill the project or require major rewrites.

---

### 1. Multi-Tenant Data Isolation Failures

**What:** Misconfigured tenant isolation allows one workspace to access another's feedback data. With Supabase RLS, a single missing `tenant_id` check becomes a data leak.

**Warning Signs:**
- RLS policies use only `auth.uid()` without workspace/tenant context
- No automated tests verifying cross-tenant data isolation
- Queries returning data without explicit workspace filter
- Using `user_metadata` (user-modifiable) instead of `app_metadata` for authorization

**Prevention:**
- Design RLS policies with workspace_id checks from day one
- Store authorization data (workspace membership, roles) in `app_metadata` not `user_metadata`
- Implement integration tests that attempt cross-tenant access and verify denial
- Index `workspace_id` and `user_id` columns for RLS performance
- Never use `service_role` keys in client code (bypasses RLS)

**Phase:** Phase 1 (Core Data Model) - Must be correct before any data exists

**Sources:** [Supabase RLS Docs](https://supabase.com/docs/guides/database/postgres/row-level-security), [Multi-Tenant RLS Deep Dive](https://dev.to/blackie360/-enforcing-row-level-security-in-supabase-a-deep-dive-into-lockins-multi-tenant-architecture-4hd2)

---

### 2. Voting System Bias Creating Misleading Priorities

**What:** Feature voting boards naturally bias toward vocal minorities, early submissions, and popular momentum. Top-voted features get built but go unused because everyone voting imagined different implementations.

**Warning Signs:**
- Same users voting on everything (squeaky wheel problem)
- Features at top of list always get more votes (position bias)
- High vote counts on vague feature requests
- Building top-voted feature only to see low adoption

**Prevention:**
- Randomize feature display order (don't sort by votes by default)
- Consider hiding vote counts from voters to reduce herd behavior
- Implement vote limits per user (forces prioritization)
- Weight votes by customer value (paid tier, MRR contribution)
- Capture context with votes (why they want it, their use case)
- Track voter engagement post-ship to validate voting signal quality

**Phase:** Phase with Voting & Engagement feature

**Sources:** [How Feature Voting Forums Failed](https://www.productboard.com/blog/how-feature-voting-forums-failed-us/), [Feature Voting Pitfalls](https://www.savio.io/blog/feature-voting/), [Why Feature Voting Creates Poor Products](https://jasonevanish.com/2021/04/23/why-feature-voting-creates-poor-products-and-what-to-do-instead/)

---

### 3. Building SDK Before Validating Core Product

**What:** Native SDKs (iOS Swift, Web JS) are significant investment. Building them before validating the feedback loop works wastes engineering time if the product concept is wrong.

**Warning Signs:**
- Spending weeks on SDK polish before first paying customer
- SDK feature requests piling up before core dashboard is stable
- Delaying user testing because "SDK isn't ready yet"

**Prevention:**
- Build web dashboard and API first
- Use simple web-based submission initially (even for mobile users)
- Validate value proposition with web-only flow
- Only invest in native SDKs after proving users want the product
- Consider starting with a single SDK (Web JS) before iOS

**Phase:** Pre-SDK phases - Validate core feedback loop first

**Sources:** [SaaS MVP Mistakes](https://dev.to/shayy/7-mistakes-every-developer-makes-when-building-their-first-saas-and-how-i-fixed-them-4mi3), [Building Enterprise SaaS Lessons](https://www.mindtheproduct.com/building-enterprise-saa-s-lessons-learned-from-several-product-launches/)

---

## Common Mistakes

Frequent errors that cause delays or technical debt.

---

### 4. Public Roadmap Commitment Overload

**What:** Treating public roadmap as promises creates pressure to ship features that should be cut. Teams lose flexibility to pivot based on learning.

**Warning Signs:**
- Unable to remove items from public roadmap without backlash
- Shipping features you know are wrong because they're "promised"
- Roadmap items have specific dates visible to users
- Customer success fielding "when will X ship?" constantly

**Prevention:**
- Use "Now / Next / Later" format without dates
- Clearly communicate roadmaps show direction, not commitments
- Build governance: items within 2 weeks can ship, 6+ months need review before public
- Enable users to subscribe for updates rather than waiting on dates
- Communicate pivots transparently with reasoning

**Phase:** Phase with Public Roadmap feature

**Sources:** [Should You Share Your Roadmap Publicly](https://www.launchnotes.com/blog/should-you-share-your-product-roadmap-publicly), [Product Roadmap Sharing Mistakes](https://www.productplan.com/learn/product-roadmap-sharing-mistakes/)

---

### 5. Notification Fatigue Destroying Engagement

**What:** Over-notifying users kills engagement. Every notification trains users whether to pay attention or ignore.

**Warning Signs:**
- Users asking how to turn off notifications
- Email unsubscribe rates climbing
- Users enabling "do not disturb" in your app
- Notification open rates declining over time

**Prevention:**
- Default to conservative notification settings
- Group related notifications (digest rather than individual)
- Let users control granularity (per-feature, per-workspace settings)
- Track and monitor notification engagement rates
- Never interrupt critical user flows with notifications
- Respect time zones for email notifications

**Phase:** Phase with Notifications feature

**Sources:** [In-App Notification Pitfalls](https://sceyt.com/blog/in-app-notification-mistakes-and-how-to-avoid-them), [Notification Design Best Practices](https://www.magicbell.com/blog/in-app-notification-design)

---

### 6. Copying Canny's UX Without Understanding Why

**What:** Canny has specific UX patterns that work for their enterprise customers but may not fit indie/startup users. Blindly copying creates complexity without value.

**Warning Signs:**
- Building features because "Canny has it"
- UI complexity growing without proportional user value
- Different editing experiences across sections (Canny's actual complaint)
- Admin UI requiring extensive onboarding

**Prevention:**
- Research why competitors made specific choices
- Validate each feature with target audience (indie devs, not enterprise)
- Prioritize simplicity over feature parity
- Build for your market position, not theirs
- User test with actual indie developers early

**Phase:** All UI/UX phases

**Sources:** [Canny UI Issues](https://quickhunt.app/blog/canny-alternatives), [Canny Alternatives Analysis](https://www.featurebase.app/blog/canny-alternatives)

---

### 7. Cold Start Latency on Cloud Run

**What:** Serverless cold starts on Cloud Run can add 300ms-1s+ latency. For SDK API calls, this creates unacceptable user experience during first interaction.

**Warning Signs:**
- First API call after idle period takes 2+ seconds
- Users complaining about "slow" feedback submission
- Sporadic timeout errors in monitoring

**Prevention:**
- Set minimum instances > 0 in Cloud Run config (1-2 minimum for production)
- Keep container image small (Go compiles to small binaries, use minimal base image)
- Move initialization logic outside request handlers (db connections, etc.)
- Monitor P95/P99 latency, not just average
- Consider "keep-warm" mechanism for low-traffic periods

**Phase:** Phase 1 (Backend Infrastructure) - Configure from start

**Sources:** [Cloud Run Cold Starts](https://medium.com/google-cloud/go-serverless-with-google-cloud-run-functions-376b065c6c48), [Mitigating Cold Starts](https://www.mindfulchase.com/explore/troubleshooting-tips/cloud-platforms-and-services/mitigating-cold-starts-and-latency-in-google-cloud-run-applications.html)

---

## Technical Pitfalls

Architecture and implementation mistakes specific to this stack.

---

### 8. RLS Performance Degradation at Scale

**What:** Supabase RLS policies with complex joins slow down every query. Performance degrades as data grows, often unnoticed until production load.

**Warning Signs:**
- Query times increasing gradually over months
- RLS policies using subqueries or complex joins
- Missing indexes on columns used in policies
- No performance testing with realistic data volumes

**Prevention:**
- Index all columns used in RLS policies (`workspace_id`, `user_id`)
- Use custom JWT claims for tenant context (avoid subqueries)
- Run `EXPLAIN ANALYZE` on queries with RLS enabled
- Load test with realistic tenant data volumes early
- Monitor query performance in production from day one

**Phase:** Phase 1 (Core Data Model) and ongoing

**Sources:** [Supabase Best Practices](https://www.leanware.co/insights/supabase-best-practices), [RLS Performance](https://medium.com/@jay.digitalmarketing09/how-to-manage-row-level-security-policies-effectively-in-supabase-98c9dfbc2c01)

---

### 9. SDK Versioning and Breaking Changes

**What:** Once SDKs ship to customers, breaking changes become expensive. Version management across iOS/Web/future-Android creates complexity.

**Warning Signs:**
- Shipped SDK with public API you want to change
- No versioning strategy for API endpoints
- Breaking changes requiring all customers to update simultaneously
- Supporting multiple SDK versions becoming maintenance burden

**Prevention:**
- Design API with versioning from day one (`/v1/feedback`, etc.)
- Keep SDK surface area minimal initially
- Use semantic versioning strictly
- Deprecation policy with sunset windows (minimum 6 months)
- Test SDK backwards compatibility in CI
- Document migration paths for breaking changes

**Phase:** Phase with SDK development

**Sources:** General SDK development best practices, [npm versioning patterns](https://docs.npmjs.com/about-semantic-versioning)

---

### 10. Feedback Data Model Inflexibility

**What:** Feedback platforms need evolving schemas (custom fields, new feedback types, metadata). Rigid data model early on makes later changes painful.

**Warning Signs:**
- Hardcoded feedback types (Bug, Feature, UI/UX) in schema
- Custom fields requested but impossible to add
- Migration-heavy changes for new feedback attributes
- Different workspaces needing different feedback structures

**Prevention:**
- Use flexible schema for feedback metadata (JSONB in Postgres)
- Keep core fields minimal (id, workspace_id, user_id, type, title, description, created_at)
- Store custom/optional fields in JSONB column with validation
- Plan for custom fields feature from architecture perspective

**Phase:** Phase 1 (Core Data Model)

---

### 11. Supabase Auth vs Custom Identity Confusion

**What:** Illoominate has two identity concepts: Supabase auth users (admins logging into dashboard) and SDK-passed identities (end users submitting feedback). Conflating these creates security and data issues.

**Warning Signs:**
- Trying to create Supabase auth accounts for every end user
- SDK identity tokens conflated with Supabase JWTs
- Unable to link anonymous feedback to later-authenticated user
- Identity lookup queries becoming complex

**Prevention:**
- Clear separation: Supabase Auth for workspace admins only
- SDK users stored in separate `end_users` table with app-specific identity
- SDK uses API keys + user identity payload, not Supabase auth
- Design identity linking flow early (same email across submissions)

**Phase:** Phase with User Identity + Auth features

---

## Product Pitfalls

Feature and UX mistakes common to feedback platforms.

---

### 12. Anonymous Feedback Later Regret

**What:** PROJECT.md specifies no anonymous feedback, but customers may demand it. Adding later requires rethinking the entire data model.

**Warning Signs:**
- Early customer requests for "quick feedback" without login
- Competitors launching anonymous submission features
- Users abandoning submission due to identity requirement

**Prevention:**
- Validate "identity required" assumption with target users
- If sticking with it, have strong rationale (quality over quantity)
- Design schema that could support anonymous if needed (nullable user_id with flag)
- Clear messaging on why identity matters (your feedback drives results)

**Phase:** Pre-MVP validation

**Note:** This is a product decision, not a mistake. But lock it in consciously.

---

### 13. Feedback Fragmentation Across Channels

**What:** Users give feedback in many places (SDK, email, social media, support). Platform only captures SDK submissions, missing the full picture.

**Warning Signs:**
- Admins manually copying feedback from email/Twitter into system
- Duplicate feedback items because user tried multiple channels
- "We already know about this" conversations with users
- Integrations (Zendesk, Intercom) deferred indefinitely

**Prevention:**
- Accept this limitation consciously for v1 (integrations are v2)
- Provide easy ways for admins to manually add feedback from other channels
- Email import feature (forward-to-feedback@)
- Clear value prop: not replacing all channels, enhancing in-app feedback

**Phase:** v1 conscious scope limitation, v2 integration planning

**Sources:** [Canny Feedback Collection Limitations](https://www.savio.io/blog/canny-alternatives/)

---

### 14. Search and Discovery Neglect

**What:** As feedback volume grows, finding related feedback becomes critical. Poor search makes admins recreate existing feedback, duplicate work, and miss patterns.

**Warning Signs:**
- Duplicate feedback items with different wording
- Admins can't find feedback they know exists
- No way to discover related feedback when submitting
- Pattern discovery requires manual review

**Prevention:**
- Implement full-text search on feedback content from start (Postgres FTS)
- Show potential duplicates during submission (for both users and admins)
- Plan for tagging/categorization to aid discovery
- Consider semantic/AI search for v2 (deferred per project scope)

**Phase:** Phase with feedback viewing/management features

---

## Business Pitfalls

Pricing and positioning mistakes.

---

### 15. Underpricing Destroys the Business

**What:** Indie developer pricing ($0-79/mo range) has thin margins. Underpricing attracts low-value users who churn fast and demand support.

**Warning Signs:**
- Support burden per customer exceeds subscription value
- Free tier users giving most feedback (but no revenue)
- Unable to afford marketing/growth due to low ARPU
- Competitors with higher prices growing faster

**Prevention:**
- Price on value, not on cost-to-serve
- Free tier must have clear limitations (not just reduced numbers)
- Annual prepay option from launch (improves cash flow)
- Track support costs per pricing tier
- Be willing to raise prices if underpriced
- Don't be afraid to be 20-30% higher than gut says

**Phase:** Pre-launch pricing validation

**Sources:** [SaaS Pricing Mistakes](https://mucker.com/blog/saas-startup-pricing-mistakes-and-how-to-fix-them-with-kyle-poyar/), [Indie Pricing Strategies](https://freemius.com/blog/micro-saas-pricing-strategies/)

---

### 16. "Users" vs "Tracked Users" Definition Drift

**What:** Illoominate's differentiation is not charging per tracked user like Canny. But without clear definition, you risk the same problem with different terminology.

**Warning Signs:**
- Customers confused about what counts toward limits
- Edge cases (API users? Voters? Commenters?) causing billing disputes
- Pricing tiers based on metrics that aren't intuitive
- Growth limiting customer willingness to promote feedback collection

**Prevention:**
- Define "user" clearly and publicly (feedback submitters only? unique identities?)
- Ensure definition aligns with customer mental model
- Limits should be on things customers can control and predict
- Test pricing model explanation with target users

**Phase:** Pre-launch pricing definition

---

### 17. No Upsell Path in Pricing

**What:** Three tiers with hard caps means customers hit ceiling and can't give you more money. Lifetime value capped artificially.

**Warning Signs:**
- Pro customers maxing out limits but unwilling to contact sales
- No revenue growth from existing customers
- Large gap between Pro ($79) and Business (custom)
- Churn when customers outgrow Pro

**Prevention:**
- Design upsell triggers beyond just limits (features, analytics, integrations)
- Consider usage-based add-ons (storage, API calls, seats)
- Self-serve upgrade path from Pro to something between Pro and Business
- Track customers approaching limits and proactively engage

**Phase:** Pricing structure phase

**Sources:** [SaaS Pricing Strategy](https://www.alpinesg.com/blog/saas-pricing-strategy-3-common-pricing-mistakes-founders-make)

---

## Phase-Specific Warning Summary

| Phase Topic | Primary Pitfall | Mitigation |
|-------------|-----------------|------------|
| Core Data Model | Multi-tenant isolation (#1), RLS performance (#8) | RLS policies with workspace_id, index columns |
| Backend Infrastructure | Cold start latency (#7) | Minimum instances, small images |
| User Identity | Auth confusion (#11), Anonymous regret (#12) | Clear separation of identity types |
| Feedback Submission | Data model inflexibility (#10) | JSONB for extensibility |
| Voting System | Voting bias (#2) | Randomize order, vote limits |
| Public Roadmap | Commitment overload (#4) | Now/Next/Later, no dates |
| Notifications | Notification fatigue (#5) | Conservative defaults, digests |
| SDK Development | Versioning (#9), Building too early (#3) | API versioning, validate first |
| Pricing Launch | Underpricing (#15), Definition drift (#16), No upsell (#17) | Value-based pricing, clear definitions |

---

## Sources Summary

**Multi-Tenant / RLS:**
- [Supabase Row Level Security Docs](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [Multi-Tenant RLS Architecture](https://dev.to/blackie360/-enforcing-row-level-security-in-supabase-a-deep-dive-into-lockins-multi-tenant-architecture-4hd2)
- [Supabase Best Practices](https://www.leanware.co/insights/supabase-best-practices)

**Voting Systems:**
- [How Feature Voting Forums Failed](https://www.productboard.com/blog/how-feature-voting-forums-failed-us/)
- [Feature Voting Pitfalls](https://www.savio.io/blog/feature-voting/)
- [Why Feature Voting Creates Poor Products](https://jasonevanish.com/2021/04/23/why-feature-voting-creates-poor-products-and-what-to-do-instead/)

**Public Roadmaps:**
- [Should You Share Your Roadmap Publicly](https://www.launchnotes.com/blog/should-you-share-your-product-roadmap-publicly)
- [Product Roadmap Sharing Mistakes](https://www.productplan.com/learn/product-roadmap-sharing-mistakes/)

**Notifications:**
- [In-App Notification Pitfalls](https://sceyt.com/blog/in-app-notification-mistakes-and-how-to-avoid-them)
- [Notification Design Best Practices](https://www.magicbell.com/blog/in-app-notification-design)

**Competitor Analysis:**
- [Canny Alternatives Analysis](https://www.featurebase.app/blog/canny-alternatives)
- [Canny Issues Discussion](https://quickhunt.app/blog/canny-alternatives)

**SaaS Pricing:**
- [SaaS Pricing Mistakes](https://mucker.com/blog/saas-startup-pricing-mistakes-and-how-to-fix-them-with-kyle-poyar/)
- [Indie Pricing Strategies](https://freemius.com/blog/micro-saas-pricing-strategies/)

**Serverless/Cloud Run:**
- [Go Serverless with Cloud Run](https://medium.com/google-cloud/go-serverless-with-google-cloud-run-functions-376b065c6c48)
- [Mitigating Cold Starts](https://www.mindfulchase.com/explore/troubleshooting-tips/cloud-platforms-and-services/mitigating-cold-starts-and-latency-in-google-cloud-run-applications.html)
