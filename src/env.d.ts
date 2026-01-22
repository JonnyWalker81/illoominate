/// <reference types="astro/client" />

type D1Database = import('@cloudflare/workers-types').D1Database;

interface Env {
  DB: D1Database;
  RESEND_API_KEY: string;
  ADMIN_TOKEN: string;
}

type Runtime = import('@astrojs/cloudflare').Runtime<Env>;

declare namespace App {
  interface Locals extends Runtime {}
}
