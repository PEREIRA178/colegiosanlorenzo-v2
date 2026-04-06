package main

import (
	"log"
	"os"

	"csl-system/internal/auth"
	"csl-system/internal/config"
	"csl-system/internal/handlers/admin"
	"csl-system/internal/handlers/fragments"
	"csl-system/internal/handlers/web"
	"csl-system/internal/handlers/ws"
	"csl-system/internal/middleware"
	"csl-system/internal/realtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gows "github.com/gofiber/websocket/v2"
	"github.com/pocketbase/pocketbase"
)

func main() {
	// ── Load config ──
	cfg := config.Load()

	// ── PocketBase (embedded) ──
	pb := pocketbase.New()

	// Register collections & hooks before starting PB
	auth.RegisterPBHooks(pb)
	realtime.RegisterPBHooks(pb)

	// Start PocketBase in background (serves its own admin UI on :8090)
	go func() {
		if err := pb.Start(); err != nil {
			log.Fatalf("PocketBase failed: %v", err)
		}
	}()

	// ── Fiber app ──
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("ERROR [%d] %s %s: %v", code, c.Method(), c.Path(), err)
			if c.Get("HX-Request") == "true" {
				return c.Status(code).SendString(`<div class="toast toast-error">Error interno</div>`)
			}
			return c.Status(code).SendString("Error interno del servidor")
		},
		BodyLimit: 50 * 1024 * 1024, // 50MB for media uploads
	})

	// ── Global middleware ──
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} ${method} ${path} (${latency})\n",
		TimeFormat: "15:04:05",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, HX-Request, HX-Trigger",
	}))

	// ── Static files ──
	app.Static("/static", "./static", fiber.Static{
		Compress:      true,
		CacheDuration: cfg.StaticCacheDuration,
	})

	// ── Realtime hub ──
	hub := realtime.NewHub()
	go hub.Run()

	// ══════════════════════════════════════════════════════
	//  PUBLIC WEB ROUTES
	// ══════════════════════════════════════════════════════
	app.Get("/", web.IndexHandler(cfg))
	app.Get("/nuestro-colegio.html", web.PageHandler(cfg, "nuestro-colegio"))
	app.Get("/admision.html", web.PageHandler(cfg, "admision"))
	app.Get("/comunicados.html", web.PageHandler(cfg, "comunicados"))
	app.Get("/edex.html", web.PageHandler(cfg, "edex"))
	app.Get("/cepad.html", web.PageHandler(cfg, "cepad"))
	app.Get("/inclusion.html", web.PageHandler(cfg, "inclusion"))
	app.Get("/ceal.html", web.PageHandler(cfg, "ceal"))

	// ── HTMX fragment endpoints (public) ──
	frag := app.Group("/fragments")
	frag.Get("/hero", fragments.HeroCarousel(cfg))
	frag.Get("/eventos", fragments.Eventos(cfg))
	frag.Get("/noticias", fragments.Noticias(cfg))
	frag.Get("/blog", fragments.Blog(cfg))

	// ── RSS feed ──
	app.Get("/rss.xml", web.RSSFeed(cfg))

	// ══════════════════════════════════════════════════════
	//  DEVICE DISPLAY ROUTES
	// ══════════════════════════════════════════════════════
	app.Get("/display/:code", web.DeviceDisplay(cfg))

	// ── WebSocket upgrade check ──
	app.Use("/ws", func(c *fiber.Ctx) error {
		if gows.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/device/:code", gows.New(ws.DeviceSocket(hub)))
	app.Get("/ws/web", gows.New(ws.WebSocket(hub)))

	// ══════════════════════════════════════════════════════
	//  ADMIN / DASHBOARD ROUTES (protected)
	// ══════════════════════════════════════════════════════
	app.Get("/admin/login", admin.LoginPage(cfg))
	app.Post("/admin/login", admin.LoginSubmit(cfg))
	app.Post("/admin/logout", admin.Logout())

	adm := app.Group("/admin", middleware.AuthRequired(cfg))

	// Dashboard
	adm.Get("/", admin.Dashboard(cfg))
	adm.Get("/dashboard", admin.Dashboard(cfg))

	// Multimedia CRUD
	adm.Get("/multimedia", admin.MultimediaList(cfg))
	adm.Get("/multimedia/new", admin.MultimediaForm(cfg))
	adm.Post("/multimedia", admin.MultimediaCreate(cfg))
	adm.Get("/multimedia/:id/edit", admin.MultimediaEdit(cfg))
	adm.Put("/multimedia/:id", admin.MultimediaUpdate(cfg))
	adm.Delete("/multimedia/:id", admin.MultimediaDelete(cfg))

	// Events CRUD
	adm.Get("/events", admin.EventsList(cfg))
	adm.Get("/events/new", admin.EventForm(cfg))
	adm.Post("/events", admin.EventCreate(cfg))
	adm.Get("/events/:id/edit", admin.EventEdit(cfg))
	adm.Put("/events/:id", admin.EventUpdate(cfg))
	adm.Delete("/events/:id", admin.EventDelete(cfg))
	adm.Post("/events/:id/publish", admin.EventPublish(cfg))

	// News CRUD
	adm.Get("/news", admin.NewsList(cfg))
	adm.Get("/news/new", admin.NewsForm(cfg))
	adm.Post("/news", admin.NewsCreate(cfg))
	adm.Get("/news/:id/edit", admin.NewsEdit(cfg))
	adm.Put("/news/:id", admin.NewsUpdate(cfg))
	adm.Delete("/news/:id", admin.NewsDelete(cfg))

	// Playlists
	adm.Get("/playlists", admin.PlaylistList(cfg))
	adm.Get("/playlists/new", admin.PlaylistForm(cfg))
	adm.Post("/playlists", admin.PlaylistCreate(cfg))
	adm.Get("/playlists/:id/edit", admin.PlaylistEdit(cfg))
	adm.Put("/playlists/:id", admin.PlaylistUpdate(cfg))
	adm.Delete("/playlists/:id", admin.PlaylistDelete(cfg))
	adm.Post("/playlists/:id/items/reorder", admin.PlaylistReorder(cfg))

	// Devices
	adm.Get("/devices", admin.DeviceList(cfg))
	adm.Post("/devices", admin.DeviceCreate(cfg))
	adm.Put("/devices/:id", admin.DeviceUpdate(cfg))
	adm.Delete("/devices/:id", admin.DeviceDelete(cfg))
	adm.Post("/devices/:id/assign-playlist", admin.DeviceAssignPlaylist(cfg))

	// Users (superadmin/director only)
	adm.Get("/users", middleware.RoleRequired("superadmin", "director"), admin.UserList(cfg))
	adm.Post("/users", middleware.RoleRequired("superadmin", "director"), admin.UserCreate(cfg))
	adm.Put("/users/:id", middleware.RoleRequired("superadmin", "director"), admin.UserUpdate(cfg))
	adm.Delete("/users/:id", middleware.RoleRequired("superadmin"), admin.UserDelete(cfg))

	// WhatsApp logs
	adm.Get("/whatsapp-logs", admin.WhatsAppLogs(cfg))

	// ── Webhook (WhatsApp inbound) ──
	app.Post("/webhook/whatsapp", web.WhatsAppWebhook(cfg))

	// ── Start server ──
	port := cfg.Port
	if port == "" {
		port = "3000"
	}

	log.Printf("🏫 CSL System en http://localhost:%s", port)
	log.Printf("📊 Dashboard: http://localhost:%s/admin", port)
	log.Printf("🔧 PocketBase Admin: http://localhost:8090/_/", port)

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Fatal(app.Listen(":" + port))
}
