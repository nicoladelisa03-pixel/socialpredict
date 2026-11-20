package startups

import (
	"context"
	"strings"
	"time"
)

// LinkedMarket is a public YES/NO market shown on a startup profile.
type LinkedMarket struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	UTCOffset               int
	IsResolved              bool
	ResolutionResult        string
	InitialProbability      float64
	CreatorUsername         string
	StewardUsername         string
	CreatedAt               time.Time
	YesLabel                string
	NoLabel                 string
}

// MarketStartupLink is the association stored for one market.
type MarketStartupLink struct {
	MarketID    int64
	StartupSlug string
}

// MarketAssociation is an admin view of one market and its current startup.
// It includes markets that are not public.
type MarketAssociation struct {
	MarketID        int64
	QuestionTitle   string
	LifecycleStatus string
	StartupSlug     string
}

// SetMarketStartup attaches, changes, or clears a market's startup.
// An empty slug clears the association. Trading fields are left unchanged.
func (s *Service) SetMarketStartup(ctx context.Context, marketID int64, startupSlug string) (MarketStartupLink, error) {
	if s == nil || s.repo == nil || marketID <= 0 {
		return MarketStartupLink{}, ErrInvalidInput
	}
	if strings.TrimSpace(startupSlug) == "" {
		if err := s.repo.SetMarketStartup(ctx, marketID, nil); err != nil {
			return MarketStartupLink{}, err
		}
		return MarketStartupLink{MarketID: marketID}, nil
	}
	slug := NormalizeSlug(startupSlug)
	if !validSlug(slug) {
		return MarketStartupLink{}, ErrInvalidInput
	}
	startup, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return MarketStartupLink{}, err
	}
	if err := s.repo.SetMarketStartup(ctx, marketID, &startup.ID); err != nil {
		return MarketStartupLink{}, err
	}
	return MarketStartupLink{MarketID: marketID, StartupSlug: startup.Slug}, nil
}

// ListLinkedMarkets returns the public markets attached to a startup.
func (s *Service) ListLinkedMarkets(ctx context.Context, slug string) ([]LinkedMarket, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	normalized := NormalizeSlug(slug)
	if !validSlug(normalized) {
		return nil, ErrNotFound
	}
	startup, err := s.repo.GetBySlug(ctx, normalized)
	if err != nil {
		return nil, err
	}
	markets, err := s.repo.ListLinkedMarkets(ctx, startup.ID)
	if err != nil {
		return nil, err
	}
	if markets == nil {
		markets = []LinkedMarket{}
	}
	return markets, nil
}

// ListMarketAssociations returns every market with its current startup slug.
// This is an admin read. Public startup pages keep using ListLinkedMarkets.
func (s *Service) ListMarketAssociations(ctx context.Context) ([]MarketAssociation, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	rows, err := s.repo.ListMarketAssociations(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []MarketAssociation{}, nil
	}
	return rows, nil
}
