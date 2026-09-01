GO        ?= go
GOMOD2NIX ?= gomod2nix
GINKGO    ?= ginkgo
CFORGO    ?= $(GO) tool c-for-go

# c-for-go resolves SourcesPaths against gnumake.yml IncludePaths, which
# cannot reference the store path the devshell exports, so the header is
# staged into the repo root where the implicit "." include path finds it.
GNUMAKE_INCLUDE ?= /usr/include

GO_SRC ?= $(shell find . -name '*.go')

GEN_SRC := $(addprefix internal/gnumake/,cgo_helpers.go cgo_helpers.h doc.go gnumake.go types.go)

build:
	nix build .#

test:
	$(GINKGO) run -r

generate gen: ${GEN_SRC}

update:
	nix flake update

check lint:
	nix flake check

format fmt:
	nix fmt

tidy: go.sum nix/gomod2nix.toml

gnumake.h: ${GNUMAKE_INCLUDE}/gnumake.h
	cp $< $@

${GEN_SRC} &: gnumake.yml gnumake.h
	$(CFORGO) -out ${CURDIR}/internal $<

go.sum: go.mod ${GO_SRC}
	$(GO) mod tidy

nix/gomod2nix.toml: go.sum ${GO_SRC}
	$(GOMOD2NIX) generate --dir ${CURDIR} --outdir ${@D}
