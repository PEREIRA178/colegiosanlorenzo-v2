package fragments

import (
	"fmt"
	"html/template"
	"strings"

	"csl-system/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// GET /fragments/hero
func HeroCarousel(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := `<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <div class="hero-slide active">
      <div class="hero-slide-bg slide-grad-1"></div>
      <div class="hero-slide-content">
        <div class="hero-text">
          <p class="hero-eyebrow">Copiapó, Atacama · desde 1990</p>
          <h1 class="hero-title">Per laborem<br/><em>ad lucem</em></h1>
          <p class="hero-desc">Formando generaciones con excelencia académica, valores humanos y el espíritu del norte de Chile.</p>
          <div class="hero-actions">
            <a href="admision.html" class="btn-filled">Admisión 2027</a>
            <a href="nuestro-colegio.html" class="btn-tonal-white">Conocer más</a>
          </div>
        </div>
        <div class="hero-float-card">
          <div class="hero-stat-row">
            <div><div class="hero-stat-num">35+</div><div class="hero-stat-label">años formando<br/>estudiantes</div></div>
            <div><div class="hero-stat-num">3</div><div class="hero-stat-label">niveles:<br/>Básica y Media</div></div>
          </div>
          <div class="hero-lema-float"><em>"Por el trabajo hacia la luz"</em></div>
        </div>
      </div>
    </div>
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

// cssClassForCategory maps content_block category to CSS class
func cssClassForCategory(category, urgency string) string {
	if urgency == "true" || category == "EMERGENCIA" {
		return "ev-urgente"
	}
	switch category {
	case "REUNIÓN":
		return "ev-reunion"
	case "EVENTO", "DEPORTIVO":
		return "ev-evento"
	case "ACADÉMICO":
		return "ev-academico"
	default:
		return "ev-info"
	}
}

type contentBlock struct {
	ID       string
	Title    string
	Desc     string
	Category string
	Date     string
	Urgency  bool
	Featured bool
	CSSClass string
}

func fetchContentBlocks(pb *pocketbase.PocketBase, filter string, limit int) []contentBlock {
	records, err := pb.FindRecordsByFilter("content_blocks", filter, "-date", limit, 0)
	if err != nil || len(records) == 0 {
		return nil
	}
	result := make([]contentBlock, 0, len(records))
	for _, r := range records {
		urgency := r.GetBool("urgency")
		urg := ""
		if urgency {
			urg = "true"
		}
		category := r.GetString("category")
		dateStr := ""
		dt := r.GetDateTime("date")
		if !dt.IsZero() {
			dateStr = dt.Time().Format("2 Jan 2006")
		}
		result = append(result, contentBlock{
			ID:       r.Id,
			Title:    r.GetString("title"),
			Desc:     r.GetString("description"),
			Category: category,
			Date:     dateStr,
			Urgency:  urgency,
			CSSClass: cssClassForCategory(category, urg),
		})
	}
	return result
}

// GET /fragments/eventos
func Eventos(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		blocks := fetchContentBlocks(pb,
			"status = 'publicado' && category != 'NOTICIA'", 6)

		if len(blocks) == 0 {
			blocks = []contentBlock{
				{ID: "1", Title: "Simulacro de Evacuación", Desc: "Jueves 2 de abril, participación obligatoria.", Category: "EMERGENCIA", Date: "2 abr 2026", Urgency: true, CSSClass: "ev-urgente"},
				{ID: "2", Title: "Reunión de Apoderados 7mos", Desc: "Reunión primer trimestre 2026. Asistencia obligatoria.", Category: "REUNIÓN", Date: "17 abr 2026", CSSClass: "ev-reunion"},
				{ID: "3", Title: "Campeonato de tenis padre-hijo", Desc: "Inscripciones abiertas. Una iniciativa que une familias.", Category: "EVENTO", Date: "3 abr 2026", CSSClass: "ev-info"},
			}
		}

		var sb strings.Builder
		sb.WriteString(`<section class="eventos-section" id="eventos">`)
		sb.WriteString(`<div class="eventos-header-row">`)
		sb.WriteString(`<div><p class="label-primary reveal visible">Comunicados y avisos</p>`)
		sb.WriteString(`<h2 class="headline-l reveal visible" style="margin-bottom:0">Últimos eventos</h2></div>`)
		sb.WriteString(`<a href="comunicados.html" class="eventos-link reveal visible">Ver todos los comunicados →</a>`)
		sb.WriteString(`</div><div class="eventos-grid">`)

		for i, ev := range blocks {
			urgLabel := ""
			if ev.Urgency {
				urgLabel = `<span style="font-size:11px;font-weight:600;color:#B71C1C">URGENTE</span>`
			}
			sb.WriteString(fmt.Sprintf(`
    <div class="evento-card %s reveal visible%s">
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
				ev.CSSClass, delayClass(i),
				template.HTMLEscapeString(ev.Category),
				template.HTMLEscapeString(ev.Title),
				template.HTMLEscapeString(ev.Desc),
				ev.Date, urgLabel,
			))
		}

		sb.WriteString(`</div></section>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// GET /fragments/noticias
func Noticias(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		blocks := fetchContentBlocks(pb,
			"status = 'publicado' && category = 'NOTICIA'", 3)

		type noticia struct{ Title, Excerpt, Category, Date string }
		var noticias []noticia

		if len(blocks) > 0 {
			for _, b := range blocks {
				noticias = append(noticias, noticia{
					Title: b.Title, Excerpt: b.Desc,
					Category: b.Category, Date: b.Date,
				})
			}
		} else {
			noticias = []noticia{
				{Title: "Resultados SIMCE 2025", Excerpt: "El Colegio San Lorenzo obtuvo resultados destacados en las pruebas SIMCE.", Category: "Institucional", Date: "28 mar 2026"},
				{Title: "Equipo de fútbol clasifica a regionales", Excerpt: "Nuestro equipo sub-14 representará a Atacama en el campeonato regional 2026.", Category: "Deportes", Date: "25 mar 2026"},
				{Title: "Festival de Arte EDEX 2026", Excerpt: "Más de 200 estudiantes participaron en la muestra artística anual.", Category: "Cultura", Date: "20 mar 2026"},
			}
		}

		bgColors := []string{
			"var(--md-primary-container)",
			"var(--md-secondary-container)",
			"var(--md-tertiary-container)",
		}
		delays := []string{"", " reveal-delay-1", " reveal-delay-2"}

		var sb strings.Builder
		sb.WriteString(`<section class="noticias-section" id="noticias"><div class="container">`)
		sb.WriteString(`<p class="label-primary reveal visible">Noticias y Prensa</p>`)
		sb.WriteString(`<h2 class="headline-l reveal visible" style="margin-bottom:var(--sp-40)">Últimas noticias</h2>`)
		sb.WriteString(`<div class="noticias-grid">`)

		for i, n := range noticias {
			sb.WriteString(fmt.Sprintf(`
      <article class="noticia-card reveal visible%s">
        <div class="noticia-img" style="background:%s"></div>
        <div class="noticia-content">
          <span class="noticia-cat">%s</span>
          <h3>%s</h3>
          <p>%s</p>
          <span class="noticia-fecha">%s</span>
        </div>
      </article>`,
				delays[i%len(delays)], bgColors[i%len(bgColors)],
				template.HTMLEscapeString(n.Category),
				template.HTMLEscapeString(n.Title),
				template.HTMLEscapeString(n.Excerpt),
				n.Date,
			))
		}

		sb.WriteString(`</div></div></section>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// GET /fragments/comunicados — full grid for /comunicados.html
func Comunicados(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := c.Query("filter", "todos")

		pbFilter := "status = 'publicado'"
		switch filter {
		case "urgente":
			pbFilter += " && urgency = true"
		case "reunion":
			pbFilter += " && category = 'REUNIÓN'"
		case "info":
			pbFilter += " && category = 'INFORMACIÓN'"
		case "academico":
			pbFilter += " && category = 'ACADÉMICO'"
		case "evento":
			pbFilter += " && (category = 'EVENTO' || category = 'DEPORTIVO')"
		}

		blocks := fetchContentBlocks(pb, pbFilter, 50)

		// Fallback to hardcoded 8 events
		if len(blocks) == 0 {
			blocks = hardcodedComunicados()
		}

		var sb strings.Builder
		sb.WriteString(`<div class="comunicados-grid" id="comunicados-grid">`)
		for i, b := range blocks {
			tipo, chip, accentClass := categoryToTipo(b.Category, b.Urgency)
			sb.WriteString(fmt.Sprintf(`
      <div class="comunicado-card %s reveal visible%s" data-tipo="%s">
        <div class="comunicado-accent"></div>
        <div class="comunicado-body">
          <span class="comunicado-chip">%s</span>
          <h3>%s</h3>
          <p>%s</p>
        </div>
        <div class="comunicado-footer">
          <span class="comunicado-fecha">%s</span>
          <a href="#" class="comunicado-btn">Ver más →</a>
        </div>
      </div>`,
				accentClass, delayClass(i%4), tipo,
				chip,
				template.HTMLEscapeString(b.Title),
				template.HTMLEscapeString(b.Desc),
				b.Date,
			))
		}
		sb.WriteString(`</div>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func categoryToTipo(category string, urgency bool) (tipo, chip, accentClass string) {
	if urgency || category == "EMERGENCIA" {
		return "urgente", "⚠ Emergencia", "tipo-urgente"
	}
	switch category {
	case "REUNIÓN":
		return "reunion", "Reunión", "tipo-reunion"
	case "ACADÉMICO":
		return "academico", "Académico", "tipo-academico"
	case "EVENTO", "DEPORTIVO":
		return "evento", "Evento", "tipo-evento"
	case "NOTICIA":
		return "info", "Noticia", "tipo-info"
	default:
		return "info", "Información", "tipo-info"
	}
}

func hardcodedComunicados() []contentBlock {
	return []contentBlock{
		{Title: "⚠ Simulacro de Evacuación — Jueves 2 de abril", Desc: "Recordamos a toda la comunidad escolar que el jueves 2 de abril se realizará el simulacro de evacuación obligatorio a las 10:00 horas. Participación de todos los cursos.", Category: "EMERGENCIA", Date: "31 de marzo, 2026", Urgency: true},
		{Title: "Reunión de Apoderados 7° Básico", Desc: "Se cita a los apoderados de 7° año básico a reunión del primer trimestre 2026. La reunión se realizará el 17 de abril a las 18:30 hrs en la sala del curso. Asistencia obligatoria.", Category: "REUNIÓN", Date: "17 de abril, 2026"},
		{Title: "Campeonato de Tenis Padre-Hijo", Desc: "Inscripciones abiertas para el campeonato de tenis padre-hijo, 3 de abril de 2026. Una iniciativa del Área Deportiva EDEX que une a las familias.", Category: "EVENTO", Date: "3 de abril, 2026"},
		{Title: "Inicio año escolar 2026 — Nuevas iniciativas pedagógicas", Desc: "El Colegio San Lorenzo inicia el año escolar 2026 con importantes cambios en su propuesta pedagógica, incorporando metodologías activas y herramientas tecnológicas.", Category: "ACADÉMICO", Date: "1 de abril, 2026"},
		{Title: "Sistema Digital Wellness — Comunicación por WhatsApp", Desc: "El colegio informa que toda la comunicación oficial con apoderados se realizará a través del sistema Digital Wellness. Los avisos llegan directamente por WhatsApp.", Category: "INFORMACIÓN", Date: "15 de marzo, 2026"},
		{Title: "Reunión General Enseñanza Media — 6 de mayo", Desc: "Se cita a apoderados de 1° a 4° Medio a reunión general informativa del primer trimestre 2026. Miércoles 6 de mayo a las 19:00 hrs en el gimnasio.", Category: "REUNIÓN", Date: "6 de mayo, 2026"},
		{Title: "Calendario de pruebas primer trimestre 2026 — CEAL", Desc: "El calendario de pruebas del primer trimestre 2026 está disponible en la sección CEAL. Incluye fechas de pruebas, integradoras y exámenes para todos los niveles.", Category: "ACADÉMICO", Date: "5 de marzo, 2026"},
		{Title: "Lista de útiles escolares 2026 disponible en CEPAD", Desc: "Las listas de útiles escolares para todos los niveles del año 2026 ya están disponibles en la sección CEPAD del sitio web del colegio.", Category: "INFORMACIÓN", Date: "1 de marzo, 2026"},
	}
}

// Blog — removed. Returns empty to avoid breaking existing route.
func Blog(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString("")
	}
}

func delayClass(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf(" reveal-delay-%d", i)
}
