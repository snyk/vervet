{
  description = "API resource versioning tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        packages = flake-utils.lib.flattenTree {
          default = pkgs.callPackage ./default.nix { };
        };
        apps.default = flake-utils.lib.mkApp { drv = packages.default; };
        devShell = pkgs.callPackage ./shell.nix { inherit pkgs; };
      }
    );
}
