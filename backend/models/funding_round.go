package models

import "gorm.io/gorm"

// FundingRound is one sourced fundraising announcement for a startup.
// AmountRaisedMinor and ValuationMinor are the smallest currency unit, such as
// cents. A nil amount means the figure was not disclosed. Zero is not used.
type FundingRound struct {
	gorm.Model
	ID                int64  `json:"id" gorm:"primary_key"`
	StartupID         int64  `json:"startupId" gorm:"not null;index"`
	RoundLabel        string `json:"roundLabel" gorm:"not null;size:80"`
	AnnouncementDate  string `json:"announcementDate" gorm:"not null;size:10"`
	AmountRaisedMinor *int64 `json:"amountRaisedMinor"`
	ValuationMinor    *int64 `json:"valuationMinor"`
	ValuationBasis    string `json:"valuationBasis" gorm:"not null;size:20"`
	Currency          string `json:"currency" gorm:"not null;size:3"`
	EvidenceURL       string `json:"evidenceUrl" gorm:"not null;size:500"`
	SourceLabel       string `json:"sourceLabel" gorm:"not null;size:120"`
	Note              string `json:"note,omitempty" gorm:"type:text"`
}
