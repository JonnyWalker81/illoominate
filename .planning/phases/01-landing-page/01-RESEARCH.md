# Phase 1: Landing Page - Research

**Researched:** 2026-01-21
**Domain:** Cloudflare Full-Stack (Pages/Workers, D1, Resend) + Astro Frontend
**Confidence:** HIGH

## Summary

This phase builds a landing page for Illoominate on the Cloudflare stack with Astro as the frontend framework. The research confirms that Cloudflare acquired Astro in January 2026, making this combination the optimal choice for content-driven sites on Cloudflare infrastructure.

The standard stack is: **Astro 6** for the frontend with Tailwind CSS v4 for styling, **Cloudflare Workers** for API endpoints (recommended over Pages for new projects), **D1** for the waitlist database, **Drizzle ORM** for type-safe database access, **Hono** for API routing, and **Resend** with **React Email** for transactional emails.

**Primary recommendation:** Use Cloudflare Workers (not Pages) as the deployment target since Cloudflare is investing all new features in Workers. Use Astro's hybrid rendering mode with static HTML for the landing page and on-demand rendering for API routes.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Astro | 6.x | Frontend framework | Cloudflare-acquired, content-driven sites, Islands Architecture |
| Tailwind CSS | v4 | Styling | Native CSS variables, dark mode support, industry standard |
| Cloudflare Workers | - | Serverless compute | Recommended over Pages for new projects, full-stack support |
| Cloudflare D1 | - | SQLite database | Serverless, edge-native, 10GB per database |
| Hono | 4.x | API routing | Cloudflare's internal choice, used by D1/KV/Queues internally |
| Drizzle ORM | latest | Database ORM | Type-safe, supports D1, better batch operations than Prisma |
| Resend | latest | Email API | Official Cloudflare tutorial integration, React Email support |
| React Email | latest | Email templates | Beautiful cross-client emails, React components |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Starwind UI | latest | Astro components | Pre-built accessible components, Tailwind v4 native |
| @astrojs/cloudflare | latest | Cloudflare adapter | Required for SSR/API routes on Cloudflare |
| wrangler | latest | Cloudflare CLI | Local dev, deployment, secrets management |
| Zod | latest | Validation | Form/API input validation, Hono middleware integration |
| Inter/Geist | - | Typography | Modern geometric sans-serif, developer-focused aesthetic |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Drizzle ORM | Prisma | Prisma lacks D1 batch operations support |
| Hono | Native Workers | Hono provides better routing, middleware, type safety |
| Starwind UI | shadcn/ui | shadcn requires React; Starwind is native Astro |
| React Email | MJML | React Email integrates seamlessly with Resend |

**Installation:**
```bash
# Create Astro project with Cloudflare
npm create cloudflare@latest illoominate-landing -- --framework=astro

# Install dependencies
npm install hono drizzle-orm resend @react-email/components zod
npm install -D drizzle-kit wrangler

# Add Cloudflare adapter
npx astro add cloudflare
```

## Architecture Patterns

### Recommended Project Structure
```
src/
├── pages/              # Astro pages (static + API routes)
│   ├── index.astro     # Landing page (static)
│   ├── admin/          # Admin dashboard (SSR, protected)
│   │   └── index.astro
│   └── api/            # API endpoints (Hono)
│       ├── waitlist.ts # POST /api/waitlist
│       └── admin/      # Admin API routes
├── components/         # Astro/React components
│   ├── Hero.astro
│   ├── Features.astro
│   ├── WaitlistForm.astro
│   └── islands/        # Interactive React islands
│       └── Quiz.tsx
├── emails/             # React Email templates
│   ├── WelcomeEmail.tsx
│   └── ReferralEmail.tsx
├── db/                 # Database layer
│   ├── schema.ts       # Drizzle schema
│   └── index.ts        # DB client setup
├── lib/                # Shared utilities
│   ├── referral.ts     # Referral code generation
│   └── validation.ts   # Zod schemas
└── styles/             # Global styles
    └── global.css      # Tailwind imports, CSS variables
```

