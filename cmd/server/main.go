package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"

	"notesapp-backend/internal/config"
	"notesapp-backend/internal/db"
	"notesapp-backend/internal/repositories"
	"notesapp-backend/internal/services"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	config   *config.Config

	authService *services.AuthService
	noteService *services.NoteService
}

func main() {

	// Loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Config
	cfg, err := config.Load()
	if err != nil {
		errorLog.Fatal(err)
	}

	port := cfg.Port
	if port == "" {
		port = "3000"
	}

	addr := ":" + port

	// MongoDB
	mongoClient, err := db.NewMongoClient(cfg.MongoURI)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background()) // close the connection at the end

	database := mongoClient.Database(cfg.MongoDBName)
	if err := db.EnsureIndexes(database); err != nil {
		errorLog.Fatal(err)
	}

	// Repositories
	userRepo := repositories.NewUserRepository(database, infoLog, errorLog)
	noteRepo := repositories.NewNoteRepository(database, infoLog, errorLog)
	tokenRepo := repositories.NewTokenRepository(database, infoLog, errorLog)

	// Services
	authService := services.NewAuthService(userRepo, tokenRepo, cfg, infoLog, errorLog)
	noteService := services.NewNoteService(noteRepo, infoLog, errorLog)

	// Application
	app := &application{
		errorLog:    errorLog,
		infoLog:     infoLog,
		config:      cfg,
		authService: authService,
		noteService: noteService,
	}

	app.infoLog.Printf("Testing App infoLogger")
	app.errorLog.Printf("Testing App errorLogger")

	allowedOrigins := []string{"http://localhost:3001"}
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),                 // allow all origins
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // all common methods
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),           // allow common headers
	)(app.routes())

	// Create a server
	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server
	go func() {
		infoLog.Printf("starting server on %s", addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			srv.ErrorLog.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	infoLog.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // at the end

	if err := srv.Shutdown(ctx); err != nil {
		errorLog.Println(err)
	}
}
