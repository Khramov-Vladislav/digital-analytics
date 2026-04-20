package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"net/url"
)

//go:embed web/*
var webFS embed.FS

type TemplateID string

const (
	TemplateClassic TemplateID = "classic"
	TemplateModern  TemplateID = "modern"
	TemplateSidebar TemplateID = "sidebar"
	TemplateElegant TemplateID = "elegant"
	TemplateNordic  TemplateID = "nordic"
	TemplateTerminal TemplateID = "terminal"
	TemplateAcademic TemplateID = "academic"
	TemplateMagazine TemplateID = "magazine"
	TemplateInfographic TemplateID = "infographic"
	TemplateMonochrome TemplateID = "monochrome"
)

var sAllowedTemplates = map[TemplateID]struct{}{
	TemplateClassic: {},
	TemplateModern:  {},
	TemplateSidebar: {},
	TemplateElegant: {},
	TemplateNordic: {},
	TemplateTerminal: {},
	TemplateAcademic: {},
	TemplateMagazine: {},
	TemplateInfographic: {},
	TemplateMonochrome: {},
}

type Education struct {
	Institution string
	Specialty   string
	Years       string
}

type Experience struct {
	Company  string
	Position string
	Period   string
	Duties   string
}

type ResumeData struct {
	Template   TemplateID
	FullName   string
	Phone      string
	Email      string
	City       string
	About      string
	Skills     string
	Additional string

	Education  []Education
	Experience []Experience

	PhotoURL string
}

type photoStore struct {
	mu    sync.RWMutex
	items map[string]photoItem
}

type photoItem struct {
	contentType string
	data        []byte
	expiresAt   time.Time
}

func (s *photoStore) put(contentType string, data []byte, ttl time.Duration) string {
	id := randomID(16)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = map[string]photoItem{}
	}
	s.items[id] = photoItem{
		contentType: contentType,
		data:        data,
		expiresAt:   time.Now().Add(ttl),
	}
	return id
}

func (s *photoStore) get(id string) (photoItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.items[id]
	if !ok {
		return photoItem{}, false
	}
	if time.Now().After(it.expiresAt) {
		return photoItem{}, false
	}
	return it, true
}

func (s *photoStore) cleanupLoop(stop <-chan struct{}) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.items {
				if now.After(v.expiresAt) {
					delete(s.items, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

type server struct {
	mux       *http.ServeMux
	uiTmpl    *template.Template
	resumeTpl map[TemplateID]*template.Template
	photos    *photoStore
}

func main() {
	s, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	addr := "127.0.0.1:8080"
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, s.routes()))
}

func newServer() (*server, error) {
	uiTmpl, err := template.ParseFS(webFS, "web/templates/ui.html")
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"splitLines": splitLines,
		"firstPosition": firstPosition,
	}

	parse := func(name string, files ...string) (*template.Template, error) {
		return template.New(name).Funcs(funcs).ParseFS(webFS, files...)
	}

	classicTml, err := parse("resume_classic.html", "web/templates/resume_classic.html")
	if err != nil {
		return nil, err
	}
	modernTml, err := parse("resume_modern.html", "web/templates/resume_modern.html")
	if err != nil {
		return nil, err
	}
	sidebarTml, err := parse("resume_sidebar.html", "web/templates/resume_sidebar.html")
	if err != nil {
		return nil, err
	}
	elegantTml, err := parse("resume_elegant.html", "web/templates/resume_elegant.html")
	if err != nil {
		return nil, err
	}
	nordicTml, err := parse("resume_nordic.html", "web/templates/resume_nordic.html")
	if err != nil {
		return nil, err
	}
	terminalTml, err := parse("resume_terminal.html", "web/templates/resume_terminal.html")
	if err != nil {
		return nil, err
	}
	academicTml, err := parse("resume_academic.html", "web/templates/resume_academic.html")
	if err != nil {
		return nil, err
	}
	magazineTml, err := parse("resume_magazine.html", "web/templates/resume_magazine.html")
	if err != nil {
		return nil, err
	}
	infographicTml, err := parse("resume_infographic.html", "web/templates/resume_infographic.html")
	if err != nil {
		return nil, err
	}
	monochromeTml, err := parse("resume_monochrome.html", "web/templates/resume_monochrome.html")
	if err != nil {
		return nil, err
	}

	return &server{
		mux:   http.NewServeMux(),
		uiTmpl: uiTmpl,
		resumeTpl: map[TemplateID]*template.Template{
			TemplateClassic: classicTml,
			TemplateModern:  modernTml,
			TemplateSidebar: sidebarTml,
			TemplateElegant: elegantTml,
			TemplateNordic: nordicTml,
			TemplateTerminal: terminalTml,
			TemplateAcademic: academicTml,
			TemplateMagazine: magazineTml,
			TemplateInfographic: infographicTml,
			TemplateMonochrome: monochromeTml,
		},
		photos: &photoStore{items: map[string]photoItem{}},
	}, nil
}

