FROM golang:1.10 as build-env

WORKDIR /go/src/goredirector
ADD main.go /go/src/goredirector

RUN go build

FROM gcr.io/distroless/base

COPY --from=build-env /go/src/goredirector/goredirector /

ENTRYPOINT ["/goredirector"]
