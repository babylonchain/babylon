FROM golang:1.19-alpine AS build
RUN apk add build-base git linux-headers libc-dev
RUN go install github.com/go-delve/delve/cmd/dlv@latest
WORKDIR /work
COPY go.mod go.sum /work/
RUN go mod download
COPY ./ /work
RUN LEDGER_ENABLED=false make COSMOS_BUILD_OPTIONS="debug,nostrip" clean build

FROM alpine:3.14 AS run
RUN apk add bash curl jq
COPY contrib/images/babylond-dlv/wrapper.sh /usr/bin/wrapper.sh

VOLUME /babylond
COPY --from=build /work/build/babylond /babylond/
COPY --from=build /go/bin/dlv /usr/local/bin
WORKDIR /babylond

EXPOSE 26656 26657 2345 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
