FROM golang:1.14.2-buster AS builder

ADD . /app
WORKDIR /app
RUN make build-no-test

FROM debian:buster

COPY --from=builder /app/build /app
WORKDIR /app
ENTRYPOINT ./go-microservice
