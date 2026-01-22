# Stack Research: Illoominate

**Domain:** User feedback collection platform (Canny/UserVoice competitor)
**Researched:** 2026-01-21
**Target:** Indie devs, startups (not enterprise)

---

## Executive Summary

Illoominate's pre-defined stack (Go + Supabase + Astro + CloudFlare + native SDKs) is well-aligned with 2025/2026 best practices for a startup-focused SaaS. This research validates the choices and fills in specific libraries, patterns, and versions.

**Key recommendations:**
- Go: Chi router v5 + sqlc + slog (standard library where possible)
- Supabase: Direct PostgreSQL via sqlc, REST API for specific use cases
- Astro: v5 with ClientRouter, Nanostores for cross-island state
- CloudFlare: D1 + Workers + Resend (already correct stack)
- Native SDKs: SwiftUI-first for iOS, Rollup + esbuild for Web JS

---

## Backend (Go on Cloud Run)

### Core Framework

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **Go** | 1.24+ | Language runtime | HIGH |
| **Chi** | v5.2.3 | HTTP router | HIGH |
| **slog** | stdlib | Structured logging | HIGH |
| **sqlc** | v1.30.0 | Type-safe SQL | HIGH |

#### Why Chi over alternatives

**Chi v5.2.3** (go get github.com/go-chi/chi/v5)

Chi is the recommended choice because:
1. **100% net/http compatible** - Uses standard library patterns, any Go middleware works
2. **Lightweight** - Under 1000 LOC, no framework lock-in
3. **Production-proven** - Used at Cloudflare, Heroku, 99Designs
4. **Go 1.22+ routing** - Supports new `http.Request.PathValue()` and `http.Request.Pattern`

**Why NOT Fiber:** Built on fasthttp (not net/http), breaks stdlib compatibility. The performance gains (10-15%) are not worth ecosystem fragmentation for a feedback platform.

**Why NOT Echo:** More opinionated, heavier. Chi's composability is better for iterating quickly.

**Why NOT Gin:** Larger community but uses custom context. Chi's stdlib alignment is preferred.

