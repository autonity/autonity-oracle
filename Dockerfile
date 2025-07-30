# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build autoracle in a stock Go builder container
FROM golang:1.24-alpine AS builder

LABEL org.opencontainers.image.source https://github.com/autonity/autonity-oracle

RUN apk add --no-cache make musl-dev linux-headers libc-dev git

ADD . /autoracle
RUN cd /autoracle && make autoracle


# Pull autoracle into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /autoracle/build/bin/autoracle /usr/local/bin/
COPY --from=builder /autoracle/build/bin/plugins ./plugins/
COPY --from=builder /autoracle/build/bin/oracle_config.yml /usr/local/bin/

ENTRYPOINT ["autoracle"]

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
