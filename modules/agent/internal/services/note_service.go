package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteService struct {
	noteRepo *repositories.NoteRepository
}

func NewNoteService(noteRepo *repositories.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

func (s *NoteService) GetNotesByEntity(entityType string, entityID int64, tenantID *int64) ([]serializers.NoteResponse, error) {
	notes, err := s.noteRepo.FindByEntity(entityType, entityID, tenantID)
	if err != nil {
		return nil, err
	}
	return serializers.NotesToResponse(notes), nil
}

func (s *NoteService) GetNote(id int64) (*serializers.NoteResponse, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}
	resp := serializers.NoteToResponse(note)
	return &resp, nil
}

type NoteCreateInput struct {
	TenantID   int64
	EntityType string
	EntityID   int64
	Content    string
	UserID     int64
}

func (s *NoteService) CreateNote(input NoteCreateInput) (*serializers.NoteResponse, error) {
	repoInput := repositories.NoteCreateInput{
		TenantID:   input.TenantID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Content:    input.Content,
		UserID:     input.UserID,
	}
	id, err := s.noteRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}
	return s.GetNote(id)
}

func (s *NoteService) UpdateNote(id int64, content string, userID int64) (*serializers.NoteResponse, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}
	updates := map[string]interface{}{
		"content":    content,
		"updated_by": userID,
	}
	if err := s.noteRepo.Update(id, updates); err != nil {
		return nil, err
	}
	return s.GetNote(id)
}

func (s *NoteService) DeleteNote(id int64) error {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return err
	}
	if note == nil {
		return ErrNoteNotFound
	}
	return s.noteRepo.Delete(id)
}
