package startups

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	maxNameLength        = 160
	maxSlugLength        = 64
	maxDescriptionLength = 2000
	maxWebsiteLength     = 500
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var (
	ErrInvalidInput   = errors.New("invalid startup input")
	ErrNotFound       = errors.New("startup not found")
	ErrMarketNotFound = errors.New("market not found")
	ErrSlugTaken      = errors.New("startup slug already exists")
)

// Startup is the company record shown on a future profile page.
type Startup struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	WebsiteURL  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateRequest is the admin payload for a new startup.
type CreateRequest struct {
	Name        string
	Slug        string
	Description string
	WebsiteURL  string
}

// UpdateRequest edits fields other than the slug. A different slug is rejected.
type UpdateRequest struct {
	Name        string
	Slug        string
	Description string
	WebsiteURL  string
}

// Repository stores startup records.
type Repository interface {
	Create(ctx context.Context, startup *Startup) error
	Update(ctx context.Context, startup *Startup) error
	GetBySlug(ctx context.Context, slug string) (*Startup, error)
	List(ctx context.Context) ([]Startup, error)
	SetMarketStartup(ctx context.Context, marketID int64, startupID *int64) error
	ListLinkedMarkets(ctx context.Context, startupID int64) ([]LinkedMarket, error)
	ListMarketAssociations(ctx context.Context) ([]MarketAssociation, error)
}

// Service validates startup records before they are stored.
type Service struct {
	repo Repository
}

// NewService returns a startup service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns startups in name order.
func (s *Service) List(ctx context.Context) ([]Startup, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx)
}

// GetBySlug returns one startup. The slug is normalized before lookup.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Startup, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	normalized := NormalizeSlug(slug)
	if !validSlug(normalized) {
		return nil, ErrNotFound
	}
	return s.repo.GetBySlug(ctx, normalized)
}

// Create stores a new startup. The slug cannot be reused.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Startup, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	startup, err := normalizeCreate(req)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &startup); err != nil {
		return nil, err
	}
	return &startup, nil
}

// Update changes the name, description, and website. The stored slug stays the same.
func (s *Service) Update(ctx context.Context, slug string, req UpdateRequest) (*Startup, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidInput
	}
	currentSlug := NormalizeSlug(slug)
	if !validSlug(currentSlug) {
		return nil, ErrNotFound
	}
	if requested := NormalizeSlug(req.Slug); requested != "" && requested != currentSlug {
		return nil, ErrInvalidInput
	}
	current, err := s.repo.GetBySlug(ctx, currentSlug)
	if err != nil {
		return nil, err
	}
	name, description, website, err := normalizeEditable(req.Name, req.Description, req.WebsiteURL)
	if err != nil {
		return nil, err
	}
	current.Name = name
	current.Description = description
	current.WebsiteURL = website
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func normalizeCreate(req CreateRequest) (Startup, error) {
	slug := NormalizeSlug(req.Slug)
	if !validSlug(slug) {
		return Startup{}, ErrInvalidInput
	}
	name, description, website, err := normalizeEditable(req.Name, req.Description, req.WebsiteURL)
	if err != nil {
		return Startup{}, err
	}
	return Startup{
		Name:        name,
		Slug:        slug,
		Description: description,
		WebsiteURL:  website,
	}, nil
}

func normalizeEditable(name, description, website string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	website = strings.TrimSpace(website)
	if name == "" || len(name) > maxNameLength {
		return "", "", "", ErrInvalidInput
	}
	if len(description) > maxDescriptionLength {
		return "", "", "", ErrInvalidInput
	}
	if err := validateWebsite(website); err != nil {
		return "", "", "", err
	}
	return name, description, website, nil
}

func validateWebsite(website string) error {
	if website == "" {
		return nil
	}
	if len(website) > maxWebsiteLength {
		return ErrInvalidInput
	}
	parsed, err := url.ParseRequestURI(website)
	if err != nil || parsed.Host == "" {
		return ErrInvalidInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidInput
	}
	return nil
}

// NormalizeSlug lowercases a slug and removes surrounding hyphens.
func NormalizeSlug(slug string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(slug)), "-")
}

func validSlug(slug string) bool {
	return slug != "" && len(slug) <= maxSlugLength && slugPattern.MatchString(slug)
}
