{ lib, buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "gopls";
  version = "0.10.1";

  src = fetchFromGitHub {
    owner = "golang";
    repo = "tools";
    rev = "8321f7bbcfd30300762661ed9188226b42e27ec1";
    sha256 = "9WDqd8Xgiov/OFAFl5yZmon4o3grbOxzZs3wnNu7pbg=";
  };

  vendorSha256 = "EZ/XPta2vQfemywoC2kbTamJ43K4tr4I7mwVzrTbRkA=";
  modRoot = "gopls";

  subPackages = [ "." ];

  meta = with lib; {
    platforms = platforms.linux ++ platforms.darwin;
  };
}
