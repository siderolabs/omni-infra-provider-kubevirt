# THIS FILE WAS AUTOMATICALLY GENERATED, PLEASE DO NOT EDIT.
#
# Generated on 2025-07-24T12:03:40Z by kres 4c6b4c0.

# common variables

SHA := $(shell git describe --match=none --always --abbrev=8 --dirty)
TAG := $(shell git describe --tag --always --dirty --match v[0-9]\*)
ABBREV_TAG := $(shell git describe --tags >/dev/null 2>/dev/null && git describe --tag --always --match v[0-9]\* --abbrev=0 || echo 'undefined')
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
ARTIFACTS := _out
IMAGE_TAG ?= $(TAG)
OPERATING_SYSTEM := $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
WITH_DEBUG ?= false
WITH_RACE ?= false
REGISTRY ?= ghcr.io
USERNAME ?= siderolabs
REGISTRY_AND_USERNAME ?= $(REGISTRY)/$(USERNAME)
PROTOBUF_GO_VERSION ?= 1.36.6
GRPC_GO_VERSION ?= 1.5.1
GRPC_GATEWAY_VERSION ?= 2.27.1
VTPROTOBUF_VERSION ?= 0.6.0
GOIMPORTS_VERSION ?= 0.34.0
GOMOCK_VERSION ?= 0.5.2
DEEPCOPY_VERSION ?= v0.5.6
GOLANGCILINT_VERSION ?= v2.2.2
GOFUMPT_VERSION ?= v0.8.0
GO_VERSION ?= 1.24.5
GO_BUILDFLAGS ?=
GO_LDFLAGS ?=
CGO_ENABLED ?= 0
GOTOOLCHAIN ?= local
GOEXPERIMENT ?= synctest
TESTPKGS ?= ./...
KRES_IMAGE ?= ghcr.io/siderolabs/kres:latest
CONFORMANCE_IMAGE ?= ghcr.io/siderolabs/conform:latest

# docker build settings

BUILD := docker buildx build
PLATFORM ?= linux/amd64
PROGRESS ?= auto
PUSH ?= false
CI_ARGS ?=
BUILDKIT_MULTI_PLATFORM ?=
COMMON_ARGS = --file=Dockerfile
COMMON_ARGS += --provenance=false
COMMON_ARGS += --progress=$(PROGRESS)
COMMON_ARGS += --platform=$(PLATFORM)
COMMON_ARGS += --build-arg=BUILDKIT_MULTI_PLATFORM=$(BUILDKIT_MULTI_PLATFORM)
COMMON_ARGS += --push=$(PUSH)
COMMON_ARGS += --build-arg=ARTIFACTS="$(ARTIFACTS)"
COMMON_ARGS += --build-arg=SHA="$(SHA)"
COMMON_ARGS += --build-arg=TAG="$(TAG)"
COMMON_ARGS += --build-arg=ABBREV_TAG="$(ABBREV_TAG)"
COMMON_ARGS += --build-arg=USERNAME="$(USERNAME)"
COMMON_ARGS += --build-arg=REGISTRY="$(REGISTRY)"
COMMON_ARGS += --build-arg=TOOLCHAIN="$(TOOLCHAIN)"
COMMON_ARGS += --build-arg=CGO_ENABLED="$(CGO_ENABLED)"
COMMON_ARGS += --build-arg=GO_BUILDFLAGS="$(GO_BUILDFLAGS)"
COMMON_ARGS += --build-arg=GO_LDFLAGS="$(GO_LDFLAGS)"
COMMON_ARGS += --build-arg=GOTOOLCHAIN="$(GOTOOLCHAIN)"
COMMON_ARGS += --build-arg=GOEXPERIMENT="$(GOEXPERIMENT)"
COMMON_ARGS += --build-arg=PROTOBUF_GO_VERSION="$(PROTOBUF_GO_VERSION)"
COMMON_ARGS += --build-arg=GRPC_GO_VERSION="$(GRPC_GO_VERSION)"
COMMON_ARGS += --build-arg=GRPC_GATEWAY_VERSION="$(GRPC_GATEWAY_VERSION)"
COMMON_ARGS += --build-arg=VTPROTOBUF_VERSION="$(VTPROTOBUF_VERSION)"
COMMON_ARGS += --build-arg=GOIMPORTS_VERSION="$(GOIMPORTS_VERSION)"
COMMON_ARGS += --build-arg=GOMOCK_VERSION="$(GOMOCK_VERSION)"
COMMON_ARGS += --build-arg=DEEPCOPY_VERSION="$(DEEPCOPY_VERSION)"
COMMON_ARGS += --build-arg=GOLANGCILINT_VERSION="$(GOLANGCILINT_VERSION)"
COMMON_ARGS += --build-arg=GOFUMPT_VERSION="$(GOFUMPT_VERSION)"
COMMON_ARGS += --build-arg=TESTPKGS="$(TESTPKGS)"
TOOLCHAIN ?= docker.io/golang:1.24-alpine

