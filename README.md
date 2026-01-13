# Illoominate

A feedback management platform for collecting, organizing, and acting on user feedback. Enable your users to submit bugs, feature requests, and general feedback through customizable portals.

## Features

- **Feedback Collection**: Collect bugs, feature requests, and general feedback from users
- **Community Portals**: Public feedback portals where users can submit and vote on items
- **Voting System**: Let users upvote feedback to surface the most requested features
- **Status Tracking**: Track feedback through its lifecycle (new, under review, planned, in progress, completed)
- **Team Collaboration**: Assign feedback to team members, add internal comments, and manage visibility
- **Tagging & Organization**: Categorize feedback with tags for easy filtering
- **Anonymous Submissions**: Support for anonymous feedback submissions

## Project Structure

```
illoominate/
├── api/          # Go backend API (Chi router, PostgreSQL)
├── web/          # React dashboard (Vite, TailwindCSS, React 19)
├── landing/      # Marketing site (Astro, Cloudflare)
├── sdks/         # Client SDKs for integration
└── supabase/     # Database migrations and Supabase config
```

## Tech Stack

### Backend
- Go 1.23
- Chi router
- PostgreSQL (via pgx)
- Supabase Auth
- Google Cloud Storage

### Web Dashboard
- React 19
- Vite
- TailwindCSS 4
- TanStack Router & Query
- Zustand

### Landing Page
- Astro 5
- React
- TailwindCSS
- Cloudflare Pages

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- PostgreSQL (or Supabase)
- [Nix](https://nixos.org/) (optional, for reproducible dev environment)

### Development

1. Clone the repository:
   ```bash
   git clone git@github.com:JonnyWalker81/illoominate.git
   cd illoominate
   ```

2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```

3. Start the API:
   ```bash
   cd api
   go run cmd/server/main.go
   ```

4. Start the web dashboard:
   ```bash
   cd web
   npm install
   npm run dev
   ```

5. Start the landing page:
   ```bash
   cd landing
   npm install
   npm run dev
   ```

### Using Just

If you have [just](https://github.com/casey/just) installed, you can use the provided justfile:

```bash
just --list  # See available commands
```

## License

MIT
