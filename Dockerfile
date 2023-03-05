FROM docker.io/golang:1.20 AS build-env

ENV GOOS=linux
ENV GOARCH=arm64

# Build Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM gcr.io/distroless/base:nonroot-arm64
#FROM gcr.io/distroless/base:nonroot-amd64

#EXPOSE 40000

WORKDIR /
COPY commands_hpsu.json .
COPY config.pi.yaml config.yaml
COPY echoctl .
#COPY --from=build-env /go/bin/linux_arm64/dlv /

ENTRYPOINT ["./echoctl"]
#CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/echoctl"]
