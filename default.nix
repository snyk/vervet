{ buildGoModule, lib, lastMod }:
buildGoModule rec {
  pname = "vervet";
  version = builtins.substring 0 8 lastMod;
  src = ./.;

  vendorSha256 = "sha256-Ttl+2nOqj1/MBu6PQsp3/PS6Q6tp3EXS40zE+1nLCzQ=";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
