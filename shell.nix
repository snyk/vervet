{
  pkgs ? import <nixpkgs> { },
}:
with builtins;
let
  goMinorVersion = head (match ".*go 1\.\([0-9]+\)\(\.[0-9]+\)?\n.*" (readFile ./go.mod));
  go = pkgs."go_1_${goMinorVersion}";
in
pkgs.mkShell {
  packages = with pkgs; [
    go
    gopls
    gotools
    golangci-lint
    envsubst
    gomod2nix
  ];
}
