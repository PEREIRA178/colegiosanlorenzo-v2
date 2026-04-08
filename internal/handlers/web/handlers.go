package web

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"csl-system/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
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
		return c.SendString("<Response></Response>")
	}
}

// NoticiaHandler renders a single news article as a blog entry page
func NoticiaHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil || r.GetString("category") != "NOTICIA" || r.GetString("status") != "publicado" {
			return c.Redirect("/", fiber.StatusFound)
		}

		title := r.GetString("title")
		desc := r.GetString("description")
		body := r.GetString("body")
		imageURL := r.GetString("image_url")

		dateStr := ""
		if dt := r.GetDateTime("date"); !dt.IsZero() {
			dateStr = dt.Time().Format("2 de January de 2006")
		}

		// Build image section
		imgHTML := `<div style="width:100%;aspect-ratio:16/6;background:var(--md-primary-container);border-radius:24px;margin-bottom:48px;display:flex;align-items:center;justify-content:center"><span style="font-family:'DM Serif Display',Georgia,serif;font-size:80px;color:rgba(155,18,48,0.15)">SL</span></div>`
		if imageURL != "" {
			imgHTML = fmt.Sprintf(`<div style="width:100%%;aspect-ratio:16/6;border-radius:24px;margin-bottom:48px;overflow:hidden"><img src="%s" style="width:100%%;height:100%%;object-fit:cover" alt="%s"/></div>`,
				template.HTMLEscapeString(imageURL), template.HTMLEscapeString(title))
		}

		// Build body content
		var bodyParts []string
		if body != "" {
			for _, p := range strings.Split(strings.TrimSpace(body), "\n\n") {
				p = strings.TrimSpace(p)
				if p != "" {
					bodyParts = append(bodyParts, "<p>"+template.HTMLEscapeString(p)+"</p>")
				}
			}
		} else if desc != "" {
			bodyParts = append(bodyParts, "<p>"+template.HTMLEscapeString(desc)+"</p>")
		}
		bodyHTML := strings.Join(bodyParts, "\n")

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>%s — Colegio San Lorenzo · Copiapó</title>
<link rel="icon" href="/static/logo.webp" type="image/webp"/>
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link href="https://fonts.googleapis.com/css2?family=DM+Serif+Display:ital@0;1&family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;0,9..40,600&display=swap" rel="stylesheet"/>
<style>
:root{--md-primary:#9B1230;--md-on-primary:#fff;--md-primary-container:#F5E0E4;--md-on-primary-container:#5C0A1E;--md-primary-dark:#6E0C22;--md-surface:#FAF8F7;--md-surface-bright:#FFFFFF;--md-on-surface:#1C1B1A;--md-on-surface-variant:#4D4B4C;--md-outline:#7E7B7C;--md-outline-variant:#CEC9CA;--md-inverse-surface:#312F30;--r-full:9999px;--font-display:'DM Serif Display',Georgia,serif;--font-body:'DM Sans',system-ui,sans-serif;--max-w:1200px;--nav-h:72px;--ease-express:cubic-bezier(0.05,0.7,0.1,1.0);--glass-bg:rgba(250,248,247,0.92);--glass-blur:blur(20px) saturate(180%);--glass-border:rgba(155,18,48,0.10);--glass-shadow:0 8px 32px rgba(155,18,48,0.08)}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
html{scroll-behavior:smooth;-webkit-font-smoothing:antialiased}
body{font-family:var(--font-body);background:var(--md-surface);color:var(--md-on-surface);line-height:1.6;overflow-x:hidden}
.nav{position:sticky;top:0;z-index:200;height:var(--nav-h);background:var(--glass-bg);backdrop-filter:var(--glass-blur);-webkit-backdrop-filter:var(--glass-blur);border-bottom:1px solid var(--glass-border);box-shadow:var(--glass-shadow)}
.nav-inner{max-width:var(--max-w);margin:0 auto;padding:0 24px;height:100%%;display:flex;align-items:center;justify-content:space-between;gap:16px}
.nav-brand{display:flex;align-items:center;gap:12px;text-decoration:none}
.nav-logo{width:44px;height:44px;border-radius:12px;object-fit:contain}
.nav-brand-name{font-family:var(--font-display);font-size:16px;color:var(--md-on-surface);line-height:1.1}
.nav-brand-sub{font-size:10px;font-weight:500;letter-spacing:.12em;color:var(--md-outline);text-transform:uppercase}
.nav-links{display:flex;align-items:center;gap:2px}
.nav-links a{font-size:13.5px;color:var(--md-on-surface-variant);text-decoration:none;padding:8px 14px;border-radius:var(--r-full);transition:background 200ms,color 200ms}
.nav-links a:hover{background:var(--md-primary-container);color:var(--md-on-primary-container)}
.nav-cta{font-size:13px;font-weight:500;color:#fff;background:var(--md-primary);padding:10px 22px;border-radius:var(--r-full);text-decoration:none;flex-shrink:0}
.article-wrap{max-width:760px;margin:64px auto 96px;padding:0 24px}
.article-back{display:inline-flex;align-items:center;gap:6px;font-size:13px;font-weight:500;color:var(--md-on-surface-variant);text-decoration:none;margin-bottom:32px;transition:color 200ms}
.article-back:hover{color:var(--md-primary)}
.article-cat{font-size:11px;font-weight:600;letter-spacing:.14em;text-transform:uppercase;color:var(--md-primary);margin-bottom:12px}
.article-title{font-family:var(--font-display);font-size:clamp(28px,4vw,46px);line-height:1.1;color:var(--md-on-surface);margin-bottom:16px}
.article-meta{font-size:13px;color:var(--md-outline);margin-bottom:40px;padding-bottom:32px;border-bottom:1px solid var(--md-outline-variant)}
.article-body p{font-size:17px;line-height:1.85;color:var(--md-on-surface-variant);margin-bottom:20px}
footer{background:var(--md-inverse-surface);color:rgba(255,255,255,0.55);padding:40px 24px;text-align:center;font-size:13px}
@media(max-width:768px){:root{--nav-h:60px}.nav-links{display:none}}
</style>
</head>
<body>
<nav class="nav">
  <div class="nav-inner">
    <a href="/" class="nav-brand">
      <img src="/static/logo.webp" class="nav-logo" alt="Logo Colegio San Lorenzo"/>
      <div><div class="nav-brand-name">Colegio San Lorenzo</div><div class="nav-brand-sub">Copiapó · Atacama</div></div>
    </a>
    <div class="nav-links">
      <a href="/nuestro-colegio.html">Nuestro Colegio</a>
      <a href="/ceal.html">CEAL</a>
      <a href="/cepad.html">CEPAD</a>
      <a href="/edex.html">EDEX</a>
      <a href="/inclusion.html">Inclusión</a>
      <a href="/comunicados.html">Comunicados</a>
    </div>
    <a href="/admision.html" class="nav-cta">Admisión 2027</a>
  </div>
</nav>
<main>
  <div class="article-wrap">
    <a href="/" class="article-back">← Volver al inicio</a>
    %s
    <p class="article-cat">Noticia</p>
    <h1 class="article-title">%s</h1>
    <p class="article-meta">%s</p>
    <div class="article-body">%s</div>
  </div>
</main>
<footer>
  <p>© 2026 Colegio San Lorenzo · Copiapó, Atacama &nbsp;·&nbsp; <a href="/comunicados.html" style="color:rgba(255,255,255,0.6);text-decoration:none">Comunicados</a></p>
</footer>
</body>
</html>`,
			template.HTMLEscapeString(title),
			imgHTML,
			template.HTMLEscapeString(title),
			template.HTMLEscapeString(dateStr),
			bodyHTML,
		)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}
