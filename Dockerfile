FROM golang:alpine AS builder
RUN mkdir /app
ADD . /app/
WORKDIR /app

ARG GOPROXY
ARG GOSUMDB

RUN go env -w GOPROXY=${GOPROXY}
RUN go env -w GOSUMDB=${GOSUMDB}
RUN go mod download
RUN go build -o bin/ trpc/trpc.go
RUN chmod +x bin/trpc

# protoc-gen-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# mockgen
RUN go install go.uber.org/mock/mockgen@latest
# goimports
RUN go install golang.org/x/tools/cmd/goimports@latest
# protoc-gen-validate
RUN go install github.com/envoyproxy/protoc-gen-validate@latest
RUN go install github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@latest

FROM golang:alpine
RUN apk update --no-cache && apk add --no-cache protoc flatc

ARG GOPROXY
ARG GOSUMDB

RUN go env -w GOPROXY=${GOPROXY}
RUN go env -w GOSUMDB=${GOSUMDB}

COPY --from=builder /app/bin/trpc /usr/local/bin/
COPY --from=builder /go/bin /go/bin
WORKDIR /workspace
RUN trpc setup

ENTRYPOINT ["trpc"]

