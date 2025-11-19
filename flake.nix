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
            };
          }
        );
    in
    {
      devShells = forEachSupportedSystem (
        { pkgs }:
        {
          default = pkgs.mkShell {
            shellHook = ''
              export LD_LIBRARY_PATH=${pkgs.lib.getLib pkgs.alsa-lib}/lib:${pkgs.lib.getLib pkgs.wayland}/lib:${pkgs.lib.getLib pkgs.libGL}/lib:${pkgs.lib.getLib pkgs.libdecor}/lib:${pkgs.lib.getLib pkgs.libxkbcommon}/lib:$LD_LIBRARY_PATH
              ${self.checks.${pkgs.system}.pre-commit-check.shellHook}
            '';
            hardeningDisable = [ "fortify" ]; # Make delve work with direnv IDE extension
            nativeBuildInputs = with pkgs; [
              go
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

      packages = forEachSupportedSystem (
        { pkgs }:
        {
          default = pkgs.buildGoModule {
            pname = "chip8-go";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-61sZUaaYdVSb3Vw0Csc9yHFNlAEhIyGeq2ujkfa6HS4=";

            doCheck = false;

            nativeBuildInputs = with pkgs; [
              makeWrapper
            ];

            postFixup = ''
              wrapProgram $out/bin/chip8-go \
                --prefix LD_LIBRARY_PATH : "${
                  pkgs.lib.makeLibraryPath (
                    with pkgs;
                    [
                      alsa-lib
                      libjack2
                      pipewire
                      wayland
                      libxkbcommon
                      libdecor
                      xorg.libX11
                      xorg.libXext
                      xorg.libXcursor
                      xorg.libXinerama
                      xorg.libXi
                      xorg.libXrandr
                      xorg.libXxf86vm
                      libGL
                      vulkan-loader
                      mesa
                    ]
                  )
                }"
            '';
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
            };
          };
        }
      );
    };
}
