package models

import "gorm.io/gorm"

// Startup is an admin-managed company profile. It is not a market and does not
// participate in trading math.
type Startup struct {
	gorm.Model
	ID          int64  `json:"id" gorm:"primary_key"`
	Name        string `json:"name" gorm:"not null;size:160"`
	Slug        string `json:"slug" gorm:"not null;uniqueIndex;size:64"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	WebsiteURL  string `json:"websiteUrl,omitempty" gorm:"size:500"`
}
