package admin

import (
	"csl-system/internal/auth"
	"csl-system/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ══════════════════════════════════════════════════════
//  LOGIN
// ══════════════════════════════════════════════════════

func LoginPage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/login.html")
	}
}

func LoginSubmit(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		remember := c.FormValue("remember") == "on"

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).SendString(
				`<div class="toast toast-error" id="login-error">Email y contraseña requeridos</div>`,
			)
		}

		// TODO: Authenticate against PocketBase users collection
		// pbResponse, err := pbClient.AuthWithPassword("users", email, password)
		// For now, hardcoded check for development:
		if email == "admin@colegiosanlorenzo.cl" && password == "csl2026admin!" {
			token, err := auth.GenerateToken(cfg, "admin-id", email, "superadmin", "Administrador")
			if err != nil {
				return c.Status(500).SendString(
					`<div class="toast toast-error">Error generando sesión</div>`,
				)
			}

			expiry := 24 * time.Hour
			if remember {
				expiry = 72 * time.Hour
			}

			c.Cookie(&fiber.Cookie{
				Name:     "csl_token",
				Value:    token,
				Expires:  time.Now().Add(expiry),
				HTTPOnly: true,
				Secure:   cfg.Env == "production",
				SameSite: "Lax",
				Path:     "/",
			})

			c.Set("HX-Redirect", "/admin/dashboard")
			return c.SendString("")
		}

		return c.Status(fiber.StatusUnauthorized).SendString(
			`<div class="toast toast-error" id="login-error">Credenciales incorrectas</div>`,
		)
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "csl_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Path:     "/",
		})
		c.Set("HX-Redirect", "/admin/login")
		return c.SendString("")
	}
}

// ══════════════════════════════════════════════════════
//  DASHBOARD
// ══════════════════════════════════════════════════════

func Dashboard(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/dashboard.html")
	}
}

// ══════════════════════════════════════════════════════
//  MULTIMEDIA CRUD
// ══════════════════════════════════════════════════════

func MultimediaList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch from PocketBase
		return c.SendFile("./internal/templates/admin/pages/multimedia.html")
	}
}

func MultimediaForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Nuevo Multimedia</h2><!-- HTMX form -->`)
	}
}

func MultimediaCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Upload to R2, create PocketBase record
		return c.SendString(`<div class="toast toast-success">Multimedia creado</div>`)
	}
}

func MultimediaEdit(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// id := c.Params("id")
		return c.SendString(`<h2>Editar Multimedia</h2><!-- form -->`)
	}
}

func MultimediaUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Multimedia actualizado</div>`)
	}
}

func MultimediaDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Multimedia eliminado</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  EVENTS CRUD
// ══════════════════════════════════════════════════════

func EventsList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/events.html")
	}
}

func EventForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Nuevo Evento</h2>`)
	}
}

func EventCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Evento creado</div>`)
	}
}

func EventEdit(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Editar Evento</h2>`)
	}
}

func EventUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Evento actualizado</div>`)
	}
}

func EventDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Evento eliminado</div>`)
	}
}

func EventPublish(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Update status to "publicado", trigger WhatsApp if targets set
		return c.SendString(`<div class="toast toast-success">Evento publicado</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  NEWS CRUD
// ══════════════════════════════════════════════════════

func NewsList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/news.html")
	}
}

func NewsForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Nueva Noticia</h2>`)
	}
}

func NewsCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Noticia creada</div>`)
	}
}

func NewsEdit(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Editar Noticia</h2>`)
	}
}

func NewsUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Noticia actualizada</div>`)
	}
}

func NewsDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Noticia eliminada</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  PLAYLISTS
// ══════════════════════════════════════════════════════

func PlaylistList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/playlists.html")
	}
}

func PlaylistForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Nueva Playlist</h2>`)
	}
}

func PlaylistCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Playlist creada</div>`)
	}
}

func PlaylistEdit(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<h2>Editar Playlist</h2>`)
	}
}

func PlaylistUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Playlist actualizada</div>`)
	}
}

func PlaylistDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Playlist eliminada</div>`)
	}
}

func PlaylistReorder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Parse new order from request body, update playlist_items.orden
		return c.SendString(`<div class="toast toast-success">Orden actualizado</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  DEVICES
// ══════════════════════════════════════════════════════

func DeviceList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/devices.html")
	}
}

func DeviceCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Dispositivo creado</div>`)
	}
}

func DeviceUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Dispositivo actualizado</div>`)
	}
}

func DeviceDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Dispositivo eliminado</div>`)
	}
}

func DeviceAssignPlaylist(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Update device.playlist_id, broadcast to device
		return c.SendString(`<div class="toast toast-success">Playlist asignada</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  USERS
// ══════════════════════════════════════════════════════

func UserList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/users.html")
	}
}

func UserCreate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Usuario creado</div>`)
	}
}

func UserUpdate(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Usuario actualizado</div>`)
	}
}

func UserDelete(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Usuario eliminado</div>`)
	}
}

// ══════════════════════════════════════════════════════
//  WHATSAPP LOGS
// ══════════════════════════════════════════════════════

func WhatsAppLogs(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/whatsapp-logs.html")
	}
}
