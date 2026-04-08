package admin

import (
	"csl-system/internal/auth"
	"csl-system/internal/config"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ── LOGIN ──

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
				`<div class="toast toast-error">Email y contraseña requeridos</div>`,
			)
		}

		if email == "admin@colegiosanlorenzo.cl" && password == "csl2026admin!" {
			token, err := auth.GenerateToken(cfg, "admin-id", email, "superadmin", "Administrador")
			if err != nil {
				return c.Status(500).SendString(`<div class="toast toast-error">Error generando sesión</div>`)
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
			`<div class="toast toast-error">Credenciales incorrectas</div>`,
		)
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:    "csl_token",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
			Path:    "/",
		})
		c.Set("HX-Redirect", "/admin/login")
		return c.SendString("")
	}
}

// ── DASHBOARD ──

func Dashboard(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/dashboard.html")
	}
}

// ── MULTIMEDIA CRUD ──

func MultimediaList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "rows" {
			return c.SendFile("./internal/templates/admin/pages/multimedia.html")
		}
		records, err := pb.FindRecordsByFilter("multimedia", "", "-created", 100, 0)
		var sb strings.Builder
		if err != nil || len(records) == 0 {
			sb.WriteString(`<tr><td colspan="5" style="text-align:center;padding:32px;color:var(--md-outline)">Sin archivos multimedia</td></tr>`)
		} else {
			for _, r := range records {
				estado := r.GetString("estado")
				badgeClass := "badge-info"
				if estado == "publicado" {
					badgeClass = "badge-success"
				} else if estado == "archivado" {
					badgeClass = "badge-warning"
				}
				dur := r.GetFloat("duracion_segundos")
				durStr := "—"
				if dur > 0 {
					durStr = fmt.Sprintf("%.0fs", dur)
				}
				sb.WriteString(fmt.Sprintf(`<tr>
          <td>%s</td><td>%s</td><td>%s</td>
          <td><span class="badge %s">%s</span></td>
          <td>
            <button class="topbar-btn topbar-btn-outline" style="padding:4px 10px;font-size:12px"
              hx-get="/admin/multimedia/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
            <button class="topbar-btn" style="padding:4px 10px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer"
              hx-delete="/admin/multimedia/%s" hx-confirm="¿Eliminar?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
          </td></tr>`,
					template.HTMLEscapeString(r.GetString("filename")),
					template.HTMLEscapeString(r.GetString("type")),
					durStr, badgeClass,
					template.HTMLEscapeString(estado),
					r.Id, r.Id,
				))
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func MultimediaForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := `<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header"><h3>Nuevo Archivo Multimedia</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form hx-post="/admin/multimedia" hx-encoding="multipart/form-data" hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field"><label>Nombre</label><input type="text" name="filename" required class="form-input" placeholder="ej: imagen-bienvenida.jpg"/></div>
      <div class="form-field"><label>Tipo</label>
        <select name="type" class="form-input">
          <option value="imagen">Imagen</option><option value="video">Video</option>
          <option value="youtube">YouTube</option><option value="vimeo">Vimeo</option>
        </select>
      </div>
      <div class="form-field"><label>URL (R2 / CDN)</label><input type="url" name="url_r2" class="form-input" placeholder="https://..."/></div>
      <div class="form-field"><label>Duración (seg)</label><input type="number" name="duracion_segundos" class="form-input" placeholder="10"/></div>
      <div class="form-field"><label>Estado</label>
        <select name="estado" class="form-input">
          <option value="borrador">Borrador</option><option value="publicado">Publicado</option>
        </select>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar</button>
      </div>
    </form>
  </div>
</div>`
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func MultimediaCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("multimedia")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		r.Set("filename", c.FormValue("filename"))
		r.Set("type", c.FormValue("type"))
		r.Set("url_r2", c.FormValue("url_r2"))
		r.Set("duracion_segundos", c.FormValue("duracion_segundos"))
		r.Set("estado", c.FormValue("estado"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error guardando</div>`)
		}
		c.Set("HX-Trigger", "mediaCreated")
		c.Set("HX-Reswap", "innerHTML")
		document := `<div class="toast toast-success" id="toast-area">✅ Archivo guardado</div>
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';document.getElementById('toast-area').remove()},1500)</script>`
		return c.SendString(document)
	}
}

func MultimediaEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("multimedia", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		html := fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header"><h3>Editar Multimedia</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form hx-put="/admin/multimedia/%s" hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field"><label>Nombre</label><input type="text" name="filename" value="%s" required class="form-input"/></div>
      <div class="form-field"><label>Tipo</label>
        <select name="type" class="form-input">
          <option value="imagen"%s>Imagen</option><option value="video"%s>Video</option>
          <option value="youtube"%s>YouTube</option><option value="vimeo"%s>Vimeo</option>
        </select>
      </div>
      <div class="form-field"><label>URL (R2)</label><input type="url" name="url_r2" value="%s" class="form-input"/></div>
      <div class="form-field"><label>Estado</label>
        <select name="estado" class="form-input">
          <option value="borrador"%s>Borrador</option><option value="publicado"%s>Publicado</option>
        </select>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Actualizar</button>
      </div>
    </form>
  </div>
</div>`,
			r.Id,
			template.HTMLEscapeString(r.GetString("filename")),
			sel(r.GetString("type"), "imagen"),
			sel(r.GetString("type"), "video"),
			sel(r.GetString("type"), "youtube"),
			sel(r.GetString("type"), "vimeo"),
			template.HTMLEscapeString(r.GetString("url_r2")),
			sel(r.GetString("estado"), "borrador"),
			sel(r.GetString("estado"), "publicado"),
		)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func MultimediaUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("multimedia", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		r.Set("filename", c.FormValue("filename"))
		r.Set("type", c.FormValue("type"))
		r.Set("url_r2", c.FormValue("url_r2"))
		r.Set("estado", c.FormValue("estado"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		c.Set("HX-Trigger", "mediaUpdated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Actualizado
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';document.getElementById('toast-area').remove()},1500)</script></div>`)
	}
}

func MultimediaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("multimedia", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		c.Set("HX-Trigger", "mediaDeleted")
		return c.SendString("")
	}
}

// ── EVENTS CRUD (content_blocks) ──

func EventsList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "rows" {
			return c.SendFile("./internal/templates/admin/pages/events.html")
		}
		records, err := pb.FindRecordsByFilter("content_blocks",
			"category != 'NOTICIA'", "-date", 100, 0)

		var sb strings.Builder
		if err != nil || len(records) == 0 {
			sb.WriteString(`<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--md-outline)">Sin eventos — agrega uno con el botón de arriba</td></tr>`)
		} else {
			for _, r := range records {
				status := r.GetString("status")
				badgeClass := "badge-warning"
				if status == "publicado" {
					badgeClass = "badge-success"
				}
				urgIcon := ""
				if r.GetBool("urgency") {
					urgIcon = `<span class="badge badge-danger">URGENTE</span>`
				}
				dateStr := "—"
				if dt := r.GetDateTime("date"); !dt.IsZero() {
					dateStr = dt.Time().Format("2 Jan 2006")
				}
				sb.WriteString(fmt.Sprintf(`<tr>
          <td>%s</td><td>%s</td><td>%s</td><td>%s</td>
          <td><span class="badge %s">%s</span></td>
          <td>
            <button class="topbar-btn topbar-btn-outline" style="padding:4px 10px;font-size:12px"
              hx-get="/admin/events/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
            <button class="topbar-btn" style="padding:4px 10px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer"
              hx-delete="/admin/events/%s" hx-confirm="¿Eliminar este evento?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
          </td></tr>`,
					template.HTMLEscapeString(r.GetString("title")),
					template.HTMLEscapeString(r.GetString("category")),
					urgIcon, dateStr,
					badgeClass, template.HTMLEscapeString(status),
					r.Id, r.Id,
				))
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func EventForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := eventFormHTML("", "", "", "", "", "", false)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func EventCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("content_blocks")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		r.Set("title", c.FormValue("title"))
		r.Set("description", c.FormValue("description"))
		r.Set("category", c.FormValue("category"))
		r.Set("urgency", c.FormValue("urgency") == "on")
		if ds := c.FormValue("date"); ds != "" {
			if t, err2 := time.Parse("2006-01-02T15:04", ds); err2 == nil {
				r.Set("date", t.UTC())
			}
		}
		r.Set("featured", c.FormValue("featured") == "on")
		r.Set("status", c.FormValue("status"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error guardando</div>`)
		}
		c.Set("HX-Trigger", "eventCreated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Evento creado
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';document.getElementById('toast-area')?.remove()},1500)</script></div>`)
	}
}

func EventEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		dateStr := ""
		if dt := r.GetDateTime("date"); !dt.IsZero() {
			dateStr = dt.Time().Format("2006-01-02T15:04")
		}
		html := eventFormHTML(
			r.Id,
			r.GetString("title"),
			r.GetString("description"),
			r.GetString("category"),
			r.GetString("status"),
			dateStr,
			r.GetBool("urgency"),
		)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func EventUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		r.Set("title", c.FormValue("title"))
		r.Set("description", c.FormValue("description"))
		r.Set("category", c.FormValue("category"))
		r.Set("urgency", c.FormValue("urgency") == "on")
		if ds := c.FormValue("date"); ds != "" {
			if t, err2 := time.Parse("2006-01-02T15:04", ds); err2 == nil {
				r.Set("date", t.UTC())
			}
		}
		r.Set("featured", c.FormValue("featured") == "on")
		r.Set("status", c.FormValue("status"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		c.Set("HX-Trigger", "eventUpdated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Evento actualizado
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';document.getElementById('toast-area')?.remove()},1500)</script></div>`)
	}
}

func EventDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		c.Set("HX-Trigger", "eventDeleted")
		return c.SendString("")
	}
}

