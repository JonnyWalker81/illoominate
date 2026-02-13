{
  description = "Illoominate - Native-first user feedback platform";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
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
            # JavaScript runtime & package managers
            bun
            nodejs_22

            # Database tools
            sqlite

            # Git & version control
            git

            # Useful dev utilities
            jq
            curl
          ];

          shellHook = ''
            echo "Illoominate dev environment loaded"
            echo "  bun: $(bun --version)"
            echo "  node: $(node --version)"
            echo ""
            echo "Run 'bun install' or 'npm install' to install dependencies"
            echo "Run 'bun run dev' or 'npm run dev' to start the dev server"
          '';
        };
      }
    );
}
