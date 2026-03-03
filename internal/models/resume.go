package models

import (
	"errors"
	"image"
	"mime/multipart"
	"strings"
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
	Template   string // classic or modern
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

// Validate performs basic checks on the resume fields according to constraints.
func (r *Resume) Validate() error {
	if len(r.FullName) == 0 || len(r.FullName) > 100 {
		return errors.New("full name is required and must be <=100 characters")
	}
	if len(r.Email) == 0 || len(r.Email) > 100 {
		return errors.New("email is required and must be <=100 characters")
	}
	if len(r.Phone) == 0 || len(r.Phone) > 50 {
		return errors.New("phone is required and must be <=50 characters")
	}
	if len(r.City) > 100 {
		return errors.New("city must be <=100 characters")
	}
	if len(r.Summary) > 1000 {
		return errors.New("summary must be <=1000 characters")
	}
	// Additional validation can be added for education/experience
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
