{ buildGoModule, lib, lastMod }:
buildGoModule rec {
  pname = "vervet";
  version = builtins.substring 0 8 lastMod;
  src = ./.;

  vendorHash = "sha256-8UOyGj1//ydQuU9PHKNRG68xkKyI4Qz3eSKj8sqNqDc=";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
