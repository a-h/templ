{
  description = "templ";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs?ref=nixos-unstable";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gitignore, xc }:
    let
      # Systems supported
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      # Helper to provide system-specific attributes
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
      });
    in rec {
      packages = forAllSystems ({ pkgs, ... }: rec {
        default = templ;

        templ = pkgs.buildGo121Module {
          name = "templ";
          src = gitignore.lib.gitignoreSource ./.;
          subPackages = [ "cmd/templ" ];
          vendorSha256 = "sha256-7QYF8BvLpTcDstkLWxR0BgBP0NUlJ20IqW/nNqMSBn4=";
          CGO_ENALBED = 0;

          flags = [
            "-trimpath"
          ];

          ldflags = [
            "-s"
            "-w"
            "-extldflags -static"
          ];
        };

        templ-docs = pkgs.buildNpmPackage {
          name = "templ-docs";
          src = gitignore.lib.gitignoreSource ./docs;
          npmDepsHash = "sha256-/9FzD38s0bER9zvR4AfU/KBrzRsWDmuFeuMgSL45/1s=";

          installPhase = ''
            mkdir -p $out/share
            cp -r build/ $out/share/docs
          '';
        };
      });

      # `nix develop` provides a shell containing required tools for development
      devShell = forAllSystems ({ system, pkgs }:
        pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_21
            goreleaser
            nodejs
            xc.packages.${system}.xc
          ];
        });

      # Allows users to install the package on their system in an easy way
      overlays.default = final: prev:
        forAllSystems ({ system, ... }: {
          templ = packages.${system}.templ;
          templ-docs = packages.${system}.templ-docs;
        });
    };
}
