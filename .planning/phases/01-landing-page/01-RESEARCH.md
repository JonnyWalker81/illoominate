# Phase 1: Landing Page - Research

**Researched:** 2026-01-22
**Domain:** Cloudflare Full-Stack (Pages/Workers, D1, Resend) + Astro Frontend + Frontend-Design Skill
**Confidence:** HIGH

## Summary

This phase builds a landing page for Illoominate on the Cloudflare stack with Astro as the frontend framework. The research confirms that Cloudflare acquired Astro in January 2026, making this combination the optimal choice for content-driven sites on Cloudflare infrastructure.

A key addition to this research is the `/frontend-design` skill for Claude Code, which creates distinctive, production-grade frontend interfaces with high design quality. This skill is specifically designed to avoid generic "AI slop" aesthetics and produce polished, memorable designs that align with the Linear/Vercel aesthetic specified in CONTEXT.md.

The standard stack is: **Astro 6** for the frontend with Tailwind CSS v4 for styling, **Cloudflare Workers** for API endpoints, **D1** for the waitlist database, **Drizzle ORM** for type-safe database access, **Hono** for API routing, and **Resend** with **React Email** for transactional emails.

**Primary recommendation:** Use the `/frontend-design` skill when creating UI components to ensure high-quality, distinctive design. Provide clear context about the Linear/Vercel aesthetic, dark mode default, and cool color palette to guide the skill toward the desired output.

## Leveraging the /frontend-design Skill

The `/frontend-design` skill is a Claude Code capability that produces distinctive, production-grade frontend interfaces. This section documents how to best leverage it for the Illoominate landing page.

### What the Skill Does

The skill creates distinctive, production-grade frontend interfaces by:
- Loading specialized design guidance dynamically when frontend tasks are detected
- Steering Claude away from "distributional convergence" (the tendency toward generic choices)
- Targeting four primary design vectors: **typography**, **color/theme**, **motion**, and **backgrounds**
- Producing code with exceptional attention to aesthetic details and creative choices

### Context to Provide for Optimal Results

Based on CONTEXT.md decisions, provide this context when invoking the skill:

**Required Context Block:**
```
Purpose: Landing page for Illoominate, a native-first user feedback platform
Target: Indie developers and startups
Aesthetic: Linear meets Vercel - clean, dark, developer-focused but approachable

Visual Identity (from CONTEXT.md):
- Dark mode default (dark backgrounds, light text)
- Cool color palette (blue/purple) for primary accents
- Geometric/modern typography - clean sans-serif like Geist
- Minimal/abstract backgrounds with product screenshots, icons, diagrams
- Linear/Vercel design inspiration

Tone: Friendly professional - warm but competent, approachable startup vibe
Differentiation: Native-first, indie-friendly alternative to Canny
```

**Aesthetic Direction (pick one extreme):**
For Illoominate, the appropriate extreme is **"luxury/refined"** or **"minimalist"** with a technical edge:
- Monochrome with few bold accents (per Linear 2025 redesign)
- Terminal-inspired precision
- Zero visual noise, maximum focus

### Typography Guidance for This Project

The skill emphasizes avoiding generic fonts. For Illoominate:

| Avoid | Use Instead | Rationale |
|-------|-------------|-----------|
| Inter | **Geist** | Vercel's font, 90% similar to Inter but more rounded, better at small sizes |
| Arial, Helvetica | **Geist Sans** | Modern, geometric, crafted for web |
| System monospace | **Geist Mono** | Perfect for code/terminal aesthetic |

**Font Loading:**
```css
/* Load from Google Fonts (Geist now available) */
@import url('https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700&display=swap');

@theme {
  --font-sans: 'Geist', system-ui, sans-serif;
}
```

**Weight/Size Guidance:**
- Use extreme weight contrasts: 400 vs 700 (not 400 vs 500)
- Use 3x+ size jumps between headings and body
- Headlines: 700 weight, large sizes
- Body: 400 weight, comfortable reading size

### Animation Guidance for This Project

The skill prioritizes high-impact, CSS-only animations:

