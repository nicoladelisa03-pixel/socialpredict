package startups

import (
	"context"
	"errors"
	"strings"

	dmarkets "socialpredict/internal/domain/markets"
	dstartups "socialpredict/internal/domain/startups"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository stores startups with GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository returns a Postgres or sqlite-backed startup repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create inserts a startup and copies the generated id back onto the domain value.
func (r *GormRepository) Create(ctx context.Context, startup *dstartups.Startup) error {
	if r == nil || r.db == nil || startup == nil {
		return dstartups.ErrInvalidInput
	}
	row := toModel(startup)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isDuplicate(err) {
			return dstartups.ErrSlugTaken
		}
		return err
	}
	*startup = fromModel(row)
	return nil
}

// Update saves editable fields for an existing startup.
func (r *GormRepository) Update(ctx context.Context, startup *dstartups.Startup) error {
	if r == nil || r.db == nil || startup == nil || startup.ID <= 0 {
		return dstartups.ErrInvalidInput
	}
	result := r.db.WithContext(ctx).Model(&models.Startup{}).Where("id = ?", startup.ID).Updates(map[string]any{
		"name":        startup.Name,
		"description": startup.Description,
		"website_url": startup.WebsiteURL,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return dstartups.ErrNotFound
	}
	var row models.Startup
	if err := r.db.WithContext(ctx).First(&row, startup.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dstartups.ErrNotFound
		}
		return err
	}
	*startup = fromModel(row)
	return nil
}

// GetBySlug loads one startup by its unique slug.
func (r *GormRepository) GetBySlug(ctx context.Context, slug string) (*dstartups.Startup, error) {
	if r == nil || r.db == nil {
		return nil, dstartups.ErrInvalidInput
	}
	var row models.Startup
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, dstartups.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	startup := fromModel(row)
	return &startup, nil
}

// List returns startups ordered by name, then slug.
func (r *GormRepository) List(ctx context.Context) ([]dstartups.Startup, error) {
	if r == nil || r.db == nil {
		return nil, dstartups.ErrInvalidInput
	}
	var rows []models.Startup
	if err := r.db.WithContext(ctx).Order("name asc, slug asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	startups := make([]dstartups.Startup, 0, len(rows))
	for _, row := range rows {
		startups = append(startups, fromModel(row))
	}
	return startups, nil
}

func toModel(startup *dstartups.Startup) models.Startup {
	return models.Startup{
		ID:          startup.ID,
		Name:        startup.Name,
		Slug:        startup.Slug,
		Description: startup.Description,
		WebsiteURL:  startup.WebsiteURL,
	}
}

func fromModel(row models.Startup) dstartups.Startup {
	return dstartups.Startup{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		WebsiteURL:  row.WebsiteURL,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// SetMarketStartup stores or clears a market's startup id without changing trading fields.
func (r *GormRepository) SetMarketStartup(ctx context.Context, marketID int64, startupID *int64) error {
	if r == nil || r.db == nil || marketID <= 0 {
		return dstartups.ErrInvalidInput
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Market{}).Where("id = ?", marketID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return dstartups.ErrMarketNotFound
	}
	query := r.db.WithContext(ctx).Model(&models.Market{}).Where("id = ?", marketID)
	var result *gorm.DB
	if startupID == nil {
		result = query.Update("startup_id", gorm.Expr("NULL"))
	} else {
		result = query.Update("startup_id", *startupID)
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ListLinkedMarkets returns markets that are public and attached to the startup.
func (r *GormRepository) ListLinkedMarkets(ctx context.Context, startupID int64) ([]dstartups.LinkedMarket, error) {
	if r == nil || r.db == nil || startupID <= 0 {
		return nil, dstartups.ErrInvalidInput
	}
	var rows []models.Market
	err := publicMarketScope(r.db.WithContext(ctx).Model(&models.Market{})).
		Where("startup_id = ?", startupID).
		Order("question_title asc, id asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	markets := make([]dstartups.LinkedMarket, 0, len(rows))
	for _, row := range rows {
		markets = append(markets, linkedMarketFromModel(row))
	}
	return markets, nil
}

type marketAssociationRow struct {
	ID              int64  `gorm:"column:id"`
	QuestionTitle   string `gorm:"column:question_title"`
	LifecycleStatus string `gorm:"column:lifecycle_status"`
	StartupSlug     string `gorm:"column:startup_slug"`
}

// ListMarketAssociations returns every market, including ones that are not public,
// with the slug of its startup when one is set.
func (r *GormRepository) ListMarketAssociations(ctx context.Context) ([]dstartups.MarketAssociation, error) {
	if r == nil || r.db == nil {
		return nil, dstartups.ErrInvalidInput
	}
	var rows []marketAssociationRow
	err := r.db.WithContext(ctx).Model(&models.Market{}).
		Select("markets.id, markets.question_title, markets.lifecycle_status, COALESCE(startups.slug, '') AS startup_slug").
		Joins("LEFT JOIN startups ON startups.id = markets.startup_id AND startups.deleted_at IS NULL").
		Order("markets.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	associations := make([]dstartups.MarketAssociation, 0, len(rows))
	for _, row := range rows {
		associations = append(associations, dstartups.MarketAssociation{
			MarketID:        row.ID,
			QuestionTitle:   row.QuestionTitle,
			LifecycleStatus: row.LifecycleStatus,
			StartupSlug:     row.StartupSlug,
		})
	}
	return associations, nil
}

func publicMarketScope(query *gorm.DB) *gorm.DB {
	return query.Where(
		"lifecycle_status = ? OR lifecycle_status = ? OR lifecycle_status = '' OR lifecycle_status IS NULL",
		dmarkets.MarketLifecyclePublished,
		dmarkets.MarketLifecycleResolved,
	)
}

func linkedMarketFromModel(row models.Market) dstartups.LinkedMarket {
	steward := strings.TrimSpace(row.StewardUsername)
	if steward == "" {
		steward = strings.TrimSpace(row.CreatorUsername)
	}
	resolved := row.IsResolved || strings.EqualFold(strings.TrimSpace(row.LifecycleStatus), dmarkets.MarketLifecycleResolved)
	return dstartups.LinkedMarket{
		ID:                      row.ID,
		QuestionTitle:           row.QuestionTitle,
		Description:             row.Description,
		OutcomeType:             row.OutcomeType,
		ResolutionDateTime:      row.ResolutionDateTime,
		FinalResolutionDateTime: row.FinalResolutionDateTime,
		UTCOffset:               row.UTCOffset,
		IsResolved:              resolved,
		ResolutionResult:        row.ResolutionResult,
		InitialProbability:      row.InitialProbability,
		CreatorUsername:         row.CreatorUsername,
		StewardUsername:         steward,
		CreatedAt:               row.CreatedAt,
		YesLabel:                row.YesLabel,
		NoLabel:                 row.NoLabel,
	}
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique") || strings.Contains(message, "duplicate")
}
