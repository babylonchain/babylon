FROM golang:1.19-alpine AS build

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers

WORKDIR /work

COPY go.mod go.sum /work/

RUN go mod download

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)

# Copy the remaining files
COPY . .

RUN GOWORK=off go build \
        -mod=readonly \
        -tags "netgo,ledger,muslc" \
        -ldflags \
            "-X github.com/cosmos/cosmos-sdk/version.Name="babylon" \
            -X github.com/cosmos/cosmos-sdk/version.AppName="babylond" \
            -X github.com/cosmos/cosmos-sdk/version.BuildTags='netgo,ledger,muslc' \
            -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
        -trimpath \
        -o /work/build/babylond \
        /work/cmd/babylond/main.go

FROM alpine:3.14 AS run
RUN apk add bash curl jq
COPY contrib/images/babylond-env/wrapper.sh /usr/bin/wrapper.sh

VOLUME /babylond
COPY --from=build /work/build/babylond /babylond/
WORKDIR /babylond

EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