func (s *server) routes() http.Handler {
	stop := make(chan struct{})
	go s.photos.cleanupLoop(stop)

	s.mux.HandleFunc("GET /", s.handleUI)
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		panic(err)
	}
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	s.mux.HandleFunc("GET /uploads/", s.handleUpload)
	s.mux.HandleFunc("POST /api/preview", s.handlePreview)
	s.mux.HandleFunc("POST /api/pdf", s.handlePDF)

	return s.wrap(s.mux)
}

func (s *server) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// Preview is rendered inside <iframe srcdoc>, which may have an opaque origin in some browsers.
		// If CORP is too strict, the browser can block loading /uploads/* images into the preview.
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

		// Reasonable timeouts for handlers doing PDF work.
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.uiTmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/uploads/")
	id = path.Clean("/" + id)
	id = strings.TrimPrefix(id, "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	it, ok := s.photos.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", it.contentType)
	w.Header().Set("Cache-Control", "private, max-age=60")
	_, _ = w.Write(it.data)
}

func (s *server) handlePreview(w http.ResponseWriter, r *http.Request) {
	data, photoContentType, photoBytes, err := parseAndValidate(r, 5<<20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(photoBytes) > 0 {
		id := s.photos.put(photoContentType, photoBytes, 10*time.Minute)
		data.PhotoURL = "/uploads/" + id
	}

	html, err := s.renderResumeHTML(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"html": html})
}

