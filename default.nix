{ buildGoModule, lib }:

buildGoModule rec {
  pname = "vervet";
  version = "4.6.6";
  src = ./.;

  vendorSha256 = null;
  proxyVendor = true;

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
