{
  description = "templ";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, nixpkgs-unstable, gitignore, xc }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
        pkgs-unstable = import nixpkgs-unstable { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ system, pkgs, ... }:
        rec {
          default = templ;

          templ = pkgs.buildGo123Module {
            name = "templ";
            subPackages = [ "cmd/templ" ];
            src = gitignore.lib.gitignoreSource ./.;
            vendorHash = "sha256-ipLn52MsgX7KQOJixYcwMR9TCeHz55kQQ7fgkIgnu7w=";
            CGO_ENABLED = 0;
            flags = [
              "-trimpath"
            ];
            ldflags = [
              "-s"
              "-w"
              "-extldflags -static"
            ];
          };
        });

      # `nix develop` provides a shell containing development tools.
      devShell = forAllSystems ({ system, pkgs, pkgs-unstable, ... }:
        pkgs.mkShell {
          buildInputs = [
            pkgs.golangci-lint
            pkgs.cosign # Used to sign container images.
            pkgs.esbuild # Used to package JS examples.
            pkgs.go
            pkgs-unstable.gopls
            pkgs.goreleaser
            pkgs.gotestsum
            pkgs.ko # Used to build Docker images.
            pkgs.nodejs # Used to build templ-docs.
            xc.packages.${system}.xc
          ];
        });

      # This flake outputs an overlay that can be used to add templ and
      # templ-docs to nixpkgs as per https://templ.guide/quick-start/installation/#nix
      #
      # Example usage:
      #
      # nixpkgs.overlays = [
      #   inputs.templ.overlays.default
      # ];
      overlays.default = final: prev: {
        templ = self.packages.${final.stdenv.system}.templ;
        templ-docs = self.packages.${final.stdenv.system}.templ-docs;
      };
    };
}

