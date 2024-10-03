{
  buildGoModule,
  lib,
  version,
}:
buildGoModule rec {
  inherit version;

  pname = "vervet";
  src = ./.;

  vendorHash = "sha256-8UOyGj1//ydQuU9PHKNRG68xkKyI4Qz3eSKj8sqNqDc=";

  preBuild = "VERSION=${version} bash ./scripts/genversion.bash";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