func (s *server) handlePDF(w http.ResponseWriter, r *http.Request) {
	data, photoContentType, photoBytes, err := parseAndValidate(r, 5<<20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(photoBytes) > 0 {
		id := s.photos.put(photoContentType, photoBytes, 10*time.Minute)
		data.PhotoURL = "/uploads/" + id
	}

	hostURL := absoluteBaseURL(r)
	html, err := s.renderResumeHTMLWithBase(data, hostURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfBytes, err := htmlToPDF(r.Context(), html)
	if err != nil {
		http.Error(w, "PDF generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := sanitizeFilename(data.FullName)
	if filename == "" {
		filename = "resume"
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", contentDispositionPDF(filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	_, _ = w.Write(pdfBytes)
}

func (s *server) renderResumeHTML(data ResumeData) (string, error) {
	return s.renderResumeHTMLWithBase(data, "")
}

func (s *server) renderResumeHTMLWithBase(data ResumeData, absoluteBase string) (string, error) {
	clone := data
	if absoluteBase != "" && strings.HasPrefix(clone.PhotoURL, "/") {
		clone.PhotoURL = absoluteBase + clone.PhotoURL
	}

	tmpl, ok := s.resumeTpl[clone.Template]
	if !ok {
		tmpl = s.resumeTpl[TemplateClassic]
	}

	var b strings.Builder
	if err := tmpl.Execute(&b, clone); err != nil {
		return "", err
	}
	return b.String(), nil
}

func parseAndValidate(r *http.Request, maxPhotoBytes int64) (ResumeData, string, []byte, error) {
	if err := r.ParseMultipartForm(12 << 20); err != nil { // includes base64 overhead on some browsers
		return ResumeData{}, "", nil, errors.New("не удалось прочитать форму")
	}

	short := func(key string) string { return strings.TrimSpace(r.FormValue(key)) }
	long := func(key string) string { return strings.TrimSpace(r.FormValue(key)) }

	data := ResumeData{
		Template:   TemplateID(short("template")),
		FullName:   short("full_name"),
		Phone:      short("phone"),
		Email:      short("email"),
		City:       short("city"),
		About:      long("about"),
		Skills:     long("skills"),
		Additional: long("additional"),
	}

	if _, ok := sAllowedTemplates[data.Template]; !ok {
		data.Template = TemplateClassic
	}

	if err := validateLen("ФИО", data.FullName, 100); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("Телефон", data.Phone, 100); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("Email", data.Email, 100); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("Город", data.City, 100); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("О себе", data.About, 1000); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("Навыки", data.Skills, 1000); err != nil {
		return ResumeData{}, "", nil, err
	}
	if err := validateLen("Доп. информация", data.Additional, 1000); err != nil {
		return ResumeData{}, "", nil, err
	}

	data.Education = parseEducation(r.MultipartForm)
	data.Experience = parseExperience(r.MultipartForm)

	for i, e := range data.Education {
		if err := validateLen("Образование: учебное заведение #"+strconv.Itoa(i+1), e.Institution, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
		if err := validateLen("Образование: специальность #"+strconv.Itoa(i+1), e.Specialty, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
		if err := validateLen("Образование: годы #"+strconv.Itoa(i+1), e.Years, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
	}
	for i, e := range data.Experience {
		if err := validateLen("Опыт: компания #"+strconv.Itoa(i+1), e.Company, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
		if err := validateLen("Опыт: должность #"+strconv.Itoa(i+1), e.Position, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
		if err := validateLen("Опыт: период #"+strconv.Itoa(i+1), e.Period, 100); err != nil {
			return ResumeData{}, "", nil, err
		}
		if err := validateLen("Опыт: обязанности #"+strconv.Itoa(i+1), e.Duties, 1000); err != nil {
			return ResumeData{}, "", nil, err
		}
	}

	photoContentType, photoBytes, err := readAndValidatePhoto(r.MultipartForm, maxPhotoBytes)
	if err != nil {
		return ResumeData{}, "", nil, err
	}
	return data, photoContentType, photoBytes, nil
}

func parseEducation(f *multipart.Form) []Education {
	if f == nil {
		return nil
	}
	inst := f.Value["edu_institution"]
	spec := f.Value["edu_specialty"]
	years := f.Value["edu_years"]
	n := max3(len(inst), len(spec), len(years))
	out := make([]Education, 0, n)
	for i := 0; i < n; i++ {
		e := Education{
			Institution: safeIndex(inst, i),
			Specialty:   safeIndex(spec, i),
			Years:       safeIndex(years, i),
		}
		if strings.TrimSpace(e.Institution+e.Specialty+e.Years) == "" {
			continue
		}
		out = append(out, trimEducation(e))
	}
	return out
}

func parseExperience(f *multipart.Form) []Experience {
	if f == nil {
		return nil
	}
	company := f.Value["exp_company"]
	position := f.Value["exp_position"]
	period := f.Value["exp_period"]
	duties := f.Value["exp_duties"]
	n := max4(len(company), len(position), len(period), len(duties))
	out := make([]Experience, 0, n)
	for i := 0; i < n; i++ {
		e := Experience{
			Company:  safeIndex(company, i),
			Position: safeIndex(position, i),
			Period:   safeIndex(period, i),
			Duties:   safeIndex(duties, i),
		}
		if strings.TrimSpace(e.Company+e.Position+e.Period+e.Duties) == "" {
			continue
		}
		out = append(out, trimExperience(e))
	}
	return out
}

func trimEducation(e Education) Education {
	e.Institution = strings.TrimSpace(e.Institution)
	e.Specialty = strings.TrimSpace(e.Specialty)
	e.Years = strings.TrimSpace(e.Years)
	return e
}

func trimExperience(e Experience) Experience {
	e.Company = strings.TrimSpace(e.Company)
	e.Position = strings.TrimSpace(e.Position)
	e.Period = strings.TrimSpace(e.Period)
	e.Duties = strings.TrimSpace(e.Duties)
	return e
}

func readAndValidatePhoto(f *multipart.Form, maxBytes int64) (string, []byte, error) {
	if f == nil || f.File == nil || len(f.File["photo"]) == 0 {
		return "", nil, nil
	}
	fh := f.File["photo"][0]

	if fh.Size > maxBytes {
		return "", nil, errors.New("фотография больше 5 МБ")
	}

	file, err := fh.Open()
	if err != nil {
		return "", nil, errors.New("не удалось прочитать фотографию")
	}
	defer file.Close()

	limited := io.LimitReader(file, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", nil, errors.New("не удалось прочитать фотографию")
	}
	if int64(len(data)) > maxBytes {
		return "", nil, errors.New("фотография больше 5 МБ")
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", nil, errors.New("не удалось декодировать фотографию")
	}

	// Be permissive: rely on actual decoding result, not file extension/mimetype.
	// We still restrict to formats supported by the requirements.
	var contentType string
	switch strings.ToLower(format) {
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	default:
		return "", nil, errors.New("поддерживаются только JPG/JPEG/PNG")
	}

	// Removed minimum resolution requirement to accept common "document style" photos.
	return contentType, data, nil
}

func htmlToPDF(ctx context.Context, html string) ([]byte, error) {
	bin, err := findChromeBinary()
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "resume-pdf-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	htmlPath := filepath.Join(dir, "resume.html")
	pdfPath := filepath.Join(dir, "resume.pdf")
	if err := os.WriteFile(htmlPath, []byte(html), 0o600); err != nil {
		return nil, err
	}

	// Chrome headless PDF.
	// For local assets we use absolute URLs in HTML (photo), so Chrome can fetch them.
	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--print-to-pdf=" + pdfPath,
		// Disable print header/footer (date/title/url/page) in the output PDF.
		// Newer Chrome: --no-pdf-header-footer, older: --print-to-pdf-no-header.
		"--no-pdf-header-footer",
		"--print-to-pdf-no-header",
		"file://" + htmlPath,
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	// Some environments need HOME/TMPDIR; keep current env.
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New("Chrome PDF error: " + msg)
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, err
	}
	return pdfBytes, nil
}

func validateLen(field, v string, max int) error {
	if len([]rune(v)) > max {
		return errors.New(field + ": слишком длинное поле (макс. " + strconv.Itoa(max) + " символов)")
	}
	return nil
}

func randomID(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func safeIndex(s []string, i int) string {
	if i < 0 || i >= len(s) {
		return ""
	}
	return s[i]
}

func max3(a, b, c int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}

func max4(a, b, c, d int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	if a < d {
		a = d
	}
	return a
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '-' || r == '_':
			return r
		case r >= 0x0400 && r <= 0x04FF: // Cyrillic
			return r
		default:
			return -1
		}
	}, s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	if len([]rune(s)) > 64 {
		s = string([]rune(s)[:64])
	}
	return s
}

func contentDispositionPDF(baseName string) string {
	// RFC 6266 + RFC 5987: provide both filename and filename* (UTF-8).
	// Some browsers ignore non-ASCII in filename=, so filename* is required for Cyrillic.
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "resume"
	}

	utf8Name := baseName + ".pdf"
	ascii := sanitizeFilenameASCII(baseName)
	if ascii == "" {
		ascii = "resume"
	}
	asciiName := ascii + ".pdf"

	// RFC 5987 encoding: UTF-8'' + pct-encoded
	escaped := url.PathEscape(utf8Name)
	return `attachment; filename="` + asciiName + `"; filename*=UTF-8''` + escaped
}

func sanitizeFilenameASCII(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '-' || r == '_':
			return r
		default:
			return -1
		}
	}, s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	if len([]rune(s)) > 64 {
		s = string([]rune(s)[:64])
	}
	return s
}

func absoluteBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "127.0.0.1:8080"
	}
	return scheme + "://" + host
}

func splitLines(s string) []string {
	// split by newline and commas/semicolons to make nice skill chips
	s = strings.ReplaceAll(s, "\r\n", "\n")
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n' || r == ',' || r == ';'
	})
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func firstPosition(exps []Experience) string {
	for _, e := range exps {
		p := strings.TrimSpace(e.Position)
		if p != "" {
			return p
		}
	}
	return ""
}

func findChromeBinary() (string, error) {
	if v := strings.TrimSpace(os.Getenv("CHROME_BIN")); v != "" {
		return v, nil
	}

	// macOS common locations
	macCandidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}
	for _, p := range macCandidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// PATH candidates (Linux/macOS via brew)
	names := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p, nil
		}
	}
	return "", errors.New("не найден Chrome/Chromium. Установите браузер или задайте переменную окружения CHROME_BIN")
}