Sources:
- [LogRocket: Go Frameworks 2025](https://blog.logrocket.com/top-go-frameworks-2025/)
- [JetBrains Go Ecosystem 2025](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)
- [Chi GitHub](https://github.com/go-chi/chi)

#### Why slog (not zap/zerolog)

**slog** (log/slog - standard library since Go 1.21)

1. **Zero dependencies** - Part of standard library
2. **Good enough performance** - 650ns/op vs zap's 420ns/op (negligible for feedback API)
3. **Future-proof** - Standard library won't break, third-party loggers may
4. **JSON handler for production** - `slog.NewJSONHandler()` for Cloud Run logs

```go
// Production setup
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)
```

**When to use zap instead:** Only if you have extreme logging throughput (>100k logs/sec).

Sources:
- [Better Stack: Logging in Go with Slog](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog)

#### Why sqlc (not GORM/ent)

**sqlc v1.30.0** (go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)

sqlc generates type-safe Go code from SQL queries. You write SQL, it generates Go.

Why this is ideal for Illoominate:
1. **Type-safe** - Compile-time errors for SQL mistakes
2. **No ORM overhead** - Raw SQL performance, no N+1 surprises
3. **PostgreSQL-native** - Full access to Supabase PostgreSQL features (JSONB, arrays, etc.)
4. **pgx/v5 support** - Modern PostgreSQL driver

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "schema/"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
```

**Why NOT GORM:** Too much magic, hard to debug, N+1 query risks.

**Why NOT ent:** GraphQL-oriented, overkill for REST API.

**Why NOT raw database/sql:** Too much boilerplate, error-prone.

Sources:
- [sqlc GitHub](https://github.com/sqlc-dev/sqlc)
- [Leapcell: Type-Safe SQL in Go](https://leapcell.io/blog/type-safe-sql-go-sqlc)

### Supabase Integration

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **pgx** | v5 | PostgreSQL driver | HIGH |
| **supabase-go** | v0.0.5 | Auth, Storage, Realtime | MEDIUM |
| **postgrest-go** | latest | REST queries (optional) | MEDIUM |

#### Database Access Strategy

**Recommended: Direct PostgreSQL connection via pgx/v5 + sqlc**

For the Go backend, bypass the Supabase REST API and connect directly to PostgreSQL:

```go
// Direct connection (recommended for backend)
connString := "postgres://user:pass@db.xxx.supabase.co:5432/postgres"
pool, err := pgxpool.New(ctx, connString)
```

**Why direct over REST:**
1. **Performance** - No HTTP overhead for database queries
2. **Type safety** - sqlc generates typed queries
3. **Full SQL** - Complex joins, CTEs, window functions
4. **Transactions** - Proper ACID support

**When to use supabase-go:**
- Authentication (GoTrue integration)
- Storage (file uploads)
- Realtime subscriptions
- RPC calls to database functions

```go
// supabase-go for auth
import "github.com/supabase-community/supabase-go"

client, _ := supabase.NewClient(url, anonKey, nil)
user, err := client.Auth.SignInWithEmailPassword(ctx, email, password)
```

Sources:
- [supabase-go GitHub](https://github.com/supabase-community/supabase-go)
- [Supabase RLS Best Practices](https://supabase.com/docs/guides/troubleshooting/rls-performance-and-best-practices-Z5Jjwv)

#### Row Level Security (RLS) Pattern

**CRITICAL: RLS must be enabled on all tables.**

83% of exposed Supabase databases involve RLS misconfigurations. In January 2025, 170+ apps were found with exposed databases due to missing RLS (CVE-2025-48757).

**Pattern for Go backend:**
1. Use **service key** (bypasses RLS) for admin operations
2. Use **anon key + user JWT** for user-context operations
3. Always verify JWT in Go middleware before using service key

```sql
-- Example: users can only see their own feedback
CREATE POLICY "Users see own feedback" ON feedback
  FOR SELECT
  USING (auth.uid() = user_id);

-- Example: org members see org feedback
CREATE POLICY "Org members see org feedback" ON feedback
  FOR SELECT
  USING (
    org_id IN (
      SELECT org_id FROM org_members WHERE user_id = auth.uid()
    )
  );
```

### Supporting Libraries

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| **golang-jwt/jwt** | v5 | JWT validation | HIGH |
| **google-uuid** | v1 | UUID generation | HIGH |
| **validator** | v10 | Request validation | HIGH |
| **testify** | v1 | Testing assertions | HIGH |

### Cloud Run Deployment

| Aspect | Recommendation | Confidence |
|--------|---------------|------------|
| **Container** | Multi-stage Dockerfile, distroless base | HIGH |
| **Secrets** | Google Secret Manager (not env vars) | HIGH |
| **Scaling** | Min 0, Max based on load testing | MEDIUM |
| **Region** | Same region as Supabase | HIGH |

```dockerfile
# Multi-stage build for minimal image
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
```

Sources:
- [Cloud Run Go Quickstart](https://docs.cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-go-service)
- [PloyCloud Go Deployment Guide 2025](https://ploy.cloud/blog/go-hosting-deployment-guide-2025/)

---

## Frontend (Astro Dashboard)

### Core Framework

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **Astro** | 5.16+ | Meta-framework | HIGH |
| **React** | 18.3+ | Interactive islands | HIGH |
| **Nanostores** | 1.1.0 | Cross-island state | HIGH |
| **TypeScript** | 5.7+ | Type safety | HIGH |

#### Why Astro for a Dashboard

Astro 5's **Server Islands** and **ClientRouter** enable SPA-like experiences:

1. **Partial hydration** - Only hydrate interactive components
2. **View Transitions** - Native SPA-like navigation (85%+ browser support in 2025)
3. **Server Islands** - Mix cached static + dynamic personalized content
4. **React compatible** - Use React for complex interactive parts

```astro
---
// src/layouts/Layout.astro
import { ClientRouter } from 'astro:transitions';
---
<html>
  <head>
    <ClientRouter />
  </head>
  <body>
    <slot />
  </body>
</html>
```

**Why NOT Next.js:** Overkill for dashboard, RSC complexity unnecessary.

**Why NOT plain React SPA:** Worse performance, no static optimization.

Sources:
- [Astro 5.0 Release](https://astro.build/blog/astro-5/)
- [Astro View Transitions](https://docs.astro.build/en/guides/view-transitions/)

#### State Management: Nanostores

**Nanostores 1.1.0** (npm install nanostores @nanostores/react)

Astro's islands architecture breaks React Context. Nanostores solves this:

```typescript
// stores/user.ts
import { atom } from 'nanostores'

export const $user = atom<User | null>(null)
export const $feedbackItems = atom<Feedback[]>([])
```

```tsx
// components/FeedbackList.tsx
import { useStore } from '@nanostores/react'
import { $feedbackItems } from '../stores/feedback'

export function FeedbackList() {
  const items = useStore($feedbackItems)
  return <ul>{items.map(item => <li key={item.id}>{item.title}</li>)}</ul>
}
```

**Why Nanostores:**
- 286 bytes minified
- Framework-agnostic (works across React/Vue/Svelte islands)
- No provider wrappers needed
- Tree-shakeable

Sources:
- [Astro: Share State Between Islands](https://docs.astro.build/en/recipes/sharing-state-islands/)
- [Nanostores GitHub](https://github.com/nanostores/nanostores)

### UI Components

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| **Tailwind CSS** | 4.0+ | Styling | HIGH |
| **shadcn/ui** | latest | Component primitives | HIGH |
| **Radix UI** | latest | Accessible primitives | HIGH |

**shadcn/ui** is recommended because:
- Copy/paste components (no npm dependency)
- Built on Radix (accessible)
- Tailwind-styled (consistent with stack)
- Customizable (not locked to design system)

---

## Landing Page (CloudFlare Pages + Workers)

### Stack

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **Astro** | 5.16+ | Static site generation | HIGH |
| **CloudFlare Pages** | - | Hosting | HIGH |
| **CloudFlare Workers** | - | API endpoints | HIGH |
| **D1** | - | Waitlist database | HIGH |
| **Resend** | - | Email notifications | HIGH |

#### CloudFlare Workers Patterns

**D1 Best Practices:**

```typescript
// Type-safe D1 queries
interface WaitlistEntry {
  email: string;
  created_at: string;
}

export default {
  async fetch(request: Request, env: Env) {
    const { results } = await env.DB.prepare(
      "SELECT * FROM waitlist WHERE email = ?"
    ).bind(email).all<WaitlistEntry>();
    return Response.json(results);
  }
}
```

**Key patterns:**
1. Use `STRICT` tables in D1 schema
2. Run `PRAGMA optimize` after schema changes
3. Use `withSession("first-primary")` for read-after-write consistency
4. Always parameterize queries (automatic with `.prepare().bind()`)

#### Resend Integration

**resend-go v3.1.0** for Go backend, **@resend/node** for Workers:

```typescript
// Worker: Waitlist confirmation
import { Resend } from 'resend';

const resend = new Resend(env.RESEND_API_KEY);

await resend.emails.send({
  from: 'Illoominate <hello@illoominate.com>',
  to: email,
  subject: 'You\'re on the waitlist!',
  html: '<p>Thanks for joining...</p>'
});
```

Sources:
- [Resend Go SDK](https://resend.com/docs/send-with-go)
- [CloudFlare D1 Docs](https://developers.cloudflare.com/d1/)

---

## Native SDKs

### iOS SDK (Swift)

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **Swift** | 6.0+ | Language | HIGH |
| **SwiftUI** | iOS 16+ | UI framework | HIGH |
| **Swift Package Manager** | - | Distribution | HIGH |
| **XCFramework** | - | Binary distribution | HIGH |

#### Architecture Recommendation

**SwiftUI-first with UIKit bridge for complex needs:**

```swift
// Public API - Simple feedback submission
public struct IlloominateSDK {
    public static func showFeedbackSheet(
        config: FeedbackConfig,
        onSubmit: @escaping (Feedback) -> Void
    ) -> some View {
        FeedbackSheetView(config: config, onSubmit: onSubmit)
    }
}

// Internal - SwiftUI view
struct FeedbackSheetView: View {
    @State private var title = ""
    @State private var description = ""

    var body: some View {
        NavigationStack {
            Form {
                TextField("Title", text: $title)
                TextEditor(text: $description)
            }
        }
    }
}
```

**Why SwiftUI-first:**
- 70% of teams use hybrid (UIKit + SwiftUI) in 2025
- SwiftUI is faster to develop
- UIKit available via `UIHostingController` when needed
- Better declarative state management

#### Distribution

Distribute as **XCFramework via Swift Package Manager:**

```swift
// Package.swift
let package = Package(
    name: "IlloominateSDK",
    platforms: [.iOS(.v16)],
    products: [
        .library(name: "IlloominateSDK", targets: ["IlloominateSDK"])
    ],
    targets: [
        .binaryTarget(
            name: "IlloominateSDK",
            url: "https://github.com/illoominate/ios-sdk/releases/download/v1.0.0/IlloominateSDK.xcframework.zip",
            checksum: "abc123..."
        )
    ]
)
```

**Why SPM + XCFramework:**
- SPM is now standard (CocoaPods declining)
- XCFramework hides implementation details
- Fast integration for end users
- No dependency conflicts

Sources:
- [Apple: Distributing Binary Frameworks](https://developer.apple.com/documentation/xcode/distributing-binary-frameworks-as-swift-packages)
- [SwiftUI vs UIKit 2025](https://www.alimertgulec.com/en/blog/swiftui-vs-uikit-2025)

### Web SDK (JavaScript)

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| **TypeScript** | 5.7+ | Source language | HIGH |
| **Rollup** | 4.x | Bundler | HIGH |
| **esbuild** | 0.24+ | Transpilation | HIGH |

#### Build Configuration

Use **Rollup + esbuild** for optimal library output:

```javascript
// rollup.config.js
import esbuild from 'rollup-plugin-esbuild';
import dts from 'rollup-plugin-dts';

export default [
  {
    input: 'src/index.ts',
    output: [
      { file: 'dist/index.cjs', format: 'cjs' },
      { file: 'dist/index.mjs', format: 'esm' },
    ],
    plugins: [esbuild({ minify: true })],
    external: [] // Bundle all dependencies
  },
  {
    input: 'src/index.ts',
    output: { file: 'dist/index.d.ts', format: 'esm' },
    plugins: [dts()]
  }
];
```

**Output formats:**
- ESM (`.mjs`) - Modern bundlers
- CJS (`.cjs`) - Legacy Node.js
- IIFE (optional) - Direct `<script>` tag usage
- Type declarations (`.d.ts`)

**Why Rollup over Vite/esbuild alone:**
- Better tree-shaking for libraries
- Multiple output formats in one config
- esbuild plugin gives speed benefits

Sources:
- [Medium: JS Bundlers 2025](https://medium.com/@Hariharasudhan_/which-javascript-bundler-is-best-in-2025-vite-vs-rollup-vs-webpack-vs-esbuild-9bca86a9b36e)
- [This Dot: 2025 Guide to JS Build Tools](https://www.thisdot.co/blog/the-2025-guide-to-js-build-tools)

---

## What NOT to Use

| Technology | Why Avoid | Use Instead |
|------------|-----------|-------------|
| **Fiber (Go)** | fasthttp breaks net/http compatibility | Chi |
| **GORM** | ORM magic, N+1 risks, hard to debug | sqlc |
| **Webpack** | Slow, complex config for library dev | Rollup + esbuild |
| **CocoaPods** | Declining, SPM is standard now | Swift Package Manager |
| **Redux/Zustand** | Overkill for Astro islands, breaks with partial hydration | Nanostores |
| **Next.js** | RSC complexity unnecessary for dashboard | Astro |
| **Raw env vars for secrets** | Security risk | Secret Manager |
| **UIKit-only iOS** | Slower development, SwiftUI is mature | SwiftUI-first hybrid |

---

## Complete Installation

### Go Backend

```bash
# Core
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/supabase-community/supabase-go

# Tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Supporting
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
go get github.com/go-playground/validator/v10

# Dev/Test
go get github.com/stretchr/testify
```

### Astro Dashboard

```bash
npm create astro@latest
npx astro add react
npx astro add tailwind

npm install nanostores @nanostores/react
npx shadcn@latest init
```

### Web SDK

```bash
npm install -D rollup rollup-plugin-esbuild rollup-plugin-dts typescript
```

---

## Confidence Assessment

| Area | Confidence | Rationale |
|------|------------|-----------|
| Go stack (Chi, slog, sqlc) | HIGH | Well-established, stdlib-aligned, production-proven |
| Supabase patterns | HIGH | Official docs, common patterns verified |
| Astro 5 | HIGH | Current stable version, official docs |
| Nanostores | HIGH | Astro-recommended, lightweight |
| iOS SwiftUI | HIGH | Apple's direction, industry adoption |
| Web SDK bundling | HIGH | Standard library development pattern |
| CloudFlare D1 | MEDIUM | GA but relatively new, patterns evolving |

---

## Sources

### Go Backend
- [LogRocket: Go Frameworks 2025](https://blog.logrocket.com/top-go-frameworks-2025/)
- [Chi GitHub](https://github.com/go-chi/chi)
- [Go Blog: slog](https://go.dev/blog/slog)
- [sqlc GitHub](https://github.com/sqlc-dev/sqlc)
- [supabase-go GitHub](https://github.com/supabase-community/supabase-go)

### Frontend
- [Astro 5.0 Release](https://astro.build/blog/astro-5/)
- [Astro View Transitions](https://docs.astro.build/en/guides/view-transitions/)
- [Nanostores GitHub](https://github.com/nanostores/nanostores)

### Native SDKs
- [Apple: Binary Frameworks Distribution](https://developer.apple.com/documentation/xcode/distributing-binary-frameworks-as-swift-packages)
- [SwiftUI vs UIKit 2025](https://www.alimertgulec.com/en/blog/swiftui-vs-uikit-2025)
- [This Dot: JS Build Tools 2025](https://www.thisdot.co/blog/the-2025-guide-to-js-build-tools)

### Infrastructure
- [CloudFlare D1 Docs](https://developers.cloudflare.com/d1/)
- [Supabase RLS Best Practices](https://supabase.com/docs/guides/troubleshooting/rls-performance-and-best-practices-Z5Jjwv)
- [Cloud Run Go Quickstart](https://docs.cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-go-service)
- [Resend Go SDK](https://resend.com/docs/send-with-go)
