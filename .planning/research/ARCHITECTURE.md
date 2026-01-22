# Architecture Research: Illoominate User Feedback Platform

**Project:** Illoominate
**Researched:** 2026-01-21
**Overall Confidence:** HIGH (verified via official Supabase and Astro documentation)

## System Overview

Illoominate follows a **multi-tier architecture** with clear separation between:

1. **Collection Layer** - SDKs that capture feedback at the source
2. **API Layer** - Go backend on Cloud Run for business logic
3. **Data Layer** - Supabase PostgreSQL with RLS for multi-tenant isolation
4. **Presentation Layer** - Astro dashboard with islands architecture
5. **Marketing Layer** - CloudFlare Workers + D1 for landing/waitlist

```
                                    +-------------------+
                                    |   Astro Dashboard |
                                    |   (CF Pages)      |
                                    +--------+----------+
                                             |
                                             v
+-------------+     +-------------+    +-----------+    +------------------+
| iOS SDK     |---->|             |    | Supabase  |<---| Supabase         |
+-------------+     |   Go API    |--->| PostgreSQL|    | Realtime         |
| Web SDK     |---->| (Cloud Run) |    | + RLS     |    | (voting updates) |
+-------------+     +-------------+    +-----------+    +------------------+
                           |
                           v
                    +-------------+
                    | Resend      |
                    | (Email)     |
                    +-------------+

+-------------------+
| CF Workers + D1   | (Landing page / waitlist - separate)
+-------------------+
```

## Components

### Backend API (Go on Cloud Run)

**Responsibilities:**
- Receive feedback submissions from SDKs
- Authenticate requests (API keys for SDKs, JWT for dashboard)
- Business logic (deduplication, spam filtering, categorization)
- Database operations via Supabase client
- Trigger email notifications via Resend

**Recommended Architecture Pattern: Clean Architecture**

