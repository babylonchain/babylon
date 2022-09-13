FROM golang:1.18-alpine AS build

RUN apk add --no-cache openssh-client git make build-base linux-headers libc-dev

WORKDIR /work

ENV GO111MODULE=on

RUN git clone https://github.com/btcsuite/btcd.git
RUN cd btcd && go install -v . ./cmd/... && cd -

RUN git clone https://github.com/btcsuite/btcwallet.git
RUN cd btcwallet && go install -v . ./cmd/... && cd -


FROM alpine:3.14 AS run
RUN apk add bash expect expect-dev
WORKDIR /bitcoinsim

COPY --from=build /go/bin/* /usr/bin/
COPY wrapper.sh /bitcoinsim/wrapper.sh
COPY btcwallet_create.exp /bitcoinsim/btcwallet_create.exp

ENV RPC_USER=rpcuser
ENV RPC_PASS=rpcpass
ENV WALLET_PASS=pass
ENV BITCOIN_CONF="/bitcoinconf"
ENV GENERATE_INTERVAL_SECS=30

EXPOSE 18554 18555 18556

ENTRYPOINT ["/bitcoinsim/wrapper.sh"]
CMD []
STOPSIGNAL SIGTERM
