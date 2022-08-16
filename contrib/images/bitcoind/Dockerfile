FROM golang:1.18-alpine AS build

RUN apk add --no-cache openssh-client git make build-base linux-headers libc-dev

WORKDIR /work

ENV GO111MODULE=on

RUN git clone https://github.com/btcsuite/btcd.git
RUN cd btcd && go install -v . ./cmd/... && cd -

RUN git clone https://github.com/btcsuite/btcwallet.git
RUN cd btcwallet && go install -v . ./cmd/... && cd -


FROM alpine:3.14 AS run
RUN apk add bash curl jq expect expect-dev
WORKDIR /bitcoind

COPY --from=build /go/bin/* /usr/bin/
COPY contrib/images/bitcoind/wrapper.sh /bitcoind/wrapper.sh
COPY contrib/images/bitcoind/btcwallet_create.exp /bitcoind/btcwallet_create.exp

ENV RPCUSER=rpcuser
ENV RPCPASS=rpcpass
ENV PASSPHRASE=pass
ENV GENERATE_INTERVAL_SECS=30

# Do the one time setup of the miner and the wallet.
RUN /bitcoind/wrapper.sh setup

EXPOSE 18554 18555 18556

ENTRYPOINT ["/bitcoind/wrapper.sh"]
CMD ["run"]
STOPSIGNAL SIGTERM
