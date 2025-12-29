package repositories

import (
	"errors"

	apperrors "agent/internal/errors"
	"agent/internal/models"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) FindByEntity(entityType string, entityID int64, tenantID *int64) ([]models.Note, error) {
	var notes []models.Note
	query := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Order("created_at DESC").Find(&notes).Error
	return notes, err
}

func (r *NoteRepository) FindByID(id int64) (*models.Note, error) {
	var note models.Note
	err := r.db.First(&note, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &note, nil
}

type NoteCreateInput struct {
	TenantID   int64
	EntityType string
	EntityID   int64
	Content    string
	UserID     int64
}

func (r *NoteRepository) Create(input NoteCreateInput) (int64, error) {
	note := &models.Note{
		TenantID:   input.TenantID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Content:    input.Content,
		CreatedBy:  input.UserID,
		UpdatedBy:  input.UserID,
	}
	err := r.db.Create(note).Error
	return note.ID, err
}

func (r *NoteRepository) Update(id int64, updates map[string]interface{}) error {
	return r.db.Model(&models.Note{}).Where("id = ?", id).Updates(updates).Error
}

func (r *NoteRepository) Delete(id int64) error {
	return r.db.Delete(&models.Note{}, id).Error
}
