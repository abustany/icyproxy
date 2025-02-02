{
  description = "icyproxy";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-24.11";
    flake-utils.url  = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };
        rev = if (self ? shortRev) then self.shortRev else "dev";
      in
      with pkgs;
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
          ];

	  GOTOOLCHAIN = "local";

          shellHook = ''
          export GOPATH="$(realpath .)/.go";
          export PATH="''\${GOPATH}/bin:''\${PATH}"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "icyproxy";
          version = rev;
          src = pkgs.lib.cleanSource self;
          vendorHash = null;
        };
      }
    );
}
