{
  description = "API resource versioning tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in rec {
        packages = flake-utils.lib.flattenTree {
          vervet = pkgs.callPackage ./default.nix { };
        };
        defaultPackage = packages.vervet;
        defaultApp = flake-utils.lib.mkApp { drv = packages.vervet; };
        devShell = pkgs.callPackage ./shell.nix { inherit pkgs; };
      });
}
