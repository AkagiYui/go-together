FROM  --platform=$BUILDPLATFORM golang:1.23.2-alpine AS builder
ARG TARGETPLATFORM
COPY . /workspace
WORKDIR /workspace
RUN GOOS=linux GOARCH=${TARGETARCH} \
    GO111MODULE=on GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 \
    go build -ldflags="-w -s" -o app

FROM scratch AS prod
COPY --from=builder /workspace/app /workspace/app
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /workspace
CMD ["./app"]
