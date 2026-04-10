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
	"unicode/utf8"
)

// list of allowed cities for autocomplete/validation
var allowedCities = []string{
	"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург",
	"Казань", "Нижний Новгород", "Челябинск", "Самара",
	"Омск", "Ростов-на-Дону", "Уфа", "Красноярск",
	"Воронеж", "Пермь", "Волгоград", "Краснодар",
	"Саратов", "Тюмень", "Тольятти", "Ижевск",
	"Барнаул", "Ульяновск", "Иркутск", "Хабаровск",
	"Ярославль", "Владивосток", "Махачкала", "Томск",
}

func isValidCity(city string) bool {
	city = strings.TrimSpace(city)
	if city == "" {
		return true // empty allowed, required constraint handled elsewhere
	}
	for _, c := range allowedCities {
		if strings.EqualFold(c, city) {
			return true
		}
	}
	return false
}

// Resume holds all fields collected from the form.
type Resume struct {
	FullName   string
	Phone      string
	Email      string
	City       string
	Summary    string
	Skills     []string
	Extra      string
	Photo      *multipart.FileHeader
	PhotoURL   template.URL // Use template.URL to avoid HTML escaping
	Template   string       // classic or modern
	Education  []EducationEntry
	Experience []ExperienceEntry
}

type EducationEntry struct {
	School string // "Санкт-Петербургский государственный университет"
	Year   string // "Март 2021 - настоящее время" or "2021" or "2015 - 2019"
}

type ExperienceEntry struct {
	Position     string   // "Senior Full-Stack Developer"
	Company      string   // "ООО TechnoSoft"
	Period       string   // "Март 2021 — настоящее время"
	Achievements []string // list of accomplishments/responsibilities
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
		// Handle hyphenated words (e.g., "Константинопольский-бердичевский")
		if strings.Contains(part, "-") {
			hyphenParts := strings.Split(part, "-")
			for j, hyphenPart := range hyphenParts {
				hyphenParts[j] = capitalizeSingleWord(hyphenPart, specialLowercasePrefixes)
			}
			parts[i] = strings.Join(hyphenParts, "-")
		} else if strings.Contains(part, "'") {
			// Handle apostrophes similarly (e.g., "O'Brien")
			apostropheParts := strings.Split(part, "'")
			for j, apostrophePart := range apostropheParts {
				apostropheParts[j] = capitalizeSingleWord(apostrophePart, specialLowercasePrefixes)
			}
			parts[i] = strings.Join(apostropheParts, "'")
		} else {
			parts[i] = capitalizeSingleWord(part, specialLowercasePrefixes)
		}
	}
	return strings.Join(parts, " ")
}

// capitalizeSingleWord capitalizes a single word, respecting special lowercase prefixes
func capitalizeSingleWord(word string, specialPrefixes map[string]bool) string {
	if len(word) == 0 {
		return word
	}

	lowerWord := strings.ToLower(word)
	if specialPrefixes[lowerWord] {
		return lowerWord
	}

	rs := []rune(word)
	first := unicode.ToUpper(rs[0])
	rest := ""
	if len(rs) > 1 {
		rest = strings.ToLower(string(rs[1:]))
	}
	return string(first) + rest
}

