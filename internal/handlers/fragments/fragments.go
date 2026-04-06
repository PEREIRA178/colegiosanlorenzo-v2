package fragments

import (
	"fmt"
	"html/template"
	"strings"

	"csl-system/internal/config"

	"github.com/gofiber/fiber/v2"
)

// ══════════════════════════════════════════════════════
//  HTMX FRAGMENT: Hero Carousel
//  GET /fragments/hero
// ══════════════════════════════════════════════════════

func HeroCarousel(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch hero slides from PocketBase multimedia collection
		// where status = "publicado" and tagged as hero
		// For now, return the static hero HTML

		html := `<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <!-- Slides loaded from PocketBase -->
    <div class="hero-slide active">
      <div class="hero-slide-bg slide-grad-1"></div>
      <div class="hero-slide-content">
        <div class="hero-text">
          <p class="hero-eyebrow">Copiapó, Atacama · desde 1968</p>
          <h1 class="hero-title">Per laborem<br/><em>ad lucem</em></h1>
          <p class="hero-desc">Formando generaciones con excelencia académica, valores humanos y el espíritu del norte de Chile.</p>
          <div class="hero-actions">
            <a href="admision.html" class="btn-filled">Admisión 2027</a>
            <a href="nuestro-colegio.html" class="btn-tonal-white">Conocer más</a>
          </div>
        </div>
        <div class="hero-float-card">
          <div class="hero-stat-row">
            <div>
              <div class="hero-stat-num">35+</div>
              <div class="hero-stat-label">años formando<br/>estudiantes</div>
            </div>
            <div>
              <div class="hero-stat-num">3</div>
              <div class="hero-stat-label">niveles:<br/>Básica y Media</div>
            </div>
          </div>
          <div class="hero-lema-float">
            <em>"Por el trabajo hacia la luz"</em>
          </div>
        </div>
      </div>
    </div>
    <!-- More slides dynamically loaded -->
  </div>
  <div class="hero-controls">
    <button class="hero-arrow prev" aria-label="Anterior">‹</button>
    <div class="hero-dots"></div>
    <button class="hero-arrow next" aria-label="Siguiente">›</button>
  </div>
</section>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// ══════════════════════════════════════════════════════
//  HTMX FRAGMENT: Eventos
//  GET /fragments/eventos
// ══════════════════════════════════════════════════════

func Eventos(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch published events from PocketBase
		// events, _ := pbClient.Records("events").GetFullList(
		//   pocketbase.Params{Filter: "status = 'publicado'", Sort: "-date", Limit: 6},
		// )

		// Placeholder: will be replaced with real PB data
		type Evento struct {
			ID        string
			Title     string
			Desc      string
			Category  string
			Date      string
			Urgencia  string
			CSSClass  string
		}

		eventos := []Evento{
			{ID: "1", Title: "Simulacro de Evacuación", Desc: "Jueves 2 de abril, participación obligatoria de todos los cursos.", Category: "Emergencia", Date: "2 abr 2026", Urgencia: "critico", CSSClass: "ev-urgente"},
			{ID: "2", Title: "Reunión de Apoderados 7mos", Desc: "Reunión del primer trimestre 2026. Asistencia obligatoria.", Category: "Reunión", Date: "17 abr 2026", Urgencia: "normal", CSSClass: "ev-reunion"},
			{ID: "3", Title: "Campeonato de tenis padre-hijo", Desc: "Inscripciones abiertas. Una iniciativa que une familias.", Category: "Información", Date: "3 abr 2026", Urgencia: "informativo", CSSClass: "ev-info"},
		}

		var sb strings.Builder
		sb.WriteString(`<section class="eventos-section" id="eventos">`)
		sb.WriteString(`<div class="eventos-header-row">`)
		sb.WriteString(`<div><p class="label-primary reveal">Comunicados y avisos</p>`)
		sb.WriteString(`<h2 class="headline-l reveal reveal-delay-1" style="margin-bottom:0">Últimos eventos</h2></div>`)
		sb.WriteString(`<a href="comunicados.html" class="eventos-link reveal">Ver todos los comunicados →</a>`)
		sb.WriteString(`</div><div class="eventos-grid">`)

		for i, ev := range eventos {
			urgLabel := ""
			if ev.Urgencia == "critico" {
				urgLabel = `<span style="font-size:11px;font-weight:600;color:#B71C1C">URGENTE</span>`
			}
			sb.WriteString(fmt.Sprintf(`
    <div class="evento-card %s reveal%s">
      <div class="evento-accent"></div>
      <div class="evento-body">
        <span class="evento-chip">%s</span>
        <h3>%s</h3>
        <p>%s</p>
      </div>
      <div class="evento-footer">
        <span class="evento-fecha"><span class="evento-fecha-dot"></span>%s</span>
        %s
      </div>
    </div>`,
				ev.CSSClass,
				delayClass(i),
				template.HTMLEscapeString(ev.Category),
				template.HTMLEscapeString(ev.Title),
				template.HTMLEscapeString(ev.Desc),
				ev.Date,
				urgLabel,
			))
		}

		sb.WriteString(`</div></section>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// ══════════════════════════════════════════════════════
//  HTMX FRAGMENT: Noticias
//  GET /fragments/noticias
// ══════════════════════════════════════════════════════

func Noticias(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch from PocketBase news_articles where status = 'publicado' and category = 'noticia'

		html := `<section class="noticias-section" id="noticias">
  <div class="container">
    <p class="label-primary reveal">Noticias y Prensa</p>
    <h2 class="headline-l reveal reveal-delay-1" style="margin-bottom:var(--sp-40)">Últimas noticias</h2>
    <div class="noticias-grid">
      <article class="noticia-card reveal">
        <div class="noticia-img" style="background:var(--md-primary-container)"></div>
        <div class="noticia-content">
          <span class="noticia-cat">Institucional</span>
          <h3>Resultados SIMCE 2025</h3>
          <p>El Colegio San Lorenzo obtuvo resultados destacados en las pruebas SIMCE de 4° y 8° básico.</p>
          <span class="noticia-fecha">28 mar 2026</span>
        </div>
      </article>
      <article class="noticia-card reveal reveal-delay-1">
        <div class="noticia-img" style="background:var(--md-secondary-container)"></div>
        <div class="noticia-content">
          <span class="noticia-cat">Deportes</span>
          <h3>Equipo de fútbol clasifica a regionales</h3>
          <p>Nuestro equipo sub-14 representará a Atacama en el campeonato regional 2026.</p>
          <span class="noticia-fecha">25 mar 2026</span>
        </div>
      </article>
      <article class="noticia-card reveal reveal-delay-2">
        <div class="noticia-img" style="background:var(--md-tertiary-container)"></div>
        <div class="noticia-content">
          <span class="noticia-cat">Cultura</span>
          <h3>Festival de Arte EDEX 2026</h3>
          <p>Más de 200 estudiantes participaron en la muestra artística anual del programa EDEX.</p>
          <span class="noticia-fecha">20 mar 2026</span>
        </div>
      </article>
    </div>
  </div>
</section>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// ══════════════════════════════════════════════════════
//  HTMX FRAGMENT: Blog / Prensa Interna
//  GET /fragments/blog
// ══════════════════════════════════════════════════════

func Blog(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch from PocketBase news_articles where category IN ('blog','prensa')

		html := `<section class="blog-section" id="blog">
  <div class="container">
    <p class="label-primary reveal">Blog & Prensa Interna</p>
    <h2 class="headline-l reveal reveal-delay-1" style="margin-bottom:var(--sp-40)">Voces de la comunidad</h2>
    <div class="blog-grid">
      <article class="blog-card reveal">
        <div class="blog-card-inner">
          <span class="blog-tag">Blog docente</span>
          <h3>Innovación pedagógica en el aula</h3>
          <p>Profesora María González comparte su experiencia implementando metodologías activas en 5° básico.</p>
          <div class="blog-meta">
            <span>María González</span>
            <span>15 mar 2026</span>
          </div>
        </div>
      </article>
      <article class="blog-card reveal reveal-delay-1">
        <div class="blog-card-inner">
          <span class="blog-tag al-aire">🔴 Al aire</span>
          <h3>Día del libro 2026</h3>
          <p>Cobertura en vivo de las actividades del Día del Libro en el Colegio San Lorenzo.</p>
          <div class="blog-meta">
            <span>Equipo CEAL</span>
            <span>Ahora</span>
          </div>
        </div>
      </article>
    </div>
  </div>
</section>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// Helper for staggered reveal classes
func delayClass(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf(" reveal-delay-%d", i)
}
