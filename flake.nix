{
  description = "An untitled display manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;
      allSystems = [ "x86_64-linux" ];
      forAllSystems = f: lib.genAttrs allSystems (system: f system nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (
        system: pkgs: {
          untitled-dm = pkgs.callPackage ./nix/package.nix { };
          default = self.packages.${system}.untitled-wm;
        }
      );
      devShells = forAllSystems (
        system: pkgs: {
          default = pkgs.mkShell {
            name = "untitled-dm";
            packages = [
              pkgs.go
              pkgs.gopls
            ];
          };
        }
      );
    };
}
