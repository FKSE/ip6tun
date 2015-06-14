FROM golang:1.4.2-wheezy

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# this will ideally be built by the ONBUILD below ;)
# CMD ["/go/bin/ip6update", "server", ".config-server.yaml"]
#CMD ["go-wrapper", "run", "server", "./config-server.yaml"]

COPY . /go/src/app
#RUN go-wrapper download
#RUN go-wrapper install

RUN go get github.com/FKSE/ip6update
RUN go install github.com/FKSE/ip6update

CMD ["ip6update", "server", "./config-server.yaml"]