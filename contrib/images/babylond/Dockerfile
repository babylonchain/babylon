FROM golang:1.20-alpine AS build-env

# Customize to your build env

# TARGETPLATFORM should be one of linux/amd64 or linux/arm64.
ARG TARGETPLATFORM="linux/amd64"
# Version to build. Default is the Git HEAD.
ARG VERSION="HEAD"

# Use muslc for static libs
ARG BUILD_TAGS="muslc"
ARG LEDGER_ENABLED="false"
# Cosmos build options
ARG COSMOS_BUILD_OPTIONS=""

# Install cli tools for building and final image
RUN apk add --update --no-cache make git bash gcc linux-headers eudev-dev ncurses-dev openssh curl jq
RUN apk add --no-cache musl-dev

# Build
WORKDIR /go/src/github.com/babylonchain/babylon
# First cache dependencies
COPY go.mod go.sum /go/src/github.com/babylonchain/babylon/
RUN go mod download
# Then copy everything else
COPY ./ /go/src/github.com/babylonchain/babylon/
RUN git checkout -f ${VERSION}

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$(uname -m) | cut -d ' ' -f 1)

RUN LEDGER_ENABLED=$LEDGER_ENABLED \
    BUILD_TAGS=$BUILD_TAGS \
    COSMOS_BUILD_OPTIONS=$COSMOS_BUILD_OPTIONS \
    LINK_STATICALLY=true \
    make build

FROM alpine:3.14 AS run
# Create a user
RUN addgroup --gid 1137 -S babylon && adduser --uid 1137 -S babylon -G babylon
RUN apk add bash curl jq

# Label should match your github repo
LABEL org.opencontainers.image.source="https://github.com/babylonchain/babylond:${VERSION}"

# Install Libraries
# COPY --from=build-env /usr/lib/libgcc_s.so.1 /lib/
# COPY --from=build-env /lib/ld-musl*.so.1* /lib

COPY --from=build-env /go/src/github.com/babylonchain/babylon/build/babylond /bin/babylond

# Set home directory and user
WORKDIR /home/babylon
RUN chown -R babylon /home/babylon
USER babylon
