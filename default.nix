{
  buildGoApplication,
  lib,
}:
with builtins;
let
  version = head (match ".*const cmdVersion = \"(.*)\"\n.*" (readFile ./internal/cmd/cmd.go));
in
buildGoApplication {
  inherit version;
  pname = "vervet";
  src = ./.;
  modules = ./gomod2nix.toml;

  meta = with lib; {
    description = "API resource versioning tool";
    homepage = "https://github.com/snyk/vervet";
    platforms = platforms.linux ++ platforms.darwin;
  };
  subPackages = [ "cmd/vervet" ];
}