**Page Load (Priority 1):**
```css
/* Staggered reveal on page load */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.hero-headline { animation: fadeInUp 0.6s ease-out; }
.hero-tagline { animation: fadeInUp 0.6s ease-out 0.1s both; }
.hero-cta { animation: fadeInUp 0.6s ease-out 0.2s both; }
```

**Hover States (Priority 2):**
```css
/* Subtle scale + glow on CTA */
.cta-button {
  transition: transform 0.2s, box-shadow 0.2s;
}
.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 0 30px rgba(99, 102, 241, 0.3);
}
```

**Motion Reduce Support:**
```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### Background/Atmosphere Guidance

Move beyond solid colors. For Linear/Vercel aesthetic:

**Option 1: Subtle Gradient with Noise**
```css
.hero-background {
  background:
    linear-gradient(135deg, #0a0a0a 0%, #1a1a2e 50%, #0a0a0a 100%);
}

/* Add noise texture via SVG filter */
.noise-overlay {
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E");
  opacity: 0.03;
  mix-blend-mode: overlay;
}
```

**Option 2: Mesh Gradient (for accent sections)**
```css
.features-background {
  background:
    radial-gradient(ellipse at 20% 50%, rgba(99, 102, 241, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 50%, rgba(139, 92, 246, 0.1) 0%, transparent 50%),
    #0a0a0a;
}
```

### Color Palette for Illoominate

Based on CONTEXT.md (cool blue/purple):

```css
@theme {
  /* Base (dark mode default) */
  --color-bg-primary: #0a0a0a;
  --color-bg-secondary: #111111;
  --color-bg-tertiary: #1a1a1a;

  /* Text */
  --color-text-primary: #ffffff;
  --color-text-secondary: #a1a1aa;
  --color-text-muted: #71717a;

  /* Accent (blue/purple) */
  --color-primary-500: #6366f1;  /* Indigo */
  --color-primary-600: #4f46e5;
  --color-primary-400: #818cf8;

  /* Accent secondary (purple) */
  --color-accent-500: #8b5cf6;   /* Violet */
  --color-accent-400: #a78bfa;

  /* Status */
  --color-success: #10b981;
  --color-error: #ef4444;
}
```

### Prompt Patterns That Work

When asking Claude to create components, use these patterns:

**Good Prompt (for Hero):**
```
Create the hero section for Illoominate's landing page.

Context:
- Native-first user feedback platform for indie devs
- Aesthetic: Linear meets Vercel (dark, minimal, developer-focused)
- Dark mode default, cool blue/purple accents
- Typography: Geist font family

Requirements:
- Headline + tagline + single CTA
- No product visuals in hero (per CONTEXT.md)
- Staggered fade-in animation on load
- Subtle gradient background with noise texture

Avoid: Inter font, purple gradients on white, generic AI aesthetics
```

**Good Prompt (for Feature Cards):**
```
Create a 3-4 feature card grid section.

Context:
- Illoominate - native SDKs, dashboard, public roadmaps
- Aesthetic: Linear meets Vercel
- Dark mode, blue/purple accents

Requirements:
- Icon + title + brief description per card
- Subtle hover effect
- Grid layout (responsive: 1 col mobile, 3-4 col desktop)
- Icons should be minimal line-style (not filled)

Features to highlight:
1. Native SDKs (iOS, Web)
2. Dashboard + organization (boards, statuses, voting)
3. Public roadmap and transparency

Avoid: Heavy shadows, gradient cards, generic SaaS look
```

### What the Skill Will NOT Do Well

Areas where manual refinement may be needed:

1. **Exact pixel perfection** - May need spacing adjustments
2. **Complex responsive breakpoints** - Review mobile behavior
3. **Accessibility edge cases** - Verify contrast ratios, focus states
4. **Performance optimization** - Check font loading, image formats

### Integration with Planning

When creating PLAN.md tasks that involve UI:

```markdown
## Task: Create Hero Section

### Actions
1. Invoke /frontend-design skill with context:
   - Purpose: Landing page hero for feedback platform
   - Aesthetic: Linear/Vercel dark minimal
   - Requirements: Headline, tagline, CTA, no product visuals

2. Create `src/components/Hero.astro`
   - Use Geist font (Google Fonts)
   - Implement staggered fade-in animation
   - Add subtle gradient + noise background

3. Verify:
   - [ ] No Inter/Roboto fonts
   - [ ] Animations respect prefers-reduced-motion
   - [ ] Contrast ratio meets WCAG AA
   - [ ] Mobile responsive
```

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
| Geist Font | latest | Typography | Primary font for Linear/Vercel aesthetic |
| @astrojs/cloudflare | latest | Cloudflare adapter | Required for SSR/API routes on Cloudflare |
| wrangler | latest | Cloudflare CLI | Local dev, deployment, secrets management |
| Zod | latest | Validation | Form/API input validation, Hono middleware integration |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Geist | Inter | Inter is more ubiquitous but less distinctive |
| Drizzle ORM | Prisma | Prisma lacks D1 batch operations support |
| Hono | Native Workers | Hono provides better routing, middleware, type safety |

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
├── components/         # Astro components
│   ├── Hero.astro
│   ├── Features.astro
│   ├── WaitlistForm.astro
│   ├── SuccessModal.astro
│   └── Quiz.astro      # Post-signup quiz
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
<html lang="en" class="dark">
  <head>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700&display=swap" rel="stylesheet">
  </head>
  <body class="bg-[#0a0a0a] text-white font-sans">
    <Hero />
    <Features />
    <WaitlistForm />
  </body>
</html>
```

### Pattern 2: Waitlist Form with Success State
**What:** Form that shows position after signup, presents optional quiz
**When to use:** Waitlist signup flow
**Example:**
```astro
---
// src/components/WaitlistForm.astro
---
<section id="waitlist" class="py-24">
  <div class="max-w-md mx-auto px-6">
    <form id="waitlist-form" class="space-y-4">
      <input type="email" name="email" required
             placeholder="you@company.com"
             class="w-full bg-[#111] border border-[#333] rounded-lg px-4 py-3
                    text-white placeholder-gray-500 focus:border-indigo-500
                    focus:ring-1 focus:ring-indigo-500 transition-colors" />
      <input type="text" name="name" placeholder="Name (optional)"
             class="w-full bg-[#111] border border-[#333] rounded-lg px-4 py-3
                    text-white placeholder-gray-500" />
      <button type="submit"
              class="w-full bg-indigo-600 hover:bg-indigo-500 text-white
                     font-medium py-3 rounded-lg transition-all
                     hover:translate-y-[-2px] hover:shadow-[0_0_30px_rgba(99,102,241,0.3)]">
        Join Waitlist
      </button>
    </form>

    <!-- Success state (hidden by default) -->
    <div id="success-state" class="hidden text-center">
      <div class="text-6xl font-bold text-indigo-400 mb-2">
        #<span id="position">--</span>
      </div>
      <p class="text-gray-400 mb-6">You're on the list!</p>
      <p class="text-sm text-gray-500 mb-4">
        Share your code to move up:
        <code id="referral-code" class="text-indigo-400 font-mono">------</code>
      </p>
      <button id="take-quiz" class="text-indigo-400 hover:text-indigo-300 text-sm">
        Help us build better (optional survey)
      </button>
    </div>
  </div>
</section>

<script>
  const form = document.getElementById('waitlist-form');
  const successState = document.getElementById('success-state');

  form?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);

    const response = await fetch('/api/waitlist', {
      method: 'POST',
      body: formData,
    });
    const data = await response.json();

    if (data.success) {
      form.classList.add('hidden');
      successState?.classList.remove('hidden');
      document.getElementById('position')!.textContent = data.position;
      document.getElementById('referral-code')!.textContent = data.referralCode;
    }
  });
</script>
```

### Pattern 3: Hono API Routes in Astro
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
  referredBy: z.string().optional(),
});

