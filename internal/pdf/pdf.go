package pdf

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	case "professional":
		return generateProfessional(r)
	case "creative":
		return generateCreative(r)
	default:
		return generateClassic(r)
	}
}

// getOptimalFontSizeForName selects appropriate font size based on name length
func getOptimalFontSizeForName(name string) float64 {
	// Adjust font size based on full name length to prevent bad hyphenation
	nameLen := len(name)
	if nameLen > 50 {
		return 12.0 // Very long names
	}
	if nameLen > 40 {
		return 14.0 // Long names
	}
	return 18.0 // Standard size
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

	// Try to add photo at top right
	hasPhoto := false
	if r.Photo != nil {
		_ = addPhotoToPage(pdf, r) // height is currently unused
		hasPhoto = true
	}

	// Determine text width based on photo presence: we must not let any text extend
	// into the horizontal region occupied by the picture. photoX = pageW - imgW - 10,
	// so available width = photoX - leftMargin (15).
	textWidth := 170.0
	nameWidth := 170.0 // Separate width for name to prevent hyphenation issues
	if hasPhoto {
		pageW, _ := pdf.GetPageSize()
		imgW := 50.0                  // must match width used in addPhotoToPage
		textWidth = pageW - imgW - 25 // left margin (15) + gap (10)
		nameWidth = pageW * 0.50      // Use 50% of page width for better word breaking
		if textWidth < 0 {
			textWidth = 0
		}
		if nameWidth < 0 {
			nameWidth = 0
		}
	}

	// Name - use optimal font size based on name length to avoid hyphenation issues
	fontSize := getOptimalFontSizeForName(r.FullName)
	pdf.SetFont(fontName, "B", fontSize)
	pdf.SetXY(15, 15)
	pdf.MultiCell(nameWidth, 7, r.FullName, "", "L", false)
	pdf.Ln(2)

	// Contact info as vertical list
	pdf.SetFont(fontName, "", 9)
	pdf.SetXY(15, pdf.GetY())
	pdf.Cell(textWidth, 4, "Телефон: "+r.Phone)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(textWidth, 4, "Email: "+r.Email)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(textWidth, 4, "Город: "+r.City)
	pdf.Ln(7)

	// Summary with bold title
	if len(r.Summary) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(textWidth, 6, "ПРОФИЛЬНАЯ ИНФОРМАЦИЯ")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 10)
		pdf.SetFillColor(255, 255, 255)
		pdf.MultiCell(textWidth, 4, r.Summary, "", "L", false)
		pdf.Ln(2)
	}

	// Experience
	if len(r.Experience) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(textWidth, 6, "ОПЫТ РАБОТЫ")
		pdf.Ln(7)
		for _, exp := range r.Experience {
			pdf.SetFont(fontName, "B", 10)
			header := exp.Position
			if exp.Company != "" {
				header = exp.Position + " — " + exp.Company
			}
			pdf.MultiCell(textWidth, 4, header, "", "L", false)
			pdf.SetFont(fontName, "", 9)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(textWidth, 4, exp.Period)
			pdf.Ln(4)
			for _, achievement := range exp.Achievements {
				pdf.MultiCell(textWidth-3, 3, "• "+achievement, "", "L", false)
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
		pdf.Cell(textWidth, 6, "ОБРАЗОВАНИЕ")
		pdf.Ln(7)
		for _, edu := range r.Education {
			pdf.SetFont(fontName, "B", 10)
			pdf.MultiCell(textWidth, 4, edu.School, "", "L", false)
			pdf.SetFont(fontName, "", 9)
			if edu.Year != "" {
				pdf.MultiCell(textWidth, 3, edu.Year, "", "L", false)
			}
			pdf.Ln(2)
		}
		pdf.Ln(2)
	}

	// Skills
	if len(r.Skills) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(textWidth, 6, "НАВЫКИ")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 9)
		skillText := ""
		for i, skill := range r.Skills {
			skill = strings.TrimSpace(skill)
			if len(skill) > 0 {
				skillText += skill
				if i < len(r.Skills)-1 {
					skillText += ", "
				}
			}
		}
		pdf.MultiCell(textWidth, 4, skillText, "", "L", false)
		pdf.Ln(2)
	}

	// Extra
	if len(r.Extra) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetXY(15, pdf.GetY())
		pdf.Cell(textWidth, 6, "ДОПОЛНИТЕЛЬНО")
		pdf.Ln(7)
		pdf.SetFont(fontName, "", 9)
		pdf.MultiCell(textWidth, 4, r.Extra, "", "L", false)
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

	// Header section with background - increased height for long names
	pdf.SetFillColor(30, 90, 150)
	pdf.Rect(0, 0, 210, 65, "F")

	// Photo in header
	if r.Photo != nil {
		addPhotoToModernHeader(pdf, r)
	}

	// Name in header
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(fontName, "B", 18)
	pdf.SetXY(15, 10)
	nameWidth := 180.0
	if r.Photo != nil {
		nameWidth = 120.0 // Narrower width when photo is present
	}
	pdf.MultiCell(nameWidth, 7, r.FullName, "", "L", false)

	// Contact in header (stacked vertically)
	pdf.SetFont(fontName, "", 9)
	yPos := pdf.GetY() + 2
	if yPos > 38 {
		yPos = 38 // Ensure contact info fits
	}
	pdf.SetXY(15, yPos)
	pdf.Cell(nameWidth, 4, "Телефон: "+r.Phone)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(nameWidth, 4, "Email: "+r.Email)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(nameWidth, 4, "Город: "+r.City)

	// Move main content start below header
	pdf.SetY(70)

	// Reset colors
	pdf.SetTextColor(0, 0, 0)
	pdf.SetMargins(15, 50, 15)
	pdf.SetFont(fontName, "", 10)

	// Summary
	if len(r.Summary) > 0 {
		sectionHeaderModern(pdf, fontName, "КРАТКАЯ ИНФОРМАЦИЯ")
		pdf.SetFont(fontName, "", 9)
		// Limit width if photo is present to avoid overlay
		summaryWidth := 170.0
		if r.Photo != nil {
			summaryWidth = 120.0 // Narrower width when photo is present
		}
		pdf.MultiCell(summaryWidth, 4, r.Summary, "", "L", false)
		pdf.Ln(4)
	}

	// Experience
	if len(r.Experience) > 0 {
		sectionHeaderModern(pdf, fontName, "ОПЫТ РАБОТЫ")
		for i, exp := range r.Experience {
			pdf.SetFont(fontName, "B", 10)
			pdf.SetTextColor(30, 90, 150)
			header := exp.Position
			if exp.Company != "" {
				header = exp.Position + " — " + exp.Company
			}
			pdf.MultiCell(170, 4, header, "", "L", false)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(100, 100, 100)
			pdf.SetXY(15, pdf.GetY())
			pdf.Cell(170, 3, exp.Period)
			pdf.Ln(3)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 9)
			for _, achievement := range exp.Achievements {
				pdf.MultiCell(167, 3, "• "+achievement, "", "L", false)
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
			pdf.MultiCell(170, 4, edu.School, "", "L", false)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 9)
			if edu.Year != "" {
				pdf.MultiCell(167, 3, edu.Year, "", "L", false)
			}
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
			skill = strings.TrimSpace(skill)
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

// photoPdfImageType returns gofpdf Image() type: "PNG" or "JPG" (required when temp path has no extension).
func photoPdfImageType(r *models.Resume) string {
	if r == nil || r.Photo == nil {
		return "JPG"
	}
	fn := strings.ToLower(r.Photo.Filename)
	if strings.HasSuffix(fn, ".png") {
		return "PNG"
	}
	ct := strings.ToLower(r.Photo.Header.Get("Content-Type"))
	if strings.Contains(ct, "png") {
		return "PNG"
	}
	return "JPG"
}

func photoTempPattern(r *models.Resume) string {
	if photoPdfImageType(r) == "PNG" {
		return "resume-photo-*.png"
	}
	return "resume-photo-*.jpg"
}

// writePhotoTempFile copies the uploaded photo to a unique temp path with correct extension for gofpdf.
func writePhotoTempFile(r *models.Resume) (path string, cleanup func(), ok bool) {
	src, err := r.Photo.Open()
	if err != nil {
		return "", nil, false
	}
	defer src.Close()
	dst, err := os.CreateTemp("", photoTempPattern(r))
	if err != nil {
		return "", nil, false
	}
	path = dst.Name()
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(path)
		return "", nil, false
	}
	if err := dst.Close(); err != nil {
		os.Remove(path)
		return "", nil, false
	}
	return path, func() { os.Remove(path) }, true
}

