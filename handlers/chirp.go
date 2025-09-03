package handlers

import (
	"Chirpy/helpers"
	"Chirpy/internal/auth"
	"Chirpy/internal/config"
	"Chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/google/uuid"
)

type ChirpHandler struct {
	*config.ApiConfig
}

func (chirpHanlder *ChirpHandler) HandlerCreateChirp(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Cache-Control", "no-cache")
	chirp := struct {
		Body string `json:"body"`
	}{}
	defer req.Body.Close()
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		helpers.RespondWithError(respWriter, 401, "Must login to create chirp.")
		return
	}
	userId, err := auth.ValidateJWT(token, chirpHanlder.JWTSecret)
	if err != nil {
		helpers.RespondWithError(respWriter, 401, "Invalid auth token")
		return
	}
	err = json.NewDecoder(req.Body).Decode(&chirp)
	if err != nil {
		chirpHanlder.Logger.Printf("Error decoding the json: %v", err)
		helpers.RespondWithError(respWriter, 400, "Something went wrong.")
		return
	}
	if len(chirp.Body) == 0 {
		helpers.RespondWithError(respWriter, 400, "Chirp cannot be empty.")
		return
	}
	if len(chirp.Body) > 140 {
		helpers.RespondWithError(respWriter, 400, "Chirp is too long.")
		return
	}
	user, err := chirpHanlder.DB.GetUser(req.Context(), userId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 400, "User not found for given user_id.")
			return
		}
		chirpHanlder.Logger.Printf("Error getting user from db: %v", err)
		helpers.RespondWithError(respWriter, 500, "500 Internal Server Error. Unable to create chirp.")
		return
	}
	if user.ID == uuid.Nil {
		helpers.RespondWithError(respWriter, 400, "Invalid user_id. User doesn't exist for given user_id.")
		return
	}
	wordsToBeReplaced := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedChirpBody := helpers.CleanString(chirp.Body, wordsToBeReplaced, "****")
	newChirpParams := database.CreateChirpParams{
		Body:   cleanedChirpBody,
		UserID: user.ID,
	}
	insertedChirp, err := chirpHanlder.DB.CreateChirp(req.Context(), database.CreateChirpParams(newChirpParams))
	if err != nil {
		chirpHanlder.Logger.Printf("Error creating the chirp: %v", err)
		helpers.RespondWithError(respWriter, 500, "500 Internal Server Error. Unable to create chirp.")
		return
	}
	helpers.RespondWithJson(respWriter, 201, insertedChirp)
}

func (chirpHanlder *ChirpHandler) HandlerGetAllCirps(respWriter http.ResponseWriter, req *http.Request) {
	queryAuthorId := req.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if queryAuthorId != "" {
		authorId, err := uuid.Parse(queryAuthorId)
		if err != nil {
			helpers.RespondWithError(respWriter, 400, "Invalid author_id")
			return
		}
		chirps, err = chirpHanlder.DB.GetChirpsByAuthor(req.Context(), authorId)
	} else {
		chirps, err = chirpHanlder.DB.GetAllChirps(req.Context())
	}
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 404, "No Chirps found.")
			return
		}
		chirpHanlder.Logger.Printf("Error getting chirps from db: %v", err)
		helpers.RespondWithError(respWriter, 500, "500 Internal Server Error. Unable to get chirps.")
		return
	}
	if len(chirps) == 0 {
		helpers.RespondWithError(respWriter, 404, "No Chirps found.")
		return
	}
	sorted := req.URL.Query().Get("sort")
	if sorted == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}
	helpers.RespondWithJson(respWriter, 200, chirps)
}

func (chirpHanlder *ChirpHandler) HandlerGetOneCirps(respWriter http.ResponseWriter, req *http.Request) {
	pathValue := req.PathValue("chirpID")
	if pathValue == "" {
		helpers.RespondWithError(respWriter, 400, "chirpID is must be provided as a path value.")
	}
	chirpId, err := uuid.Parse(pathValue)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid chirpID.")
		return
	}
	chirp, err := chirpHanlder.DB.GetOneChirp(req.Context(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 404, "No Chirp found for given chirpId")
			return
		}
		chirpHanlder.Logger.Printf("Error getting chirp from db: %v", err)
		helpers.RespondWithError(respWriter, 500, "500 Internal Server Error. Unable to get chirp.")
		return
	}
	helpers.RespondWithJson(respWriter, 200, chirp)
}

func (chirpHanlder *ChirpHandler) HandlerDeleteCirp(respWriter http.ResponseWriter, req *http.Request) {
	id := req.PathValue("chirpID")
	chirpId, err := uuid.Parse(id)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid Chirp ID.")
		return
	}
	deletedChirp, err := chirpHanlder.DB.DeleteChirp(req.Context(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 404, "No chirp found for the given chirpID.")
			return
		}
		chirpHanlder.Logger.Printf("Error trying to delete chirp: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal server error.")
		return
	}
	if deletedChirp.ID == uuid.Nil {
		helpers.RespondWithError(respWriter, 404, "No chirp found for the given chirpID.")
		return
	}
	helpers.RespondWithJson(respWriter, 204, "")
}
