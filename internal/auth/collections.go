package auth

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RegisterPBHooks sets up PocketBase collections and auth hooks.
func RegisterPBHooks(pb *pocketbase.PocketBase) {
	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		log.Println("📦 PocketBase: Verificando colecciones...")
		if err := ensureCollections(se.App); err != nil {
			log.Printf("⚠️  Error creando colecciones: %v", err)
		}
		return se.Next()
	})
}

func ensureCollections(app core.App) error {
	// ── 1. USERS ──
	if _, err := app.FindCollectionByNameOrId("users"); err != nil {
		col := core.NewAuthCollection("users")
		col.Fields.Add(
			&core.TextField{Name: "role", Required: true},
			&core.TextField{Name: "nombre"},
			&core.TextField{Name: "telefono"},
			&core.TextField{Name: "rut"},
			&core.BoolField{Name: "activo"},
		)
		col.AuthToken.Duration = 259200
		col.PasswordAuth.Enabled = true
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'users' created")
	}

	// ── 2. MEDIA (biblioteca multimedia → R2) ──
	if _, err := app.FindCollectionByNameOrId("media"); err != nil {
		col := core.NewBaseCollection("media")
		col.Fields.Add(
			&core.TextField{Name: "filename", Required: true},
			&core.URLField{Name: "url_r2"},
			&core.TextField{Name: "type", Required: true}, // imagen|video|youtube|vimeo
			&core.NumberField{Name: "size"},
			&core.TextField{Name: "uploaded_by"},
			&core.TextField{Name: "status"}, // borrador|publicado|archivado
			&core.TextField{Name: "description"},
			&core.NumberField{Name: "duration_seconds"},
			&core.FileField{Name: "thumbnail", MaxSelect: 1, MaxSize: 5242880},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'media' created")
	}

	// ── 3. CONTENT_BLOCKS (unified: eventos, noticias, comunicados) ──
	if _, err := app.FindCollectionByNameOrId("content_blocks"); err != nil {
		col := core.NewBaseCollection("content_blocks")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.EditorField{Name: "description"},
			// EMERGENCIA|REUNIÓN|INFORMACIÓN|ACADÉMICO|EVENTO|DEPORTIVO|NOTICIA
			&core.TextField{Name: "category"},
			&core.BoolField{Name: "urgency"},
			&core.DateField{Name: "date"},
			&core.BoolField{Name: "featured"},
			&core.TextField{Name: "status"}, // borrador|publicado|archivado
			&core.TextField{Name: "media_ids"}, // comma-separated media IDs
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'content_blocks' created")
	}

	// ── 4. MULTIMEDIA (legacy, kept for playlists) ──
	if _, err := app.FindCollectionByNameOrId("multimedia"); err != nil {
		col := core.NewBaseCollection("multimedia")
		col.Fields.Add(
			&core.TextField{Name: "filename", Required: true},
			&core.URLField{Name: "url_r2"},
			&core.TextField{Name: "type", Required: true},
			&core.NumberField{Name: "size"},
			&core.TextField{Name: "uploaded_by"},
			&core.TextField{Name: "estado"},
			&core.TextField{Name: "descripcion"},
			&core.NumberField{Name: "duracion_segundos"},
			&core.FileField{Name: "thumbnail", MaxSelect: 1, MaxSize: 5242880},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'multimedia' created")
	}

	// ── 5. PLAYLISTS ──
	if _, err := app.FindCollectionByNameOrId("playlists"); err != nil {
		col := core.NewBaseCollection("playlists")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "status"},
			&core.TextField{Name: "created_by"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlists' created")
	}

	// ── 6. PLAYLIST_ITEMS ──
	if _, err := app.FindCollectionByNameOrId("playlist_items"); err != nil {
		col := core.NewBaseCollection("playlist_items")
		col.Fields.Add(
			&core.TextField{Name: "playlist_id", Required: true},
			&core.TextField{Name: "multimedia_id"},
			&core.TextField{Name: "tipo"},
			&core.NumberField{Name: "orden", Required: true},
			&core.NumberField{Name: "duracion_segundos"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlist_items' created")
	}

	// ── 7. DEVICES ──
	if _, err := app.FindCollectionByNameOrId("devices"); err != nil {
		col := core.NewBaseCollection("devices")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "type", Required: true},
			&core.TextField{Name: "code", Required: true},
			&core.TextField{Name: "layout"},
			&core.TextField{Name: "ubicacion"},
			&core.TextField{Name: "playlist_id"},
			&core.TextField{Name: "status"},
			&core.DateField{Name: "last_seen"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'devices' created")
	}

	// ── 8. FORM_RESPONSES ──
	if _, err := app.FindCollectionByNameOrId("form_responses"); err != nil {
		col := core.NewBaseCollection("form_responses")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "user_id"},
			&core.TextField{Name: "tipo"},
			&core.TextField{Name: "valor"},
			&core.TextField{Name: "mensaje"},
			&core.BoolField{Name: "leido"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'form_responses' created")
	}

	// ── 9. WHATSAPP_LOGS ──
	if _, err := app.FindCollectionByNameOrId("whatsapp_logs"); err != nil {
		col := core.NewBaseCollection("whatsapp_logs")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "message_sid"},
			&core.TextField{Name: "status"},
			&core.TextField{Name: "direction"},
			&core.TextField{Name: "body"},
			&core.TextField{Name: "error_message"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'whatsapp_logs' created")
	}

	// ── 10. Default superadmin ──
	users, _ := app.FindCollectionByNameOrId("users")
	if users != nil {
		records, err := app.FindRecordsByFilter(users, "role = 'superadmin'", "", 1, 0)
		if err != nil || len(records) == 0 {
			record := core.NewRecord(users)
			record.Set("email", "admin@colegiosanlorenzo.cl")
			record.Set("password", "csl2026admin!")
			record.Set("passwordConfirm", "csl2026admin!")
			record.Set("nombre", "Administrador")
			record.Set("role", "superadmin")
			record.Set("activo", true)
			record.Set("verified", true)
			if err := app.Save(record); err != nil {
				log.Printf("⚠️  Error creating superadmin: %v", err)
			} else {
				log.Println("  ✅ Default superadmin created")
			}
		}
	}

	// ── 11. Seed content_blocks ──
	if err := seedContentBlocks(app); err != nil {
		log.Printf("⚠️  Error seeding content_blocks: %v", err)
	}

	return nil
}

type seedBlock struct {
	title       string
	description string
	category    string
	urgency     bool
	date        string
	featured    bool
}

func seedContentBlocks(app core.App) error {
	col, err := app.FindCollectionByNameOrId("content_blocks")
	if err != nil {
		return err
	}

	existing, _ := app.FindRecordsByFilter(col, "status = 'publicado'", "", 1, 0)
	if len(existing) > 0 {
		return nil // already seeded
	}

	events := []seedBlock{
		{
			title:       "⚠ Simulacro de Evacuación — Jueves 2 de abril",
			description: "Recordamos a toda la comunidad escolar que el jueves 2 de abril se realizará el simulacro de evacuación obligatorio a las 10:00 horas. Participación de todos los cursos.",
			category:    "EMERGENCIA",
			urgency:     true,
			date:        "2026-04-02 10:00:00",
			featured:    true,
		},
		{
			title:       "Reunión de Apoderados 7° Básico",
			description: "Se cita a los apoderados de 7° año básico a reunión del primer trimestre 2026. La reunión se realizará el 17 de abril a las 18:30 hrs en la sala del curso. Asistencia obligatoria.",
			category:    "REUNIÓN",
			urgency:     false,
			date:        "2026-04-17 18:30:00",
			featured:    true,
		},
		{
			title:       "Campeonato de Tenis Padre-Hijo",
			description: "Inscripciones abiertas para el campeonato de tenis padre-hijo, 3 de abril de 2026. Una iniciativa del Área Deportiva EDEX que une a las familias. Inscripciones en secretaría.",
			category:    "EVENTO",
			urgency:     false,
			date:        "2026-04-03 09:00:00",
			featured:    true,
		},
		{
			title:       "Inicio año escolar 2026 — Nuevas iniciativas pedagógicas",
			description: "El Colegio San Lorenzo inicia el año escolar 2026 con importantes cambios en su propuesta pedagógica, incorporando metodologías activas, trabajo por proyectos y herramientas tecnológicas en el aula.",
			category:    "ACADÉMICO",
			urgency:     false,
			date:        "2026-04-01 08:00:00",
			featured:    false,
		},
		{
			title:       "Sistema Digital Wellness — Comunicación por WhatsApp",
			description: "El colegio informa que toda la comunicación oficial con apoderados se realizará a través del sistema Digital Wellness. Los avisos llegan directamente por WhatsApp. Consulta en secretaría.",
			category:    "INFORMACIÓN",
			urgency:     false,
			date:        "2026-03-15 00:00:00",
			featured:    false,
		},
		{
			title:       "Reunión General Enseñanza Media — 6 de mayo",
			description: "Se cita a apoderados de 1° a 4° Medio a reunión general informativa del primer trimestre 2026. Información sobre evaluaciones integradoras y proceso PAES 2026. Miércoles 6 de mayo a las 19:00 hrs en el gimnasio.",
			category:    "REUNIÓN",
			urgency:     false,
			date:        "2026-05-06 19:00:00",
			featured:    false,
		},
		{
			title:       "Calendario de pruebas primer trimestre 2026 — CEAL",
			description: "Se informa a los apoderados que el calendario de pruebas del primer trimestre 2026 está disponible en la sección CEAL del sitio web. Incluye fechas de pruebas, integradoras y exámenes para todos los niveles.",
			category:    "ACADÉMICO",
			urgency:     false,
			date:        "2026-03-05 00:00:00",
			featured:    false,
		},
		{
			title:       "Lista de útiles escolares 2026 disponible en CEPAD",
			description: "Las listas de útiles escolares para todos los niveles del año 2026 ya están disponibles en la sección CEPAD del sitio web del colegio. Descarga la lista correspondiente al nivel de tu hijo/a.",
			category:    "INFORMACIÓN",
			urgency:     false,
			date:        "2026-03-01 00:00:00",
			featured:    false,
		},
	}

	news := []seedBlock{
		{
			title:       "Resultados SIMCE 2025 — Colegio San Lorenzo entre los mejores de Atacama",
			description: "El Colegio San Lorenzo obtuvo resultados destacados en las pruebas SIMCE de 4° y 8° básico, posicionándose entre los establecimientos de mejor rendimiento en la Región de Atacama.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-28 00:00:00",
			featured:    true,
		},
		{
			title:       "Equipo sub-14 clasifica al Campeonato Regional de Fútbol",
			description: "Nuestro equipo de fútbol sub-14 representará a Atacama en el campeonato regional 2026 tras ganar la etapa comunal con resultados históricos.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-25 00:00:00",
			featured:    true,
		},
		{
			title:       "Festival de Arte EDEX 2026 — Más de 200 estudiantes en escena",
			description: "Más de 200 estudiantes participaron en la muestra artística anual del programa EDEX, mostrando sus talentos en música, danza, teatro y artes visuales.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-20 00:00:00",
			featured:    false,
		},
		{
			title:       "Nuevo laboratorio de ciencias equipado con tecnología 2026",
			description: "El Colegio San Lorenzo inaugura su nuevo laboratorio de ciencias con equipamiento moderno, beneficiando a más de 400 estudiantes de enseñanza media.",
			category:    "NOTICIA",
			urgency:     false,
			date:        "2026-03-10 00:00:00",
			featured:    false,
		},
	}

	all := append(events, news...)
	for _, b := range all {
		r := core.NewRecord(col)
		r.Set("title", b.title)
		r.Set("description", b.description)
		r.Set("category", b.category)
		r.Set("urgency", b.urgency)
		r.Set("date", b.date)
		r.Set("featured", b.featured)
		r.Set("status", "publicado")
		if err := app.Save(r); err != nil {
			log.Printf("⚠️  seed block error: %v", err)
		}
	}
	log.Printf("  ✅ Seeded %d content_blocks", len(all))
	return nil
}

var _ = types.DateTime{}
