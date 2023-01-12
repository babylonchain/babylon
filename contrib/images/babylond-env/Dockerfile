FROM golang:1.19-alpine AS build
RUN apk add build-base git linux-headers
WORKDIR /work
COPY go.mod go.sum /work/

RUN go mod download
COPY ./ /work
RUN LEDGER_ENABLED=false make clean build

FROM alpine:3.14 AS run
RUN apk add bash curl jq
COPY contrib/images/babylond-env/wrapper.sh /usr/bin/wrapper.sh

VOLUME /babylond
COPY --from=build /work/build/babylond /babylond/
WORKDIR /babylond

EXPOSE 26656 26657 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
