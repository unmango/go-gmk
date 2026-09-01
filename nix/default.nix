{
  buildGoApplication,
  lib,
  ginkgo,
  gnumake,
  version,
}:
buildGoApplication {
  pname = "gnumake-go";
  inherit version;

  src = lib.cleanSource ../.;
  modules = ./gomod2nix.toml;

  # gnumake.h lives in the gnumake out output, there is no dev output.
  buildInputs = [ gnumake ];
  CGO_CFLAGS = "-I${gnumake}/include";

  nativeCheckInputs = [ ginkgo ];

  checkPhase = ''
    ginkgo run ./...
  '';
}
