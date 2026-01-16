package main

import (
	"encoding/json"
	"net/http"

	"notesapp-backend/internal/utils"
)

// ------------------ Health Check --------------------------------------------


func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(
		w, 
		http.StatusOK, 
		map[string]string{"status": "ok"}, 
		nil
	)
}


// ------------------ Auth Handlers --------------------------------------------


// Data types to extract from json
type registerRequest struct { 
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type googleLoginRequest struct {
	Token string `json:"token"`
}
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Auth : Register
func (app *application) register(w http.ResponseWriter, r *http.Request) {

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Decode Json into req
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// Do the register
	user, err := app.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add response
	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id": user.ID.Hex(),
		"name":    user.Name,
		"email":   user.Email,
	}, nil)
}

// Auth : Login
func (app *application) login(w http.ResponseWriter, r *http.Request) {

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// Get the tokens
	tokens, err := app.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.ErrorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, nil)
}

// Auth : googleLogin
func (app *application) googleLogin(w http.ResponseWriter, r *http.Request) {
	var req googleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	tokens, err := app.authService.GoogleLogin(r.Context(), req.Token)
	if err != nil {
		utils.ErrorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, nil)
}

// Auth: Refresh Token
func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	tokens, err := app.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		utils.ErrorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, nil)
}

// Auth : logout
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	err := app.authService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"}, nil)
}


// ------------------ Notes Handlers --------------------------------------------


// Data types to extract from json
type createNoteRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Note : Create
func (app *application) createNote(w http.ResponseWriter, r *http.Request) {
	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	userID := r.Context().Value(contextKeyUserID).(string)

	note, err := app.noteService.Create(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, note, nil)
}

// Note : GetAll
func (app *application) listNotes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKeyUserID).(string)

	notes, err := app.noteService.List(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, notes, nil)
}

// Note : GetOne
func (app *application) getNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	userID := r.Context().Value(contextKeyUserID).(string)

	note, err := app.noteService.Get(r.Context(), userID, noteID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, note, nil)
}

// Note : Update
func (app *application) updateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	userID := r.Context().Value(contextKeyUserID).(string)

	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	note, err := app.noteService.Update(r.Context(), userID, noteID, req.Title, req.Description)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, note, nil)
}

// Note : Delete
func (app *application) deleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID := vars["id"]
	userID := r.Context().Value(contextKeyUserID).(string)

	err := app.noteService.Delete(r.Context(), userID, noteID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "note deleted"}, nil)
}
