package main



type application struct{
	errorLog *log.Logger
	infoLog *log.Logger
	config *config.Config

	authService services.AuthService
	noteService services.NoteService
}


func main(){
	
	addr := flag.String("addr",":3000","HTTP network address")
	flag.Parse()


	// Loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)


	// Config
	cfg, err := config.Load()
	if err!=nil {
		errorLog.Fatal(err)
	}


	// MongoDB
	mongoClient, err := db.NewMongoClient(cfg.MongoURI)
	if err!=nil {
		errorLog.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background()) // close the connection at the end

	database := mongoClient.Database(cfg.MongoDBName)
	if err := db.EnsureIndexes(database); err != nil {
		errorLog.Fatal(err)
	}

	
	// Repositories
	userRepo := repositories.NewUserRepository(database)
	noteRepo := repositories.NewNoteRepository(database)
	tokenRepo := repositories.NewTokenRepository(database)


	// Services
	authService := services.NewAuthService(userRepo, tokenRepo, cfg)
	noteService := services.NewNoteService(noteRepo)


	// Application
	app := &application{
		errorLog:   errorLog,
		infoLog:    infoLog,
		config:     cfg,
		authService: authService,
		noteService: noteService,
	}


	// Create a server
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}


	// Start the server
	go func() {
		infoLog.Printf("starting server on %s", *addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errorLog.Fatal(err)
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