{
  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      pre-commit-hooks,
    }:
    let
      goVersion = 24; # Change this to update the whole stack

      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forEachSupportedSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import nixpkgs {
              inherit system;
              overlays = [ self.overlays.default ];
            };
          }
        );
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      devShells = forEachSupportedSystem (
        { pkgs }:
        {
          default = pkgs.mkShell {
            shellHook = ''
              export LD_LIBRARY_PATH=${pkgs.wayland}/lib:${pkgs.lib.getLib pkgs.libGL}/lib:${pkgs.lib.getLib pkgs.libGL}/lib:$LD_LIBRARY_PATH
              ${self.checks.${pkgs.system}.pre-commit-check.shellHook}
            '';
            nativeBuildInputs = with pkgs; [
              go
              libGL
              xorg.libX11
              xorg.libXrandr
              xorg.libXcursor
              xorg.libXinerama
              xorg.libXi
              xorg.libXxf86vm
            ];
            packages = with pkgs; [
              air
              gotools
              gopls
              self.checks.${system}.pre-commit-check.enabledPackages
            ];
          };
        }
      );

      checks = forEachSupportedSystem (
        { pkgs }:
        {
          pre-commit-check = pre-commit-hooks.lib.${pkgs.system}.run {
            src = ./.;
            hooks = {
              gofmt.enable = true;
              golangci-lint.enable = true;
              govet.enable = true;
              staticcheck.enable = true;
            };
          };
        }
      );
    };
}
