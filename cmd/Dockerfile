FROM golang:1.4.2-wheezy

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# this will ideally be built by the ONBUILD below ;)
#CMD ["/go/bin/ip6update", "server", "./server/config.yaml"]
CMD ["ls",  "-la", "/go/src/app", "/go/src"]

COPY . /go/src/app
#RUN go-wrapper download
RUN go-wrapper install