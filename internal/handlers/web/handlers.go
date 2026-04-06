package web

import (
	"fmt"
	"time"

	"csl-system/internal/config"

	"github.com/gofiber/fiber/v2"
)

// IndexHandler serves the main index.html (with HTMX fragment placeholders)
func IndexHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./web/index.html")
	}
}

// PageHandler serves static sub-pages
func PageHandler(cfg *config.Config, page string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile(fmt.Sprintf("./web/%s.html", page))
	}
}

// DeviceDisplay serves the kiosk mode display for a device
func DeviceDisplay(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" || len(code) != 4 {
			return c.Status(404).SendString("Código de dispositivo inválido")
		}
		// TODO: Fetch device config from PocketBase, render display template
		return c.SendFile("./internal/templates/devices/display.html")
	}
}

// RSSFeed generates an RSS feed from published events and news
func RSSFeed(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch from PocketBase events + news_articles where status = 'publicado'
		now := time.Now().Format(time.RFC1123Z)

		rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Colegio San Lorenzo — Noticias y Eventos</title>
    <link>%s</link>
    <description>Comunicados, eventos y noticias del Colegio San Lorenzo de Copiapó</description>
    <language>es-cl</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss.xml" rel="self" type="application/rss+xml"/>
    <!-- Items from PocketBase -->
    <item>
      <title>Simulacro de Evacuación — 2 de abril</title>
      <link>%s/comunicados.html</link>
      <description>Recordamos a toda la comunidad escolar el simulacro de evacuación obligatorio.</description>
      <pubDate>%s</pubDate>
      <guid>%s/events/1</guid>
    </item>
  </channel>
</rss>`, cfg.BaseURL, now, cfg.BaseURL, cfg.BaseURL, now, cfg.BaseURL)

		c.Set("Content-Type", "application/rss+xml; charset=utf-8")
		return c.SendString(rss)
	}
}

// WhatsAppWebhook handles inbound WhatsApp messages
func WhatsAppWebhook(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Parse Twilio webhook, process response, log to whatsapp_logs
		// Ollama hook: process message with AI for auto-response
		return c.SendString("<Response></Response>")
	}
}
