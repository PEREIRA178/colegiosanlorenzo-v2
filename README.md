# 🏫 CSL System — Colegio San Lorenzo

Sistema integral de gestión de contenido y comunicación para el Colegio San Lorenzo de Copiapó, Atacama.

**Stack:** Go 1.22 + Fiber v2 + PocketBase (embedded) + HTMX + Cloudflare R2 + WebSockets

---

## Arquitectura

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENTS                              │
│  Browsers · Pantallas · Totems (kiosk) · Mobile         │
└──────────┬──────────┬──────────┬────────────────────────┘
           │ HTTP     │ HTMX     │ WebSocket
┌──────────▼──────────▼──────────▼────────────────────────┐
│              Go + Fiber v2 (API Gateway)                │
│  ┌──────────┐ ┌───────────┐ ┌────────┐ ┌─────────────┐ │
│  │Auth MW   │ │HTMX       │ │WS Hub  │ │Static Server│ │
│  │JWT+RBAC  │ │Fragments  │ │Realtime│ │HTML+Assets  │ │
│  └──────────┘ └───────────┘ └────────┘ └─────────────┘ │
└──────────┬──────────┬──────────┬────────────────────────┘
           │          │          │
┌──────────▼──────────▼──────────▼────────────────────────┐
│              PocketBase (embedded)                       │
│  ┌──────┐ ┌───────────┐ ┌────────┐ ┌──────┐ ┌───────┐  │
│  │Auth  │ │Collections│ │Realtime│ │S3    │ │Hooks  │  │
│  │Users │ │10 tablas  │ │SSE     │ │Hook  │ │Events │  │
│  └──────┘ └───────────┘ └────────┘ └──────┘ └───────┘  │
└──────────┬──────────┬──────────┬────────────────────────┘
           │          │          │
     ┌─────▼────┐ ┌───▼──────┐ ┌▼──────────────┐
     │ SQLite   │ │ CF R2    │ │ External      │
     │ pb_data  │ │ S3 media │ │ WhatsApp/LLM  │
     └──────────┘ └──────────┘ └───────────────┘
```

---

## Requisitos

- Go 1.22+
- PocketBase (viene embebido)
- Ollama (opcional, para auto-respuestas AI)
- Cuenta Cloudflare R2 (para almacenamiento de medios)
- Cuenta Twilio (para WhatsApp)

---

## Instalación paso a paso

### 1. Clonar y configurar

```bash
git clone https://github.com/tu-org/csl-system.git
cd csl-system
cp .env.example .env
# Editar .env con tus credenciales
```

### 2. Configurar Cloudflare R2

Necesitas los siguientes datos de tu cuenta Cloudflare:

```env
R2_ACCOUNT_ID=tu-account-id
R2_ACCESS_KEY_ID=tu-access-key
R2_SECRET_ACCESS_KEY=tu-secret-key
R2_BUCKET_NAME=csl-media
R2_REGION=auto
R2_PUBLIC_URL=https://media.colegiosanlorenzo.cl
```

Para obtenerlos:
1. Ir a Cloudflare Dashboard → R2
2. Crear bucket "csl-media"
3. Ir a "Manage R2 API Tokens" → crear token con permisos de lectura/escritura
4. Copiar Account ID, Access Key ID, Secret Access Key

### 3. Configurar WhatsApp (Twilio)

```env
TWILIO_ACCOUNT_SID=tu-sid
TWILIO_AUTH_TOKEN=tu-token
TWILIO_FROM_NUMBER=whatsapp:+14155238886
```

### 4. Compilar y ejecutar

```bash
# Instalar dependencias
go mod tidy

# Ejecutar en desarrollo
go run cmd/server/main.go

