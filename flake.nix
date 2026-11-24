{
  description = "Local Beach, a Docker-based development environment for Neos CMS and Flow Framework";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
      version = self.shortRev or self.dirtyShortRev or "dev";
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "localbeach";
          inherit version;
          src = self;

          # The full module cache is needed, because "go generate" runs
          # assets_generate.go, which imports vfsgen.
          proxyVendor = true;
          vendorHash = "sha256-Zgj0cokYdeUX1FsnByEwBxVlqAnu7CSG0GNEgu9cVMY=";

          subPackages = [ "cmd/beach" ];

          env.CGO_ENABLED = 0;

          ldflags = [
            "-X github.com/flownative/localbeach/pkg/version.Version=${version}"
          ];

          preBuild = ''
            rm -f assets/compiled.go
            go generate -v
          '';

          nativeBuildInputs = [ pkgs.makeWrapper ];

          # Like the Homebrew formula, provide mkcert and certutil for "beach setup-https".
          # Docker and Docker Compose are expected to be installed separately.
          postInstall = ''
            wrapProgram $out/bin/beach \
              --suffix PATH : ${
                pkgs.lib.makeBinPath [
                  pkgs.mkcert
                  pkgs.nss.tools
                ]
              }
          '';

          meta = {
            description = "Docker-based development environment for Neos CMS and Flow Framework";
            homepage = "https://www.flownative.com/localbeach";
            license = pkgs.lib.licenses.asl20;
            mainProgram = "beach";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          inputsFrom = [ self.packages.${pkgs.stdenv.hostPlatform.system}.default ];
          packages = [
            pkgs.gnumake
            pkgs.mkcert
          ];
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}
