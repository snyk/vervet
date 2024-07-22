{ pkgs ? import <nixpkgs> { }, nodeEnv }:
let
  spectral = pkgs.callPackage ./spectral.nix {
    inherit (pkgs) fetchurl;
    buildNodePackage = nodeEnv.buildNodePackage;
  };
in pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    go_1_22
    gopls
    gotools
    golangci-lint
    envsubst
    spectral
  ];
  shellHook = ''
    export GOPATH="$HOME/.cache/gopaths/$(sha256sum <<<$(pwd) | awk '{print $1}')"
  '';
}