app.post('/', zValidator('form', waitlistSchema), async (c) => {
  const { email, name, source, referredBy } = c.req.valid('form');
  const db = drizzle(c.env.DB);
  const resend = new Resend(c.env.RESEND_API_KEY);

  // Generate referral code (6 chars, alphanumeric)
  const referralCode = generateReferralCode();

  // Insert into D1
  const [entry] = await db.insert(waitlist).values({
    email,
    name,
    source,
    referralCode,
    referredBy,
    createdAt: new Date().toISOString(),
  }).returning();

  // If referred, increment referrer's count
  if (referredBy) {
    await db.update(waitlist)
      .set({ referralCount: sql`referral_count + 1` })
      .where(eq(waitlist.referralCode, referredBy));
  }

  // Get position (ordered by referrals DESC, then created_at ASC)
  const position = await getWaitlistPosition(db, entry.id);

  // Send confirmation email
  await resend.emails.send({
    from: 'Illoominate <hello@illoominate.com>',
    to: email,
    subject: `You're #${position} on the Illoominate waitlist!`,
    react: WelcomeEmail({ name, position, referralCode }),
  });

  return c.json({ success: true, position, referralCode });
});

function generateReferralCode(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'; // Exclude confusing chars
  let code = '';
  for (let i = 0; i < 6; i++) {
    code += chars[Math.floor(Math.random() * chars.length)];
  }
  return code;
}

