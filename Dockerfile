# ====================== BUILD STAGE ======================
FROM golang:1.23-bookworm AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Compilamos
RUN go build -v -o /run-app ./cmd/server

# ====================== RUNTIME STAGE ======================
FROM debian:bookworm-slim

# Certificados + herramientas mínimas
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Binario de la app
COPY --from=builder /run-app /usr/local/bin/run-app

# ←←← ARCHIVOS ESTÁTICOS Y TEMPLATES (esto es lo que faltaba)
COPY --from=builder /usr/src/app/web /web
COPY --from=builder /usr/src/app/internal/templates /internal/templates

EXPOSE 3000

CMD ["run-app", "serve"]