Based on [Go Clean Architecture patterns](https://dev.to/kittipat1413/structuring-a-go-project-with-clean-architecture-a-practical-example-3b3f), structure the Go service with clear layer separation:

```
cmd/
  api/
    main.go              # Entry point, wire dependencies
internal/
  domain/                # Entities - core business objects
    feedback.go          # Feedback, Vote, Category entities
    workspace.go         # Workspace, App, User entities
  usecase/               # Application business rules
    feedback/
      submit.go          # Submit feedback use case
      vote.go            # Vote on feedback use case
      list.go            # List/filter feedback use case
    workspace/
      manage.go          # Workspace management
  repository/            # Interfaces (ports)
    feedback_repo.go     # Feedback repository interface
    workspace_repo.go    # Workspace repository interface
  infrastructure/        # External implementations (adapters)
    supabase/            # Supabase PostgreSQL implementation
    resend/              # Email notification implementation
  delivery/
    http/
      handler/           # HTTP handlers
      middleware/        # Auth, logging, CORS
      router.go          # Route definitions
config/
  config.go              # Environment configuration
```

**Key Design Decisions:**

1. **Stateless** - Cloud Run containers scale to zero; no in-memory state
2. **Request-scoped DB connections** - Use Supabase REST API or pooled connections
3. **Idempotent operations** - SDKs may retry; design for it
4. **Structured logging** - JSON logs for Cloud Run observability

**API Authentication Strategy:**

| Client | Auth Method | Implementation |
|--------|-------------|----------------|
| SDKs | API Key | `X-API-Key` header, validate against `api_keys` table |
| Dashboard | JWT | Supabase Auth JWT, verify signature |
| Webhooks | Signature | HMAC signature verification |

### Database (Supabase PostgreSQL)

**Schema Approach: Shared Database with Tenant ID Column**

Based on [Supabase RLS documentation](https://supabase.com/docs/guides/database/postgres/row-level-security), use the **tenant_id column pattern** with JWT claims for workspace isolation.

**Core Schema Design:**

```sql
-- Tenant hierarchy
CREATE TABLE workspaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  settings JSONB DEFAULT '{}'
);

CREATE TABLE apps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  platform TEXT NOT NULL, -- 'ios', 'web', 'android'
  api_key TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  settings JSONB DEFAULT '{}'
);

-- User membership (links Supabase Auth users to workspaces)
CREATE TABLE workspace_members (
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'member', -- 'owner', 'admin', 'member'
  created_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (workspace_id, user_id)
);

-- Core feedback data
CREATE TABLE feedback (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT,
  category TEXT, -- 'bug', 'feature', 'improvement', 'question'
  status TEXT DEFAULT 'open', -- 'open', 'under_review', 'planned', 'in_progress', 'completed', 'declined'
  submitter_email TEXT,
  submitter_user_id TEXT, -- From client app's user system
  vote_count INTEGER DEFAULT 0, -- Denormalized for performance
  metadata JSONB DEFAULT '{}', -- Device info, app version, etc.
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Vote tracking (prevents duplicates, enables vote changes)
CREATE TABLE votes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  voter_identifier TEXT NOT NULL, -- Email or anonymous ID
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(feedback_id, voter_identifier)
);

-- Comments on feedback
CREATE TABLE comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id UUID REFERENCES auth.users(id), -- Null for anonymous
  author_name TEXT,
  content TEXT NOT NULL,
  is_internal BOOLEAN DEFAULT false, -- Internal team notes
  created_at TIMESTAMPTZ DEFAULT now()
);

-- API key audit log
CREATE TABLE api_key_usage (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  endpoint TEXT NOT NULL,
  ip_address INET,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Indexes for common queries
CREATE INDEX idx_feedback_workspace ON feedback(workspace_id);
CREATE INDEX idx_feedback_app ON feedback(app_id);
CREATE INDEX idx_feedback_status ON feedback(workspace_id, status);
CREATE INDEX idx_feedback_created ON feedback(workspace_id, created_at DESC);
CREATE INDEX idx_votes_feedback ON votes(feedback_id);
CREATE INDEX idx_comments_feedback ON comments(feedback_id);
```

**Row Level Security Policies:**

```sql
-- Enable RLS on all tables
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE apps ENABLE ROW LEVEL SECURITY;
ALTER TABLE feedback ENABLE ROW LEVEL SECURITY;
ALTER TABLE votes ENABLE ROW LEVEL SECURITY;
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

-- Workspace access: User must be a member
CREATE POLICY workspace_member_access ON workspaces
  FOR ALL
  USING (
    id IN (
      SELECT workspace_id FROM workspace_members
      WHERE user_id = (SELECT auth.uid())
    )
  );

-- Apps access: User must be member of parent workspace
CREATE POLICY apps_workspace_member ON apps
  FOR ALL
  USING (
    workspace_id IN (
      SELECT workspace_id FROM workspace_members
      WHERE user_id = (SELECT auth.uid())
    )
  );

-- Feedback: Workspace members can read all, SDK can insert via service role
CREATE POLICY feedback_read ON feedback
  FOR SELECT
  USING (
    workspace_id IN (
      SELECT workspace_id FROM workspace_members
      WHERE user_id = (SELECT auth.uid())
    )
  );

-- Similar policies for votes, comments
```

**Performance Optimization:**

Per [Supabase best practices](https://www.leanware.co/insights/supabase-best-practices):

1. **Index RLS columns** - `workspace_id` is indexed on all tables
2. **Wrap `auth.uid()` in `SELECT`** - Prevents re-evaluation per row
3. **Denormalize vote_count** - Avoid COUNT(*) on every feedback query
4. **Use service role for SDK writes** - Bypass RLS, enforce in API layer

### Realtime Voting Updates

Based on [Supabase Realtime documentation](https://supabase.com/docs/guides/realtime/postgres-changes), implement voting updates via Postgres Changes:

```typescript
// Dashboard subscription
const channel = supabase
  .channel('feedback-changes')
  .on('postgres_changes', {
    event: 'UPDATE',
    schema: 'public',
    table: 'feedback',
    filter: `workspace_id=eq.${workspaceId}`
  }, (payload) => {
    // Update local state with new vote_count
    updateFeedbackItem(payload.new)
  })
  .subscribe()
```

**Scaling Considerations:**

Per documentation, Realtime processes changes on a single thread. For scale:

1. **Selective subscriptions** - Filter by workspace_id to reduce broadcast scope
2. **Vote count denormalization** - Trigger updates `vote_count` so only one field changes
3. **Consider Broadcast for high-volume** - If >500 concurrent users per workspace

```sql
-- Trigger to update vote_count on vote insert/delete
CREATE OR REPLACE FUNCTION update_vote_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE feedback SET vote_count = vote_count + 1 WHERE id = NEW.feedback_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE feedback SET vote_count = vote_count - 1 WHERE id = OLD.feedback_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER vote_count_trigger
AFTER INSERT OR DELETE ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_count();
```

### Web Frontend (Astro on CloudFlare Pages)

**Architecture Pattern: Islands with React Components**

Based on [Astro Islands documentation](https://docs.astro.build/en/concepts/islands/), structure the dashboard as:

- **Static shell** - Layout, navigation, headers (pure Astro)
- **Interactive islands** - Data grids, forms, real-time components (React)

```
src/
  layouts/
    DashboardLayout.astro    # Static shell
  pages/
    dashboard/
      index.astro            # Dashboard home (static wrapper)
      feedback/
        index.astro          # Feedback list page
        [id].astro           # Feedback detail page
  components/
    static/                  # Astro components (no JS)
      Header.astro
      Sidebar.astro
    islands/                 # React components (hydrated)
      FeedbackTable.tsx      # client:load - needs immediate interactivity
      VoteButton.tsx         # client:visible - load when scrolled into view
      FilterPanel.tsx        # client:idle - load when browser is idle
      RealtimeBadge.tsx      # client:load - real-time connection indicator
  lib/
    supabase.ts              # Supabase client initialization
    api.ts                   # API wrapper functions
  stores/
    feedback.ts              # Shared state (nanostores or signals)
```

**Hydration Strategy:**

| Component | Directive | Rationale |
|-----------|-----------|-----------|
| FeedbackTable | `client:load` | Core interactivity, needs to work immediately |
| VoteButton | `client:visible` | Many on page, hydrate only when visible |
| FilterPanel | `client:idle` | Not critical path, can wait |
| RealtimeBadge | `client:load` | WebSocket connection needed immediately |
| Charts | `client:visible` | Heavy, only load when user scrolls to them |

**State Management:**

For cross-island state sharing, use [nanostores](https://github.com/nanostores/nanostores):

```typescript
// stores/feedback.ts
import { atom, map } from 'nanostores'

export const $feedbackItems = atom<Feedback[]>([])
export const $filters = map<FilterState>({
  status: 'all',
  category: 'all',
  sortBy: 'votes'
})
```

### Native SDKs

**iOS SDK (Swift)**

Following [Azure SDK design guidelines](https://azure.github.io/azure-sdk/ios_design.html) and Swift best practices:

```swift
// Public API
public class IlloominateSDK {
    public static let shared = IlloominateSDK()

    private var apiKey: String?
    private var baseURL: URL

    public func configure(apiKey: String, environment: Environment = .production) {
        self.apiKey = apiKey
        self.baseURL = environment.baseURL
    }

    public func submitFeedback(
        title: String,
        description: String? = nil,
        category: FeedbackCategory = .feature,
        userEmail: String? = nil,
        metadata: [String: Any]? = nil
    ) async throws -> FeedbackResponse {
        // Implementation
    }

    public func vote(feedbackId: String, voterIdentifier: String) async throws {
        // Implementation
    }

    public func showFeedbackWidget(from viewController: UIViewController) {
        // Present built-in UI
    }
}

// Types
public enum FeedbackCategory: String, Codable {
    case bug, feature, improvement, question
}

public enum Environment {
    case production, staging
    var baseURL: URL { /* ... */ }
}
```

**Design Patterns:**

1. **Singleton with configuration** - `IlloominateSDK.shared.configure()`
2. **Async/await** - Modern Swift concurrency
3. **Optional UI component** - Can use API only or built-in widget
4. **Offline queue** - Buffer submissions when offline, sync when connected

**Web SDK (JavaScript/TypeScript)**

```typescript
// @illoominate/sdk
export interface IlloominateConfig {
  apiKey: string
  environment?: 'production' | 'staging'
  autoCapture?: boolean // Capture errors automatically
}

export interface FeedbackSubmission {
  title: string
  description?: string
  category?: 'bug' | 'feature' | 'improvement' | 'question'
  userEmail?: string
  metadata?: Record<string, unknown>
}

class Illoominate {
  private config: IlloominateConfig | null = null

  init(config: IlloominateConfig): void {
    this.config = config
    if (config.autoCapture) {
      this.setupErrorCapture()
    }
  }

  async submitFeedback(feedback: FeedbackSubmission): Promise<FeedbackResponse> {
    // POST to API
  }

  async vote(feedbackId: string, voterIdentifier: string): Promise<void> {
    // POST to API
  }

  showWidget(options?: WidgetOptions): void {
    // Inject and display feedback widget
  }

  private setupErrorCapture(): void {
    // Listen to window.onerror, unhandledrejection
  }
}

export const illoominate = new Illoominate()
export default illoominate

// Usage
import illoominate from '@illoominate/sdk'
illoominate.init({ apiKey: 'illo_xxx' })
illoominate.submitFeedback({ title: 'Add dark mode' })
```

**Bundle Considerations:**

1. **Tree-shakeable** - Widget UI separate from core API
2. **Small core** - < 5KB gzipped for API-only usage
3. **CDN distribution** - UMD build for script tag usage
4. **TypeScript-first** - Full type definitions

### Landing Page (CloudFlare Workers + D1)

**Architecture: Edge-First with D1**

Per [Cloudflare D1 documentation](https://developers.cloudflare.com/d1/), D1 is ideal for this use case:

```
src/
  index.ts          # Worker entry point
  routes/
    waitlist.ts     # POST /waitlist - add email
    stats.ts        # GET /stats - waitlist count (optional)
  lib/
    d1.ts           # D1 queries
    validation.ts   # Email validation
```

**D1 Schema:**

```sql
CREATE TABLE waitlist (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  referral_code TEXT,
  created_at TEXT DEFAULT (datetime('now')),
  converted BOOLEAN DEFAULT false
);

CREATE INDEX idx_waitlist_email ON waitlist(email);
CREATE INDEX idx_waitlist_referral ON waitlist(referral_code);
```

**Worker Implementation:**

```typescript
// src/index.ts
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)

    if (url.pathname === '/waitlist' && request.method === 'POST') {
      const { email, referralCode } = await request.json()

      // Validate email
      if (!isValidEmail(email)) {
        return Response.json({ error: 'Invalid email' }, { status: 400 })
      }

      // Insert into D1
      try {
        await env.DB.prepare(
          'INSERT INTO waitlist (email, referral_code) VALUES (?, ?)'
        ).bind(email, referralCode || null).run()

        return Response.json({ success: true })
      } catch (e) {
        if (e.message.includes('UNIQUE constraint')) {
          return Response.json({ error: 'Already registered' }, { status: 409 })
        }
        throw e
      }
    }

    return new Response('Not found', { status: 404 })
  }
}
```

**Why D1 over Supabase for landing:**

1. **Edge latency** - D1 is globally distributed, < 50ms reads
2. **Cost** - D1 free tier is generous for waitlist scale
3. **Simplicity** - No auth complexity, just email collection
4. **Isolation** - Landing page concerns separate from product

## Data Flow

### Feedback Submission Flow

```
1. User submits feedback via SDK
   iOS/Web App → SDK → HTTP POST

2. SDK sends to Go API
   SDK → Cloud Run (Go API)
   Headers: X-API-Key: illo_xxx
   Body: { title, description, category, userEmail, metadata }

3. API validates and stores
   Go API → Validate API key against `apps` table
         → Extract workspace_id from app
         → Insert into `feedback` table
         → Queue email notification (optional)
         → Return feedback ID

4. Dashboard receives update (real-time)
   Supabase Postgres → Realtime → Dashboard WebSocket
   Dashboard updates UI with new feedback item

5. Email notification sent (async)
   Go API → Resend API → Workspace admins
```

### Voting Flow

```
1. User votes on feedback
   Dashboard/Public Board → HTTP POST /api/feedback/{id}/vote

2. API processes vote
   Go API → Validate user can vote (not duplicate)
         → Insert into `votes` table
         → Trigger updates `feedback.vote_count`
         → Return success

3. Real-time update broadcast
   Postgres trigger → Realtime channel
   All dashboard clients subscribed to workspace see new vote_count
```

### Authentication Flow

```
Dashboard Login:
1. User clicks "Sign in with Google/GitHub"
2. Supabase Auth handles OAuth
3. User returned to dashboard with session
4. Dashboard reads JWT, extracts user_id
5. RLS policies enforce workspace access

SDK Submission:
1. SDK includes X-API-Key header
2. Go API validates key against `apps.api_key`
3. Go API uses service role to bypass RLS
4. Workspace isolation enforced in application code
```

## Multi-Tenancy

**Isolation Strategy: Workspace ID Column + RLS**

Per [Supabase multi-tenant best practices](https://dev.to/blackie360/-enforcing-row-level-security-in-supabase-a-deep-dive-into-lockins-multi-tenant-architecture-4hd2):

| Level | Purpose | Example |
|-------|---------|---------|
| Workspace | Billing/organization boundary | "Acme Corp" |
| App | Product/platform boundary | "Acme iOS", "Acme Web" |
| User | Individual contributor | john@acme.com |

**Data Isolation:**

- All data tables include `workspace_id`
- RLS policies check membership via `workspace_members`
- SDK submissions route through app, which belongs to workspace
- Cross-workspace queries impossible at database level

**Why Not Schema-Per-Tenant:**

1. **Complexity** - Requires dynamic schema management
2. **Scale** - Supabase doesn't optimize for thousands of schemas
3. **Cost** - Shared schema is more connection-efficient
4. **Adequate isolation** - RLS provides sufficient security

## Build Order

Based on component dependencies, recommended build order:

### Phase 1: Foundation (Build First)

**Database Schema** - Everything depends on data structure
- Create tables with RLS policies
- Set up Supabase project
- Configure auth providers

**Why first:** Cannot build API or dashboard without knowing data shape

### Phase 2: Core API

**Go Backend** - Central orchestrator
- Implement feedback CRUD endpoints
- API key authentication
- Supabase integration

**Why second:** SDKs and dashboard both need API to function

### Phase 3: Dashboard MVP

**Astro Dashboard** - See data, manage feedback
- Auth integration with Supabase
- Feedback list view
- Basic filtering

**Why third:** Need to visualize data before refining SDK

### Phase 4: SDKs

**Web SDK first** - Faster iteration
- Core API methods
- Basic widget

**iOS SDK** - Platform-specific
- Swift Package Manager distribution
- SwiftUI widget

**Why after dashboard:** Can test submissions end-to-end

### Phase 5: Real-time & Polish

**Voting system** - Requires foundation in place
**Real-time updates** - Enhancement on working system
**Email notifications** - Non-critical path

### Phase 6: Landing Page

**CloudFlare Workers + D1** - Marketing independent of product
- Can be built in parallel
- No dependencies on main system

## Anti-Patterns to Avoid

### 1. Direct Database Access from SDKs
**Problem:** Exposing Supabase credentials in client SDKs
**Why bad:** API keys in mobile apps can be extracted
**Instead:** All SDK requests go through Go API with server-validated API keys

### 2. Monolithic Frontend
**Problem:** Building dashboard as full React SPA
**Why bad:** Slower initial load, harder to optimize
**Instead:** Astro islands - static shell, interactive components

### 3. Synchronous Email Sending
**Problem:** Sending emails in request path
**Why bad:** Adds latency, fails if email service down
**Instead:** Queue notifications, process async (or use Cloud Tasks)

### 4. Over-relying on Realtime for Critical Data
**Problem:** Only using WebSocket for feedback list
**Why bad:** Connection drops, missed updates
**Instead:** Initial fetch via REST, realtime for updates; periodic refresh

### 5. Global Vote Count Queries
**Problem:** `SELECT COUNT(*) FROM votes WHERE feedback_id = ?`
**Why bad:** Expensive at scale, locks table
**Instead:** Denormalized `vote_count` column with trigger

## Scalability Considerations

| Concern | At 100 users | At 10K users | At 1M users |
|---------|--------------|--------------|-------------|
| API | Single Cloud Run instance | Auto-scale, multiple instances | Regional deployment |
| Database | Supabase free tier | Supabase Pro, connection pooling | Read replicas, query optimization |
| Realtime | Direct Postgres Changes | Filter subscriptions | Broadcast with separate tables |
| SDKs | Direct API calls | Request batching | Local queue + batch sync |

## Sources

**HIGH Confidence (Official Documentation):**
- [Supabase Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [Supabase Realtime Postgres Changes](https://supabase.com/docs/guides/realtime/postgres-changes)
- [Astro Islands Architecture](https://docs.astro.build/en/concepts/islands/)
- [Cloudflare D1 Overview](https://developers.cloudflare.com/d1/)

**MEDIUM Confidence (Verified Patterns):**
- [Go Clean Architecture](https://dev.to/kittipat1413/structuring-a-go-project-with-clean-architecture-a-practical-example-3b3f)
- [Azure iOS SDK Guidelines](https://azure.github.io/azure-sdk/ios_design.html)
- [Supabase Multi-Tenant Architecture](https://dev.to/blackie360/-enforcing-row-level-security-in-supabase-a-deep-dive-into-lockins-multi-tenant-architecture-4hd2)
- [Supabase Best Practices](https://www.leanware.co/insights/supabase-best-practices)

**LOW Confidence (Community Patterns):**
- [System Design: Voting System](https://medium.com/@bugfreeai/system-design-deep-dive-designing-a-voting-system-bff917dbdcd2)
- [Feature Voting Database Schema](https://github.com/sempedia/UpVote)