# O compilar binario
go build -o csl-system cmd/server/main.go
./csl-system
```

### 5. Primer acceso

- **Web pública:** http://localhost:3000
- **Dashboard admin:** http://localhost:3000/admin
- **PocketBase admin:** http://localhost:8090/_/

**Credenciales por defecto:**
- Email: `admin@colegiosanlorenzo.cl`
- Password: `csl2026admin!`

⚠️ **Cambiar inmediatamente en producción.**

---

## Estructura de carpetas

```
csl-system/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── auth/
│   │   ├── collections.go      # PocketBase collections (10 tablas)
│   │   └── jwt.go              # JWT generation/validation
│   ├── config/config.go        # Environment config
│   ├── handlers/
│   │   ├── admin/handlers.go   # Dashboard CRUD handlers
│   │   ├── web/handlers.go     # Public web + RSS + webhooks
│   │   ├── ws/websocket.go     # WebSocket handlers
│   │   └── fragments/          # HTMX fragment handlers
│   │       └── fragments.go    # Hero, Eventos, Noticias, Blog
│   ├── middleware/auth.go      # JWT auth + RBAC middleware
│   ├── realtime/hub.go         # WebSocket hub + PB hooks
│   ├── services/
│   │   ├── r2/r2.go            # Cloudflare R2 storage
│   │   ├── whatsapp/           # Twilio WhatsApp
│   │   └── ollama/             # Local AI service
│   └── templates/
│       ├── admin/pages/        # Dashboard HTML pages
│       ├── devices/display.html # Kiosk mode template
│       └── web/fragments/      # HTMX fragment templates
├── web/                        # Public HTML files (modified)
│   ├── index.html              # Main page (HTMX-enhanced)
│   ├── comunicados.html
│   ├── nuestro-colegio.html
│   ├── admision.html
│   ├── edex.html
│   ├── cepad.html
│   ├── inclusion.html
│   └── ceal.html
├── static/                     # CSS, JS, images
├── pb_data/                    # PocketBase SQLite (gitignored)
├── go.mod
├── .env.example
├── DESIGN_SYSTEM.md
└── README.md
```

---

## PocketBase Collections (10)

| Collection | Type | Campos |
|---|---|---|
| `users` | Auth | email, password, role, nombre, telefono, rut, activo |
| `multimedia` | Base | filename, url_r2, type, size, uploaded_by, estado, descripcion, duracion_segundos, thumbnail |
| `events` | Base | title, description, date, category, urgencia, status, targets_whatsapp, whatsapp_sent, whatsapp_sent_at, created_by |
| `news_articles` | Base | title, slug, body, excerpt, category, status, cover_image, author, featured, al_aire |
| `playlists` | Base | name, description, status, created_by |
| `playlist_items` | Base | playlist_id, multimedia_id, tipo, orden, duracion_segundos |
| `devices` | Base | name, type, code, layout, ubicacion, playlist_id, status, last_seen |
| `form_responses` | Base | event_id, user_id, tipo, valor, mensaje, leido |
| `whatsapp_logs` | Base | event_id, phone, message_sid, status, direction, body, error_message |

---

## Rutas del Sistema

### Web pública
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/` | Index con HTMX fragments |
| GET | `/{page}.html` | Sub-páginas estáticas |
| GET | `/fragments/hero` | Fragment: hero carousel |
| GET | `/fragments/eventos` | Fragment: eventos |
| GET | `/fragments/noticias` | Fragment: noticias |
| GET | `/fragments/blog` | Fragment: blog/prensa |
| GET | `/rss.xml` | RSS feed |
| GET | `/display/:code` | Display kiosk para dispositivos |

### WebSocket
| Ruta | Descripción |
|---|---|
| `/ws/web` | Conexión realtime para web pública |
| `/ws/device/:code` | Conexión realtime para dispositivos |

### Admin (protegidas con JWT)
| Método | Ruta | Descripción |
|---|---|---|
| GET/POST | `/admin/login` | Login |
| POST | `/admin/logout` | Logout |
| GET | `/admin/dashboard` | Dashboard principal |
| CRUD | `/admin/events/*` | Gestión de eventos |
| CRUD | `/admin/news/*` | Gestión de noticias |
| CRUD | `/admin/multimedia/*` | Gestión de multimedia |
| CRUD | `/admin/playlists/*` | Gestión de playlists |
| CRUD | `/admin/devices/*` | Gestión de dispositivos |
| CRUD | `/admin/users/*` | Gestión de usuarios (superadmin/director) |
| GET | `/admin/whatsapp-logs` | Logs de WhatsApp |

---

## Flujo Realtime

1. Admin modifica un evento/multimedia/playlist en el dashboard
2. PocketBase hook detecta el cambio (`OnRecordAfterCreateSuccess`, etc.)
3. Hook envía mensaje al WebSocket Hub
4. Hub difunde a todos los clientes conectados:
   - **Web:** `refresh_web` → HTMX re-fetch de fragments
   - **Devices:** `playlist_update` → recarga de playlist
   - **All:** `refresh_all` → todo se actualiza

---

## Modo Kiosk (Dispositivos)

1. Registrar dispositivo en `/admin/devices` con código de 4 dígitos
2. Acceder a `http://tu-server/display/XXXX` desde el dispositivo
3. Hacer clic para entrar en fullscreen
4. El dispositivo:
   - Se conecta por WebSocket (`/ws/device/XXXX`)
   - Auto-refresh cada 30 segundos
   - Aplica el layout CSS Grid configurado
   - Mantiene la pantalla activa (Wake Lock API)

---

## Producción

```bash
# Compilar binario optimizado
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o csl-system cmd/server/main.go

# Ejecutar con systemd
sudo cp csl-system.service /etc/systemd/system/
sudo systemctl enable csl-system
sudo systemctl start csl-system
```

### Nginx reverse proxy (recomendado)

```nginx
server {
    listen 443 ssl;
    server_name colegiosanlorenzo.cl;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Licencia

Propiedad del Colegio San Lorenzo de Copiapó. Todos los derechos reservados.
