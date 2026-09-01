GO        ?= go
GOMOD2NIX ?= gomod2nix
GINKGO    ?= ginkgo
CFORGO    ?= $(GO) tool c-for-go

GO_SRC ?= $(shell find . -name '*.go')

CGO_SRC     := $(addprefix cgo_helpers,.h .go .c)
GEN_SRC     := $(addsuffix .go,types gnumake)
GNUMAKE_SRC := $(addprefix gnumake/,${CGO_SRC} ${GEN_SRC})

build:
	nix build .#

test:
	$(GINKGO) run -r

generate gen: ${CGO_SRC} ${GEN_SRC}

update:
	nix flake update

check lint:
	nix flake check

format fmt:
	nix fmt

tidy: go.sum nix/gomod2nix.toml

${GNUMAKE_SRC}: gnumake.yml
	$(CFORGO) -out ${CURDIR} $<
${GEN_SRC} ${CGO_SRC}: ${GNUMAKE_SRC}
	cp $? .

go.sum: go.mod ${GO_SRC}
	$(GO) mod tidy

nix/gomod2nix.toml: go.sum ${GO_SRC}
	$(GOMOD2NIX) generate --dir ${CURDIR} --outdir ${@D}
