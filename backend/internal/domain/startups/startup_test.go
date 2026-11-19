package startups

import (
	"context"
	"errors"
	"testing"
)

type memoryRepo struct {
	items map[string]Startup
	next  int64
}

func (m *memoryRepo) Create(_ context.Context, startup *Startup) error {
	if _, ok := m.items[startup.Slug]; ok {
		return ErrSlugTaken
	}
	m.next++
	startup.ID = m.next
	if m.items == nil {
		m.items = map[string]Startup{}
	}
	m.items[startup.Slug] = *startup
	return nil
}

func (m *memoryRepo) Update(_ context.Context, startup *Startup) error {
	if _, ok := m.items[startup.Slug]; !ok {
		return ErrNotFound
	}
	m.items[startup.Slug] = *startup
	return nil
}

func (m *memoryRepo) GetBySlug(_ context.Context, slug string) (*Startup, error) {
	startup, ok := m.items[slug]
	if !ok {
		return nil, ErrNotFound
	}
	return &startup, nil
}

func (m *memoryRepo) List(context.Context) ([]Startup, error) {
	out := make([]Startup, 0, len(m.items))
	for _, startup := range m.items {
		out = append(out, startup)
	}
	return out, nil
}

func (m *memoryRepo) SetMarketStartup(context.Context, int64, *int64) error {
	return nil
}

func (m *memoryRepo) ListLinkedMarkets(context.Context, int64) ([]LinkedMarket, error) {
	return []LinkedMarket{}, nil
}

func (m *memoryRepo) ListMarketAssociations(context.Context) ([]MarketAssociation, error) {
	return []MarketAssociation{}, nil
}

func TestCreateNormalizesAndRejectsInvalidFields(t *testing.T) {
	svc := NewService(&memoryRepo{})
	created, err := svc.Create(context.Background(), CreateRequest{
		Name:        "  Northwind  ",
		Slug:        " Northwind ",
		Description: " A test company ",
		WebsiteURL:  "https://northwind.example",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Slug != "northwind" || created.Name != "Northwind" || created.WebsiteURL != "https://northwind.example" {
		t.Fatalf("unexpected startup: %+v", created)
	}

	if _, err := svc.Create(context.Background(), CreateRequest{Name: "Again", Slug: "northwind"}); !errors.Is(err, ErrSlugTaken) {
		t.Fatalf("duplicate slug error = %v", err)
	}
	if _, err := svc.Create(context.Background(), CreateRequest{Name: "", Slug: "empty-name"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty name error = %v", err)
	}
	if _, err := svc.Create(context.Background(), CreateRequest{Name: "Bad", Slug: "Bad Slug"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid slug error = %v", err)
	}
	if _, err := svc.Create(context.Background(), CreateRequest{Name: "Bad URL", Slug: "bad-url", WebsiteURL: "javascript:alert(1)"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("bad website error = %v", err)
	}
}

func TestUpdateKeepsSlugAndEditsOtherFields(t *testing.T) {
	repo := &memoryRepo{}
	svc := NewService(repo)
	if _, err := svc.Create(context.Background(), CreateRequest{Name: "Northwind", Slug: "northwind"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.Update(context.Background(), "Northwind", UpdateRequest{
		Name:        "Northwind Labs",
		Description: "Updated",
		WebsiteURL:  "https://labs.example",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Slug != "northwind" || updated.Name != "Northwind Labs" || updated.Description != "Updated" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	if _, err := svc.Update(context.Background(), "northwind", UpdateRequest{Name: "Other", Slug: "other"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("slug change error = %v", err)
	}
	if _, err := svc.GetBySlug(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing slug error = %v", err)
	}
}
