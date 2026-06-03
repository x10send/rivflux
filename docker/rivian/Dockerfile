FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY src/rivian/go.mod src/rivian/go.sum ./
RUN go mod download
COPY src/rivian/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/auth

FROM alpine:latest
RUN addgroup -S rivflux && adduser -S rivflux -G rivflux
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/auth .
RUN mkdir -p /data && chown -R rivflux:rivflux /data
USER rivflux
EXPOSE 8888
ENTRYPOINT ["./server"]
