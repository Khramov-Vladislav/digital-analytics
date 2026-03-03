package pdf

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Khramov-Vladislav/digital-analytics/internal/models"
	gofpdf "github.com/jung-kurt/gofpdf"
)

// getFontPath returns the absolute path to font files.
func getFontPath() string {
	// Try relative path from working directory
	if _, err := os.Stat("fonts"); err == nil {
		return "fonts"
	}
	// Try relative to binary location
	if ex, err := os.Executable(); err == nil {
		exePath := filepath.Dir(ex)
		if _, err := os.Stat(filepath.Join(exePath, "fonts")); err == nil {
			return filepath.Join(exePath, "fonts")
		}
	}
	return "fonts" // fallback
}

// tryAddUTF8Font attempts to add a UTF8 font, falling back to Helvetica if not available.
func tryAddUTF8Font(pdf *gofpdf.Fpdf, fontName, fontFile, boldFile, fontPath string) string {
	regularPath := filepath.Join(fontPath, fontFile)
	boldPath := filepath.Join(fontPath, boldFile)

	// Check if regular font exists and is valid
	if info, err := os.Stat(regularPath); err == nil && info.Size() > 100000 {
		// File exists and is larger than 100KB - likely valid TTF
		pdf.AddUTF8Font(fontName, "", regularPath)
		if boldFile != "" {
			// Try to add bold variant
			pdf.AddUTF8Font(fontName, "B", boldPath)
		}
		log.Printf("Using UTF8 font: %s\n", fontName)
		return fontName
	}

	log.Printf("Warning: UTF8 font '%s' not found or invalid, using Helvetica fallback\n", fontName)
	return "Helvetica"
}

// Generate takes a resume model and returns a PDF byte slice.
func Generate(r *models.Resume) ([]byte, error) {
	switch r.Template {
	case "modern":
		return generateModern(r)
	default:
		return generateClassic(r)
	}
}

