FROM golang:1.10-alpine3.7
WORKDIR /go/src/github.com/MadVikingGod/mockmetrics/
RUN apk update
RUN apk add git

RUN go get -u github.com/golang/dep/...

ADD . /go/src/github.com/MadVikingGod/mockmetrics/

RUN dep ensure -v
RUN GOOS=linux GOCC=gcc go build -o app main.go && mv ./app /go/bin

FROM alpine:3.7

COPY --from=0 /go/bin/app .

EXPOSE 8080

CMD ["./app"]
