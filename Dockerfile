FROM golang:1.14.2-buster

ADD . ./app

WORKDIR ./app

ENTRYPOINT make run
