FROM golang:1.22.1-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
# use cache mount if building takes too long https://docs.docker.com/build/guide/mounts/
RUN go mod download
COPY . .
RUN go build -o bin/http-log-processor ./cmd/processor

FROM debian:bookworm-slim
COPY --from=builder /app/bin/http-log-processor /usr/local/bin/
ENTRYPOINT ["http-log-processor"]