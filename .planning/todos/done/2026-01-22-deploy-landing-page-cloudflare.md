created: 2026-01-22T00:00
title: Deploy landing page to CloudFlare with Resend and D1
area: deployment
files:
  - src/pages/index.astro
  - src/pages/api/waitlist.ts
  - wrangler.toml

## Problem

Phase 1 landing page is complete but not yet deployed to production. Need to:
- Deploy Astro site to CloudFlare Pages
- Configure Resend for email sending in production
- Set up D1 database for waitlist storage
- Connect all services together

## Solution

1. Configure wrangler.toml for production D1 database
2. Set up Resend API key in CloudFlare environment variables
3. Deploy to CloudFlare Pages (either via CLI or GitHub integration)
4. Verify email sending and database writes work in production
5. Set up custom domain if needed