### Pattern 1: Astro Islands for Interactive Components
**What:** Static HTML by default, hydrate only interactive components
**When to use:** Forms, quizzes, any user interaction
**Example:**
```astro
---
// src/pages/index.astro - Landing page (static)
import Hero from '../components/Hero.astro';
import Features from '../components/Features.astro';
import WaitlistForm from '../components/WaitlistForm.astro';
---
<html>
  <body class="dark bg-gray-950 text-white">
    <Hero />
    <Features />
    <WaitlistForm />
  </body>
</html>
```

```astro
---
// src/components/WaitlistForm.astro
// Form is static HTML, JavaScript handles submission
---
<form id="waitlist-form" class="space-y-4">
  <input type="email" name="email" required
         class="bg-gray-900 border-gray-700" />
  <input type="text" name="name" placeholder="Name (optional)"
         class="bg-gray-900 border-gray-700" />
  <button type="submit">Join Waitlist</button>
</form>

<script>
  const form = document.getElementById('waitlist-form');
  form?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const response = await fetch('/api/waitlist', {
      method: 'POST',
      body: formData,
    });
    const data = await response.json();
    // Show success/quiz modal
  });
</script>
```

### Pattern 2: Hono API Routes in Astro
**What:** Use Hono for type-safe API routing within Astro
**When to use:** All API endpoints
**Example:**
```typescript
// src/pages/api/waitlist.ts
export const prerender = false;
import { Hono } from 'hono';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';
import { drizzle } from 'drizzle-orm/d1';
import { waitlist } from '../../db/schema';
import { Resend } from 'resend';
import WelcomeEmail from '../../emails/WelcomeEmail';

const app = new Hono<{ Bindings: { DB: D1Database; RESEND_API_KEY: string } }>();

const waitlistSchema = z.object({
  email: z.string().email(),
  name: z.string().optional(),
  source: z.string().optional(),
});

app.post('/', zValidator('form', waitlistSchema), async (c) => {
  const { email, name, source } = c.req.valid('form');
  const db = drizzle(c.env.DB);
  const resend = new Resend(c.env.RESEND_API_KEY);

  // Generate referral code
  const referralCode = generateReferralCode();

  // Insert into D1
  const [entry] = await db.insert(waitlist).values({
    email,
    name,
    source,
    referralCode,
    createdAt: new Date().toISOString(),
  }).returning();

  // Get position
  const position = await db.select({ count: count() })
    .from(waitlist)
    .where(lte(waitlist.id, entry.id));

  // Send confirmation email
  await resend.emails.send({
    from: 'Illoominate <hello@illoominate.com>',
    to: email,
    subject: 'Welcome to the Illoominate Waitlist!',
    react: WelcomeEmail({ name, position: position[0].count, referralCode }),
  });

  return c.json({ success: true, position: position[0].count, referralCode });
});

export const POST = app.fetch;
```

### Pattern 3: D1 with Drizzle Schema
**What:** Type-safe database schema and queries
**When to use:** All database operations
**Example:**
```typescript
// src/db/schema.ts
import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';

export const waitlist = sqliteTable('waitlist', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  email: text('email').notNull().unique(),
  name: text('name'),
  source: text('source'), // "how did you hear about us"
  referralCode: text('referral_code').notNull().unique(),
  referredBy: text('referred_by'), // referral code of referrer
  referralCount: integer('referral_count').default(0),
  createdAt: text('created_at').notNull(),
});

export const quizResponses = sqliteTable('quiz_responses', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  waitlistId: integer('waitlist_id').references(() => waitlist.id),
  platform: text('platform'), // iOS, Android, Web
  teamSize: text('team_size'),
  painPoints: text('pain_points'),
  createdAt: text('created_at').notNull(),
});
```