export const POST = app.fetch;
```

### Pattern 4: D1 with Drizzle Schema
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

### Pattern 5: React Email Templates
**What:** Beautiful, cross-client compatible email templates
**When to use:** All transactional emails
**Example:**
```tsx
// src/emails/WelcomeEmail.tsx
import {
  Body, Container, Head, Heading, Html,
  Preview, Section, Text, Hr
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
      <Body style={{
        backgroundColor: '#0a0a0a',
        color: '#ffffff',
        fontFamily: 'system-ui, sans-serif'
      }}>
        <Container style={{ padding: '40px 20px', maxWidth: '480px' }}>
          <Heading style={{
            fontSize: '24px',
            fontWeight: '600',
            marginBottom: '16px'
          }}>
            Welcome{name ? `, ${name}` : ''}!
          </Heading>

          <Text style={{ color: '#a1a1aa', lineHeight: '1.6' }}>
            You're #{position} on the Illoominate waitlist. We're building
            native-first user feedback for indie developers and startups.
          </Text>

          <Hr style={{ borderColor: '#333', margin: '24px 0' }} />

          <Section style={{ textAlign: 'center' }}>
            <Text style={{ color: '#a1a1aa', fontSize: '14px', marginBottom: '8px' }}>
              Move up the list! Share your referral code:
            </Text>
            <Text style={{
              fontSize: '32px',
              fontWeight: 'bold',
              fontFamily: 'monospace',
              color: '#818cf8',
              letterSpacing: '0.1em'
            }}>
              {referralCode}
            </Text>
            <Text style={{ color: '#71717a', fontSize: '12px' }}>
              Each friend who joins moves you up the list
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
}
```

### Anti-Patterns to Avoid
- **Generic AI aesthetics:** Using Inter, purple gradients on white, predictable layouts
- **Hand-rolling form validation:** Use Zod with Hono's zValidator middleware
- **Separate API server:** Keep API routes in Astro's `/api` directory
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
| Dark mode toggle | Custom localStorage logic | CSS class + inline script | SSR flash prevention |
| Referral codes | Sequential IDs | Random alphanumeric (6 chars) | Non-guessable, URL-safe, short |
| Admin auth | Session cookies from scratch | Bearer token with wrangler secrets | Simple for admin-only, no user auth needed |
| UI design direction | Guessing aesthetic | /frontend-design skill | Distinctive, production-grade output |

**Key insight:** The Cloudflare + Astro ecosystem has mature solutions for every common pattern. The /frontend-design skill elevates UI quality beyond typical AI outputs.

## Common Pitfalls

### Pitfall 1: Generic AI Design (AI Slop)
**What goes wrong:** UI looks like every other AI-generated site (Inter font, purple gradient, white background)
**Why it happens:** LLMs converge to most common patterns in training data
**How to avoid:** Use /frontend-design skill with explicit context about Linear/Vercel aesthetic
**Warning signs:** Purple gradients on white, Inter or Roboto fonts, generic SaaS layout

### Pitfall 2: Using Cloudflare Pages Instead of Workers
**What goes wrong:** Pages is in maintenance mode; new features go to Workers only
**Why it happens:** Pages was the previous recommendation; outdated tutorials still suggest it
**How to avoid:** Use Workers with `@astrojs/cloudflare` adapter, configure wrangler.jsonc correctly
**Warning signs:** Following tutorials that use `pages_build_output_dir` instead of `assets.directory`

### Pitfall 3: D1 Single-Thread Bottleneck
**What goes wrong:** Database becomes unresponsive under load
**Why it happens:** D1 is single-threaded; each query blocks the next
**How to avoid:** Keep queries fast (<10ms), use batch operations, avoid locking operations
**Warning signs:** "overloaded" errors, queries taking >100ms

### Pitfall 4: Forgetting prerender = false for API Routes
**What goes wrong:** API routes return 404 or static content
**Why it happens:** Astro defaults to static generation
**How to avoid:** Add `export const prerender = false;` to every API route
**Warning signs:** API returns HTML instead of JSON, POST methods don't work

### Pitfall 5: Hydration Mismatch with Dark Mode
**What goes wrong:** Flash of wrong theme, hydration errors
**Why it happens:** Server doesn't know client's theme preference
**How to avoid:** Dark mode by default (`class="dark"` on `<html>`), no JS toggle needed for this phase
**Warning signs:** White flash on dark mode pages, console hydration warnings

### Pitfall 6: Font Loading Flash
**What goes wrong:** Text renders in fallback font, then jumps when Geist loads
**Why it happens:** Web fonts load after HTML
**How to avoid:** Use `font-display: swap`, preconnect to Google Fonts
**Warning signs:** Layout shift on page load

## Code Examples

Verified patterns from official sources:

### Wrangler Configuration for Astro + Workers
```jsonc
// wrangler.jsonc
{
  "name": "illoominate-landing",
  "main": "dist/_worker.js/index.js",
  "compatibility_date": "2026-01-21",
  "compatibility_flags": ["nodejs_compat"],
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
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  output: 'hybrid', // Static by default, opt-in to SSR
  adapter: cloudflare(),
  vite: {
    plugins: [tailwindcss()],
  },
});
```
Source: [Astro Cloudflare docs](https://docs.astro.build/en/guides/deploy/cloudflare/)

### Tailwind v4 Dark Mode Setup with Geist
```css
/* src/styles/global.css */
@import 'tailwindcss';

