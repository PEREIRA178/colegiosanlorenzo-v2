# ====================== BUILD STAGE ======================
FROM golang:1.23-bookworm AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /run-app ./cmd/server

# ====================== RUNTIME STAGE ======================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /run-app /usr/local/bin/run-app

EXPOSE 3000

# ←←←← ESTA ES LA CORRECCIÓN IMPORTANTE
CMD ["run-app", "serve"]