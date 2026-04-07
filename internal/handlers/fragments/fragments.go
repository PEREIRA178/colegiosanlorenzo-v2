package fragments

import (
	"fmt"
	"html/template"
	"strings"

	"csl-system/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
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

func Eventos(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Evento struct {
			ID       string
			Title    string
			Desc     string
			Category string
			Date     string
			Urgencia string
			CSSClass string
		}

		// Fallback hardcoded data
		fallback := []Evento{
			{ID: "1", Title: "Simulacro de Evacuación", Desc: "Jueves 2 de abril, participación obligatoria de todos los cursos.", Category: "Emergencia", Date: "2 abr 2026", Urgencia: "critico", CSSClass: "ev-urgente"},
			{ID: "2", Title: "Reunión de Apoderados 7mos", Desc: "Reunión del primer trimestre 2026. Asistencia obligatoria.", Category: "Reunión", Date: "17 abr 2026", Urgencia: "normal", CSSClass: "ev-reunion"},
			{ID: "3", Title: "Campeonato de tenis padre-hijo", Desc: "Inscripciones abiertas. Una iniciativa que une familias.", Category: "Información", Date: "3 abr 2026", Urgencia: "informativo", CSSClass: "ev-info"},
		}

		var eventos []Evento

		records, err := pb.FindRecordsByFilter("events", "status = 'publicado'", "-date", 6, 0)
		if err == nil && len(records) > 0 {
			for _, record := range records {
				urgencia := record.GetString("urgencia")
				cssClass := "ev-info"
				switch urgencia {
				case "critico":
					cssClass = "ev-urgente"
				case "normal":
					cssClass = "ev-reunion"
				}
				dateTime := record.GetDateTime("date")
				dateStr := dateTime.Time().Format("2 Jan 2006")

				eventos = append(eventos, Evento{
					ID:       record.Id,
					Title:    record.GetString("title"),
					Desc:     record.GetString("description"),
					Category: record.GetString("category"),
					Date:     dateStr,
					Urgencia: urgencia,
					CSSClass: cssClass,
				})
			}
		} else {
			eventos = fallback
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

func Noticias(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Noticia struct {
			Title    string
			Excerpt  string
			Category string
			Date     string
		}

		fallback := []Noticia{
			{Title: "Resultados SIMCE 2025", Excerpt: "El Colegio San Lorenzo obtuvo resultados destacados en las pruebas SIMCE de 4° y 8° básico.", Category: "Institucional", Date: "28 mar 2026"},
			{Title: "Equipo de fútbol clasifica a regionales", Excerpt: "Nuestro equipo sub-14 representará a Atacama en el campeonato regional 2026.", Category: "Deportes", Date: "25 mar 2026"},
			{Title: "Festival de Arte EDEX 2026", Excerpt: "Más de 200 estudiantes participaron en la muestra artística anual del programa EDEX.", Category: "Cultura", Date: "20 mar 2026"},
		}

		var noticias []Noticia

		records, err := pb.FindRecordsByFilter("news_articles", "status = 'publicado' && category != 'blog'", "-created", 3, 0)
		if err == nil && len(records) > 0 {
			for _, record := range records {
				dateTime := record.GetDateTime("created")
				dateStr := dateTime.Time().Format("2 Jan 2006")
				noticias = append(noticias, Noticia{
					Title:    record.GetString("title"),
					Excerpt:  record.GetString("excerpt"),
					Category: record.GetString("category"),
					Date:     dateStr,
				})
			}
		} else {
			noticias = fallback
		}

		bgColors := []string{
			"var(--md-primary-container)",
			"var(--md-secondary-container)",
			"var(--md-tertiary-container)",
		}
		delayClasses := []string{"", " reveal-delay-1", " reveal-delay-2"}

		var sb strings.Builder
		sb.WriteString(`<section class="noticias-section" id="noticias">`)
		sb.WriteString(`  <div class="container">`)
		sb.WriteString(`    <p class="label-primary reveal">Noticias y Prensa</p>`)
		sb.WriteString(`    <h2 class="headline-l reveal reveal-delay-1" style="margin-bottom:var(--sp-40)">Últimas noticias</h2>`)
		sb.WriteString(`    <div class="noticias-grid">`)

		for i, n := range noticias {
			bg := bgColors[i%len(bgColors)]
			dc := delayClasses[i%len(delayClasses)]
			sb.WriteString(fmt.Sprintf(`
      <article class="noticia-card reveal%s">
        <div class="noticia-img" style="background:%s"></div>
        <div class="noticia-content">
          <span class="noticia-cat">%s</span>
          <h3>%s</h3>
          <p>%s</p>
          <span class="noticia-fecha">%s</span>
        </div>
      </article>`,
				dc,
				bg,
				template.HTMLEscapeString(n.Category),
				template.HTMLEscapeString(n.Title),
				template.HTMLEscapeString(n.Excerpt),
				n.Date,
			))
		}

		sb.WriteString(`    </div>`)
		sb.WriteString(`  </div>`)
		sb.WriteString(`</section>`)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// ══════════════════════════════════════════════════════
//  HTMX FRAGMENT: Blog / Prensa Interna
//  GET /fragments/blog
// ══════════════════════════════════════════════════════

func Blog(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type BlogPost struct {
			Title   string
			Excerpt string
			Author  string
			Date    string
			AlAire  bool
		}

		fallback := []BlogPost{
			{Title: "Innovación pedagógica en el aula", Excerpt: "Profesora María González comparte su experiencia implementando metodologías activas en 5° básico.", Author: "María González", Date: "15 mar 2026", AlAire: false},
			{Title: "Día del libro 2026", Excerpt: "Cobertura en vivo de las actividades del Día del Libro en el Colegio San Lorenzo.", Author: "Equipo CEAL", Date: "Ahora", AlAire: true},
		}

		var posts []BlogPost

		records, err := pb.FindRecordsByFilter("news_articles", "status = 'publicado' && category = 'blog'", "-created", 2, 0)
		if err == nil && len(records) > 0 {
			for _, record := range records {
				alAire := record.GetBool("al_aire")
				dateStr := "Ahora"
				if !alAire {
					dateTime := record.GetDateTime("created")
					dateStr = dateTime.Time().Format("2 Jan 2006")
				}
				posts = append(posts, BlogPost{
					Title:   record.GetString("title"),
					Excerpt: record.GetString("excerpt"),
					Author:  record.GetString("author"),
					Date:    dateStr,
					AlAire:  alAire,
				})
			}
		} else {
			posts = fallback
		}

		delayClasses := []string{"", " reveal-delay-1"}

		var sb strings.Builder
		sb.WriteString(`<section class="blog-section" id="blog">`)
		sb.WriteString(`  <div class="container">`)
		sb.WriteString(`    <p class="label-primary reveal">Blog & Prensa Interna</p>`)
		sb.WriteString(`    <h2 class="headline-l reveal reveal-delay-1" style="margin-bottom:var(--sp-40)">Voces de la comunidad</h2>`)
		sb.WriteString(`    <div class="blog-grid">`)

		for i, p := range posts {
			dc := delayClasses[i%len(delayClasses)]
			tagHTML := `<span class="blog-tag">Blog docente</span>`
			if p.AlAire {
				tagHTML = `<span class="blog-tag al-aire">🔴 Al aire</span>`
			}
			sb.WriteString(fmt.Sprintf(`
      <article class="blog-card reveal%s">
        <div class="blog-card-inner">
          %s
          <h3>%s</h3>
          <p>%s</p>
          <div class="blog-meta">
            <span>%s</span>
            <span>%s</span>
          </div>
        </div>
      </article>`,
				dc,
				tagHTML,
				template.HTMLEscapeString(p.Title),
				template.HTMLEscapeString(p.Excerpt),
				template.HTMLEscapeString(p.Author),
				p.Date,
			))
		}

		sb.WriteString(`    </div>`)
		sb.WriteString(`  </div>`)
		sb.WriteString(`</section>`)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// Helper for staggered reveal classes
func delayClass(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf(" reveal-delay-%d", i)
}
