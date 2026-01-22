# Phase 2: Foundation - Research

**Researched:** 2026-01-22
**Domain:** Supabase PostgreSQL, Multi-Tenant Database Design, Row Level Security
**Confidence:** HIGH

## Summary

Phase 2 establishes the database foundation for Illoominate using Supabase PostgreSQL. The core challenge is implementing proper multi-tenant isolation where workspaces are completely isolated from each other. Supabase's Row Level Security (RLS) is the standard approach for this, with policies that check `workspace_id` on every query.

The schema must support the full product scope: workspaces, apps, users (linked to auth.users), feedback items, votes, comments, boards, and status workflows. The database design follows a workspace-centric model where nearly every table has a `workspace_id` foreign key used in RLS policies.

**Primary recommendation:** Use Supabase CLI for local development with migration files version-controlled in `supabase/migrations/`. Define all tables in the `public` schema with RLS enabled, referencing `auth.users` for user identity and using `auth.uid()` in policies.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Supabase | Latest | Managed PostgreSQL + Auth | Project constraint, excellent RLS support |
| Supabase CLI | v2.x | Local development, migrations | Official tooling for version-controlled schemas |
| PostgreSQL | 15 | Database engine | Supabase default, mature RLS implementation |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| uuid-ossp | - | UUID generation | Built-in extension, use for primary keys |
| pg_cron | - | Scheduled jobs | If background processing needed later |
| pgTAP | - | Database unit testing | For testing RLS policies |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Supabase RLS | Application-level isolation | RLS is defense-in-depth, app-level alone risks data leaks |
| UUID primary keys | BIGINT identity | UUIDs prevent enumeration attacks, better for public APIs |
| Single `profiles` table | Separate user metadata tables | Keep it simple for v1, profiles work well |

**Installation:**

```bash
# Install Supabase CLI (macOS)
brew install supabase/tap/supabase

# Initialize project (if not done)
supabase init

# Start local Supabase stack
supabase start
```

## Architecture Patterns

### Recommended Project Structure

```
supabase/
├── migrations/           # Version-controlled SQL migrations
│   ├── 20260122000000_initial_schema.sql
│   ├── 20260122000001_rls_policies.sql
│   └── ...
├── seed.sql              # Development seed data
└── config.toml           # Supabase local config
```

### Pattern 1: Workspace-Centric Multi-Tenancy

**What:** Every resource belongs to a workspace. Tables have `workspace_id` column. RLS policies filter by workspace membership.

**When to use:** SaaS applications with organization/team isolation.

**Example:**

```sql
-- Source: Supabase official docs
-- Core pattern: workspace ownership checked via membership
create table public.workspaces (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  created_at timestamptz default now()
);

create table public.workspace_members (
  id uuid primary key default gen_random_uuid(),
  workspace_id uuid references public.workspaces on delete cascade not null,
  user_id uuid references auth.users on delete cascade not null,
  role text not null default 'member',
  unique (workspace_id, user_id)
);

-- RLS policy using membership table
create policy "Users can view workspaces they belong to"
on public.workspaces for select
to authenticated
using (
  id in (
    select workspace_id from public.workspace_members
    where user_id = (select auth.uid())
  )
);
```

### Pattern 2: User Profile Trigger

**What:** Automatically create a profile row when a user signs up via auth.

**When to use:** When you need public user data separate from auth.users.

**Example:**

```sql
-- Source: Supabase official docs (Managing User Data)
create table public.profiles (
  id uuid not null references auth.users on delete cascade,
  email text,
  full_name text,
  avatar_url text,
  primary key (id)
);

alter table public.profiles enable row level security;

-- Trigger to auto-create profile
create function public.handle_new_user()
returns trigger
language plpgsql
security definer set search_path = ''
as $$
begin
  insert into public.profiles (id, email, full_name)
  values (
    new.id,
    new.email,
    new.raw_user_meta_data ->> 'full_name'
  );
  return new;
end;
$$;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute procedure public.handle_new_user();
```

### Pattern 3: Performant RLS with Wrapped Functions

**What:** Wrap `auth.uid()` and other functions in `select` to enable query plan caching.

**When to use:** Always, for significant performance improvement.

**Example:**

