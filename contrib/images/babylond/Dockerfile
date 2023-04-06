FROM golang:1.19-alpine3.16 AS build-env

# Customize to your build env

# TARGETPLATFORM should be one of linux/amd64 or linux/arm64.
ARG TARGETPLATFORM="linux/amd64"
# Version to build. Default is the Git HEAD.
ARG VERSION="HEAD"

# Use muslc for static libs
ARG BUILD_TAGS="muslc"
ARG LEDGER_ENABLED="false"
# Debug mode
ARG DLV_ENABLED="false"
# Cosmos build options
ARG COSMOS_BUILD_OPTIONS=""

# Install cli tools for building and final image
RUN apk add --update --no-cache make git bash gcc linux-headers eudev-dev ncurses-dev openssh curl jq
RUN apk add --no-cache musl-dev


# Build
WORKDIR /go/src/github.com/babylonchain
RUN git clone https://github.com/babylonchain/babylon.git
WORKDIR /go/src/github.com/babylonchain/babylon
RUN git fetch
RUN git checkout ${VERSION}

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)

RUN if [ "${DLV_ENABLED}" = "true" ] ; then \
    go install github.com/go-delve/delve/cmd/dlv@latest ; \
fi

RUN LEDGER_ENABLED=$LEDGER_ENABLED BUILD_TAGS=$BUILD_TAGS COSMOS_BUILD_OPTIONS=$COSMOS_BUILD_OPTIONS make install

# Remove extra binaries we don't want in the image.
# We do this because later we have to copy all binaries from /go/bin
RUN rm -rf /go/bin/chain
RUN rm -rf /go/bin/node

# Use busybox to create a user
FROM busybox:stable-musl AS busybox
RUN addgroup --gid 1137 -S babylon && adduser --uid 1137 -S babylon -G babylon

# Use scratch for the final image
FROM scratch
WORKDIR /bin

# Label should match your github repo
LABEL org.opencontainers.image.source="https://github.com/babylonchain/babylond:${VERSION}"

# Installs all binaries built with go.
COPY --from=build-env /go/bin /bin

# Other binaries we want to keep.
COPY --from=build-env /usr/bin/ldd /bin/ldd
COPY --from=build-env /usr/bin/curl /bin/curl
COPY --from=build-env /usr/bin/jq /bin/jq

# Install Libraries
COPY --from=build-env /usr/lib/libgcc_s.so.1 /lib/
COPY --from=build-env /lib/ld-musl*.so.1* /lib

# jq Libraries
COPY --from=build-env /usr/lib/libonig.so.5 /lib

# curl Libraries
COPY --from=build-env /usr/lib/libcurl.so.4 /lib
COPY --from=build-env /lib/libz.so.1 /lib
COPY --from=build-env /usr/lib/libnghttp2.so.14 /lib
COPY --from=build-env /lib/libssl.so.1.1 /lib
COPY --from=build-env /lib/libcrypto.so.1.1 /lib
COPY --from=build-env /usr/lib/libbrotlidec.so.1 /lib
COPY --from=build-env /usr/lib/libbrotlicommon.so.1 /lib

# Install trusted CA certificates for curl
COPY --from=build-env /etc/ssl/cert.pem /etc/ssl/cert.pem

# Install cli tools from busybox
COPY --from=busybox /bin/ln /bin/ln
COPY --from=busybox /bin/ln /bin/sh

# Link each binary you want to sh to save space.
RUN for binary in \
dd \
vi \
chown \
id \
cp \
ls \
sh \
cat \
less \
grep \
sleep \
env \
tar \
tee \
rm \
du \
sort \
head \
df \
nc \
netstat \
mkdir \
; do ln sh $binary; done

# Copy user from busybox to scratch
COPY --from=busybox /etc/passwd /etc/passwd
COPY --from=busybox --chown=1137:1137 /home/babylon /home/babylon

# Set home directory and user
WORKDIR /home/babylon
RUN chown -R babylon /home/babylon
USER babylon
