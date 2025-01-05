package models

import "time"

// type User struct {
// 	ID           uint           // Standard field for the primary key
// 	Name         string         // A regular string field
// 	Email        *string        // A pointer to a string, allowing for null values
// 	Age          uint8          // An unsigned 8-bit integer
// 	Birthday     *time.Time     // A pointer to time.Time, can be null
// 	MemberNumber sql.NullString // Uses sql.NullString to handle nullable strings
// 	ActivatedAt  sql.NullTime   // Uses sql.NullTime for nullable time fields
// 	CreatedAt    time.Time      // Automatically managed by GORM for creation time
// 	UpdatedAt    time.Time      // Automatically managed by GORM for update time
// }

type User struct {
	ID      string `gorm:"primaryKey"`
	Name    string
	Email   string
	Picture string
}

type URL struct {
	ID        uint    `gorm:"primaryKey"`
	ShortCode string  `gorm:"uniqueIndex;not null"`          // Unique short URL code
	LongURL   string  `gorm:"type:text;not null"`            // Original long URL
	UserID    *string `gorm:"index"`                         // Optional, links the URL to a user
	User      *User   `gorm:"constraint:OnDelete:SET NULL;"` // Optional user association
	// Clicks      uint       `gorm:"default:0"`                     // Click count for analytics
	CreatedAt   time.Time // When the URL was created
	Description string    `gorm:"type:text"` // Optional description for the URL
}

/*
id
userid
slug	// needs to be indexed

createdAt
updatedAt
*/
