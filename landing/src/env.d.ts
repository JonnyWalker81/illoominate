/// <reference path="../.astro/types.d.ts" />
/// <reference types="astro/client" />

type Runtime = import('@astrojs/cloudflare').Runtime<Env>;

interface Env {
  DB: D1Database;
  TURNSTILE_SECRET_KEY: string;
  RESEND_API_KEY: string;
  RESEND_SEGMENT_ID: string;
  RESEND_WEBHOOK_SECRET: string;
  VERIFICATION_BASE_URL: string;
  ADMIN_EMAIL: string;
  FROM_EMAIL: string;
}

declare namespace App {
  interface Locals extends Runtime {}
}
