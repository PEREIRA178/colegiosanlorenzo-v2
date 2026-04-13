package admin

import (
	"csl-system/internal/auth"
	"csl-system/internal/config"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
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
		records, err := pb.FindRecordsByFilter("multimedia", "", "filename", 100, 0)
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
      <div class="form-field"><label>URL del archivo</label><input type="url" name="url_r2" class="form-input" placeholder="https://... URL completa del video o imagen"/></div>
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
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Archivo guardado
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/multimedia?fragment=rows',{target:'#media-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
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
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';var _ta=document.getElementById('toast-area');if(_ta)_ta.innerHTML=''},1500)</script></div>`)
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
			// Direct browser access (F5 / bookmark): serve full SPA shell.
			// HTMX navigation sets HX-Request; the JS in dashboard auto-loads this panel.
			if c.Get("HX-Request") != "true" {
				return c.SendFile("./internal/templates/admin/pages/dashboard.html")
			}
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
		html := eventFormHTML("", "", "", "", "", "", "")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func EventCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		title := c.FormValue("title")
		description := c.FormValue("description")
		if len([]rune(title)) > 120 {
			return c.SendString(`<div class="toast toast-error">El título no puede superar 120 caracteres</div>`)
		}
		if len([]rune(description)) > 600 {
			return c.SendString(`<div class="toast toast-error">La descripción no puede superar 600 caracteres</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("content_blocks")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		r.Set("title", title)
		r.Set("description", description)
		r.Set("category", c.FormValue("category"))
		r.Set("urgency", c.FormValue("category") == "EMERGENCIA")
		if ds := c.FormValue("date"); ds != "" {
			if t, err2 := time.Parse("2006-01-02T15:04", ds); err2 == nil {
				r.Set("date", t.UTC())
			}
		}
		r.Set("featured", c.FormValue("featured") == "on")
		r.Set("status", c.FormValue("status"))
		r.Set("pdf_url", c.FormValue("pdf_url"))
		r.Set("image_url", c.FormValue("image_url"))
		r.Set("body", c.FormValue("body"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error guardando</div>`)
		}
		c.Set("HX-Trigger", "eventCreated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Evento creado
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';var _ta=document.getElementById('toast-area');if(_ta)_ta.innerHTML=''},1500)</script></div>`)
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
		var html string
		if r.GetString("category") == "NOTICIA" {
			html = newsFormHTML(r.Id, r.GetString("title"), r.GetString("description"),
				r.GetString("status"), dateStr, r.GetString("image_url"), r.GetString("body"))
		} else {
			html = eventFormHTML(r.Id, r.GetString("title"), r.GetString("description"),
				r.GetString("category"), r.GetString("status"), dateStr,
				r.GetString("pdf_url"))
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func EventUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		title := c.FormValue("title")
		description := c.FormValue("description")
		if len([]rune(title)) > 120 {
			return c.SendString(`<div class="toast toast-error">El título no puede superar 120 caracteres</div>`)
		}
		if len([]rune(description)) > 600 {
			return c.SendString(`<div class="toast toast-error">La descripción no puede superar 600 caracteres</div>`)
		}
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		r.Set("title", title)
		r.Set("description", description)
		r.Set("category", c.FormValue("category"))
		r.Set("urgency", c.FormValue("category") == "EMERGENCIA")
		if ds := c.FormValue("date"); ds != "" {
			if t, err2 := time.Parse("2006-01-02T15:04", ds); err2 == nil {
				r.Set("date", t.UTC())
			}
		}
		r.Set("featured", c.FormValue("featured") == "on")
		r.Set("status", c.FormValue("status"))
		r.Set("pdf_url", c.FormValue("pdf_url"))
		r.Set("image_url", c.FormValue("image_url"))
		r.Set("body", c.FormValue("body"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		c.Set("HX-Trigger", "eventUpdated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Evento actualizado
<script>setTimeout(()=>{document.getElementById('modal-container').innerHTML='';var _ta=document.getElementById('toast-area');if(_ta)_ta.innerHTML=''},1500)</script></div>`)
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

// eventFormHTML builds the create/edit modal form for events (all categories except NOTICIA).
// Urgency is auto-derived from category=EMERGENCIA — no checkbox needed.
func eventFormHTML(id, title, description, category, status, date, pdfUrl string) string {
	method := `hx-post="/admin/events"`
	if id != "" {
		method = fmt.Sprintf(`hx-put="/admin/events/%s"`, id)
	}

	cats := []string{"EMERGENCIA", "REUNIÓN", "INFORMACIÓN", "ACADÉMICO", "EVENTO", "DEPORTIVO"}
	var catOpts strings.Builder
	for _, cat := range cats {
		selected := ""
		if cat == category {
			selected = " selected"
		}
		catOpts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, cat, selected, cat))
	}

	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field">
        <label>Título <small id="ev-title-cnt" style="float:right;font-size:11px;color:var(--md-outline)">0/120</small></label>
        <input type="text" name="title" value="%s" required class="form-input" maxlength="120" placeholder="Título del comunicado"
          oninput="(function(el){var c=document.getElementById('ev-title-cnt');if(c){c.textContent=el.value.length+'/120';c.style.color=el.value.length>100?'#B71C1C':''}})(this)"
          onchange="(function(el){var c=document.getElementById('ev-title-cnt');if(c)c.textContent=el.value.length+'/120'})(this)"/>
      </div>
      <div class="form-field">
        <label>Descripción <small id="ev-desc-cnt" style="float:right;font-size:11px;color:var(--md-outline)">0/600</small></label>
        <textarea name="description" class="form-input" rows="3" maxlength="600" placeholder="Descripción breve..."
          oninput="(function(el){var c=document.getElementById('ev-desc-cnt');if(c){c.textContent=el.value.length+'/600';c.style.color=el.value.length>500?'#B71C1C':''}})( this)">%s</textarea>
      </div>
      <div class="form-field"><label>Link PDF (opcional)</label><input type="url" name="pdf_url" value="%s" class="form-input" placeholder="https://... enlace al comunicado PDF"/></div>
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
      <div class="form-field"><label>Fecha</label><input type="datetime-local" name="date" value="%s" class="form-input"/></div>
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
		template.HTMLEscapeString(pdfUrl),
		catOpts.String(),
		sel(status, "borrador"),
		sel(status, "publicado"),
		sel(status, "archivado"),
		date,
	)
}

// newsFormHTML builds the create/edit modal form for noticias (category=NOTICIA only)
func newsFormHTML(id, title, description, status, date, imageUrl, body string) string {
	method := `hx-post="/admin/news"`
	if id != "" {
		method = fmt.Sprintf(`hx-put="/admin/news/%s"`, id)
	}
	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <input type="hidden" name="category" value="NOTICIA"/>
      <div class="form-field">
        <label>Título <small id="nw-title-cnt" style="float:right;font-size:11px;color:var(--md-outline)">0/120</small></label>
        <input type="text" name="title" value="%s" required class="form-input" maxlength="120" placeholder="Título de la noticia"
          oninput="(function(el){var c=document.getElementById('nw-title-cnt');if(c){c.textContent=el.value.length+'/120';c.style.color=el.value.length>100?'#B71C1C':''}})(this)"/>
      </div>
      <div class="form-field">
        <label>Resumen <small id="nw-desc-cnt" style="float:right;font-size:11px;color:var(--md-outline)">0/600</small></label>
        <textarea name="description" class="form-input" rows="2" maxlength="600" placeholder="Resumen breve..."
          oninput="(function(el){var c=document.getElementById('nw-desc-cnt');if(c){c.textContent=el.value.length+'/600';c.style.color=el.value.length>500?'#B71C1C':''}})( this)">%s</textarea>
      </div>
      <div class="form-field"><label>Contenido completo</label><textarea name="body" class="form-input" rows="6" placeholder="Escribe el artículo completo aquí...">%s</textarea></div>
      <div class="form-field"><label>URL Foto principal</label><input type="url" name="image_url" value="%s" class="form-input" placeholder="https://... enlace a la imagen"/></div>
      <div class="form-row">
        <div class="form-field"><label>Fecha</label><input type="datetime-local" name="date" value="%s" class="form-input"/></div>
        <div class="form-field"><label>Estado</label>
          <select name="status" class="form-input">
            <option value="borrador"%s>Borrador</option>
            <option value="publicado"%s>Publicado</option>
            <option value="archivado"%s>Archivado</option>
          </select>
        </div>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar</button>
      </div>
    </form>
  </div>
</div>`,
		map[bool]string{true: "Editar Noticia", false: "Nueva Noticia"}[id != ""],
		method,
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		template.HTMLEscapeString(body),
		template.HTMLEscapeString(imageUrl),
		date,
		sel(status, "borrador"),
		sel(status, "publicado"),
		sel(status, "archivado"),
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
		html := newsFormHTML("", "", "", "borrador", "", "", "")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

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

func PlaylistList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch c.Query("fragment") {
		case "list":
			records, _ := pb.FindRecordsByFilter("playlists", "", "name", 100, 0)
			var sb strings.Builder
			if len(records) == 0 {
				sb.WriteString(`<tr><td colspan="4" style="text-align:center;padding:32px;color:var(--md-outline)">Sin playlists — crea una nueva con el botón de arriba</td></tr>`)
			} else {
				for _, r := range records {
					status := r.GetString("status")
					badgeClass := "badge-warning"
					if status == "activa" {
						badgeClass = "badge-success"
					}
					items, _ := pb.FindRecordsByFilter("playlist_items", "playlist_id='"+r.Id+"'", "", 100, 0)
					sb.WriteString(fmt.Sprintf(`<tr>
						<td style="font-weight:500">%s</td>
						<td><span class="badge %s">%s</span></td>
						<td style="color:var(--md-outline)">%d items</td>
						<td><div style="display:flex;gap:6px">
							<button class="topbar-btn topbar-btn-outline" style="padding:3px 8px;font-size:12px"
								hx-get="/admin/playlists/%s/edit" hx-target="#playlist-editor-area" hx-swap="innerHTML">Editar</button>
							<button style="padding:3px 8px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer;border-radius:9999px;font-family:inherit"
								hx-delete="/admin/playlists/%s" hx-confirm="¿Eliminar playlist?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
						</div></td>
					</tr>`,
						template.HTMLEscapeString(r.GetString("name")),
						badgeClass, template.HTMLEscapeString(status),
						len(items),
						r.Id, r.Id,
					))
				}
			}
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(sb.String())

		case "content":
			return buildContentPool(c, pb)

		default:
			if c.Get("HX-Request") != "true" {
				return c.SendFile("./internal/templates/admin/pages/dashboard.html")
			}
			return c.SendFile("./internal/templates/admin/pages/playlists.html")
		}
	}
}

func PlaylistForm(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(playlistEditorHTML("", "", "[]"))
	}
}

func PlaylistCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := strings.TrimSpace(c.FormValue("name"))
		if name == "" {
			return c.SendString(`<div class="toast toast-error">El nombre es requerido</div>`)
		}
		var items []plItemInput
		if ij := c.FormValue("items_json"); ij != "" {
			if err := json.Unmarshal([]byte(ij), &items); err != nil {
				log.Printf("⚠️  PlaylistCreate items_json parse: %v", err)
			}
		}
		plCol, err := pb.FindCollectionByNameOrId("playlists")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		pl := core.NewRecord(plCol)
		pl.Set("name", name)
		pl.Set("status", "activa")
		if err := pb.Save(pl); err != nil {
			return c.SendString(`<div class="toast toast-error">Error guardando playlist</div>`)
		}
		savePlItems(pb, pl.Id, items)
		c.Set("HX-Trigger", "playlistCreated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Playlist creada
<script>document.getElementById('playlist-editor-area').innerHTML='';htmx.trigger(document.body,'playlistCreated');setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000)</script></div>`)
	}
}

func PlaylistEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("playlists", id)
		if err != nil {
			return c.Status(404).SendString(`<p style="padding:24px;color:var(--md-outline)">Playlist no encontrada</p>`)
		}
		items, _ := pb.FindRecordsByFilter("playlist_items", "playlist_id='"+id+"'", "orden", 100, 0)
		type existItem struct {
			Tipo     string `json:"tipo"`
			RefID    string `json:"ref_id"`
			Orden    int    `json:"orden"`
			Duracion int    `json:"duracion"`
			Name     string `json:"name"`
		}
		var existing []existItem
		for _, it := range items {
			d := existItem{
				Tipo:     it.GetString("tipo"),
				Orden:    it.GetInt("orden"),
				Duracion: it.GetInt("duracion_segundos"),
			}
			switch d.Tipo {
			case "content_block":
				d.RefID = it.GetString("content_block_id")
				if cb, e := pb.FindRecordById("content_blocks", d.RefID); e == nil {
					d.Name = cb.GetString("title")
				}
			default:
				d.RefID = it.GetString("multimedia_id")
				if mm, e := pb.FindRecordById("multimedia", d.RefID); e == nil {
					d.Name = mm.GetString("filename")
				}
			}
			existing = append(existing, d)
		}
		existJSON, _ := json.Marshal(existing)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(playlistEditorHTML(r.Id, r.GetString("name"), string(existJSON)))
	}
}

func PlaylistUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("playlists", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrada</div>`)
		}
		if name := strings.TrimSpace(c.FormValue("name")); name != "" {
			r.Set("name", name)
		}
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		// Replace items
		old, _ := pb.FindRecordsByFilter("playlist_items", "playlist_id='"+id+"'", "", 100, 0)
		for _, it := range old {
			pb.Delete(it)
		}
		var items []plItemInput
		if ij := c.FormValue("items_json"); ij != "" {
			json.Unmarshal([]byte(ij), &items)
		}
		savePlItems(pb, id, items)
		c.Set("HX-Trigger", "playlistUpdated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Playlist actualizada
<script>document.getElementById('playlist-editor-area').innerHTML='';htmx.trigger(document.body,'playlistUpdated');setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000)</script></div>`)
	}
}

func PlaylistDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("playlists", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		items, _ := pb.FindRecordsByFilter("playlist_items", "playlist_id='"+id+"'", "", 100, 0)
		for _, it := range items {
			pb.Delete(it)
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		c.Set("HX-Trigger", "playlistDeleted")
		return c.SendString("")
	}
}

func PlaylistReorder(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`<div class="toast toast-success">Orden guardado</div>`)
	}
}

// ── DEVICES ──

func DeviceList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "table" {
			if c.Get("HX-Request") != "true" {
				return c.SendFile("./internal/templates/admin/pages/dashboard.html")
			}
			return c.SendFile("./internal/templates/admin/pages/devices.html")
		}
		// Build playlist lookup map
		plRecs, _ := pb.FindRecordsByFilter("playlists", "", "name", 100, 0)
		plMap := make(map[string]string)
		for _, pl := range plRecs {
			plMap[pl.Id] = pl.GetString("name")
		}
		records, err := pb.FindRecordsByFilter("devices", "", "name", 100, 0)
		var sb strings.Builder
		sb.WriteString(`<table><thead><tr>`)
		sb.WriteString(`<th>Nombre</th><th>Tipo</th><th>Código</th><th>Ubicación</th><th>Playlist</th><th>Estado</th><th>Vista actual</th><th>Acciones</th>`)
		sb.WriteString(`</tr></thead><tbody>`)
		if err != nil || len(records) == 0 {
			sb.WriteString(`<tr><td colspan="8" style="text-align:center;padding:32px;color:var(--md-outline)">Sin dispositivos registrados</td></tr>`)
		} else {
			for _, r := range records {
				dtype := r.GetString("type")
				code := r.GetString("code")
				plName := plMap[r.GetString("playlist_id")]
				if plName == "" {
					plName = "—"
				}
				status := r.GetString("status")
				badgeClass := "badge-warning"
				if status == "activo" {
					badgeClass = "badge-success"
				}
				typeLabel := map[string]string{
					"web_hero":   "Web Hero",
					"vertical":   "Totem",
					"horizontal": "Pantalla",
				}[dtype]
				if typeLabel == "" {
					typeLabel = dtype
				}
				displayLink := ""
				if code != "" && dtype != "web_hero" {
					displayLink = fmt.Sprintf(`<a href="/display/%s" target="_blank" class="topbar-btn topbar-btn-outline" style="padding:3px 8px;font-size:11px;text-decoration:none">Ver pantalla</a>`,
						template.HTMLEscapeString(code))
				}
				currentView := r.GetString("current_view")
				if currentView == "" {
					currentView = "—"
				}
				sb.WriteString(fmt.Sprintf(`<tr>
					<td style="font-weight:500">%s</td>
					<td><span class="badge badge-info" style="font-size:11px">%s</span></td>
					<td><code style="background:var(--md-surface-container-high);padding:2px 6px;border-radius:4px;font-size:12px">%s</code></td>
					<td style="color:var(--md-on-surface-variant);font-size:13px">%s</td>
					<td style="font-size:13px">%s</td>
					<td><span class="badge %s">%s</span></td>
					<td style="font-size:12px;color:var(--md-outline)">%s</td>
					<td><div style="display:flex;gap:6px;align-items:center;flex-wrap:wrap">
						%s
						<button class="topbar-btn topbar-btn-outline" style="padding:3px 8px;font-size:12px"
							hx-get="/admin/devices/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
						<button style="padding:3px 8px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer;border-radius:9999px;font-family:inherit"
							hx-delete="/admin/devices/%s" hx-confirm="¿Eliminar dispositivo?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
					</div></td>
				</tr>`,
					template.HTMLEscapeString(r.GetString("name")),
					template.HTMLEscapeString(typeLabel),
					template.HTMLEscapeString(code),
					template.HTMLEscapeString(r.GetString("ubicacion")),
					template.HTMLEscapeString(plName),
					badgeClass, template.HTMLEscapeString(status),
					template.HTMLEscapeString(currentView),
					displayLink,
					r.Id, r.Id,
				))
			}
		}
		sb.WriteString(`</tbody></table>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func DeviceForm(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		opts := playlistOptsHTML(pb, "")
		html := fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header"><h3>Registrar Dispositivo</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form hx-post="/admin/devices" hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field"><label>Nombre</label><input type="text" name="name" required class="form-input" placeholder="ej: Totem Entrada Principal"/></div>
      <div class="form-row">
        <div class="form-field"><label>Tipo</label>
          <select name="type" class="form-input">
            <option value="horizontal">Pantalla Horizontal</option>
            <option value="vertical">Totem Vertical</option>
            <option value="web_hero">Web Hero</option>
          </select>
        </div>
        <div class="form-field"><label>Código único</label><input type="text" name="code" class="form-input" placeholder="ej: T003"/></div>
      </div>
      <div class="form-field"><label>Ubicación</label><input type="text" name="ubicacion" class="form-input" placeholder="ej: Gimnasio"/></div>
      <div class="form-field"><label>Playlist asignada</label>
        <select name="playlist_id" class="form-input">
          <option value="">— Sin playlist —</option>%s
        </select>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Registrar</button>
      </div>
    </form>
  </div>
</div>`, opts)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func DeviceEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("devices", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Dispositivo no encontrado</div>`)
		}
		dtype := r.GetString("type")
		opts := playlistOptsHTML(pb, r.GetString("playlist_id"))
		html := fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()">
    <div class="modal-header"><h3>Editar Dispositivo</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form hx-put="/admin/devices/%s" hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field"><label>Nombre</label><input type="text" name="name" value="%s" required class="form-input"/></div>
      <div class="form-row">
        <div class="form-field"><label>Tipo</label>
          <select name="type" class="form-input">
            <option value="horizontal"%s>Pantalla Horizontal</option>
            <option value="vertical"%s>Totem Vertical</option>
            <option value="web_hero"%s>Web Hero</option>
          </select>
        </div>
        <div class="form-field"><label>Código único</label><input type="text" name="code" value="%s" class="form-input"/></div>
      </div>
      <div class="form-field"><label>Ubicación</label><input type="text" name="ubicacion" value="%s" class="form-input"/></div>
      <div class="form-field"><label>Playlist asignada</label>
        <select name="playlist_id" class="form-input">
          <option value="">— Sin playlist —</option>%s
        </select>
      </div>
      <div class="form-field"><label>Estado</label>
        <select name="status" class="form-input">
          <option value="activo"%s>Activo</option>
          <option value="inactivo"%s>Inactivo</option>
        </select>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar cambios</button>
      </div>
    </form>
  </div>
