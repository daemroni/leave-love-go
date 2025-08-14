package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/example/leaf-love-go/internal/data"
	"github.com/example/leaf-love-go/internal/models"
)

var (
	layoutHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Leaf Love Advisor (Go)</title>
  <link rel="stylesheet" href="/static/styles.css">
  <style>
    body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans, Helvetica, Arial, sans-serif; margin: 0; background: #0b1215; color: #e6f1f5; }
    header { padding: 2rem 1rem; text-align: center; }
    main { max-width: 960px; margin: 0 auto; padding: 1rem; }
    .card { background: #0d1a1f; border: 1px solid #11303a; border-radius: 16px; padding: 1rem; box-shadow: 0 10px 25px rgba(0,0,0,.25); }
    .grid { display: grid; gap: 1rem; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); }
    .btn { display: inline-block; padding: .75rem 1rem; border-radius: 12px; border: 1px solid #2a6e7f; background: #0f2a32; color: #b7ecff; text-decoration: none; cursor: pointer; }
    .btn.primary { background: #124a58; border-color: #2995ae; color: #d8f7ff; }
    label { font-weight: 600; display:block; margin-bottom: .5rem; }
    select { width: 100%; padding: .5rem; border-radius: 8px; background: #0b1418; color: #d0e7ee; border: 1px solid #1b3b47; }
    h1 { font-size: 1.75rem; margin: 0; }
    h2 { font-size: 1.25rem; margin-top: 0; }
    .muted { color: #8fb8c4; }
    .pill { display:inline-block; padding: .25rem .5rem; border-radius: 999px; border:1px solid #1b3b47; margin-right: .25rem; font-size: .8rem; }
    img { max-width: 100%; border-radius: 12px; border:1px solid #11303a; }
  </style>
</head>
<body>
  <header>
    <h1>🌿 Leaf Love Advisor — Go Edition</h1>
    <p class="muted">Answer a few questions and get beginner-friendly plant recommendations.</p>
  </header>
  <main>{{.Content}}</main>
</body>
</html>`

	indexHTML = `
<div class="card">
  <h2>Tell us your preferences</h2>
  <form method="POST" action="/recommend" class="grid">
    <div>
      <label for="light">Light Conditions</label>
      <select id="light" name="lightCondition">
        <option value="partial-shade" selected>Partial shade</option>
        <option value="full-sun">Full sun</option>
        <option value="low-light">Low light</option>
      </select>
    </div>
    <div>
      <label for="care">Care Level</label>
      <select id="care" name="careLevel">
        <option value="medium" selected>Medium</option>
        <option value="low">Low</option>
        <option value="high">High</option>
      </select>
    </div>
    <div>
      <label for="type">Plant Type</label>
      <select id="type" name="plantType">
        <option value="any" selected>Any</option>
        <option value="foliage">Foliage</option>
        <option value="flowering">Flowering</option>
        <option value="succulent">Succulent</option>
      </select>
    </div>
    <div>
      <label for="loc">Location</label>
      <select id="loc" name="location">
        <option value="both" selected>Both</option>
        <option value="indoor">Indoor</option>
        <option value="outdoor">Outdoor</option>
      </select>
    </div>
    <div>
      <label for="size">Size</label>
      <select id="size" name="size">
        <option value="any" selected>Any</option>
        <option value="small">Small</option>
        <option value="medium">Medium</option>
        <option value="large">Large</option>
      </select>
    </div>
    <div style="align-self:end">
      <button class="btn primary" type="submit">Get Recommendations</button>
    </div>
  </form>
</div>`

	resultsHTML = `
<div class="card">
  <a class="btn" href="/">← Back</a>
  <h2 style="margin-top:1rem">Recommended Plants ({{.Count}})</h2>
  {{if eq .Count 0}}
    <p class="muted">No exact matches. Try relaxing one of your preferences.</p>
  {{else}}
    <div class="grid">
      {{range .Plants}}
        <div class="card">
          <img src="{{.Image}}" alt="{{.Name}}">
          <h3 style="margin:.5rem 0">{{.Name}}</h3>
          <p class="muted"><em>{{.ScientificName}}</em></p>
          <p>{{.Description}}</p>
          <div style="margin:.5rem 0">
            <span class="pill">{{.CareLevel}} care</span>
            <span class="pill">{{.PlantType}}</span>
            <span class="pill">{{.Location}}</span>
            <span class="pill">{{.Size}}</span>
          </div>
          <div class="muted" style="font-size:.9rem">
            <div>💧 {{.Care.Watering}}</div>
            <div>☀️ {{.Care.Light}}</div>
            <div>🌡️ {{.Care.Temperature}}</div>
            <div>💨 {{.Care.Humidity}}</div>
          </div>
        </div>
      {{end}}
    </div>
  {{end}}
</div>`

	// Compiled templates controlled from same place
	tplLayout  = template.Must(template.New("layout").Parse(layoutHTML))
	tplIndex   = template.Must(template.New("index").Parse(indexHTML))
	tplResults = template.Must(template.New("results").Parse(resultsHTML))

	// Global mutable state (routing + metrics + config all here).
	requestCount   uint64
	lastStatusCode int32
	startTime      = time.Now()
)

// omniHandler: centralized router + controller
func omniHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&requestCount, 1)
	start := time.Now()
	defer func() {
		// Log timing and last status after we handled the request
		d := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, atomic.LoadInt32(&lastStatusCode), d)
	}()

	switch r.URL.Path {
	case "/health":
		writeStatus(w, http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return

	case "/metrics":
		if !methodAllowed(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		writeStatus(w, http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		uptime := time.Since(startTime).Seconds()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(
			"# HELP leaflove_requests_total Total HTTP requests.\n" +
				"# TYPE leaflove_requests_total counter\n" +
				"leaflove_requests_total " + strconv.FormatUint(atomic.LoadUint64(&requestCount), 10) + "\n" +
				"# HELP leaflove_uptime_seconds Process uptime in seconds.\n" +
				"# TYPE leaflove_uptime_seconds gauge\n" +
				"leaflove_uptime_seconds " + strconv.FormatFloat(uptime, 'f', 0, 64) + "\n"))
		return

	case "/":
		if !methodAllowed(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		writeStatus(w, http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		renderHTML(w, tplIndex, nil)
		return

	case "/recommend":
		if !methodAllowed(w, r, http.MethodPost) {
			return
		}
		if err := r.ParseForm(); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid form: "+err.Error())
			return
		}
		prefs := models.PlantPreferences{
			LightCondition: r.FormValue("lightCondition"),
			CareLevel:      r.FormValue("careLevel"),
			PlantType:      r.FormValue("plantType"),
			Location:       r.FormValue("location"),
			Size:           r.FormValue("size"),
		}
		recs := filterPlants(prefs)
		writeStatus(w, http.StatusOK)
		renderHTML(w, tplResults, map[string]any{
			"Plants":      recs,
			"Preferences": prefs,
			"Count":       len(recs),
		})
		return

	case "/api/recommend":
		if !methodAllowed(w, r, http.MethodGet, http.MethodHead) {
			return
		}
		q := r.URL.Query()
		prefs := models.PlantPreferences{
			LightCondition: q.Get("lightCondition"),
			CareLevel:      q.Get("careLevel"),
			PlantType:      q.Get("plantType"),
			Location:       q.Get("location"),
			Size:           q.Get("size"),
		}
		recs := filterPlants(prefs)

		// Basic caching hints (safe since results depend only on query)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")

		writeStatus(w, http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		_ = json.NewEncoder(w).Encode(recs)
		return
	}

	// Fallback 404
	writeJSONError(w, http.StatusNotFound, "route not found")
}


// filterPlants orchestrates the matching + sorting.
func filterPlants(p models.PlantPreferences) []models.Plant {
	p = normalizePrefs(p)

	var out []models.Plant
	for _, pl := range data.Plants {
		if matchesAll(pl, p) {
			out = append(out, pl)
		}
	}
	sortPlantsByName(out)
	return out
}

// matchesAll checks if a plant matches all preferences.
func matchesAll(pl models.Plant, p models.PlantPreferences) bool {
	return matchesLight(pl, p) &&
		matchesCare(pl, p) &&
		matchesType(pl, p) &&
		matchesLocation(pl, p) &&
		matchesSize(pl, p)
}

func matchesLight(pl models.Plant, p models.PlantPreferences) bool {
	if p.LightCondition == "" {
		return true
	}
	return has(pl.LightCondition, p.LightCondition)
}

func matchesCare(pl models.Plant, p models.PlantPreferences) bool {
	return p.CareLevel == "" || pl.CareLevel == p.CareLevel
}

func matchesType(pl models.Plant, p models.PlantPreferences) bool {
	if p.PlantType == "" || p.PlantType == "any" {
		return true
	}
	return pl.PlantType == p.PlantType
}

func matchesLocation(pl models.Plant, p models.PlantPreferences) bool {
	if p.Location == "" || p.Location == "both" {
		return true
	}
	// also match if the plant itself is usable in both locations
	return pl.Location == p.Location || pl.Location == "both"
}

func matchesSize(pl models.Plant, p models.PlantPreferences) bool {
	if p.Size == "" || p.Size == "any" {
		return true
	}
	return pl.Size == p.Size
}

func sortPlantsByName(plants []models.Plant) {
	slices.SortFunc(plants, func(a, b models.Plant) int {
		return strings.Compare(a.Name, b.Name)
	})
}

func normalizePrefs(p models.PlantPreferences) models.PlantPreferences {
	p.LightCondition = strings.ToLower(strings.TrimSpace(p.LightCondition))
	p.CareLevel = strings.ToLower(strings.TrimSpace(p.CareLevel))
	p.PlantType = strings.ToLower(strings.TrimSpace(p.PlantType))
	p.Location = strings.ToLower(strings.TrimSpace(p.Location))
	p.Size = strings.ToLower(strings.TrimSpace(p.Size))
	return p
}

func has(list []string, val string) bool {
	if val == "" {
		return true
	}
	for _, x := range list {
		if strings.EqualFold(strings.TrimSpace(x), val) {
			return true
		}
	}
	return false
}


func renderHTML(w http.ResponseWriter, t *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// One-off composition inside the renderer
	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tplLayout.Execute(w, map[string]any{"Content": template.HTML(sb.String())}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Logging setup
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "server.log"
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		log.SetOutput(f) // Redirect logs to file
	} else {
		log.Printf("failed to open log file %q: %v", logFile, err)
	}

	mux := http.NewServeMux()

	// Static file serving also configured here.
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", omniHandler)
	mux.HandleFunc("/recommend", omniHandler)
	mux.HandleFunc("/api/recommend", omniHandler)
	mux.HandleFunc("/health", omniHandler)
	mux.HandleFunc("/metrics", omniHandler)

	addr := ":8080"
	log.Printf("Leaf Love Advisor (Go) listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