// generateClassic creates a traditional ГОСТ-compliant resume style.
func generateClassic(r *models.Resume) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath := getFontPath()

	// Try to use UTF8 font, fall back to Helvetica
	fontName := tryAddUTF8Font(pdf, "DejaVu", "DejaVuSans.ttf", "DejaVuSans-Bold.ttf", fontPath)

	pdf.AddPage()

	// Margins and setup
	pdf.SetMargins(15, 15, 15)
	pdf.SetFont(fontName, "", 10)

	// Try to add photo in top right
	if r.Photo != nil {
		addPhotoToPage(pdf, r)
	}

	// Name
	pdf.SetFont(fontName, "B", 18)
	pdf.SetXY(15, 15)
	pdf.Cell(150, 10, r.FullName)
	pdf.Ln(14)

	// Contact info
	pdf.SetFont(fontName, "", 9)
	contact := fmt.Sprintf("Телефон: %s | Email: %s | Город: %s", r.Phone, r.Email, r.City)
	pdf.MultiCell(170, 4, contact, "", "L", false)
	pdf.Ln(3)

	// Summary with bold title
	if len(r.Summary) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(170, 6, "ПРОФИЛЬНАЯ ИНФОРМАЦИЯ")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 10)
		pdf.SetFillColor(255, 255, 255)
		pdf.MultiCell(170, 4, r.Summary, "", "L", false)
		pdf.Ln(2)
	}

	// Experience
	if len(r.Experience) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(170, 6, "ОПЫТ РАБОТЫ")
		pdf.Ln(7)
		for _, exp := range r.Experience {
			pdf.SetFont(fontName, "B", 10)
			pdf.MultiCell(170, 4, exp.Company+" - "+exp.Position, "", "L", false)
			pdf.SetFont(fontName, "", 9)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 4, exp.Period)
			pdf.Ln(4)
			if len(exp.Description) > 0 {
				pdf.MultiCell(170, 3, exp.Description, "", "L", false)
			}
			pdf.Ln(2)
		}
		pdf.Ln(2)
	}

	// Education
	if len(r.Education) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(170, 6, "ОБРАЗОВАНИЕ")
		pdf.Ln(7)
		for _, edu := range r.Education {
			pdf.SetFont(fontName, "B", 10)
			pdf.MultiCell(170, 4, edu.Institution, "", "L", false)
			pdf.SetFont(fontName, "", 9)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 4, edu.Degree)
			pdf.Ln(4)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 4, edu.StartYear+" - "+edu.EndYear)
			pdf.Ln(4)
		}
		pdf.Ln(2)
	}

	// Skills
	if len(r.Skills) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(170, 6, "НАВЫКИ")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 9)
		skillText := ""
		for i, skill := range r.Skills {
			skill = stripSpace(skill)
			if len(skill) > 0 {
				skillText += skill
				if i < len(r.Skills)-1 {
					skillText += ", "
				}
			}
		}
		pdf.MultiCell(170, 4, skillText, "", "L", false)
		pdf.Ln(2)
	}

	// Extra
	if len(r.Extra) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(170, 6, "ДОПОЛНИТЕЛЬНО")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 9)
		pdf.MultiCell(170, 4, r.Extra, "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generateModern creates a modern resume with colors and style.
func generateModern(r *models.Resume) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath := getFontPath()

	// Try to use UTF8 font, fall back to Helvetica
	fontName := tryAddUTF8Font(pdf, "DejaVu", "DejaVuSans.ttf", "DejaVuSans-Bold.ttf", fontPath)

	pdf.AddPage()

	// Header section with background
	pdf.SetFillColor(30, 90, 150)
	pdf.Rect(0, 0, 210, 50, "F")

	// Photo in header
	if r.Photo != nil {
		addPhotoToModernHeader(pdf, r)
	}

	// Name in header
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(fontName, "B", 20)
	pdf.SetXY(15, 18)
	pdf.Cell(150, 12, r.FullName)

	// Contact in header
	pdf.SetFont(fontName, "", 9)
	pdf.SetXY(15, 32)
	contact := fmt.Sprintf("%s | %s | %s", r.Phone, r.Email, r.City)
	pdf.MultiCell(150, 4, contact, "", "L", false)

	// Reset colors
	pdf.SetTextColor(0, 0, 0)
	pdf.SetMargins(15, 50, 15)
	pdf.SetFont(fontName, "", 10)

	// Summary
	if len(r.Summary) > 0 {
		sectionHeaderModern(pdf, fontName, "О СЕБЕ")
		pdf.SetFont(fontName, "", 9)
		pdf.MultiCell(170, 4, r.Summary, "", "L", false)
		pdf.Ln(4)
	}

	// Experience
	if len(r.Experience) > 0 {
		sectionHeaderModern(pdf, fontName, "ОПЫТ РАБОТЫ")
		for i, exp := range r.Experience {
			pdf.SetFont(fontName, "B", 10)
			pdf.SetTextColor(30, 90, 150)
			pdf.MultiCell(170, 4, exp.Company+" - "+exp.Position, "", "L", false)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(100, 100, 100)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 3, exp.Period)
			pdf.Ln(3)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 9)
			if len(exp.Description) > 0 {
				pdf.MultiCell(170, 3, exp.Description, "", "L", false)
			}
			if i < len(r.Experience)-1 {
				pdf.Ln(2)
			}
		}
		pdf.Ln(2)
	}

	// Education
	if len(r.Education) > 0 {
		sectionHeaderModern(pdf, fontName, "ОБРАЗОВАНИЕ")
		for i, edu := range r.Education {
			pdf.SetFont(fontName, "B", 10)
			pdf.SetTextColor(30, 90, 150)
			pdf.MultiCell(170, 4, edu.Institution, "", "L", false)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 9)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 4, edu.Degree)
			pdf.Ln(4)
			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(100, 100, 100)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 3, edu.StartYear+" - "+edu.EndYear)
			pdf.Ln(3)
			pdf.SetTextColor(0, 0, 0)
			if i < len(r.Education)-1 {
				pdf.Ln(2)
			}
		}
		pdf.Ln(2)
	}

	// Skills with tags
	if len(r.Skills) > 0 {
		sectionHeaderModern(pdf, fontName, "НАВЫКИ")
		x := 15.0
		y := pdf.GetY()
		for _, skill := range r.Skills {
			skill = stripSpace(skill)
			if len(skill) == 0 {
				continue
			}
			w := pdf.GetStringWidth(skill) + 4
			if x+w > 185 {
				y += 8
				x = 15
				pdf.SetXY(x, y)
			}
			pdf.SetFillColor(220, 235, 255)
			pdf.SetTextColor(30, 90, 150)
			pdf.SetFont(fontName, "", 8)
			pdf.SetXY(x, y)
			pdf.Cell(w, 6, skill)
			x += w + 2
			pdf.SetXY(x, y)
		}
		pdf.SetXY(15, y+8)
		pdf.Ln(4)
	}

	// Extra
	if len(r.Extra) > 0 {
		sectionHeaderModern(pdf, fontName, "ДОПОЛНИТЕЛЬНО")
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		pdf.MultiCell(170, 4, r.Extra, "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// sectionHeaderModern adds a styled section header for modern template.
func sectionHeaderModern(pdf *gofpdf.Fpdf, fontName, title string) {
	pdf.SetFont(fontName, "B", 11)
	pdf.SetTextColor(30, 90, 150)
	pdf.SetXY(15, pdf.GetY())
	pdf.Cell(170, 6, title)
	pdf.Ln(6)
}

// addPhotoToPage adds a photo to the classic template.
func addPhotoToPage(pdf *gofpdf.Fpdf, r *models.Resume) {
	file, err := r.Photo.Open()
	if err != nil {
		return
	}
	defer file.Close()

	tmpFile := filepath.Join(os.TempDir(), r.Photo.Filename)
	f, err := os.Create(tmpFile)
	if err != nil {
		return
	}
	_, _ = io.Copy(f, file)
	f.Close()

	// Add image: path, x, y, width, height, flow, type, link, filestr
	pdf.Image(tmpFile, 165, 15, 25, 0, false, "", 0, "")
	os.Remove(tmpFile)
}

// addPhotoToModernHeader adds a photo to the modern header.
func addPhotoToModernHeader(pdf *gofpdf.Fpdf, r *models.Resume) {
	file, err := r.Photo.Open()
	if err != nil {
		return
	}
	defer file.Close()

	tmpFile := filepath.Join(os.TempDir(), r.Photo.Filename)
	f, err := os.Create(tmpFile)
	if err != nil {
		return
	}
	_, _ = io.Copy(f, file)
	f.Close()

	// Add image in header
	pdf.Image(tmpFile, 165, 10, 28, 0, false, "", 0, "")
	os.Remove(tmpFile)
}

// stripSpace removes leading/trailing whitespace from a string.
func stripSpace(s string) string {
	var result string
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result += string(c)
		}
	}
	return result
}