</div>`,
			r.Id,
			template.HTMLEscapeString(r.GetString("name")),
			sel(dtype, "horizontal"), sel(dtype, "vertical"), sel(dtype, "web_hero"),
			template.HTMLEscapeString(r.GetString("code")),
			template.HTMLEscapeString(r.GetString("ubicacion")),
			opts,
			sel(r.GetString("status"), "activo"),
			sel(r.GetString("status"), "inactivo"),
		)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func DeviceCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("devices")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		r.Set("name", c.FormValue("name"))
		r.Set("type", c.FormValue("type"))
		r.Set("code", c.FormValue("code"))
		r.Set("ubicacion", c.FormValue("ubicacion"))
		r.Set("playlist_id", c.FormValue("playlist_id"))
		r.Set("status", "activo")
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error guardando</div>`)
		}
		c.Set("HX-Trigger", "deviceCreated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Dispositivo registrado
<script>document.getElementById('modal-container').innerHTML='';htmx.ajax('GET','/admin/devices?fragment=table',{target:'#devices-table-wrap',swap:'innerHTML'});setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000)</script></div>`)
	}
}

func DeviceUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("devices", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		r.Set("name", c.FormValue("name"))
		r.Set("type", c.FormValue("type"))
		r.Set("code", c.FormValue("code"))
		r.Set("ubicacion", c.FormValue("ubicacion"))
		r.Set("playlist_id", c.FormValue("playlist_id"))
		r.Set("status", c.FormValue("status"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		c.Set("HX-Trigger", "deviceUpdated")
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Dispositivo actualizado
<script>document.getElementById('modal-container').innerHTML='';htmx.ajax('GET','/admin/devices?fragment=table',{target:'#devices-table-wrap',swap:'innerHTML'});setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000)</script></div>`)
	}
}

func DeviceDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("devices", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		c.Set("HX-Trigger", "deviceDeleted")
		return c.SendString("")
	}
}

func DeviceAssignPlaylist(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("devices", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Dispositivo no encontrado</div>`)
		}
		r.Set("playlist_id", c.FormValue("playlist_id"))
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error asignando playlist</div>`)
		}
		c.Set("HX-Trigger", "deviceUpdated")
		return c.SendString(`<div class="toast toast-success">✅ Playlist asignada</div>`)
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

// ── PLAYLIST HELPERS ──────────────────────────────────────────────────────────

// plItemInput is the JSON shape sent from the playlist editor frontend.
type plItemInput struct {
	Tipo     string `json:"tipo"`
	RefID    string `json:"ref_id"`
	Orden    int    `json:"orden"`
	Duracion int    `json:"duracion"`
	Name     string `json:"name"`
}

// playlistOptsHTML returns <option> elements for every playlist in PocketBase.
func playlistOptsHTML(pb *pocketbase.PocketBase, selectedID string) string {
	records, _ := pb.FindRecordsByFilter("playlists", "", "name", 100, 0)
	var sb strings.Builder
	for _, r := range records {
		s := ""
		if r.Id == selectedID {
			s = " selected"
		}
		sb.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`,
			r.Id, s, template.HTMLEscapeString(r.GetString("name"))))
	}
	return sb.String()
}

// savePlItems creates playlist_item records in PocketBase for the given playlist.
func savePlItems(pb *pocketbase.PocketBase, playlistID string, items []plItemInput) {
	col, err := pb.FindCollectionByNameOrId("playlist_items")
	if err != nil {
		log.Printf("⚠️  savePlItems: %v", err)
		return
	}
	for i, it := range items {
		r := core.NewRecord(col)
		r.Set("playlist_id", playlistID)
		r.Set("tipo", it.Tipo)
		r.Set("orden", i+1)
		r.Set("duracion_segundos", it.Duracion)
		switch it.Tipo {
		case "content_block":
			r.Set("content_block_id", it.RefID)
		default:
			r.Set("multimedia_id", it.RefID)
		}
		if err := pb.Save(r); err != nil {
			log.Printf("⚠️  savePlItems[%d]: %v", i, err)
		}
	}
}

// buildContentPool returns the content-card grid for the playlist editor.
func buildContentPool(c *fiber.Ctx, pb *pocketbase.PocketBase) error {
	var sb strings.Builder
	sb.WriteString(`<div class="content-grid">`)

	cbs, _ := pb.FindRecordsByFilter("content_blocks", "", "-date", 100, 0)
	for _, r := range cbs {
		cat := r.GetString("category")
		dtype := "eventos"
		if cat == "NOTICIA" {
			dtype = "noticias"
		}
		icon := "📋"
		if r.GetBool("urgency") {
			icon = "🚨"
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="content-card" data-id="%s" data-tipo="content_block" data-type="%s" data-name="%s" onclick="addCardToPlaylist(this)"><div class="content-card-thumb">%s</div><div class="content-card-name">%s</div><div class="content-card-meta">%s</div></div>`,
			r.Id, dtype,
			template.HTMLEscapeString(r.GetString("title")),
			icon,
			template.HTMLEscapeString(r.GetString("title")),
			template.HTMLEscapeString(cat),
		))
	}

	mms, _ := pb.FindRecordsByFilter("multimedia", "", "filename", 100, 0)
	for _, r := range mms {
		mtype := r.GetString("type")
		tipo, dtype, icon := "image", "imagen", "🖼️"
		if mtype == "video" {
			tipo, dtype, icon = "video", "video", "🎬"
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="content-card" data-id="%s" data-tipo="%s" data-type="%s" data-name="%s" onclick="addCardToPlaylist(this)"><div class="content-card-thumb">%s</div><div class="content-card-name">%s</div><div class="content-card-meta">%s</div></div>`,
			r.Id, tipo, dtype,
			template.HTMLEscapeString(r.GetString("filename")),
			icon,
			template.HTMLEscapeString(r.GetString("filename")),
			template.HTMLEscapeString(mtype),
		))
	}

	if len(cbs) == 0 && len(mms) == 0 {
		sb.WriteString(`<p style="padding:20px;color:var(--md-outline);font-size:14px;grid-column:1/-1">Sin contenido. Agrega imágenes, videos o eventos primero.</p>`)
	}

	sb.WriteString(`</div>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(sb.String())
}

// playlistEditorHTML builds the playlist editor fragment (no DOCTYPE/html/head/body).
// Pass id="" and name="" for create mode; populated for edit mode.
func playlistEditorHTML(id, name, existingItemsJSON string) string {
	title := "Nueva Playlist"
	if id != "" {
		title = "Editar Playlist"
	}
	formAction := `hx-post="/admin/playlists"`
	if id != "" {
		formAction = fmt.Sprintf(`hx-put="/admin/playlists/%s"`, id)
	}

	return fmt.Sprintf(`<div class="pl-editor-wrap">
<script src="https://cdn.jsdelivr.net/npm/sortablejs@1.15.0/Sortable.min.js"></script>
<style>
.pl-editor{display:grid;grid-template-columns:1fr 360px;gap:24px;margin-top:16px}
.pl-pool{background:var(--md-surface-bright);border-radius:var(--r-lg);border:1px solid var(--md-outline-variant);padding:20px}
.pl-pool h3{font-family:var(--font-display);font-size:18px;margin-bottom:14px}
.content-tabs{display:flex;gap:4px;margin-bottom:16px;flex-wrap:wrap}
.content-tab{padding:7px 14px;border-radius:var(--r-full);font-size:12px;font-weight:500;border:1px solid var(--md-outline-variant);background:transparent;color:var(--md-on-surface-variant);cursor:pointer;font-family:var(--font-body)}
.content-tab.active{background:var(--md-primary);color:var(--md-on-primary);border-color:var(--md-primary)}
.content-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(140px,1fr));gap:10px}
.content-card{background:var(--md-surface);border:1.5px solid var(--md-outline-variant);border-radius:var(--r-md);padding:14px;cursor:pointer;transition:all 180ms}
.content-card:hover{border-color:var(--md-primary);box-shadow:var(--elev-2)}
.content-card-thumb{width:100%%;height:64px;border-radius:6px;background:var(--md-surface-container-high);margin-bottom:8px;display:flex;align-items:center;justify-content:center;font-size:26px}
.content-card-name{font-size:12px;font-weight:500;color:var(--md-on-surface);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.content-card-meta{font-size:11px;color:var(--md-outline)}
.pl-right{display:flex;flex-direction:column;gap:14px}
.pl-name-box{background:var(--md-surface-bright);border-radius:var(--r-lg);border:1px solid var(--md-outline-variant);padding:18px}
.pl-name-box label{font-size:13px;font-weight:500;color:var(--md-on-surface-variant);display:block;margin-bottom:7px}
.pl-seq{background:var(--md-surface-bright);border-radius:var(--r-lg);border:1px solid var(--md-outline-variant);padding:18px;flex:1;min-height:260px}
.pl-seq h4{font-size:13px;font-weight:500;color:var(--md-on-surface-variant);margin-bottom:10px}
.pl-empty{height:140px;display:flex;align-items:center;justify-content:center;border:2px dashed var(--md-outline-variant);border-radius:var(--r-md);color:var(--md-outline);font-size:13px;text-align:center}
.pl-item{display:flex;align-items:center;gap:8px;padding:9px 10px;background:var(--md-surface);border:1px solid var(--md-outline-variant);border-radius:var(--r-sm);margin-bottom:7px;cursor:grab}
.pl-item:hover{box-shadow:var(--elev-2)}
.pl-item.sortable-ghost{opacity:0.4}
.pl-handle{color:var(--md-outline);cursor:grab;font-size:17px;user-select:none}
.pl-info{flex:1;min-width:0}
.pl-iname{font-size:12px;font-weight:500;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.pl-itype{font-size:11px;color:var(--md-outline)}
.pl-dur{width:46px;padding:3px 5px;border:1px solid var(--md-outline-variant);border-radius:5px;font-size:11px;text-align:center}
.pl-rm{background:none;border:none;cursor:pointer;color:var(--md-outline);padding:3px;border-radius:50%%;line-height:1}
.pl-rm:hover{color:#B71C1C}
.pl-save-bar{display:flex;gap:8px}
.pl-btn-save{padding:11px 0;border-radius:var(--r-full);font-size:13px;font-weight:500;border:none;cursor:pointer;font-family:var(--font-body);background:var(--md-primary);color:var(--md-on-primary);flex:1}
.pl-btn-cancel{padding:11px 20px;border-radius:var(--r-full);font-size:13px;cursor:pointer;font-family:var(--font-body);background:transparent;color:var(--md-on-surface-variant);border:1px solid var(--md-outline-variant)}
@media(max-width:860px){.pl-editor{grid-template-columns:1fr}}
</style>
<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px">
  <h2 style="font-family:var(--font-display);font-size:22px">%s</h2>
</div>
<form id="pl-form" %s hx-target="#toast-area" hx-swap="innerHTML">
  <input type="hidden" id="pl-items-json" name="items_json" value="[]"/>
  <div class="pl-editor">
    <div class="pl-pool">
      <h3>Contenido disponible</h3>
      <div class="content-tabs">
        <button type="button" class="content-tab active" onclick="plTab('all',this)">Todo</button>
        <button type="button" class="content-tab" onclick="plTab('imagen',this)">Imágenes</button>
        <button type="button" class="content-tab" onclick="plTab('video',this)">Videos</button>
        <button type="button" class="content-tab" onclick="plTab('eventos',this)">Eventos</button>
        <button type="button" class="content-tab" onclick="plTab('noticias',this)">Noticias</button>
      </div>
      <div id="pl-content-grid"
           hx-get="/admin/playlists?fragment=content"
           hx-trigger="load"
           hx-swap="innerHTML">
        <p style="color:var(--md-outline);font-size:13px;padding:12px">Cargando...</p>
      </div>
    </div>
    <div class="pl-right">
      <div class="pl-name-box">
        <label>Nombre de la playlist</label>
        <input type="text" name="name" value="%s" required class="form-input" placeholder="ej: Hero Principal 2026" style="margin:0"/>
      </div>
      <div class="pl-seq">
        <h4>Secuencia de reproducción</h4>
        <div id="pl-sortable">
          <div class="pl-empty">Haz clic en las tarjetas<br/>para agregar</div>
        </div>
      </div>
      <div class="pl-save-bar">
        <button type="button" class="pl-btn-cancel" onclick="document.getElementById('playlist-editor-area').innerHTML=''">Cancelar</button>
        <button type="submit" class="pl-btn-save" onclick="plPrepareSubmit()">Guardar playlist</button>
      </div>
    </div>
  </div>
</form>
<script>
(function(){
var plItems = [];
var EXISTING = %s;
if(EXISTING && EXISTING.length){
  EXISTING.forEach(function(it){
    plItems.push({tipo:it.tipo,refID:it.ref_id||'',name:it.name||it.ref_id||'',duracion:it.duracion||15});
  });
  renderPl();
}
function renderPl(){
  var wrap = document.getElementById('pl-sortable');
  if(!wrap) return;
  if(!plItems.length){
    wrap.innerHTML = '<div class="pl-empty">Haz clic en las tarjetas<br/>para agregar</div>';
    return;
  }
  wrap.innerHTML = '';
  plItems.forEach(function(it,idx){
    var d = document.createElement('div');
    d.className = 'pl-item';
    d.dataset.idx = idx;
    var handle = document.createElement('span');
    handle.className = 'material-symbols-outlined pl-handle';
    handle.textContent = 'drag_indicator';
    var info = document.createElement('div');
    info.className = 'pl-info';
    var nm = document.createElement('div');
    nm.className = 'pl-iname';
    nm.textContent = it.name;
    var tp = document.createElement('div');
    tp.className = 'pl-itype';
    tp.textContent = it.tipo;
    info.appendChild(nm);
    info.appendChild(tp);
    var dur = document.createElement('input');
    dur.type = 'number';
    dur.className = 'pl-dur';
    dur.value = it.duracion||15;
    dur.min = 5; dur.max = 300;
    dur.title = 'Duracion (seg)';
    dur.addEventListener('change',(function(i){ return function(){ plItems[i].duracion=parseInt(this.value)||15; }; })(idx));
    var lbl = document.createElement('span');
    lbl.style.cssText = 'font-size:10px;color:var(--md-outline)';
    lbl.textContent = 's';
    var rm = document.createElement('button');
    rm.type = 'button';
    rm.className = 'pl-rm';
    rm.title = 'Quitar';
    var rmI = document.createElement('span');
    rmI.className = 'material-symbols-outlined';
    rmI.style.fontSize = '16px';
    rmI.textContent = 'close';
    rm.appendChild(rmI);
    rm.addEventListener('click',(function(i){ return function(){ plItems.splice(i,1); renderPl(); }; })(idx));
    d.appendChild(handle);
    d.appendChild(info);
    d.appendChild(dur);
    d.appendChild(lbl);
    d.appendChild(rm);
    wrap.appendChild(d);
  });
  if(window.Sortable){
    if(wrap._sortable){ wrap._sortable.destroy(); }
    wrap._sortable = new Sortable(wrap,{
      animation:180,handle:'.pl-handle',ghostClass:'sortable-ghost',
      onEnd:function(evt){
        var moved = plItems.splice(evt.oldIndex,1)[0];
        plItems.splice(evt.newIndex,0,moved);
      }
    });
  }
}
window.addCardToPlaylist = function(card){
  var id = card.dataset.id;
  for(var i=0;i<plItems.length;i++){ if(plItems[i].refID===id){ return; } }
  plItems.push({tipo:card.dataset.tipo,refID:id,name:card.dataset.name||id,duracion:15});
  renderPl();
};
window.plTab = function(type,btn){
  document.querySelectorAll('.content-tab').forEach(function(t){ t.classList.remove('active'); });
  btn.classList.add('active');
  var grid = document.getElementById('pl-content-grid');
  if(!grid) return;
  grid.querySelectorAll('.content-card').forEach(function(c){
    c.style.display=(type==='all'||c.dataset.type===type)?'':'none';
  });
};
window.plPrepareSubmit = function(){
  var data = plItems.map(function(it,i){
    return {tipo:it.tipo,ref_id:it.refID,orden:i+1,duracion:it.duracion||15,name:it.name};
  });
  var inp = document.getElementById('pl-items-json');
  if(inp){ inp.value = JSON.stringify(data); }
};
})();
</script>
</div>`,
		title, formAction, template.HTMLEscapeString(name), existingItemsJSON)
}