@theme {
  --font-sans: 'Geist', system-ui, sans-serif;

  /* Cool color palette (blue/purple) */
  --color-primary-500: #6366f1;
  --color-primary-600: #4f46e5;
  --color-primary-400: #818cf8;

  --color-accent-500: #8b5cf6;
  --color-accent-400: #a78bfa;
}

/* Dark mode by default */
@custom-variant dark (&:where(.dark, .dark *));
```

```html
<!-- In base layout - dark class by default, preload fonts -->
<!DOCTYPE html>
<html lang="en" class="dark">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700&display=swap" rel="stylesheet" />
  </head>
  <body class="bg-[#0a0a0a] text-white antialiased">
    <slot />
  </body>
</html>
```
Source: [Tailwind CSS docs](https://tailwindcss.com/docs/dark-mode)

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

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Generic AI UI | /frontend-design skill | 2025-2026 | Distinctive, production-grade designs |
| Inter font | Geist font | 2024 | Better at small sizes, more refined |
| Cloudflare Pages | Cloudflare Workers | 2025-2026 | Workers is now recommended deployment target |
| Tailwind v3 config | Tailwind v4 CSS-first | 2024-2025 | Native CSS variables, simpler config |
| Astro independent | Astro owned by Cloudflare | Jan 2026 | First-class Cloudflare integration |
| Prisma with D1 | Drizzle with D1 | 2024 | Drizzle supports batch operations |

**Deprecated/outdated:**
- **Cloudflare Pages as primary target:** Use Workers instead; Pages is maintenance-only
- **Inter font for developer tools:** Geist is now the modern choice
- **MailChannels for email:** Use Resend; MailChannels integration deprecated
- **@astrojs/tailwind integration:** Use Tailwind v4 Vite plugin directly

## Waitlist Landing Page Design Patterns

Based on high-converting waitlist pages:

### Page Structure (from CONTEXT.md)
1. **Hero** - Headline + tagline + single CTA (no product visuals)
2. **Features** - Icon cards grid (3-4 features)
3. **Waitlist Form** - Email (required), name (optional), source (optional)

### Conversion Best Practices

| Element | Best Practice | Rationale |
|---------|--------------|-----------|
| CTA Text | "Join the Waitlist" not "Get Early Access" | Sets correct expectation |
| Position Display | Show exact number (#47) | Creates engagement, urgency |
| Referral Code | 6-char alphanumeric code | Easy to share, non-URL |
| Post-Signup | Show position, then offer quiz | Confirms success, gathers data |
| Quiz | Entirely optional, single page | Lowest friction |

### Form Implementation
- Single column layout
- Clear field labels
- Immediate validation feedback
- Submit button spans full width
- Loading state during submission
- Success state replaces form (don't navigate)

## Open Questions

Things that couldn't be fully resolved:

1. **Geist Font Availability on Google Fonts**
   - What we know: Geist was recently added to Google Fonts
   - What's unclear: All weights availability, variable font support
   - Recommendation: Test with Google Fonts; fall back to Vercel CDN if needed

2. **Optimal Quiz Implementation**
   - What we know: Quiz should be optional, single-page, after signup
   - What's unclear: Modal vs. inline expansion vs. separate page
   - Recommendation: Start with modal/overlay; iterate based on completion rates

3. **Referral Position Calculation**
   - What we know: Position should favor referrers
   - What's unclear: Exact algorithm (referrals * X + signup order?)
   - Recommendation: Start simple (referral_count DESC, created_at ASC)

## Sources

### Primary (HIGH confidence)
- [Cloudflare Workers + Resend Tutorial](https://developers.cloudflare.com/workers/tutorials/send-emails-with-resend/) - Email integration
- [Cloudflare D1 Getting Started](https://developers.cloudflare.com/d1/get-started/) - Database setup
- [Astro Cloudflare Deployment](https://docs.astro.build/en/guides/deploy/cloudflare/) - Framework integration
- [Drizzle D1 Documentation](https://orm.drizzle.team/docs/connect-cloudflare-d1) - ORM setup
- [Tailwind Dark Mode](https://tailwindcss.com/docs/dark-mode) - Styling setup
- [Claude Frontend-Design Skill](https://github.com/anthropics/claude-code/blob/main/plugins/frontend-design/skills/frontend-design/SKILL.md) - UI generation
- [Improving Frontend Design Through Skills](https://claude.com/blog/improving-frontend-design-through-skills) - Skill usage patterns

### Secondary (MEDIUM confidence)
- [Vercel Geist Font](https://vercel.com/font) - Typography reference
- [Geist on Google Fonts](https://fonts.google.com/specimen/Geist) - Font loading
- [Linear Design Trend](https://blog.logrocket.com/ux-design/linear-design/) - Aesthetic reference
- [Waitlist Landing Page Best Practices](https://moosend.com/blog/waitlist-landing-page/) - Conversion patterns
- [Grainy Gradients](https://css-tricks.com/grainy-gradients/) - Background effects

### Tertiary (LOW confidence)
- Various WebSearch results on referral systems - need validation with implementation
- Community examples of Linear/Vercel aesthetics - subjective

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Official Cloudflare tutorials, official Astro docs, verified integrations
- Architecture: HIGH - Based on official documentation patterns
- Frontend-design skill: HIGH - Official Claude documentation, verified patterns
- Pitfalls: MEDIUM - Mix of official warnings and community experience
- Design aesthetic: MEDIUM - Based on public references to Linear/Vercel, subjective

**Research date:** 2026-01-22
**Valid until:** 2026-02-22 (30 days - Cloudflare stack is stable)
