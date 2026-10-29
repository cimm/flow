{
  description = "Flow Development Environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "flow";
          version = "0.1.0";
          src = ./.;

          vendorHash = "sha256-6fV4G9sv1DAWhZLRxHYVpAGM9N/nD5oLLepV1+ZgN88=";

          nativeBuildInputs = with pkgs; [
            pkg-config
            gobject-introspection
            wrapGAppsHook4
          ];

          buildInputs = with pkgs; [
            gtk4
            libadwaita
          ];

          # Desktop launcher entry
          desktopItems = [
            (pkgs.makeDesktopItem {
              name = "flow";
              desktopName = "Flow";
              comment = "HomeWizard energy monitor";
              exec = "flow";
              icon = "utilities-system-monitor";
              categories = [ "Utility" "Monitor" ];
            })
          ];

          # Allow gotk4's unsafe assume-no-moving-gc
          tags = [ "purego" ];

          meta = {
            description = "HomeWizard energy monitor";
            mainProgram = "flow";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gtk4
            libadwaita
            pkg-config
            gobject-introspection
          ];
        };
      });
}
