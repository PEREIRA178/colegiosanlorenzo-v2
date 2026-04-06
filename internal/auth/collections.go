package auth

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RegisterPBHooks sets up PocketBase collections and auth hooks.
// Called before pb.Start() to ensure schema exists on first boot.
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
	// ══════════════════════════════════════════════════════
	//  1. USERS (auth collection with role field)
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("users"); err != nil {
		col := core.NewAuthCollection("users")
		col.Fields.Add(
			&core.TextField{
				Name:     "role",
				Required: true,
			},
			&core.TextField{Name: "nombre"},
			&core.TextField{Name: "telefono"},
			&core.TextField{Name: "rut"},
			&core.BoolField{Name: "activo"},
		)
		// Auth settings
		col.AuthToken.Duration = 259200 // 72h in seconds
		col.PasswordAuth.Enabled = true
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'users' created")
	}

	// ══════════════════════════════════════════════════════
	//  2. MULTIMEDIA
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("multimedia"); err != nil {
		col := core.NewBaseCollection("multimedia")
		col.Fields.Add(
			&core.TextField{Name: "filename", Required: true},
			&core.URLField{Name: "url_r2"},
			&core.TextField{Name: "type", Required: true},        // "imagen"|"video"|"youtube"|"vimeo"
			&core.NumberField{Name: "size"},                       // bytes
			&core.TextField{Name: "uploaded_by"},                  // user id
			&core.TextField{Name: "estado"},                       // "borrador"|"publicado"|"archivado"
			&core.TextField{Name: "descripcion"},
			&core.NumberField{Name: "duracion_segundos"},          // display duration
			&core.FileField{Name: "thumbnail", MaxSelect: 1, MaxSize: 5242880},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'multimedia' created")
	}

	// ══════════════════════════════════════════════════════
	//  3. EVENTS
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("events"); err != nil {
		col := core.NewBaseCollection("events")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.EditorField{Name: "description"},
			&core.DateField{Name: "date"},
			&core.TextField{Name: "category"},                     // "reunion"|"emergencia"|"noticia"|"informacion"|"concurso"
			&core.TextField{Name: "urgencia"},                     // "critico"|"normal"|"informativo"
			&core.TextField{Name: "status"},                       // "borrador"|"publicado"|"archivado"
			&core.TextField{Name: "targets_whatsapp"},             // comma-separated phone numbers or group ids
			&core.BoolField{Name: "whatsapp_sent"},
			&core.DateField{Name: "whatsapp_sent_at"},
			&core.TextField{Name: "created_by"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'events' created")
	}

	// ══════════════════════════════════════════════════════
	//  4. NEWS_ARTICLES
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("news_articles"); err != nil {
		col := core.NewBaseCollection("news_articles")
		col.Fields.Add(
			&core.TextField{Name: "title", Required: true},
			&core.TextField{Name: "slug"},
			&core.EditorField{Name: "body"},
			&core.TextField{Name: "excerpt"},
			&core.TextField{Name: "category"},                     // "noticia"|"blog"|"prensa"
			&core.TextField{Name: "status"},                       // "borrador"|"publicado"
			&core.FileField{Name: "cover_image", MaxSelect: 1, MaxSize: 10485760},
			&core.TextField{Name: "author"},
			&core.BoolField{Name: "featured"},
			&core.BoolField{Name: "al_aire"},                      // "on air" flag for live news
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'news_articles' created")
	}

	// ══════════════════════════════════════════════════════
	//  5. PLAYLISTS
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("playlists"); err != nil {
		col := core.NewBaseCollection("playlists")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "description"},
			&core.TextField{Name: "status"},                       // "activa"|"inactiva"
			&core.TextField{Name: "created_by"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlists' created")
	}

	// ══════════════════════════════════════════════════════
	//  6. PLAYLIST_ITEMS (junction table)
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("playlist_items"); err != nil {
		col := core.NewBaseCollection("playlist_items")
		col.Fields.Add(
			&core.TextField{Name: "playlist_id", Required: true},
			&core.TextField{Name: "multimedia_id"},
			&core.TextField{Name: "tipo"},                         // "multimedia"|"eventos"|"noticias"
			&core.NumberField{Name: "orden", Required: true},
			&core.NumberField{Name: "duracion_segundos"},          // display duration
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'playlist_items' created")
	}

	// ══════════════════════════════════════════════════════
	//  7. DEVICES
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("devices"); err != nil {
		col := core.NewBaseCollection("devices")
		col.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "type", Required: true},         // "pantalla"|"totem"
			&core.TextField{Name: "code", Required: true},         // 4-digit unique code
			&core.TextField{Name: "layout"},                       // "1col"|"2col-50"|"3col-33"
			&core.TextField{Name: "ubicacion"},
			&core.TextField{Name: "playlist_id"},                  // assigned playlist
			&core.TextField{Name: "status"},                       // "online"|"offline"
			&core.DateField{Name: "last_seen"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'devices' created")
	}

	// ══════════════════════════════════════════════════════
	//  8. FORM_RESPONSES
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("form_responses"); err != nil {
		col := core.NewBaseCollection("form_responses")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "user_id"},
			&core.TextField{Name: "tipo"},                         // "confirmacion"|"duda"|"no-asistencia"
			&core.TextField{Name: "valor"},
			&core.TextField{Name: "mensaje"},
			&core.BoolField{Name: "leido"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'form_responses' created")
	}

	// ══════════════════════════════════════════════════════
	//  9. WHATSAPP_LOGS
	// ══════════════════════════════════════════════════════
	if _, err := app.FindCollectionByNameOrId("whatsapp_logs"); err != nil {
		col := core.NewBaseCollection("whatsapp_logs")
		col.Fields.Add(
			&core.TextField{Name: "event_id"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "message_sid"},
			&core.TextField{Name: "status"},                       // "queued"|"sent"|"delivered"|"read"|"failed"
			&core.TextField{Name: "direction"},                    // "outbound"|"inbound"
			&core.TextField{Name: "body"},
			&core.TextField{Name: "error_message"},
		)
		if err := app.Save(col); err != nil {
			return err
		}
		log.Println("  ✅ Collection 'whatsapp_logs' created")
	}

	// ══════════════════════════════════════════════════════
	//  10. Create default superadmin if none exists
	// ══════════════════════════════════════════════════════
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
				log.Println("  ✅ Default superadmin created (admin@colegiosanlorenzo.cl)")
			}
		}
	}

	return nil
}

// Unused but kept for reference:
// Available role values: "superadmin", "director", "secretaria", "profesor", "invitado"
var _ = types.DateTime{}
