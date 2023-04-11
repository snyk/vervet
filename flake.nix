{
  description = "API resource versioning tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let lastMod = self.lastModifiedDate or self.lastModified or "19700101";
    in flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        ne = nixpkgs + "/pkgs/development/node-packages/node-env.nix";
        nodeEnv = import ne {
          inherit (pkgs)
            stdenv lib python2 runCommand writeTextFile writeShellScript;
          inherit pkgs;
          libtool = pkgs.darwin.cctools;
          nodejs = pkgs.nodejs-14_x;
        };
      in rec {
        packages = flake-utils.lib.flattenTree {
          default = pkgs.callPackage ./default.nix { inherit lastMod; };
        };
        apps.default = flake-utils.lib.mkApp { drv = packages.default; };
        devShell = pkgs.callPackage ./shell.nix { inherit pkgs nodeEnv; };
      });
}
