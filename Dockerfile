FROM golang:1.10 as build-env

WORKDIR /go/src/goredirector
ADD . /go/src/goredirector

RUN go get && go build

FROM gcr.io/distroless/base

COPY --from=build-env /go/src/goredirector/goredirector /
COPY redirs.yaml /

ENTRYPOINT ["/goredirector"]
