package models

import "time"

type User struct {
	ID      string `gorm:"primaryKey"`
	Name    string
	Email   string
	Picture string
}

type Link struct {
	ID          string  `gorm:"primaryKey" validate:"required"`
	Name        string  `gorm:"not null" validate:"required,min=3,max=100" errormgs:"Name is required & has to between 3-100 characters"`
	ShortCode   string  `gorm:"uniqueIndex;not null" validate:"required,min=3,max=50,url_friendly" errormgs:"Short Code must be URL friendly & has to between 3-50 characters"`
	LongURL     string  `gorm:"type:text;not null" validate:"required,http_url,max=250" errormgs:"URL needs to be a valid HTTP URL"`
	UserID      *string `gorm:"index" validate:"required"`
	User        *User   `gorm:"constraint:OnDelete:SET NULL;"`
	CreatedAt   time.Time
	Description string `gorm:"type:text"`
	// todo: add State
}
