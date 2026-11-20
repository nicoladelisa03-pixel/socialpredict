package startups

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	BasisPreMoney       = "pre-money"
	BasisPostMoney      = "post-money"
	BasisUnspecified    = "unspecified"
	maxRoundLabel       = 80
	maxSourceLabel      = 120
	maxFundingNote      = 500
	minAnnouncementYear = 1980
	maxAnnouncementYear = 2100
)

var (
	ErrFundingNotFound = errors.New("funding round not found")
	currencyPattern    = regexp.MustCompile(`^[A-Z]{3}$`)
)

// FundingRound is one disclosed fundraising announcement.
// Nil money fields mean the figure was not disclosed.
type FundingRound struct {
	ID                int64
	StartupID         int64
	RoundLabel        string
	AnnouncementDate  string
	AmountRaisedMinor *int64
	ValuationMinor    *int64
	ValuationBasis    string
	Currency          string
	EvidenceURL       string
	SourceLabel       string
	Note              string
}

// FundingInput is the admin payload for creating or correcting a round.
type FundingInput struct {
	RoundLabel        string
	AnnouncementDate  string
	AmountRaisedMinor *int64
	ValuationMinor    *int64
	ValuationBasis    string
	Currency          string
	EvidenceURL       string
	SourceLabel       string
	Note              string
}

// FundingHistory is the public funding read for one startup.
type FundingHistory struct {
	Rounds                  []FundingRound
	LatestReportedValuation *FundingRound
}

// CreateFundingRound stores a sourced round for a startup.
func (s *Service) CreateFundingRound(ctx context.Context, slug string, input FundingInput) (*FundingRound, error) {
	startup, err := s.startupBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	round, err := normalizeFunding(input)
	if err != nil {
		return nil, err
	}
	round.StartupID = startup.ID
	if err := s.repo.CreateFundingRound(ctx, &round); err != nil {
		return nil, err
	}
	return &round, nil
}

// UpdateFundingRound corrects an existing round. The startup slug must own it.
func (s *Service) UpdateFundingRound(ctx context.Context, slug string, id int64, input FundingInput) (*FundingRound, error) {
	startup, err := s.startupBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, ErrFundingNotFound
	}
	current, err := s.repo.GetFundingRound(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrFundingNotFound
		}
		return nil, err
	}
	if current == nil || current.StartupID != startup.ID {
		return nil, ErrFundingNotFound
	}
	round, err := normalizeFunding(input)
	if err != nil {
		return nil, err
	}
	round.ID = current.ID
	round.StartupID = startup.ID
	if err := s.repo.UpdateFundingRound(ctx, &round); err != nil {
		return nil, err
	}
	return &round, nil
}

// ListFundingHistory returns rounds from oldest to newest and the latest disclosed valuation.
func (s *Service) ListFundingHistory(ctx context.Context, slug string) (*FundingHistory, error) {
	startup, err := s.startupBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	rounds, err := s.repo.ListFundingRounds(ctx, startup.ID)
	if err != nil {
		return nil, err
	}
	if rounds == nil {
		rounds = []FundingRound{}
	}
	sort.SliceStable(rounds, func(i, j int) bool {
		if rounds[i].AnnouncementDate == rounds[j].AnnouncementDate {
			return rounds[i].ID < rounds[j].ID
		}
		return rounds[i].AnnouncementDate < rounds[j].AnnouncementDate
	})
	var latest *FundingRound
	for i := range rounds {
		if rounds[i].ValuationMinor != nil {
			copy := rounds[i]
			latest = &copy
		}
	}
	return &FundingHistory{Rounds: rounds, LatestReportedValuation: latest}, nil
}

func (s *Service) startupBySlug(ctx context.Context, slug string) (*Startup, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	normalized := NormalizeSlug(slug)
	if !validSlug(normalized) {
		return nil, ErrNotFound
	}
	return s.repo.GetBySlug(ctx, normalized)
}

func normalizeFunding(input FundingInput) (FundingRound, error) {
	label := strings.TrimSpace(input.RoundLabel)
	if label == "" || len(label) > maxRoundLabel {
		return FundingRound{}, ErrInvalidInput
	}
	announcement := strings.TrimSpace(input.AnnouncementDate)
	parsed, err := time.Parse("2006-01-02", announcement)
	if err != nil || parsed.Format("2006-01-02") != announcement {
		return FundingRound{}, ErrInvalidInput
	}
	if parsed.Year() < minAnnouncementYear || parsed.Year() > maxAnnouncementYear {
		return FundingRound{}, ErrInvalidInput
	}
	if input.AmountRaisedMinor != nil && *input.AmountRaisedMinor <= 0 {
		return FundingRound{}, ErrInvalidInput
	}
	if input.ValuationMinor != nil && *input.ValuationMinor <= 0 {
		return FundingRound{}, ErrInvalidInput
	}
	basis := strings.TrimSpace(input.ValuationBasis)
	if basis == "" {
		basis = BasisUnspecified
	}
	switch basis {
	case BasisPreMoney, BasisPostMoney, BasisUnspecified:
	default:
		return FundingRound{}, ErrInvalidInput
	}
	if input.ValuationMinor == nil {
		basis = BasisUnspecified
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if !currencyPattern.MatchString(currency) {
		return FundingRound{}, ErrInvalidInput
	}
	evidence := strings.TrimSpace(input.EvidenceURL)
	if evidence == "" || validateWebsite(evidence) != nil {
		return FundingRound{}, ErrInvalidInput
	}
	source := strings.TrimSpace(input.SourceLabel)
	if source == "" || len(source) > maxSourceLabel {
		return FundingRound{}, ErrInvalidInput
	}
	note := strings.TrimSpace(input.Note)
	if len(note) > maxFundingNote {
		return FundingRound{}, ErrInvalidInput
	}
	amount := cloneOptionalAmount(input.AmountRaisedMinor)
	valuation := cloneOptionalAmount(input.ValuationMinor)
	return FundingRound{
		RoundLabel:        label,
		AnnouncementDate:  announcement,
		AmountRaisedMinor: amount,
		ValuationMinor:    valuation,
		ValuationBasis:    basis,
		Currency:          currency,
		EvidenceURL:       evidence,
		SourceLabel:       source,
		Note:              note,
	}, nil
}

func cloneOptionalAmount(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
