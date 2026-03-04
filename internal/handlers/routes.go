package handlers

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Khramov-Vladislav/digital-analytics/internal/models"
	"github.com/Khramov-Vladislav/digital-analytics/internal/pdf"
)

// RegisterRoutes mounts the HTTP handlers on the provided mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", formHandler)
	mux.HandleFunc("/preview", previewHandler)
	mux.HandleFunc("/generate", generateHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// formHandler serves the main form page.
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, "form.html", nil)
}

// previewHandler returns an HTML preview of the resume data submitted via POST.
func previewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}
	resume := &models.Resume{
		FullName: r.FormValue("fullname"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		City:     r.FormValue("city"),
		Summary:  r.FormValue("summary"),
		Skills:   strings.Split(r.FormValue("skills"), ","),
		Extra:    r.FormValue("extra"),
		Template: r.FormValue("template"),
	}
	// parse dynamic education entries
	eduInstitutions := r.MultipartForm.Value["edu_institution"]
	for i := range eduInstitutions {
		var startYear, endYear string
		if len(r.MultipartForm.Value["edu_start_year"]) > i {
			startYear = r.MultipartForm.Value["edu_start_year"][i]
		}
		if len(r.MultipartForm.Value["edu_end_year"]) > i {
			endYear = r.MultipartForm.Value["edu_end_year"][i]
		}
		var degree string
		if len(r.MultipartForm.Value["edu_degree"]) > i {
			degree = r.MultipartForm.Value["edu_degree"][i]
		}
		resume.Education = append(resume.Education, models.EducationEntry{
			Institution: eduInstitutions[i],
			Degree:      degree,
			StartYear:   startYear,
			EndYear:     endYear,
		})
	}

	// parse dynamic experience entries
	expCompanies := r.MultipartForm.Value["exp_company"]
	for i := range expCompanies {
		var position, period, description string
		if len(r.MultipartForm.Value["exp_position"]) > i {
			position = r.MultipartForm.Value["exp_position"][i]
		}
		if len(r.MultipartForm.Value["exp_period"]) > i {
			period = r.MultipartForm.Value["exp_period"][i]
		}
		if len(r.MultipartForm.Value["exp_desc"]) > i {
			description = r.MultipartForm.Value["exp_desc"][i]
		}
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Company:     expCompanies[i],
			Position:    position,
			Period:      period,
			Description: description,
		})
	}

	// Handle photo upload - convert to Data URL for preview
	if file, hdr, err := r.FormFile("photo"); err == nil {
		defer file.Close()
		resume.Photo = hdr

		// Read file and convert to base64 Data URL
		data, err := io.ReadAll(file)
		if err == nil {
			mimeType := hdr.Header.Get("Content-Type")
			if mimeType == "" {
				// Guess MIME type from filename
				if strings.HasSuffix(strings.ToLower(hdr.Filename), ".png") {
					mimeType = "image/png"
				} else {
					mimeType = "image/jpeg"
				}
			}
			photoDataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
			resume.PhotoURL = template.URL(photoDataURL)
		}
	}

	// Debug: log raw bytes/runes before normalization
	fmt.Printf("DEBUG preview BEFORE NormalizeData - fullname bytes=%v runes=%v\n", []byte(resume.FullName), []rune(resume.FullName))
	fmt.Printf("DEBUG preview BEFORE NormalizeData - city bytes=%v runes=%v\n", []byte(resume.City), []rune(resume.City))
	fmt.Printf("DEBUG preview BEFORE NormalizeData - summary bytes=%v runes=%v\n", []byte(resume.Summary), []rune(resume.Summary))

	// Apply formatting and normalization
	resume.NormalizeData()

	// Debug: log raw bytes/runes after normalization
	fmt.Printf("DEBUG preview AFTER NormalizeData - fullname=%q bytes=%v runes=%v\n", resume.FullName, []byte(resume.FullName), []rune(resume.FullName))
	fmt.Printf("DEBUG preview AFTER NormalizeData - city=%q bytes=%v runes=%v\n", resume.City, []byte(resume.City), []rune(resume.City))
	fmt.Printf("DEBUG preview AFTER NormalizeData - summary=%q bytes=%v runes=%v\n", resume.Summary, []byte(resume.Summary), []rune(resume.Summary))

	data := map[string]interface{}{"Resume": resume}
	renderTemplate(w, "preview.html", data)
}

// generateHandler processes the form, generates a PDF, and sends it back to the client.
func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}
	resume := &models.Resume{
		FullName: r.FormValue("fullname"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		City:     r.FormValue("city"),
		Summary:  r.FormValue("summary"),
		Skills:   strings.Split(r.FormValue("skills"), ","),
		Extra:    r.FormValue("extra"),
		Template: r.FormValue("template"),
	}
	for i := range r.MultipartForm.Value["edu_institution"] {
		var startYear, endYear, degree string
		if len(r.MultipartForm.Value["edu_degree"]) > i {
			degree = r.MultipartForm.Value["edu_degree"][i]
		}
		if len(r.MultipartForm.Value["edu_start_year"]) > i {
			startYear = r.MultipartForm.Value["edu_start_year"][i]
		}
		if len(r.MultipartForm.Value["edu_end_year"]) > i {
			endYear = r.MultipartForm.Value["edu_end_year"][i]
		}
		resume.Education = append(resume.Education, models.EducationEntry{
			Institution: r.MultipartForm.Value["edu_institution"][i],
			Degree:      degree,
			StartYear:   startYear,
			EndYear:     endYear,
		})
	}
	for i := range r.MultipartForm.Value["exp_company"] {
		var position, period, description string
		if len(r.MultipartForm.Value["exp_position"]) > i {
			position = r.MultipartForm.Value["exp_position"][i]
		}
		if len(r.MultipartForm.Value["exp_period"]) > i {
			period = r.MultipartForm.Value["exp_period"][i]
		}
		if len(r.MultipartForm.Value["exp_desc"]) > i {
			description = r.MultipartForm.Value["exp_desc"][i]
		}
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Company:     r.MultipartForm.Value["exp_company"][i],
			Position:    position,
			Period:      period,
			Description: description,
		})
	}
	if file, hdr, err := r.FormFile("photo"); err == nil {
		resume.Photo = hdr
		file.Close()
	}

	// Apply formatting and normalization
	resume.NormalizeData()

	if err := resume.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pdfBytes, err := pdf.Generate(resume)
	if err != nil {
		log.Printf("PDF generation error: %v", err)
		http.Error(w, fmt.Sprintf("failed to generate pdf: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
	w.Write(pdfBytes)
}

// helper to render templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// provide helper functions for templates
	funcMap := template.FuncMap{
		"join": strings.Join,
	}
	t, err := template.New(tmpl).Funcs(funcMap).ParseFiles("templates/" + tmpl)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
