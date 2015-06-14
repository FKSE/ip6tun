FROM golang:1.4.2-wheezy

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install -y 6tunnel

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY ./config-server.yaml /go/src/app/config.yml

RUN go get github.com/FKSE/ip6update
RUN go install github.com/FKSE/ip6update

CMD ["ip6update", "server", "./config.yml"]