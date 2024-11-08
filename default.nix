{
  buildGoModule,
  lib,
}:
with builtins;
let
  version = head (match ".*const cmdVersion = \"(.*)\"\n.*" (readFile ./internal/cmd/cmd.go));
in
buildGoModule {
  inherit version;
  pname = "vervet";
  src = ./.;

  vendorHash = "sha256-9BxGg0tOToOJhuMaVBgW89qVxEOTLGJy7h8rwKHsDkE=";

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