```sql
-- Source: Supabase RLS Performance Guide
-- BAD: Calls auth.uid() per row
create policy "bad_policy" on items
using ( auth.uid() = user_id );

-- GOOD: Caches auth.uid() result
create policy "good_policy" on items
to authenticated
using ( (select auth.uid()) = user_id );
```

### Anti-Patterns to Avoid

- **RLS without indexes:** Always add indexes on columns used in RLS policies (e.g., `workspace_id`, `user_id`)
- **Joining in RLS policies:** Avoid joins to the source table; restructure as `IN (SELECT ...)` queries
- **Missing `to authenticated`:** Always specify the role; prevents unnecessary policy evaluation for anon
- **`auth.uid()` without `IS NOT NULL` check:** For unauthenticated requests, `auth.uid()` returns `null`, causing silent failures
- **Referencing auth.users columns other than id:** Only the `id` column is guaranteed stable

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| UUID generation | Random string functions | `gen_random_uuid()` | Built-in, proper randomness |
| User auth | Custom auth tables | Supabase Auth (auth.users) | Security, session management, OAuth |
| Permission checks | Manual IF/THEN in queries | RLS policies | Declarative, enforced at database level |
| Migration versioning | Manual schema changes | Supabase CLI migrations | Reproducibility, rollback support |
| API exposure | Custom REST endpoints | Supabase PostgREST | Automatic API from schema with RLS |

**Key insight:** Supabase provides integrated auth, automatic REST API, and managed PostgreSQL. Custom solutions introduce security risks and maintenance burden.

## Common Pitfalls

### Pitfall 1: RLS Disabled After CREATE TABLE

**What goes wrong:** Tables created via SQL Editor or migrations don't have RLS enabled by default. Data is publicly accessible via API.

**Why it happens:** PostgreSQL's default is RLS disabled. Supabase Dashboard Table Editor enables it automatically, but raw SQL doesn't.

**How to avoid:** Always include `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` immediately after `CREATE TABLE`.

**Warning signs:** Unauthenticated API requests return data they shouldn't.

### Pitfall 2: Missing Policies = No Access

**What goes wrong:** After enabling RLS, all queries return empty results.

**Why it happens:** RLS defaults to "deny all" when no policies exist.

**How to avoid:** Create policies for each operation (SELECT, INSERT, UPDATE, DELETE) immediately after enabling RLS.

**Warning signs:** API returns empty arrays, dashboard shows no data.

### Pitfall 3: auth.uid() Returns NULL for Anon

**What goes wrong:** Policies like `using (auth.uid() = user_id)` silently fail for unauthenticated requests.

**Why it happens:** `null = user_id` is always false in SQL.

**How to avoid:** Explicitly check: `using (auth.uid() IS NOT NULL AND auth.uid() = user_id)`

**Warning signs:** Anonymous users see nothing when they should be blocked with an error.

### Pitfall 4: Cascade Delete Confusion

**What goes wrong:** Deleting a workspace doesn't delete associated apps, feedback, etc., or vice versa - deleting a user leaves orphan records.

**Why it happens:** Missing `ON DELETE CASCADE` in foreign key definitions.

**How to avoid:** Always specify cascade behavior explicitly:
```sql
workspace_id uuid references public.workspaces on delete cascade
user_id uuid references auth.users on delete cascade
```

**Warning signs:** Orphan records accumulate, foreign key constraint errors on delete.

### Pitfall 5: RLS Performance Degradation

**What goes wrong:** Queries become slow as data grows, especially with `LIMIT/OFFSET`.

**Why it happens:** RLS policies run on every row before filtering.

**How to avoid:**
1. Add indexes on policy columns
2. Wrap functions in `SELECT`: `(select auth.uid())`
3. Use `TO authenticated` to skip policies for anon role
4. Add explicit filters in queries (not just rely on RLS)

**Warning signs:** Query times increase linearly with table size.

## Code Examples

Verified patterns from official sources:

### Complete Workspace Setup

