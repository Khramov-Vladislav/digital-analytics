package handlers

import (
	"fmt"
	"html/template"
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
	for i := range r.MultipartForm.Value["edu_institution"] {
		resume.Education = append(resume.Education, models.EducationEntry{
			Institution: r.MultipartForm.Value["edu_institution"][i],
			Degree:      r.MultipartForm.Value["edu_degree"][i],
			StartYear:   "", // maybe further parsing
			EndYear:     r.MultipartForm.Value["edu_years"][i],
		})
	}
	for i := range r.MultipartForm.Value["exp_company"] {
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Company:     r.MultipartForm.Value["exp_company"][i],
			Position:    r.MultipartForm.Value["exp_position"][i],
			Period:      r.MultipartForm.Value["exp_period"][i],
			Description: r.MultipartForm.Value["exp_desc"][i],
		})
	}
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
		resume.Education = append(resume.Education, models.EducationEntry{
			Institution: r.MultipartForm.Value["edu_institution"][i],
			Degree:      r.MultipartForm.Value["edu_degree"][i],
			StartYear:   "",
			EndYear:     r.MultipartForm.Value["edu_years"][i],
		})
	}
	for i := range r.MultipartForm.Value["exp_company"] {
		resume.Experience = append(resume.Experience, models.ExperienceEntry{
			Company:     r.MultipartForm.Value["exp_company"][i],
			Position:    r.MultipartForm.Value["exp_position"][i],
			Period:      r.MultipartForm.Value["exp_period"][i],
			Description: r.MultipartForm.Value["exp_desc"][i],
		})
	}
	if file, hdr, err := r.FormFile("photo"); err == nil {
		resume.Photo = hdr
		file.Close()
	}
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
