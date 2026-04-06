# ====================== BUILD STAGE ======================
FROM golang:1.23-bookworm AS builder

WORKDIR /usr/src/app

# Copiar dependencias primero (cache)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar todo el código
COPY . .

# Compilar desde la carpeta correcta
RUN go build -v -o /run-app ./cmd/server

# ====================== RUNTIME STAGE ======================
FROM debian:bookworm-slim

# Instalar certificados para HTTPS y R2
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /run-app /usr/local/bin/run-app

EXPOSE 3000

CMD ["run-app"]