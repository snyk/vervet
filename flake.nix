{
  description = "API resource versioning tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let lastMod = self.lastModifiedDate or self.lastModified or "19700101";
    in flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in rec {
        packages = flake-utils.lib.flattenTree {
          default = pkgs.callPackage ./default.nix { inherit lastMod; };
        };
        apps.default = flake-utils.lib.mkApp { drv = packages.default; };
        devShell = pkgs.callPackage ./shell.nix { inherit pkgs; };
      });
}
