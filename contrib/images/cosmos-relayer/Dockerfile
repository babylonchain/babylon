FROM debian:bullseye-slim AS build-env

RUN apt-get update && apt-get install -y git make jq gcc make wget

WORKDIR /work

ARG TARGETARCH

ENV MACHINE $TARGETARCH

# Download and install Go
ENV GOLANG_VERSION 1.21.4
RUN wget -q https://golang.org/dl/go${GOLANG_VERSION}.linux-${MACHINE}.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-${MACHINE}.tar.gz && \
    rm go${GOLANG_VERSION}.linux-${MACHINE}.tar.gz
# Set Go environment variables
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

WORKDIR /work

ENV GO111MODULE on
#ENV RELAYER_TAG main
ENV RELAYER_TAG v2.4.2

# Install the relayer
RUN git clone https://github.com/cosmos/relayer.git
RUN cd relayer && git fetch origin && git checkout ${RELAYER_TAG} && make install && cd -

FROM debian:bullseye-slim AS run
# Create a user
RUN addgroup --system rly && adduser --system --gid 101 rly

COPY --from=build-env /go/bin/rly /usr/bin/rly

# Set home directory and user
WORKDIR /home/rly
RUN chown -R rly /home/rly
USER rly
