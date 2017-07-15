FROM golang:1.8.1-alpine

RUN mkdir /go/src/github.com
RUN mkdir /go/src/github.com/bor3ham
RUN mkdir /go/src/github.com/bor3ham/reja

WORKDIR /go/src/github.com/bor3ham/reja
ADD . /go/src/github.com/bor3ham/reja
ADD entrypoint.sh /entrypoint.sh

RUN apk add --no-cache git mercurial \
  && go get -u github.com/lib/pq \
  && go get -u github.com/gorilla/mux \
  && go get -u github.com/gorilla/context \
  && go get -u github.com/davecgh/go-spew/spew \
  && go get -u github.com/mailru/easyjson/... \
  && apk del git mercurial
