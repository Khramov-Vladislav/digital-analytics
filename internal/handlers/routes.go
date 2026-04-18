package handlers

import (
	"encoding/base64"
	"encoding/json"
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
	mux.HandleFunc("/autocomplete/cities", autocompleteCitiesHandler)
	mux.HandleFunc("/autocomplete/institutions", autocompleteInstitutionsHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// filterEmptyStrings removes empty strings from slice and trims whitespace
func filterEmptyStrings(s []string) []string {
	var result []string
	for _, v := range s {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
		Skills:   filterEmptyStrings(strings.Split(r.FormValue("skills"), ",")),
		Extra:    r.FormValue("extra"),
		Template: r.FormValue("template"),
	}

	// Parse education
	for i := range r.MultipartForm.Value["edu_school"] {
		school := r.MultipartForm.Value["edu_school"][i]
		if school == "" {
			continue
		}
		year := ""
		if len(r.MultipartForm.Value["edu_year"]) > i {
			year = strings.TrimSpace(r.MultipartForm.Value["edu_year"][i])
		}
		resume.Education = append(resume.Education, models.EducationEntry{
			School: school,
			Year:   year,
		})
	}

	// Parse experience
	for i := range r.MultipartForm.Value["exp_position"] {
		position := r.MultipartForm.Value["exp_position"][i]
		if position == "" {
			continue
		}
		var company, period string
		if len(r.MultipartForm.Value["exp_company"]) > i {
			company = r.MultipartForm.Value["exp_company"][i]
		}
		if len(r.MultipartForm.Value["exp_period"]) > i {
			period = r.MultipartForm.Value["exp_period"][i]
		}
		achievements := []string{}
		if len(r.MultipartForm.Value["exp_achievements"]) > i {
			bullets := strings.Split(r.MultipartForm.Value["exp_achievements"][i], "\n")
			for _, bullet := range bullets {
				bullet = strings.TrimSpace(bullet)
				if bullet != "" {
					achievements = append(achievements, bullet)
				}
			}
		}
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Position:     position,
			Company:      company,
			Period:       period,
			Achievements: achievements,
		})
	}

	// Handle photo with base64 conversion for preview
	if file, hdr, err := r.FormFile("photo"); err == nil {
		resume.Photo = hdr

		// Read file and convert to base64 Data URL
		data, err := io.ReadAll(file)
		file.Close()
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

	// Apply formatting and normalization
	resume.NormalizeData()

	// Skip validation in preview - just show what user has entered
	// Validation will happen when generating PDF

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
		Skills:   filterEmptyStrings(strings.Split(r.FormValue("skills"), ",")),
		Extra:    r.FormValue("extra"),
		Template: r.FormValue("template"),
	}
	for i := range r.MultipartForm.Value["edu_school"] {
		school := r.MultipartForm.Value["edu_school"][i]
		if school == "" {
			continue
		}
		year := ""
		if len(r.MultipartForm.Value["edu_year"]) > i {
			year = strings.TrimSpace(r.MultipartForm.Value["edu_year"][i])
		}
		resume.Education = append(resume.Education, models.EducationEntry{
			School: school,
			Year:   year,
		})
	}
	for i := range r.MultipartForm.Value["exp_position"] {
		position := r.MultipartForm.Value["exp_position"][i]
		if position == "" {
			continue
		}
		var company, period string
		if len(r.MultipartForm.Value["exp_company"]) > i {
			company = r.MultipartForm.Value["exp_company"][i]
		}
		if len(r.MultipartForm.Value["exp_period"]) > i {
			period = r.MultipartForm.Value["exp_period"][i]
		}
		achievements := []string{}
		if len(r.MultipartForm.Value["exp_achievements"]) > i {
			bullets := strings.Split(r.MultipartForm.Value["exp_achievements"][i], "\n")
			for _, bullet := range bullets {
				bullet = strings.TrimSpace(bullet)
				if bullet != "" {
					achievements = append(achievements, bullet)
				}
			}
		}
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Position:     position,
			Company:      company,
			Period:       period,
			Achievements: achievements,
		})
	}
	if file, hdr, err := r.FormFile("photo"); err == nil {
		resume.Photo = hdr
		file.Close()
	}

	// Apply formatting and normalization
	resume.NormalizeData()

	if err := resume.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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

// autocompleteCitiesHandler handles city autocomplete requests
func autocompleteCitiesHandler(w http.ResponseWriter, r *http.Request) {
	models.AutocompleteCities(w, r)
}

// autocompleteInstitutionsHandler handles institutions autocomplete requests
func autocompleteInstitutionsHandler(w http.ResponseWriter, r *http.Request) {
	models.AutocompleteInstitutions(w, r)
}
