package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"notesapp-backend/internal/models"
)

var ErrNoteNotFound = errors.New("note not found")

// -------------------- INTERFACE --------------------

type NoteRepository interface {
	Create(ctx context.Context, note *models.Note) error
	List(ctx context.Context, userID primitive.ObjectID) ([]*models.Note, error)
	Get(ctx context.Context, userID, noteID primitive.ObjectID) (*models.Note, error)
	Update(ctx context.Context, note *models.Note) error
	Delete(ctx context.Context, userID, noteID primitive.ObjectID) error
}

// -------------------- IMPLEMENTATION --------------------

type noteRepo struct {
	collection *mongo.Collection
	infoLog    *log.Logger
	errorLog   *log.Logger
}

func NewNoteRepository(
	db *mongo.Database,
	infoLog, errorLog *log.Logger,
) NoteRepository {
	return &noteRepo{
		collection: db.Collection("notes"),
		infoLog:    infoLog,
		errorLog:   errorLog,
	}
}

func (r *noteRepo) Create(ctx context.Context, note *models.Note) error {
	note.ID = primitive.NewObjectID()
	now := time.Now().UTC()
	note.CreatedAt = now
	note.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, note)
	if err != nil {
		r.errorLog.Printf("noteRepo.Create: %v", err)
		return err
	}
	return nil
}

func (r *noteRepo) List(ctx context.Context, userID primitive.ObjectID) ([]*models.Note, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		r.errorLog.Printf("noteRepo.List: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []*models.Note
	if err := cursor.All(ctx, &notes); err != nil {
		r.errorLog.Printf("noteRepo.List decode: %v", err)
		return nil, err
	}

	return notes, nil
}

func (r *noteRepo) Get(ctx context.Context, userID, noteID primitive.ObjectID) (*models.Note, error) {
	var note models.Note
	err := r.collection.FindOne(ctx, bson.M{
		"_id":    noteID,
		"userId": userID,
	}).Decode(&note)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoteNotFound
		}
		r.errorLog.Printf("noteRepo.Get: %v", err)
		return nil, err
	}

	return &note, nil
}

func (r *noteRepo) Update(ctx context.Context, note *models.Note) error {
	note.UpdatedAt = time.Now().UTC()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":    note.ID,
			"userId": note.UserID,
		},
		bson.M{
			"$set": bson.M{
				"title":       note.Title,
				"description": note.Description,
				"updatedAt":   note.UpdatedAt,
			},
		},
	)

	if err != nil {
		r.errorLog.Printf("noteRepo.Update: %v", err)
		return err
	}

	if result.MatchedCount == 0 {
		return ErrNoteNotFound
	}

	return nil
}

func (r *noteRepo) Delete(ctx context.Context, userID, noteID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"_id":    noteID,
		"userId": userID,
	})

	if err != nil {
		r.errorLog.Printf("noteRepo.Delete: %v", err)
		return err
	}

	if result.DeletedCount == 0 {
		return ErrNoteNotFound
	}

	return nil
}