# extra variables

REMOVE_VOLUMES ?= false

# help menu

export define HELP_MENU_HEADER
# Getting Started

To build this project, you must have the following installed:

- git
- make
- docker (19.03 or higher)

## Creating a Builder Instance

The build process makes use of experimental Docker features (buildx).
To enable experimental features, add 'experimental: "true"' to '/etc/docker/daemon.json' on
Linux or enable experimental features in Docker GUI for Windows or Mac.

To create a builder instance, run:

	docker buildx create --name local --use

If running builds that needs to be cached aggresively create a builder instance with the following:

	docker buildx create --name local --use --config=config.toml

config.toml contents:

[worker.oci]
  gc = true
  gckeepstorage = 50000

  [[worker.oci.gcpolicy]]
    keepBytes = 10737418240
    keepDuration = 604800
    filters = [ "type==source.local", "type==exec.cachemount", "type==source.git.checkout"]
  [[worker.oci.gcpolicy]]
    all = true
    keepBytes = 53687091200

If you already have a compatible builder instance, you may use that instead.

## Artifacts

All artifacts will be output to ./$(ARTIFACTS). Images will be tagged with the
registry "$(REGISTRY)", username "$(USERNAME)", and a dynamic tag (e.g. $(IMAGE):$(IMAGE_TAG)).
The registry and username can be overridden by exporting REGISTRY, and USERNAME
respectively.

endef

ifneq (, $(filter $(WITH_RACE), t true TRUE y yes 1))
GO_BUILDFLAGS += -race
CGO_ENABLED := 1
GO_LDFLAGS += -linkmode=external -extldflags '-static'
endif

ifneq (, $(filter $(WITH_DEBUG), t true TRUE y yes 1))
GO_BUILDFLAGS += -tags sidero.debug
else
GO_LDFLAGS += -s
endif

all: unit-tests omni-infra-provider-kubevirt image-omni-infra-provider-kubevirt lint

$(ARTIFACTS):  ## Creates artifacts directory.
	@mkdir -p $(ARTIFACTS)

.PHONY: clean
clean:  ## Cleans up all artifacts.
	@rm -rf $(ARTIFACTS)

target-%:  ## Builds the specified target defined in the Dockerfile. The build result will only remain in the build cache.
	@$(BUILD) --target=$* $(COMMON_ARGS) $(TARGET_ARGS) $(CI_ARGS) .

registry-%:  ## Builds the specified target defined in the Dockerfile and the output is an image. The image is pushed to the registry if PUSH=true.
	@$(MAKE) target-$* TARGET_ARGS="--tag=$(REGISTRY)/$(USERNAME)/$(IMAGE_NAME):$(IMAGE_TAG)" BUILDKIT_MULTI_PLATFORM=1

local-%:  ## Builds the specified target defined in the Dockerfile using the local output type. The build result will be output to the specified local destination.
	@$(MAKE) target-$* TARGET_ARGS="--output=type=local,dest=$(DEST) $(TARGET_ARGS)"
	@PLATFORM=$(PLATFORM) DEST=$(DEST) bash -c '\
	  for platform in $$(tr "," "\n" <<< "$$PLATFORM"); do \
	    directory="$${platform//\//_}"; \
	    if [[ -d "$$DEST/$$directory" ]]; then \
		  echo $$platform; \
	      mv -f "$$DEST/$$directory/"* $$DEST; \
	      rmdir "$$DEST/$$directory/"; \
	    fi; \
	  done'

generate:  ## Generate .proto definitions.
	@$(MAKE) local-$@ DEST=./

lint-golangci-lint:  ## Runs golangci-lint linter.
	@$(MAKE) target-$@

lint-gofumpt:  ## Runs gofumpt linter.
	@$(MAKE) target-$@

.PHONY: fmt
fmt:  ## Formats the source code
	@docker run --rm -it -v $(PWD):/src -w /src golang:$(GO_VERSION) \
		bash -c "export GOTOOLCHAIN=local; \
		export GO111MODULE=on; export GOPROXY=https://proxy.golang.org; \
		go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION) && \
		gofumpt -w ."

lint-govulncheck:  ## Runs govulncheck linter.
	@$(MAKE) target-$@

.PHONY: base
base:  ## Prepare base toolchain
	@$(MAKE) target-$@

.PHONY: unit-tests
unit-tests:  ## Performs unit tests
	@$(MAKE) local-$@ DEST=$(ARTIFACTS)

.PHONY: unit-tests-race
unit-tests-race:  ## Performs unit tests with race detection enabled.
	@$(MAKE) target-$@

.PHONY: $(ARTIFACTS)/omni-infra-provider-kubevirt-linux-amd64
$(ARTIFACTS)/omni-infra-provider-kubevirt-linux-amd64:
	@$(MAKE) local-omni-infra-provider-kubevirt-linux-amd64 DEST=$(ARTIFACTS)

