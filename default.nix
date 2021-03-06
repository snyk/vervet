{ buildGoModule, lib, lastMod }:
buildGoModule rec {
  pname = "vervet";
  version = builtins.substring 0 8 lastMod;
  src = ./.;

  vendorSha256 = "sha256-KcxaAOp9+d2DN4qFmmNDH3Di+TmiqhTbLsujh2aKRmk=";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
