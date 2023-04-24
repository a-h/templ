{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "nixpkgs/nixos-22.11";
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    neovim-nightly-overlay = {
      url = "github:nix-community/neovim-nightly-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
      # Neovim 0.9.0
      inputs.neovim-flake.url = "github:neovim/neovim?dir=contrib&rev=040f1459849ab05b04f6bb1e77b3def16b4c2f2b";
    };
  };

  outputs = { self, flake-utils, nixpkgs, xc, neovim-nightly-overlay }:
    flake-utils.lib.eachDefaultSystem (system:
    let 
      pkgsDefault = import nixpkgs { overlays = [ neovim-nightly-overlay.overlay ]; };
      pkgs = import nixpkgs {
        inherit system; overlays = [
          (self: super: {
            xc = xc.packages.${system}.xc;
            neovim = import ./nix/nvim.nix { pkgs = pkgsDefault; };
            gopls = pkgs.callPackage ./nix/gopls.nix { };
            templ = pkgs.callPackage ./nix/templ.nix {
              go = pkgs.go_1_20;
              xc = self.xc;
            };
            nerdfonts = (pkgsDefault.nerdfonts.override { fonts = [ "IBMPlexMono" ]; });
          })
        ];
      };
      shell = pkgs.mkShell {
        packages = [
          pkgs.asciinema
          pkgs.git
          pkgs.go
          pkgs.gopls
          pkgs.gotools
          pkgs.ibm-plex
          pkgs.neovim
          pkgs.nerdfonts
          pkgs.ripgrep
          pkgs.silver-searcher
          pkgs.templ
          pkgs.tmux
          pkgs.wget
          pkgs.xc
          pkgs.zip
        ];
      };
    in
      {
        defaultPackage = pkgs.templ;
        packages = {
          templ = pkgs.templ;
        };
        devShells = {
          default = shell;
        };
      }
  );
}