.PHONY: omni-infra-provider-kubevirt-linux-amd64
omni-infra-provider-kubevirt-linux-amd64: $(ARTIFACTS)/omni-infra-provider-kubevirt-linux-amd64  ## Builds executable for omni-infra-provider-kubevirt-linux-amd64.

.PHONY: omni-infra-provider-kubevirt
omni-infra-provider-kubevirt: omni-infra-provider-kubevirt-linux-amd64  ## Builds executables for omni-infra-provider-kubevirt.

.PHONY: lint-markdown
lint-markdown:  ## Runs markdownlint.
	@$(MAKE) target-$@

.PHONY: lint
lint: lint-golangci-lint lint-gofumpt lint-govulncheck lint-markdown  ## Run all linters for the project.

.PHONY: image-omni-infra-provider-kubevirt
image-omni-infra-provider-kubevirt:  ## Builds image for omni-infra-provider-kubevirt.
	@$(MAKE) registry-$@ IMAGE_NAME="omni-infra-provider-kubevirt"

.PHONY: docker-compose-up
docker-compose-up:
	ARTIFACTS="$(ARTIFACTS)" SHA="$(SHA)" TAG="$(TAG)" USERNAME="$(USERNAME)" REGISTRY="$(REGISTRY)" PROTOBUF_TS_VERSION="$(PROTOBUF_TS_VERSION)" NODE_BUILD_ARGS="$(NODE_BUILD_ARGS)" TOOLCHAIN="$(TOOLCHAIN)" CGO_ENABLED="$(CGO_ENABLED)" GO_BUILDFLAGS="$(GO_BUILDFLAGS)" GOLANGCILINT_VERSION="$(GOLANGCILINT_VERSION)" GOFUMPT_VERSION="$(GOFUMPT_VERSION)" GOIMPORTS_VERSION="$(GOIMPORTS_VERSION)" PROTOBUF_GO_VERSION="$(PROTOBUF_GO_VERSION)" GRPC_GO_VERSION="$(GRPC_GO_VERSION)" GRPC_GATEWAY_VERSION="$(GRPC_GATEWAY_VERSION)" VTPROTOBUF_VERSION="$(VTPROTOBUF_VERSION)" DEEPCOPY_VERSION="$(DEEPCOPY_VERSION)" TESTPKGS="$(TESTPKGS)" COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 GO_LDFLAGS="$(GO_LDFLAGS)" docker compose -p talemu --file ./hack/compose/docker-compose.yml --file ./hack/compose/docker-compose.override.yml up --build

.PHONY: docker-compose-down
docker-compose-down:
	ARTIFACTS="$(ARTIFACTS)" SHA="$(SHA)" TAG="$(TAG)" USERNAME="$(USERNAME)" REGISTRY="$(REGISTRY)" PROTOBUF_TS_VERSION="$(PROTOBUF_TS_VERSION)" NODE_BUILD_ARGS="$(NODE_BUILD_ARGS)" TOOLCHAIN="$(TOOLCHAIN)" CGO_ENABLED="$(CGO_ENABLED)" GO_BUILDFLAGS="$(GO_BUILDFLAGS)" GOLANGCILINT_VERSION="$(GOLANGCILINT_VERSION)" GOFUMPT_VERSION="$(GOFUMPT_VERSION)" GOIMPORTS_VERSION="$(GOIMPORTS_VERSION)" PROTOBUF_GO_VERSION="$(PROTOBUF_GO_VERSION)" GRPC_GO_VERSION="$(GRPC_GO_VERSION)" GRPC_GATEWAY_VERSION="$(GRPC_GATEWAY_VERSION)" VTPROTOBUF_VERSION="$(VTPROTOBUF_VERSION)" DEEPCOPY_VERSION="$(DEEPCOPY_VERSION)" TESTPKGS="$(TESTPKGS)" COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 GO_LDFLAGS="$(GO_LDFLAGS)" docker compose -p talemu --file ./hack/compose/docker-compose.yml --file ./hack/compose/docker-compose.override.yml down --rmi local --remove-orphans --volumes=$(REMOVE_VOLUMES)

run-integration-test: omni-infra-provider-kubevirt
	@hack/test/integration.sh

.PHONY: rekres
rekres:
	@docker pull $(KRES_IMAGE)
	@docker run --rm --net=host --user $(shell id -u):$(shell id -g) -v $(PWD):/src -w /src -e GITHUB_TOKEN $(KRES_IMAGE)

.PHONY: help
help:  ## This help menu.
	@echo "$$HELP_MENU_HEADER"
	@grep -E '^[a-zA-Z%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: release-notes
release-notes: $(ARTIFACTS)
	@ARTIFACTS=$(ARTIFACTS) ./hack/release.sh $@ $(ARTIFACTS)/RELEASE_NOTES.md $(TAG)

.PHONY: conformance
conformance:
	@docker pull $(CONFORMANCE_IMAGE)
	@docker run --rm -it -v $(PWD):/src -w /src $(CONFORMANCE_IMAGE) enforce

