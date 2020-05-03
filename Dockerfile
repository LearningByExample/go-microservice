FROM golang:1.14.2-buster

ADD build/config config
ADD build/go-microservice go-microservice

ENTRYPOINT ./go-microservice
