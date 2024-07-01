FROM golang:1.21.4-alpine as builder

# Use muslc for static libs
ARG BUILD_TAGS="muslc"

# Ledger Settings
ARG LEDGER_ENABLED="false"

# Cosmos build options
ARG COSMOS_BUILD_OPTIONS=""

RUN apk add --no-cache --update openssh git make build-base linux-headers libc-dev \
                                pkgconfig zeromq-dev musl-dev alpine-sdk libsodium-dev \
                                libzmq-static libsodium-static gcc

# Build
WORKDIR /go/src/github.com/babylonchain/babylon
# Cache dependencies
COPY go.mod go.sum /go/src/github.com/babylonchain/babylon/
RUN go mod download
# Copy the rest of the files
COPY ./ /go/src/github.com/babylonchain/babylon/

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$(uname -m) | cut -d ' ' -f 1)

RUN CGO_LDFLAGS="$CGO_LDFLAGS -lstdc++ -lm -lsodium" \
    LEDGER_ENABLED=$LEDGER_ENABLED \
    COSMOS_BUILD_OPTIONS=$COSMOS_BUILD_OPTIONS \
    CGO_ENABLED=1 \
    BUILD_TAGS=$BUILD_TAGS \
    LINK_STATICALLY=true \
    make build

# FINAL IMAGE
FROM alpine:3.16 AS run

RUN addgroup --gid 1137 -S babylon && adduser --uid 1137 -S babylon -G babylon

RUN apk add bash curl jq

COPY --from=builder /go/src/github.com/babylonchain/babylon/build/babylond /bin/babylond

# Set home directory and user
WORKDIR /home/babylon
RUN chown -R babylon /home/babylon
USER babylon