### Pattern 4: React Email Templates
**What:** Beautiful, cross-client compatible email templates
**When to use:** All transactional emails
**Example:**
```tsx
// src/emails/WelcomeEmail.tsx
import {
  Body, Container, Head, Heading, Html,
  Preview, Section, Text, Button
} from '@react-email/components';

interface WelcomeEmailProps {
  name?: string;
  position: number;
  referralCode: string;
}

export default function WelcomeEmail({ name, position, referralCode }: WelcomeEmailProps) {
  return (
    <Html>
      <Head />
      <Preview>You're #{position} on the Illoominate waitlist!</Preview>
      <Body style={{ backgroundColor: '#0a0a0a', color: '#ffffff' }}>
        <Container>
          <Heading>Welcome to Illoominate{name ? `, ${name}` : ''}!</Heading>
          <Text>You're #{position} on our waitlist.</Text>
          <Section>
            <Text>Share your referral code to move up the list:</Text>
            <Text style={{ fontSize: '24px', fontWeight: 'bold' }}>
              {referralCode}
            </Text>
          </Section>
          <Button href="https://illoominate.com"
                  style={{ backgroundColor: '#6366f1', color: '#ffffff' }}>
            Visit Illoominate
          </Button>
        </Container>
      </Body>
    </Html>
  );
}
```

### Anti-Patterns to Avoid
- **Hand-rolling form validation:** Use Zod with Hono's zValidator middleware
- **Separate API server:** Keep API routes in Astro's `/api` directory, not a separate Worker
- **Global React hydration:** Use Astro Islands, not client:load on everything
- **Raw SQL queries:** Use Drizzle ORM for type safety and SQL injection prevention
- **Storing secrets in code:** Use wrangler secrets and .dev.vars for local development

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Email templates | HTML strings with inline styles | React Email components | Cross-client compatibility, dark mode, responsive |
| Form validation | Manual if/else checks | Zod + zValidator | Type inference, error messages, schema reuse |
| Database queries | Template string SQL | Drizzle ORM | SQL injection prevention, TypeScript types |
| API routing | if/else on request.url | Hono | Middleware, validation, type safety |
| Dark mode toggle | Custom localStorage logic | Astro's theme pattern or Starwind | SSR flash prevention, system preference |
| Referral codes | Sequential IDs | nanoid or custom base62 | Non-guessable, URL-safe, short |
| Admin auth | Session cookies from scratch | Bearer token with wrangler secrets | Simple for admin-only, no user auth needed |

**Key insight:** The Cloudflare + Astro ecosystem has mature solutions for every common pattern. Custom solutions add maintenance burden without benefit.

## Common Pitfalls

### Pitfall 1: Using Cloudflare Pages Instead of Workers
**What goes wrong:** Pages is in maintenance mode; new features go to Workers only
**Why it happens:** Pages was the previous recommendation; outdated tutorials still suggest it
**How to avoid:** Use Workers with `@astrojs/cloudflare` adapter, configure wrangler.jsonc correctly
**Warning signs:** Following tutorials that use `pages_build_output_dir` instead of `assets.directory`

### Pitfall 2: D1 Single-Thread Bottleneck
**What goes wrong:** Database becomes unresponsive under load
**Why it happens:** D1 is single-threaded; each query blocks the next
**How to avoid:** Keep queries fast (<10ms), use batch operations, avoid locking operations
**Warning signs:** "overloaded" errors, queries taking >100ms

### Pitfall 3: Cold Start Misconceptions
**What goes wrong:** First requests are slower than expected (200-500ms)
**Why it happens:** While Workers have ~5ms cold starts, D1/KV calls hit network on first request
**How to avoid:** Accept first-request latency; subsequent requests are fast; consider prewarming
**Warning signs:** Inconsistent response times in testing

