{
  description = "screeps-launcher flake";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };
  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };
          screeps-launcher = pkgs.buildGoApplication {
            pname = "screeps-launcher";
            version = "1.16.0";
            src = ./.;
            modules = ./gomod2nix.toml;
            subPackages = [ "cmd/screeps-launcher" ];
            # try 1.26, 1.23 was previously broken due to modules.txt not being generated
            go = pkgs.go_1_26; 
          };
        in
        {
          packages = {
            default = screeps-launcher;
            # Disable for now, this needs to be considered due to needing build
            # dependencies for screeps
            # docker = pkgs.dockerTools.buildLayeredImage {
            #   name = "screeps-launcher";
            #   config = {
            #      Cmd = [ "${screeps-launcher}/bin/screeps-launcher"  ];
            #   };
            # };
          };
          devShells.default = pkgs.mkShell {
            packages = [
              (pkgs.mkGoEnv { pwd = ./.; })
              pkgs.gomod2nix
            ];
          };
        })
    );
}
