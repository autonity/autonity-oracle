# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build autoracle in a stock Go builder container
FROM golang:1.19-alpine as builder

LABEL org.opencontainers.image.source https://github.com/clearmatics/autonity-oracle

RUN apk add --no-cache make gcc musl-dev linux-headers libc-dev git perl-utils

ADD . /autoracle
RUN cd /autoracle && make autoracle


# Pull Autonity into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /autoracle/build/bin/autoracle /usr/local/bin/
COPY --from=builder /autoracle/build/bin/plugins /usr/local/bin/plugins/
COPY --from=builder /autoracle/e2e_test/plugins/simulator_plugins/ /usr/local/bin/plugins/


ENTRYPOINT ["autoracle"]

# Add some metadata labels to help programatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
