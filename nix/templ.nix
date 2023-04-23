{ pkgs, stdenv, go, xc }:

stdenv.mkDerivation {
  pname = "templ";
  version = "devel";
  src = ./..;
  builder = ./templ-install.sh;
  system = builtins.currentSystem;
  nativeBuildInputs = [ go xc ];
}

