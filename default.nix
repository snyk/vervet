{ buildGoModule, lib, lastMod }:
buildGoModule rec {
  pname = "vervet";
  version = builtins.substring 0 8 lastMod;
  src = ./.;

  vendorHash = "sha256-sumj5H1MOnsMPn5YML9co/kwU3WSiKbapfHXRIg0Xp4=";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
