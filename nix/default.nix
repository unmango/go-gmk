{
  buildGoApplication,
  lib,
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

  # There are no suites yet and ginkgo treats that as a failure. The test
  # layer adds the suite and turns this back on.
  doCheck = false;
}
