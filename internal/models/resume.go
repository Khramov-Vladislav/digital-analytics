package models

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Resume holds all information needed to generate a CV.
type Resume struct {
	FullName   string
	Phone      string
	Email      string
	City       string
	Summary    string
	Education  []EducationEntry
	Experience []ExperienceEntry
	Skills     []string
	Extra      string
	Photo      *multipart.FileHeader
	PhotoURL   template.URL // Use template.URL to avoid HTML escaping
	Template   string       // classic or modern
}

type EducationEntry struct {
	Institution string
	Degree      string
	StartYear   string
	EndYear     string
}

type ExperienceEntry struct {
	Company     string
	Position    string
	Period      string
	Description string
}

// isValidEmail checks if email format is correct
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidPhone checks if phone number is valid (international format support)
// Accepts formats like: +7 (921) 123-45-67, +1-201-555-0171, 79211234567, etc.
func isValidPhone(phone string) bool {
	// Remove all non-digit characters
	digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Phone number should have at least 10 digits
	if len(digits) < 10 {
		return false
	}

	// Phone number should not exceed 15 digits (international standard)
	if len(digits) > 15 {
		return false
	}

	// Must contain only digits in cleaned form
	digitRegex := regexp.MustCompile(`^\d+$`)
	return digitRegex.MatchString(digits)
}

// capitalizeName capitalizes name parts, respecting special cases like "van", "von", "de"
func capitalizeName(fullName string) string {
	if len(fullName) == 0 {
		return fullName
	}

	parts := strings.Fields(fullName)
	specialLowercasePrefixes := map[string]bool{
		"van": true, "von": true, "de": true, "der": true, "den": true,
		"la": true, "le": true, "du": true, "et": true, "mc": true,
	}

	for i, part := range parts {
		lowerPart := strings.ToLower(part)
		if specialLowercasePrefixes[lowerPart] {
			// Keep these prefixes in lowercase
			parts[i] = lowerPart
		} else {
			// Properly capitalize Unicode string: uppercase first rune, lowercase rest
			if len(part) > 0 {
				rs := []rune(part)
				first := unicode.ToUpper(rs[0])
				rest := ""
				if len(rs) > 1 {
					rest = strings.ToLower(string(rs[1:]))
				}
				parts[i] = string(first) + rest
			}
		}
	}
	return strings.Join(parts, " ")
}

// capitalizeCity capitalizes city name
func capitalizeCity(city string) string {
	if len(city) == 0 {
		return city
	}
	parts := strings.Split(city, "-")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			rs := []rune(part)
			first := unicode.ToUpper(rs[0])
			rest := ""
			if len(rs) > 1 {
				rest = strings.ToLower(string(rs[1:]))
			}
			parts[i] = string(first) + rest
		}
	}
	return strings.Join(parts, "-")
}

// validateYears checks if education years are logically correct
func validateYears(startYearStr, endYearStr string) error {
	if startYearStr == "" && endYearStr == "" {
		return nil // Both empty is ok
	}

	startYear, errStart := strconv.Atoi(strings.TrimSpace(startYearStr))
	endYear, errEnd := strconv.Atoi(strings.TrimSpace(endYearStr))

	if errStart != nil || errEnd != nil {
		return errors.New("education years must be valid numbers")
	}

	if startYear >= endYear {
		return errors.New("education start year must be before end year")
	}

	yearsDiff := endYear - startYear
	if yearsDiff > 6 {
		return errors.New("education duration cannot exceed 6 years")
	}

	currentYear := time.Now().Year()
	if endYear > currentYear {
		return fmt.Errorf("education end year cannot be in the future (current year: %d)", currentYear)
	}

	return nil
}

// validateWorkPeriod checks if work period is valid
func validateWorkPeriod(period string) error {
	if period == "" {
		return nil
	}

	// Check if it's "XXXX-н.в." format (current/present)
	if strings.HasSuffix(period, "н.в.") || strings.HasSuffix(period, "по наст. время") {
		parts := strings.Split(period, "-")
		if len(parts) < 2 {
			return errors.New("invalid work period format")
		}
		startYearStr := strings.TrimSpace(parts[0])
		startYear, err := strconv.Atoi(startYearStr)
		if err != nil {
			return errors.New("work period start year must be a valid number")
		}
		currentYear := time.Now().Year()
		if startYear > currentYear {
			return errors.New("work start year cannot be in the future")
		}
		return nil
	}

	// Regular format: XXXX-XXXX
	parts := strings.Split(period, "-")
	if len(parts) != 2 {
		return errors.New("invalid work period format, use 'XXXX-XXXX' or 'XXXX-н.в.'")
	}

	startYear, errStart := strconv.Atoi(strings.TrimSpace(parts[0]))
	endYear, errEnd := strconv.Atoi(strings.TrimSpace(parts[1]))

	if errStart != nil || errEnd != nil {
		return errors.New("work period years must be valid numbers")
	}

	if startYear >= endYear {
		return errors.New("work start year must be before end year")
	}

	currentYear := time.Now().Year()
	if endYear > currentYear {
		return errors.New("work end year cannot be in the future")
	}

	return nil
}

// Validate performs basic checks on the resume fields according to constraints.
func (r *Resume) Validate() error {
	if len(r.FullName) == 0 || len(r.FullName) > 100 {
		return errors.New("full name is required and must be <=100 characters")
	}

	if len(r.Email) == 0 || len(r.Email) > 100 {
		return errors.New("email is required and must be <=100 characters")
	}
	if !isValidEmail(r.Email) {
		return errors.New("email format is invalid")
	}

	if len(r.Phone) == 0 || len(r.Phone) > 50 {
		return errors.New("phone is required and must be <=50 characters")
	}
	if !isValidPhone(r.Phone) {
		return errors.New("phone number must contain 10-15 digits and be in valid international format")
	}

	if len(r.City) > 100 {
		return errors.New("city must be <=100 characters")
	}
	if len(r.Summary) > 1000 {
		return errors.New("summary must be <=1000 characters")
	}

	// Validate education years
	for _, edu := range r.Education {
		if edu.StartYear != "" || edu.EndYear != "" {
			if err := validateYears(edu.StartYear, edu.EndYear); err != nil {
				return err
			}
		}
	}

	// Validate work period
	for _, exp := range r.Experience {
		if exp.Period != "" {
			if err := validateWorkPeriod(exp.Period); err != nil {
				return err
			}
		}
	}

	if r.Photo != nil {
		// size limit 5MB
		if r.Photo.Size > 5*1024*1024 {
			return errors.New("photo must be smaller than 5MB")
		}
		// basic check for allowed extensions
		fn := strings.ToLower(r.Photo.Filename)
		if !strings.HasSuffix(fn, ".jpg") && !strings.HasSuffix(fn, ".jpeg") && !strings.HasSuffix(fn, ".png") {
			return errors.New("photo must be JPG/JPEG/PNG")
		}
		// check minimal dimensions
		file, err := r.Photo.Open()
		if err == nil {
			defer file.Close()
			if cfg, _, err := image.DecodeConfig(file); err == nil {
				if cfg.Width < 200 || cfg.Height < 200 {
					return errors.New("photo resolution must be at least 200x200")
				}
			}
		}
	}
	return nil
}

// NormalizeData applies formatting to all fields
func (r *Resume) NormalizeData() {
	r.FullName = capitalizeName(r.FullName)
	r.City = capitalizeCity(r.City)
	// Note: Validation of years will happen in Validate()
}