func EventPublish(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		r.Set("status", "publicado")
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error publicando</div>`)
		}
		c.Set("HX-Trigger", "eventUpdated")
		return c.SendString(`<div class="toast toast-success">✅ Publicado</div>`)
	}
}

// eventFormHTML builds the create/edit modal form
func eventFormHTML(id, title, description, category, status, date string, urgency bool) string {
	action := "/admin/events"
	method := `hx-post="/admin/events"`
	if id != "" {
		action = fmt.Sprintf("/admin/events/%s", id)
		method = fmt.Sprintf(`hx-put="/admin/events/%s"`, id)
	}
	_ = action

	cats := []string{"EMERGENCIA", "REUNIÓN", "INFORMACIÓN", "ACADÉMICO", "EVENTO", "DEPORTIVO", "NOTICIA"}
	var catOpts strings.Builder
	for _, cat := range cats {
		selected := ""
		if cat == category {
			selected = " selected"
		}
		catOpts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, cat, selected, cat))
	}

	urgChecked := ""
	if urgency {
		urgChecked = " checked"
	}

	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field"><label>Título</label><input type="text" name="title" value="%s" required class="form-input" placeholder="Título del comunicado"/></div>
      <div class="form-field"><label>Descripción</label><textarea name="description" class="form-input" rows="3" placeholder="Descripción...">%s</textarea></div>
      <div class="form-row">
        <div class="form-field"><label>Categoría</label><select name="category" class="form-input">%s</select></div>
        <div class="form-field"><label>Estado</label>
          <select name="status" class="form-input">
            <option value="borrador"%s>Borrador</option>
            <option value="publicado"%s>Publicado</option>
            <option value="archivado"%s>Archivado</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Fecha</label><input type="datetime-local" name="date" value="%s" class="form-input"/></div>
        <div class="form-field" style="justify-content:flex-end;flex-direction:row;align-items:center;gap:8px;padding-top:24px">
          <input type="checkbox" name="urgency" id="urgency-check"%s style="width:16px;height:16px"/>
          <label for="urgency-check" style="font-size:13px;font-weight:500">Urgente</label>
        </div>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar</button>
      </div>
    </form>
  </div>
</div>`,
		map[bool]string{true: "Editar Evento", false: "Nuevo Evento"}[id != ""],
		method,
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		catOpts.String(),
		sel(status, "borrador"),
		sel(status, "publicado"),
		sel(status, "archivado"),
		date,
		urgChecked,
	)
}

// ── NEWS (delegates to content_blocks with category=NOTICIA) ──

func NewsList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "rows" {
			return c.SendFile("./internal/templates/admin/pages/news.html")
		}
		records, err := pb.FindRecordsByFilter("content_blocks",
			"category = 'NOTICIA'", "-date", 50, 0)

		var sb strings.Builder
		if err != nil || len(records) == 0 {
			sb.WriteString(`<tr><td colspan="5" style="text-align:center;padding:32px;color:var(--md-outline)">Sin noticias</td></tr>`)
		} else {
			for _, r := range records {
				status := r.GetString("status")
				badgeClass := "badge-warning"
				if status == "publicado" {
					badgeClass = "badge-success"
				}
				feat := ""
				if r.GetBool("featured") {
					feat = `<span class="badge badge-info">⭐</span>`
				}
				dateStr := "—"
				if dt := r.GetDateTime("date"); !dt.IsZero() {
					dateStr = dt.Time().Format("2 Jan 2006")
				}
				sb.WriteString(fmt.Sprintf(`<tr>
          <td>%s</td><td>%s</td>
          <td><span class="badge %s">%s</span></td><td>%s</td>
          <td>
            <button class="topbar-btn topbar-btn-outline" style="padding:4px 10px;font-size:12px"
              hx-get="/admin/events/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
            <button class="topbar-btn" style="padding:4px 10px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer"
              hx-delete="/admin/events/%s" hx-confirm="¿Eliminar?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
          </td></tr>`,
					template.HTMLEscapeString(r.GetString("title")),
					dateStr, badgeClass,
					template.HTMLEscapeString(status),
					feat, r.Id, r.Id,
				))
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func NewsForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := eventFormHTML("", "", "", "NOTICIA", "borrador", "", false)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// News create/edit/delete/update reuse Events handlers since both use content_blocks
func NewsCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return EventCreate(cfg, pb)
}
func NewsEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return EventEdit(cfg, pb)
}
func NewsUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return EventUpdate(cfg, pb)
}
func NewsDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return EventDelete(cfg, pb)
}

// ── PLAYLISTS ──

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
		return c.SendString(`<div class="toast toast-success">Orden actualizado</div>`)
	}
}

// ── DEVICES ──

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
		return c.SendString(`<div class="toast toast-success">Playlist asignada</div>`)
	}
}

// ── USERS ──

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

// ── WHATSAPP LOGS ──

func WhatsAppLogs(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./internal/templates/admin/pages/whatsapp-logs.html")
	}
}

// sel returns " selected" if val == target
func sel(val, target string) string {
	if val == target {
		return " selected"
	}
	return ""
}