// addPhotoToPage adds a photo to the classic template.
// Returns the height of the photo in millimeters.
func addPhotoToPage(pdf *gofpdf.Fpdf, r *models.Resume) float64 {
	tmpFile, cleanup, ok := writePhotoTempFile(r)
	if !ok {
		return 0
	}
	defer cleanup()

	// Фото в правом верхнем углу: ширина 50мм (высота вычисляется автоматически)
	// чуть больше, чтобы соответствовать увеличенному превью
	imgW := 50.0 // мм (~200px)
	pageW, _ := pdf.GetPageSize()
	x := pageW - imgW - 10 // 10мм от правого края
	y := 15.0
	pdf.Image(tmpFile, x, y, imgW, 0, false, photoPdfImageType(r), 0, "")
	// возвращаем условную высоту (для учета при сужении текста)
	return 60.0
}

// addPhotoToModernHeader adds a photo to the modern header.
// Returns the height of the photo in millimeters.
func addPhotoToModernHeader(pdf *gofpdf.Fpdf, r *models.Resume) float64 {
	tmpFile, cleanup, ok := writePhotoTempFile(r)
	if !ok {
		return 0
	}
	defer cleanup()

	// Фото в правом верхнем углу: ширина 50мм (высота вычисляется автоматически)
	imgW := 50.0 // мм (~200px)
	pageW, _ := pdf.GetPageSize()
	x := pageW - imgW - 10 // 10мм от правого края
	y := 10.0
	pdf.Image(tmpFile, x, y, imgW, 0, false, photoPdfImageType(r), 0, "")
	return 60.0
}

