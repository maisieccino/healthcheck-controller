ARG GOVERSION=1.13.4
FROM golang:${GOVERSION}-alpine3.10 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /src

RUN apk add --no-cache \
    bash \
    git \
    openssh

ADD go.mod go.sum ./
RUN go mod download
ADD . .
RUN go mod vendor && \
    /src/hack/verify-codegen.sh && \
    go vet ./... && \
    go test ./... && \
    go build -o /src/hc-controller ./cmd/hc-controller/main.go

FROM scratch
COPY --from=builder /src/hc-controller /hc-controller
ENTRYPOINT [ "/hc-controller" ]
