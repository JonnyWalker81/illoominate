{
  description = "Illoominate - Multi-tenant Feedback + Feature Voting Platform";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go (latest)
            go
            golangci-lint
            sqlc
            go-migrate

            # Node.js (latest LTS: 22)
            nodejs_22
            pnpm
            nodePackages.typescript
            nodePackages.typescript-language-server

            # Supabase
            supabase-cli

            # PostgreSQL (latest: 17)
            postgresql_17

            # Google Cloud (optional)
            google-cloud-sdk

            # Cloudflare (for landing page)
            nodePackages.wrangler

            # Tools
            just
            jq
            curl
            watchexec
            httpie

            # Development
            git
            gnumake
          ];

          shellHook = ''
            echo ""
            echo "╔══════════════════════════════════════════════════════════════╗"
            echo "║           Illoominate Development Environment                ║"
            echo "║    Multi-tenant Feedback + Feature Voting Platform           ║"
            echo "╚══════════════════════════════════════════════════════════════╝"
            echo ""
            echo "Versions:"
            echo "  Go:        $(go version | cut -d' ' -f3)"
            echo "  Node.js:   $(node --version)"
            echo "  pnpm:      $(pnpm --version)"
            echo "  PostgreSQL: $(psql --version | cut -d' ' -f3)"
            echo ""
            echo "Available commands:"
            echo "  just dev        - Start all services (API + Web + Supabase)"
            echo "  just api        - Start Go API only"
            echo "  just web        - Start React frontend only"
            echo "  just db-start   - Start Supabase local"
            echo "  just db-migrate - Run database migrations"
            echo "  just db-reset   - Reset database (WARNING: destroys data)"
            echo "  just lint       - Run all linters"
            echo "  just test       - Run all tests"
            echo "  just fmt        - Format all code"
            echo "  just gen        - Generate sqlc code"
            echo "  just landing    - Start landing page dev server"
            echo "  just landing-dev - Start landing with Cloudflare Workers"
            echo ""
            echo "Setup steps (first time):"
            echo "  1. cp .env.example .env"
            echo "  2. just db-start"
            echo "  3. just db-migrate"
            echo "  4. just dev"
            echo ""

            # Set up Go environment
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"

            # Load .env if exists
            if [ -f .env ]; then
              set -a
              source .env
              set +a
              echo "Loaded environment from .env"
            else
              echo "Note: No .env file found. Copy .env.example to .env to get started."
            fi
            echo ""
          '';

          # Environment variables for development
          SUPABASE_URL = "http://127.0.0.1:54321";
          SUPABASE_ANON_KEY = "";
          DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:54322/postgres";
        };

        # Packages for CI/CD
        packages = {
          api = pkgs.buildGoModule {
            pname = "illoominate-api";
            version = "0.1.0";
            src = ./api;
            vendorHash = null; # Will be set after first build
          };
        };
      }
    );
}
