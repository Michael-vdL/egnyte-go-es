FROM golang:1.17.6-alpine as builder
RUN apk --update upgrade
RUN apk add --no-cache git make musl-dev gcc libc6-compat
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go clean --modcache
RUN go mod download
RUN GOOS=linux CGO_ENABLED=1 go build -a -o egnyte ./cmd

FROM alpine:latest
RUN apk --update upgrade
RUN mkdir /config
COPY ./config /config
WORKDIR /
COPY --from=builder /app/egnyte .

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
CMD ["/egnyte"]