```sql
-- Source: Supabase official docs, adapted for multi-tenant
-- Enable UUID extension
create extension if not exists "uuid-ossp";

-- Workspaces (organizations)
create table public.workspaces (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  slug text unique not null,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

alter table public.workspaces enable row level security;

-- Workspace members (join table)
create table public.workspace_members (
  id uuid primary key default gen_random_uuid(),
  workspace_id uuid references public.workspaces on delete cascade not null,
  user_id uuid references auth.users on delete cascade not null,
  role text not null default 'member' check (role in ('owner', 'admin', 'member')),
  created_at timestamptz default now(),
  unique (workspace_id, user_id)
);

alter table public.workspace_members enable row level security;

-- Helper function for workspace membership check
create or replace function public.is_workspace_member(ws_id uuid)
returns boolean
language sql
security definer
stable
as $$
  select exists (
    select 1 from public.workspace_members
    where workspace_id = ws_id
    and user_id = (select auth.uid())
  );
$$;

-- RLS policies for workspaces
create policy "Users can view their workspaces"
on public.workspaces for select
to authenticated
using ( (select public.is_workspace_member(id)) );

create policy "Users can create workspaces"
on public.workspaces for insert
to authenticated
with check ( true );

-- RLS policies for workspace_members
create policy "Members can view workspace membership"
on public.workspace_members for select
to authenticated
using ( (select public.is_workspace_member(workspace_id)) );
```

### Apps Within Workspace

```sql
-- Source: Pattern derived from Supabase docs
create table public.apps (
  id uuid primary key default gen_random_uuid(),
  workspace_id uuid references public.workspaces on delete cascade not null,
  name text not null,
  api_key text unique default encode(gen_random_bytes(32), 'hex'),
  created_at timestamptz default now()
);

alter table public.apps enable row level security;

create index idx_apps_workspace on public.apps(workspace_id);

create policy "Workspace members can view apps"
on public.apps for select
to authenticated
using ( (select public.is_workspace_member(workspace_id)) );

create policy "Workspace members can create apps"
on public.apps for insert
to authenticated
with check ( (select public.is_workspace_member(workspace_id)) );
```

### Migration File Structure

```sql
-- Source: Supabase CLI docs
-- File: supabase/migrations/20260122000000_initial_schema.sql

-- This is an example migration structure
-- 1. Extensions first
create extension if not exists "uuid-ossp";

-- 2. Core tables
create table public.workspaces (...);

-- 3. Enable RLS immediately
alter table public.workspaces enable row level security;

-- 4. Create indexes
create index idx_workspaces_slug on public.workspaces(slug);

-- 5. Policies in separate migration or same file
create policy "..." on public.workspaces ...;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| pgjwt for JWT handling | Built-in auth.jwt() function | Supabase v2 | Don't install pgjwt extension |
| Custom auth tables | Supabase Auth (auth.users) | Always | Use built-in auth |
| Dashboard-only changes | CLI + migrations | Current best practice | Version control, reproducibility |
| Raw SQL auth checks | RLS policies | PostgreSQL 9.5+ | Database-level enforcement |

**Deprecated/outdated:**
- pgjwt extension: Use auth.jwt() built-in function instead
- timescaledb extension: Deprecated in Supabase
- pgsodium: Pending deprecation

## Open Questions

1. **API Key Storage Security**
   - What we know: API keys for apps stored in plain text in `apps` table
   - What's unclear: Should we hash them? How does SDK send them?
   - Recommendation: Store hash, compare with `crypt()`. Planner to decide exact approach.

2. **Soft Delete vs Hard Delete**
   - What we know: Cascade delete is standard
   - What's unclear: Business requirement for audit trail
   - Recommendation: Proceed with hard delete for v1; soft delete adds complexity.

## Sources

### Primary (HIGH confidence)
- Supabase official docs: Row Level Security - https://supabase.com/docs/guides/database/postgres/row-level-security
- Supabase official docs: Managing User Data - https://supabase.com/docs/guides/auth/managing-user-data
- Supabase official docs: Local Development - https://supabase.com/docs/guides/cli/local-development
- Supabase official docs: Custom Claims & RBAC - https://supabase.com/docs/guides/database/postgres/custom-claims-and-role-based-access-control-rbac

### Secondary (MEDIUM confidence)
- Supabase GitHub discussions: RLS Best Practices

### Tertiary (LOW confidence)
- None - all findings verified with official docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Supabase is project constraint with excellent documentation
- Architecture: HIGH - Patterns from official Supabase docs and examples
- Pitfalls: HIGH - Documented in official performance guide and RLS docs

**Research date:** 2026-01-22
**Valid until:** ~60 days (Supabase has stable, mature documentation)
