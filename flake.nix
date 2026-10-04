{
  description = "templ";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    version = {
      url = "github:a-h/version";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      gitignore,
      version,
    }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs allSystems (
          system:
          f {
            inherit system;
            pkgs =
              let
                pkgs-unstable = import nixpkgs-unstable { inherit system; };
              in
              import nixpkgs {
                inherit system;
                overlays = [
                  (final: prev: {
                    gopls = pkgs-unstable.gopls;
                    version = version.packages.${system}.default; # Used to apply version numbers to the repo.
                  })
                ];
              };
          }
        );
    in
    {
      packages = forAllSystems (
        { pkgs, ... }:
        rec {
          default = templ;

          templ = pkgs.buildGoModule {
            pname = "templ";
            version = builtins.readFile ./.version;
            subPackages = [ "cmd/templ" ];
            src = gitignore.lib.gitignoreSource ./.;
            vendorHash = "sha256-WXUlbUR+5a0BRStFF4A9mMjGiYFL5ULu1yrRhIG0vwc=";
            env = {
              CGO_ENABLED = 0;
            };
            flags = [
              "-trimpath"
            ];
            ldflags = [
              "-s"
              "-w"
              "-extldflags -static"
            ];
          };

          # `nix build .#docs` builds the documentation site.
          docs = pkgs.buildNpmPackage {
            pname = "templ-docs";
            version = builtins.readFile ./.version;
            src = gitignore.lib.gitignoreSource ./docs;
            npmDeps = pkgs.importNpmLock {
              npmRoot = gitignore.lib.gitignoreSource ./docs;
            };
            npmConfigHook = pkgs.importNpmLock.npmConfigHook;
            installPhase = ''
              runHook preInstall
              cp -r build $out
              runHook postInstall
            '';
          };
        }
        # `nix build .#docker-image` builds the Docker image. It requires a Linux system.
        # The image runs as the nonroot user (UID 65532) and has no shell.
        // pkgs.lib.optionalAttrs pkgs.stdenv.isLinux {
          docker-image = pkgs.dockerTools.buildLayeredImage {
            name = "ghcr.io/a-h/templ";
            tag = "latest";
            # Produce an uncompressed tarball for `crane push`.
            compressor = "none";
            contents = [
              pkgs.dockerTools.caCertificates
              self.packages.${pkgs.stdenv.hostPlatform.system}.templ
              (pkgs.writeTextDir "etc/passwd" "nonroot:x:65532:65532:nonroot:/home/nonroot:/sbin/nologin\n")
              (pkgs.writeTextDir "etc/group" "nonroot:x:65532:\n")
            ];
            fakeRootCommands = ''
              mkdir -p home/nonroot
              chown 65532:65532 home/nonroot
              mkdir -m 1777 tmp
            '';
            config = {
              Entrypoint = [ "/bin/templ" ];
              Env = [
                "PATH=/bin"
                "HOME=/home/nonroot"
              ];
              User = "65532:65532";
            };
          };
        }
      );

      # `nix develop` provides a shell containing development tools.
      devShell = forAllSystems (
        { pkgs, ... }:
        pkgs.mkShell {
          buildInputs = [
            pkgs.golangci-lint
            pkgs.cosign # Used to sign container images.
            pkgs.crane # Used to push and tag Docker images.
            pkgs.esbuild # Used to package JS examples.
            pkgs.go
            pkgs.gopls
            pkgs.goreleaser
            pkgs.gotestsum
            pkgs.govulncheck
            pkgs.nodejs # Used to build templ-docs.
            pkgs.prettier # Used for formatting JS and CSS.
            pkgs.syft # Used to generate SBOMs for Docker images.
            pkgs.version
            pkgs.xc
          ];
        }
      );

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
      };
    };
}
