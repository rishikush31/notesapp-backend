package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"notesapp-backend/internal/models"
	"notesapp-backend/internal/repositories"
)

var (
	ErrInvalidNoteInput = errors.New("invalid note input")
	ErrNoteNotFound     = errors.New("note not found")
)

type NoteService struct {
	noteRepo repositories.NoteRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func NewNoteService(
	noteRepo repositories.NoteRepository,
	infoLog *log.Logger,
	errorLog *log.Logger,
) *NoteService {
	return &NoteService{
		noteRepo: noteRepo,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

//
// -------------------- CREATE --------------------
//

func (s *NoteService) Create(
	ctx context.Context,
	userID string,
	title, description string,
) (*models.Note, error) {

	// 1. Validate user id
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	// 2. Validate input
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return nil, ErrInvalidNoteInput
	}

	// 3. Build model
	note := &models.Note{
		ID:          primitive.NewObjectID(),
		UserID:      uid,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// 4. Persist
	if err := s.noteRepo.Create(ctx, note); err != nil {
		s.errorLog.Printf("note.Create: %v", err)
		return nil, err
	}

	s.infoLog.Printf("note.Create: user=%s note=%s", uid.Hex(), note.ID.Hex())
	return note, nil
}

//
// -------------------- LIST --------------------
//

func (s *NoteService) List(
	ctx context.Context,
	userID string,
) ([]*models.Note, error) {

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	notes, err := s.noteRepo.List(ctx, uid)
	if err != nil {
		s.errorLog.Printf("note.List: %v", err)
		return nil, err
	}

	return notes, nil
}

//
// -------------------- GET --------------------
//

func (s *NoteService) Get(
	ctx context.Context,
	userID, noteID string,
) (*models.Note, error) {

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	nid, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	note, err := s.noteRepo.Get(ctx, uid, nid)
	if err != nil {
		if err == repositories.ErrNoteNotFound {
			return nil, ErrNoteNotFound
		}
		s.errorLog.Printf("note.Get: %v", err)
		return nil, err
	}

	return note, nil
}

//
// -------------------- UPDATE --------------------
//

func (s *NoteService) Update(
	ctx context.Context,
	userID, noteID string,
	title, description string,
) (*models.Note, error) {

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	nid, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return nil, ErrInvalidNoteInput
	}

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return nil, ErrInvalidNoteInput
	}

	note := &models.Note{
		ID:          nid,
		UserID:      uid,
		Title:       title,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.noteRepo.Update(ctx, note); err != nil {
		if err == repositories.ErrNoteNotFound {
			return nil, ErrNoteNotFound
		}
		s.errorLog.Printf("note.Update: %v", err)
		return nil, err
	}

	s.infoLog.Printf("note.Update: user=%s note=%s", uid.Hex(), nid.Hex())

	return s.noteRepo.Get(ctx, uid, nid)
}

//
// -------------------- DELETE --------------------
//

func (s *NoteService) Delete(
	ctx context.Context,
	userID, noteID string,
) error {

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrInvalidNoteInput
	}

	nid, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return ErrInvalidNoteInput
	}

	if err := s.noteRepo.Delete(ctx, uid, nid); err != nil {
		if err == repositories.ErrNoteNotFound {
			return ErrNoteNotFound
		}
		s.errorLog.Printf("note.Delete: %v", err)
		return err
	}

	s.infoLog.Printf("note.Delete: user=%s note=%s", uid.Hex(), nid.Hex())
	return nil
}
