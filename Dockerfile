FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/server \
    ./cmd/main


FROM scratch

COPY --from=builder /bin/server /server
COPY --from=builder /app/local/local.yaml /local/local.yaml
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8082

ENTRYPOINT ["/server"]