// addPhotoToCreativeSidebar adds a photo to the left in the creative header.
func addPhotoToCreativeSidebar(pdf *gofpdf.Fpdf, r *models.Resume) float64 {
	tmpFile, cleanup, ok := writePhotoTempFile(r)
	if !ok {
		return 0
	}
	defer cleanup()

	// Place photo on the left: width 50mm
	imgW := 50.0
	x := 15.0
	y := 10.0
	pdf.Image(tmpFile, x, y, imgW, 0, false, photoPdfImageType(r), 0, "")
	// Return height for layout purposes
	return 60.0
}

// addPhotoToProfessionalSidebar adds a smaller photo inside the left sidebar.
func addPhotoToProfessionalSidebar(pdf *gofpdf.Fpdf, r *models.Resume) float64 {
	tmpFile, cleanup, ok := writePhotoTempFile(r)
	if !ok {
		return 0
	}
	defer cleanup()

	// Place photo in sidebar: width 40mm
	imgW := 40.0
	x := 10.0
	y := 10.0
	pdf.Image(tmpFile, x, y, imgW, 0, false, photoPdfImageType(r), 0, "")
	// approximate bottom of image for layout purposes
	return y + imgW + 10
}

// generateProfessional creates a professional resume with sidebar layout.
func generateProfessional(r *models.Resume) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath := getFontPath()
	fontName := tryAddUTF8Font(pdf, "DejaVu", "DejaVuSans.ttf", "DejaVuSans-Bold.ttf", fontPath)

	pdf.AddPage()

	// Left sidebar with dark background
	pdf.SetFillColor(45, 45, 48)
	pdf.Rect(0, 0, 70, 297, "F")

	// Optional photo in sidebar
	sidebarStartY := 10.0
	if r.Photo != nil {
		bottom := addPhotoToProfessionalSidebar(pdf, r)
		sidebarStartY = bottom + 4
	}

	// Sidebar: Contact (white text on dark background)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(fontName, "B", 9)
	pdf.SetXY(5, sidebarStartY)
	pdf.Cell(60, 4, "КОНТАКТЫ")
	pdf.Ln(4)

	pdf.SetFont(fontName, "", 8)
	pdf.SetXY(5, pdf.GetY())
	pdf.MultiCell(60, 3, "Телефон: "+r.Phone, "", "L", false)
	pdf.SetXY(5, pdf.GetY())
	pdf.MultiCell(60, 3, "Email: "+r.Email, "", "L", false)
	pdf.SetXY(5, pdf.GetY())
	pdf.MultiCell(60, 3, "Город: "+r.City, "", "L", false)

	// Sidebar: Skills
	if len(r.Skills) > 0 {
		pdf.SetXY(5, pdf.GetY()+2)
		pdf.SetFont(fontName, "B", 9)
		pdf.Cell(60, 4, "НАВЫКИ")
		pdf.Ln(4)
		pdf.SetFont(fontName, "", 8)
		for _, skill := range r.Skills {
			pdf.SetXY(5, pdf.GetY())
			pdf.MultiCell(60, 3, "• "+skill, "", "L", false)
		}
	}

	// Main content area (no duplicate contacts here)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(75, 10)

	// Header with name only
	pdf.SetFont(fontName, "B", 18)
	pdf.MultiCell(130, 7, r.FullName, "", "L", false)
	pdf.Ln(2)

	// Summary
	if r.Summary != "" {
		pdf.SetXY(75, pdf.GetY())
		pdf.SetFont(fontName, "B", 10)
		pdf.SetTextColor(45, 45, 48)
		pdf.Cell(130, 4, "ПРОФИЛЬ")
		pdf.Ln(3)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		pdf.SetXY(75, pdf.GetY())
		pdf.MultiCell(130, 3, r.Summary, "", "L", false)
		pdf.Ln(2)
	}

	// Experience
	if len(r.Experience) > 0 {
		pdf.SetXY(75, pdf.GetY())
		pdf.SetFont(fontName, "B", 10)
		pdf.SetTextColor(45, 45, 48)
		pdf.Cell(130, 4, "ОПЫТ РАБОТЫ")
		pdf.Ln(3)
		pdf.SetTextColor(0, 0, 0)

		for _, exp := range r.Experience {
			pdf.SetXY(75, pdf.GetY())
			pdf.SetFont(fontName, "B", 9)
			header := exp.Position
			if exp.Company != "" {
				header = exp.Position + " • " + exp.Company
			}
			pdf.MultiCell(130, 4, header, "", "L", false)

			pdf.SetXY(75, pdf.GetY())
			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(100, 100, 100)
			pdf.MultiCell(130, 3, exp.Period, "", "L", false)
			pdf.SetTextColor(0, 0, 0)

			for _, achievement := range exp.Achievements {
				pdf.SetXY(75, pdf.GetY())
				pdf.SetFont(fontName, "", 8)
				pdf.MultiCell(130, 3, "• "+achievement, "", "L", false)
			}
			pdf.Ln(1)
		}
	}

	// Education
	if len(r.Education) > 0 {
		pdf.SetXY(75, pdf.GetY()+1)
		pdf.SetFont(fontName, "B", 10)
		pdf.SetTextColor(45, 45, 48)
		pdf.Cell(130, 4, "ОБРАЗОВАНИЕ")
		pdf.Ln(4)
		pdf.SetTextColor(0, 0, 0)

		for _, edu := range r.Education {
			currentY := pdf.GetY()
			// Check if we need a new page
			if currentY > 270 {
				pdf.AddPage()
				pdf.SetXY(75, 15)
				currentY = 15
			}

			pdf.SetXY(75, currentY)
			pdf.SetFont(fontName, "B", 8)
			pdf.MultiCell(130, 3, edu.School, "", "L", false)

			if edu.Year != "" {
				currentY = pdf.GetY()
				if currentY > 275 {
					pdf.AddPage()
					pdf.SetXY(75, 15)
				}
				pdf.SetXY(75, pdf.GetY())
				pdf.SetFont(fontName, "", 7)
				pdf.MultiCell(127, 2, edu.Year, "", "L", false)
			}
			pdf.Ln(1)
		}
	}

	if len(r.Extra) > 0 {
		pdf.SetXY(75, pdf.GetY()+2)
		pdf.SetFont(fontName, "B", 10)
		pdf.SetTextColor(45, 45, 48)
		pdf.Cell(130, 4, "ДОПОЛНИТЕЛЬНО")
		pdf.Ln(3)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		pdf.SetXY(75, pdf.GetY())
		pdf.MultiCell(130, 3, r.Extra, "", "L", false)
	}

	return closePDF(pdf)
}

