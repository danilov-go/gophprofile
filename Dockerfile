FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,https://goproxy.io,direct
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/worker .
COPY --from=builder /app/web ./web/
COPY --from=builder /app/migrations ./migrations/

EXPOSE 8080

CMD ["./server"]
