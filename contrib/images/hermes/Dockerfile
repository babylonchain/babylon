# Adapted from https://github.com/informalsystems/hermes/blob/master/ci/release/hermes.Dockerfile

FROM rust:1-buster AS build-env

#ENV RELAYER_TAG master
ENV RELAYER_TAG romac/packet-data-hex

WORKDIR /root

# Install the relayer
RUN git clone https://github.com/informalsystems/hermes.git
RUN cd hermes && git fetch origin && git checkout ${RELAYER_TAG} && cargo build --release && cd -

FROM ubuntu:latest
LABEL maintainer="hello@informal.systems"
ARG UID=1000
ARG GID=1000

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
RUN update-ca-certificates
RUN groupadd -g ${GID} hermes && useradd -l -m hermes -s /bin/bash -u ${UID} -g ${GID}

WORKDIR /home/hermes
USER hermes:hermes
ENTRYPOINT ["/usr/bin/hermes"]

COPY --chown=hermes:hermes --from=build-env /root/hermes/target/release/hermes /usr/bin/hermes