### Pitfall 4: Forgetting prerender = false for API Routes
**What goes wrong:** API routes return 404 or static content
**Why it happens:** Astro defaults to static generation
**How to avoid:** Add `export const prerender = false;` to every API route
**Warning signs:** API returns HTML instead of JSON, POST methods don't work

### Pitfall 5: JavaScript 52-bit Number Precision in D1
**What goes wrong:** Large integer IDs become incorrect
**Why it happens:** SQLite stores int64, but JavaScript only has 52-bit precision
**How to avoid:** Use autoincrement IDs (stay small) or text UUIDs for large numbers
**Warning signs:** ID mismatches when IDs exceed 2^52

### Pitfall 6: Hydration Mismatch with Dark Mode
**What goes wrong:** Flash of wrong theme, hydration errors
**Why it happens:** Server doesn't know client's theme preference
**How to avoid:** Inline script in `<head>` to set `dark` class before render; use Astro's pattern
**Warning signs:** White flash on dark mode pages, console hydration warnings

### Pitfall 7: Missing CORS for Admin Dashboard
**What goes wrong:** Admin API calls fail from browser
**Why it happens:** Forgot to configure CORS for admin routes
**How to avoid:** Add Hono CORS middleware to admin API routes
**Warning signs:** "CORS policy" errors in browser console

## Code Examples

Verified patterns from official sources:

