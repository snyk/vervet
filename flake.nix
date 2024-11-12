{
  description = "API resource versioning tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            gomod2nix.overlays.default
          ];
        };
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