// generateCreative creates a colorful resume with accent colors for each section.
func generateCreative(r *models.Resume) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath := getFontPath()
	fontName := tryAddUTF8Font(pdf, "DejaVu", "DejaVuSans.ttf", "DejaVuSans-Bold.ttf", fontPath)

	pdf.AddPage()

	// Colorful top bars (thin accent)
	colors := []struct{ R, G, B int }{
		{255, 107, 107}, // red
		{255, 193, 7},   // yellow
		{66, 133, 244},  // blue
	}
	barWidth := 210.0 / float64(len(colors))
	for i, col := range colors {
		pdf.SetFillColor(col.R, col.G, col.B)
		pdf.Rect(float64(i)*barWidth, 0, barWidth, 4, "F")
	}

	// Header section with light background - increased height for long names
	pdf.SetFillColor(248, 250, 252)
	pdf.Rect(0, 4, 210, 70, "F")

	// Photo in header (left side)
	hasPhoto := false
	if r.Photo != nil {
		_ = addPhotoToCreativeSidebar(pdf, r)
		hasPhoto = true
	}

	// Name in header
	pdf.SetTextColor(255, 107, 107)
	pdf.SetFont(fontName, "B", 20)
	textX := 70.0
	if !hasPhoto {
		textX = 15.0
	}
	pdf.SetXY(textX, 8)
	textWidth := 125.0
	if !hasPhoto {
		textWidth = 180.0
	}
	pdf.MultiCell(textWidth, 7, r.FullName, "", "L", false)

	// Contact in header
	pdf.SetFont(fontName, "", 9)
	pdf.SetTextColor(66, 133, 244)
	yPos := pdf.GetY() + 1
	if yPos > 40 {
		yPos = 40 // Ensure contact info doesn't go too far down
	}
	pdf.SetXY(textX, yPos)
	contactText := "Город: " + r.City + " | Телефон: " + r.Phone + " | Email: " + r.Email
	pdf.MultiCell(textWidth, 4, contactText, "", "L", false)

	// Move main content start below header
	if hasPhoto {
		pdf.SetY(95)
	} else {
		pdf.SetY(52)
	}

	// Reset colors
	pdf.SetTextColor(0, 0, 0)
	pdf.SetMargins(15, 50, 15)
	pdf.SetFont(fontName, "", 10)

	// Define section colors
	summaryColor := struct{ R, G, B int }{255, 107, 107} // red
	expColor := struct{ R, G, B int }{66, 133, 244}      // blue
	eduColor := struct{ R, G, B int }{255, 193, 7}       // orange
	skillsColor := struct{ R, G, B int }{40, 167, 69}    // green

	// Summary
	if len(r.Summary) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetTextColor(summaryColor.R, summaryColor.G, summaryColor.B)
		pdf.Cell(170, 5, "КРАТКАЯ ИНФОРМАЦИЯ")
		pdf.Ln(6)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		// Limit width if photo is present to avoid overlay
		summaryWidth := 170.0
		if r.Photo != nil {
			summaryWidth = 120.0 // Narrower width when photo is present
		}
		pdf.MultiCell(summaryWidth, 4, r.Summary, "", "L", false)
		pdf.Ln(3)
	}

	// Experience
	if len(r.Experience) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetTextColor(expColor.R, expColor.G, expColor.B)
		pdf.Cell(170, 5, "ОПЫТ РАБОТЫ")
		pdf.Ln(5)

		for i, exp := range r.Experience {
			// Colored underline for experience entry
			y := pdf.GetY()
			pdf.SetDrawColor(expColor.R, expColor.G, expColor.B)
			pdf.Line(15, y+4, 185, y+4)
			pdf.SetDrawColor(0, 0, 0)

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "B", 10)
			header := exp.Position
			if exp.Company != "" {
				header = exp.Position + " • " + exp.Company
			}
			pdf.MultiCell(170, 5, header, "", "L", false)

			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(120, 120, 120)
			pdf.MultiCell(170, 3, exp.Period, "", "L", false)

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "", 9)
			for _, achievement := range exp.Achievements {
				pdf.MultiCell(170, 3, "• "+achievement, "", "L", false)
			}
			if i < len(r.Experience)-1 {
				pdf.Ln(2)
			} else {
				pdf.Ln(3)
			}
		}
	}

	// Education
	if len(r.Education) > 0 {
		pdf.SetFont(fontName, "B", 10)
		pdf.SetTextColor(eduColor.R, eduColor.G, eduColor.B)
		pdf.Cell(170, 4, "ОБРАЗОВАНИЕ")
		pdf.Ln(4)

		for _, edu := range r.Education {
			currentY := pdf.GetY()
			// Check if we need a new page
			if currentY > 270 {
				pdf.AddPage()
				pdf.SetY(15)
			}

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont(fontName, "B", 9)
			pdf.MultiCell(167, 3, edu.School, "", "L", false)

			pdf.SetFont(fontName, "", 8)
			pdf.SetTextColor(0, 0, 0)
			if edu.Year != "" {
				currentY = pdf.GetY()
				if currentY > 275 {
					pdf.AddPage()
					pdf.SetY(15)
				}
				pdf.MultiCell(165, 2, edu.Year, "", "L", false)
				pdf.Ln(2)
			}
		}
	}

	// Skills
	if len(r.Skills) > 0 {
		pdf.SetFont(fontName, "B", 11)
		pdf.SetTextColor(skillsColor.R, skillsColor.G, skillsColor.B)
		pdf.Cell(170, 5, "НАВЫКИ")
		pdf.Ln(5)

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		x := 15.0
		y := pdf.GetY()
		for _, skill := range r.Skills {
			skill = strings.TrimSpace(skill)
			if len(skill) == 0 {
				continue
			}
			w := pdf.GetStringWidth(skill) + 4
			if x+w > 185 {
				y += 6
				x = 15
				pdf.SetXY(x, y)
			}
			pdf.SetFillColor(230, 240, 255)
			pdf.SetTextColor(skillsColor.R, skillsColor.G, skillsColor.B)
			pdf.SetFont(fontName, "", 8)
			pdf.SetXY(x, y)
			pdf.Cell(w, 5, skill)
			x += w + 2
			pdf.SetXY(x, y)
		}
		pdf.SetXY(15, y+8)
		pdf.SetTextColor(0, 0, 0)
	}

	// Extra
	if len(r.Extra) > 0 {
		pdf.Ln(2)
		pdf.SetFont(fontName, "B", 11)
		pdf.SetTextColor(255, 107, 107)
		pdf.Cell(170, 5, "ДОПОЛНИТЕЛЬНО")
		pdf.Ln(5)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(fontName, "", 9)
		pdf.MultiCell(170, 4, r.Extra, "", "L", false)
	}

	return closePDF(pdf)
}

// closePDF converts a PDF to bytes.
func closePDF(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