// capitalizeCity capitalizes city name
func capitalizeCity(city string) string {
	if len(city) == 0 {
		return city
	}
	// capitalize every word, preserving spaces and hyphens
	var b strings.Builder
	capitalizeNext := true
	for _, r := range city {
		if capitalizeNext && unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
		if r == ' ' || r == '-' || r == '\t' {
			capitalizeNext = true
		}
	}
	return b.String()
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

// validateEducationYear checks if education year/period is valid
func validateEducationYear(year string) error {
	if year == "" {
		return nil
	}

	trimmedYear := strings.TrimSpace(year)
	currentYear := time.Now().Year()

	// Check for "месяц YYYY - н.в." or "месяц YYYY - настоящее время" format (current/present)
	if strings.Contains(trimmedYear, "н.в.") || strings.Contains(trimmedYear, "настоящее время") {
		// Extract start year from format like "Март 2021 - н.в."
		parts := strings.Split(trimmedYear, "-")
		if len(parts) < 2 {
			return errors.New("invalid education period format")
		}
		startPart := strings.TrimSpace(parts[0])
		startYear := extractYear(startPart)
		if startYear == 0 {
			return errors.New("education start year must be a valid number in format YYYY")
		}
		if startYear > currentYear {
			return errors.New("education start year cannot be in the future")
		}
		return nil
	}

	// Check for "YYYY - YYYY" or "Month YYYY - Month YYYY" format
	if strings.Contains(trimmedYear, "-") {
		parts := strings.Split(trimmedYear, "-")
		if len(parts) != 2 {
			return errors.New("invalid education period format, use 'YYYY - YYYY' or 'Month YYYY - Month YYYY' or 'Month YYYY - н.в.'")
		}

		startYear := extractYear(strings.TrimSpace(parts[0]))
		endYear := extractYear(strings.TrimSpace(parts[1]))

		if startYear == 0 || endYear == 0 {
			return errors.New("education years must contain valid numbers in format YYYY")
		}

		if startYear >= endYear {
			return errors.New("education start year must be before end year")
		}

		if endYear > currentYear {
			return errors.New("education end year cannot be in the future")
		}

		if startYear > currentYear {
			return errors.New("education start year cannot be in the future")
		}

		return nil
	}

	// Single year format "YYYY"
	year_int, err := strconv.Atoi(trimmedYear)
	if err != nil {
		return errors.New("education year must be a valid number or period")
	}

	if year_int > currentYear {
		return errors.New("education year cannot be in the future")
	}

	return nil
}

// extractYear extracts a 4-digit year from a string (handles both "YYYY" and "Month YYYY" formats)
func extractYear(str string) int {
	// Look for 4 consecutive digits
	re := regexp.MustCompile(`\b\d{4}\b`)
	match := re.FindString(str)
	if match == "" {
		return 0
	}
	year, err := strconv.Atoi(match)
	if err != nil {
		return 0
	}
	return year
}

// validateWorkPeriod checks if work period is valid
func validateWorkPeriod(period string) error {
	if period == "" {
		return nil
	}

	trimmedPeriod := strings.TrimSpace(period)
	currentYear := time.Now().Year()

	// Check for "месяц YYYY - н.в." or "месяц YYYY - настоящее время" format (current/present)
	if strings.Contains(trimmedPeriod, "н.в.") || strings.Contains(trimmedPeriod, "настоящее время") {
		// Extract start year from format like "Март 2021 - н.в."
		parts := strings.Split(trimmedPeriod, "-")
		if len(parts) < 2 {
			return errors.New("invalid work period format")
		}
		startPart := strings.TrimSpace(parts[0])
		startYear := extractYear(startPart)
		if startYear == 0 {
			return errors.New("work start year must be a valid number in format YYYY")
		}
		if startYear > currentYear {
			return errors.New("work start year cannot be in the future")
		}
		return nil
	}

	// Check for "YYYY - YYYY" or "Month YYYY - Month YYYY" format
	if strings.Contains(trimmedPeriod, "-") {
		parts := strings.Split(trimmedPeriod, "-")
		if len(parts) != 2 {
			return errors.New("invalid work period format, use 'YYYY - YYYY' or 'Month YYYY - Month YYYY' or 'Month YYYY - н.в.'")
		}

		startYear := extractYear(strings.TrimSpace(parts[0]))
		endYear := extractYear(strings.TrimSpace(parts[1]))

		if startYear == 0 || endYear == 0 {
			return errors.New("work period years must contain valid numbers in format YYYY")
		}

		if startYear >= endYear {
			return errors.New("work start year must be before end year")
		}

		if endYear > currentYear {
			return errors.New("work end year cannot be in the future")
		}

		if startYear > currentYear {
			return errors.New("work start year cannot be in the future")
		}

		return nil
	}

	// Single year format "YYYY"
	year_int, err := strconv.Atoi(trimmedPeriod)
	if err != nil {
		return errors.New("work period must be a valid year or period")
	}

	if year_int > currentYear {
		return errors.New("work year cannot be in the future")
	}

	return nil
}

// Validate performs basic checks on the resume fields according to constraints.
func (r *Resume) Validate() error {
	if utf8.RuneCountInString(r.FullName) == 0 || utf8.RuneCountInString(r.FullName) > 100 {
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
	if len(r.City) == 0 {
		return errors.New("city is required")
	}
	if !isValidCity(r.City) {
		return errors.New("city is not in the allowed list")
	}
	if len(r.Summary) > 1000 {
		return errors.New("summary must be <=1000 characters")
	}

	// Validate education years
	for _, edu := range r.Education {
		if edu.Year != "" {
			if err := validateEducationYear(edu.Year); err != nil {
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