### Wrangler Configuration for Astro + Workers
```jsonc
// wrangler.jsonc
{
  "name": "illoominate-landing",
  "main": "dist/_worker.js/index.js",
  "compatibility_date": "2026-01-21",
  "compatibility_flags": [
    "nodejs_compat"
  ],
  "assets": {
    "binding": "ASSETS",
    "directory": "./dist",
    "not_found_handling": "404-page"
  },
  "d1_databases": [
    {
      "binding": "DB",
      "database_name": "illoominate-waitlist",
      "database_id": "<your-database-id>"
    }
  ],
  "observability": {
    "enabled": true
  }
}
```
Source: [Cloudflare Workers docs](https://developers.cloudflare.com/workers/framework-guides/web-apps/astro/)

### Astro Config with Cloudflare Adapter
```typescript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import cloudflare from '@astrojs/cloudflare';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';

export default defineConfig({
  output: 'hybrid', // Static by default, opt-in to SSR
  adapter: cloudflare(),
  integrations: [
    tailwind(),
    react(), // For islands architecture
  ],
});
```
Source: [Astro Cloudflare docs](https://docs.astro.build/en/guides/deploy/cloudflare/)

### Drizzle Kit Configuration
```typescript
// drizzle.config.ts
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  out: './drizzle',
  schema: './src/db/schema.ts',
  dialect: 'sqlite',
  driver: 'd1-http',
  dbCredentials: {
    accountId: process.env.CLOUDFLARE_ACCOUNT_ID!,
    databaseId: process.env.CLOUDFLARE_DATABASE_ID!,
    token: process.env.CLOUDFLARE_D1_TOKEN!,
  },
});
```
Source: [Drizzle D1 docs](https://orm.drizzle.team/docs/connect-cloudflare-d1)

### Local Development Environment
```bash
# .dev.vars (local secrets - DO NOT COMMIT)
RESEND_API_KEY=re_xxxxxxxxxxxx
ADMIN_TOKEN=your-secure-admin-token

# Commands
npx wrangler d1 create illoominate-waitlist  # Create D1 database
npx drizzle-kit push                          # Apply schema to local D1
npx wrangler dev                              # Start local dev server
```
Source: [Cloudflare D1 docs](https://developers.cloudflare.com/d1/get-started/)

### Tailwind v4 Dark Mode Setup
```css
/* src/styles/global.css */
@import 'tailwindcss';

@theme {
  --font-sans: 'Inter', 'Geist', system-ui, sans-serif;

  /* Cool color palette (blue/purple) */
  --color-primary-500: #6366f1;
  --color-primary-600: #4f46e5;
  --color-primary-400: #818cf8;
}

/* Dark mode by default */
@custom-variant dark (&:where(.dark, .dark *));
```

```html
<!-- In base layout - add dark class by default -->
<html lang="en" class="dark">
  <head>
    <!-- Prevent flash by checking preference before render -->
    <script is:inline>
      if (localStorage.theme === 'light') {
        document.documentElement.classList.remove('dark');
      }
    </script>
  </head>
</html>
```
Source: [Tailwind CSS docs](https://tailwindcss.com/docs/dark-mode)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Cloudflare Pages | Cloudflare Workers | 2025-2026 | Workers is now the recommended deployment target |
| Pages Functions | Workers with static assets | 2025 | Unified developer experience |
| Astro independent | Astro owned by Cloudflare | Jan 2026 | First-class Cloudflare integration |
| Prisma with D1 | Drizzle with D1 | 2024 | Drizzle supports batch operations |
| Manual email HTML | React Email | 2023-2024 | Cross-client compatible templates |
| Workers + separate frontend | Full-stack Astro on Workers | 2025 | Single deployment, simpler architecture |

**Deprecated/outdated:**
- **Cloudflare Pages as primary target:** Use Workers instead; Pages is maintenance-only
- **MailChannels for email:** Use Resend; MailChannels integration deprecated
- **wrangler pages:** Use `wrangler deploy` with Workers configuration
- **pages_build_output_dir:** Use `assets.directory` in wrangler.jsonc

## Open Questions

Things that couldn't be fully resolved:

1. **Starwind UI Maturity**
   - What we know: It's Astro-native, Tailwind v4, 50+ components
   - What's unclear: Production stability, community adoption size
   - Recommendation: Start with it; fall back to Tailwind primitives if issues arise

2. **Cloudflare Email Service vs Resend**
   - What we know: Cloudflare Email Service launched private beta Sept 2025
   - What's unclear: Whether it's in public beta/GA now, feature parity with Resend
   - Recommendation: Use Resend (proven, documented); evaluate Cloudflare Email later

3. **Admin Dashboard Authentication**
   - What we know: Bearer token auth is simple and sufficient
   - What's unclear: Whether to add proper admin auth (Auth.js) for future features
   - Recommendation: Start with Bearer token; admin is single-user for now

## Sources

### Primary (HIGH confidence)
- [Cloudflare Workers + Resend Tutorial](https://developers.cloudflare.com/workers/tutorials/send-emails-with-resend/) - Email integration
- [Cloudflare D1 Getting Started](https://developers.cloudflare.com/d1/get-started/) - Database setup
- [Astro Cloudflare Deployment](https://docs.astro.build/en/guides/deploy/cloudflare/) - Framework integration
- [Drizzle D1 Documentation](https://orm.drizzle.team/docs/connect-cloudflare-d1) - ORM setup
- [Hono Best Practices](https://hono.dev/docs/guides/best-practices) - API patterns
- [Tailwind Dark Mode](https://tailwindcss.com/docs/dark-mode) - Styling setup

### Secondary (MEDIUM confidence)
- [Cloudflare acquires Astro announcement](https://blog.cloudflare.com/astro-joins-cloudflare/) - Jan 2026 acquisition
- [Starwind UI Getting Started](https://starwind.dev/docs/getting-started/) - Astro component library
- [React Email documentation](https://react.email) - Email templates
- [Resend Cloudflare Workers docs](https://resend.com/docs/send-with-cloudflare-workers) - Email sending

### Tertiary (LOW confidence)
- Various Medium/DEV articles on waitlist patterns - community patterns only
- WebSearch results on referral systems - need validation with implementation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Official Cloudflare tutorials, official Astro docs, verified integrations
- Architecture: HIGH - Based on official documentation patterns
- Pitfalls: MEDIUM - Mix of official warnings and community experience

**Research date:** 2026-01-21
**Valid until:** 2026-02-21 (30 days - Cloudflare stack is stable)
