package handlers

import (
	"Chirpy/helpers"
	"Chirpy/internal/auth"
	"Chirpy/internal/config"
	"Chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UsersHandler struct {
	*config.ApiConfig
}

func (usersHandler *UsersHandler) HandleCreateUser(respWriter http.ResponseWriter, req *http.Request) {
	expectedBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	defer req.Body.Close()
	err := json.NewDecoder(req.Body).Decode(&expectedBody)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid request.")
	}
	if expectedBody.Email == "" {
		helpers.RespondWithError(respWriter, 400, "Email cannot be empty.")
		return
	}
	if expectedBody.Password == "" {
		helpers.RespondWithError(respWriter, 400, "Password cannot be empty.")
		return
	}
	hashedPassword, err := auth.HashPassword(expectedBody.Password)
	if err != nil {
		usersHandler.Logger.Printf("Error Happened while trying to hash password: User Input : %v : Error : %v", expectedBody, err)
		helpers.RespondWithError(respWriter, 500, "Unable to create user.")
		return
	}
	createdUser, err := usersHandler.DB.CreateUser(req.Context(), database.CreateUserParams{Email: expectedBody.Email, HashedPassword: hashedPassword})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			helpers.RespondWithError(respWriter, 409, "User already exists.")
			return
		}
		usersHandler.Logger.Printf("Error Happened while trying to create user: User Input : %v : Error : %v", expectedBody, err)
		helpers.RespondWithError(respWriter, 500, "Unable to create user.")
		return
	}
	newUser := struct {
		Id          uuid.UUID `json:"id"`
		Created_At  time.Time `json:"created_at"`
		Updated_At  time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		Id:          createdUser.ID,
		Created_At:  createdUser.CreatedAt,
		Updated_At:  createdUser.UpdatedAt,
		Email:       createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed,
	}
	helpers.RespondWithJson(respWriter, 201, newUser)
}
func (usersHandler *UsersHandler) HandlerLogin(respWriter http.ResponseWriter, req *http.Request) {
	loginBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	defer req.Body.Close()
	err := json.NewDecoder(req.Body).Decode(&loginBody)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, fmt.Sprintf("Error decoding the request: %v", err))
		return
	}
	if loginBody.Email == "" {
		helpers.RespondWithError(respWriter, 400, "Email cannot be empty.")
		return
	}
	if loginBody.Password == "" {
		helpers.RespondWithError(respWriter, 400, "Password cannot be empty.")
		return
	}
	user, err := usersHandler.DB.GetUserByEmail(req.Context(), loginBody.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 401, "Incorrect Email or Password.")
			return
		}
		usersHandler.ApiConfig.Logger.Printf("Error Trying to get user by Email: %v", err)
		helpers.RespondWithError(respWriter, 500, "500 Something went wrong.")
		return
	}
	err = auth.CheckPasswordHash(loginBody.Password, user.HashedPassword)
	if err != nil {
		helpers.RespondWithError(respWriter, 401, "Incorrect Email or Password.")
		return
	}
	tokenExpiry := time.Duration(1) * time.Hour
	token, err := auth.MakeJWT(user.ID, usersHandler.ApiConfig.JWTSecret, tokenExpiry)
	if err != nil {
		usersHandler.ApiConfig.Logger.Printf("Error trying to generate jwt token: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal server error.")
		return
	}
	refreshToken, err := generateRefreshTokenForUser(user.ID, usersHandler, req)
	if err != nil {
		helpers.RespondWithError(respWriter, 500, "Internal server error.")
		return
	}
	LogedInUser := struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	}
	helpers.RespondWithJson(respWriter, 200, LogedInUser)
}

func (usersHandler *UsersHandler) HandlerUpdateUser(respWriter http.ResponseWriter, req *http.Request) {
	authToken := req.Header.Get("Authorization")
	userId, err := usersHandler.validateAuthToken(authToken, &respWriter)
	if err != nil {
		return
	}
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	defer req.Body.Close()
	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, fmt.Sprintf("Error decoding the request: %v", err))
		return
	}
	if reqBody.Email == "" {
		helpers.RespondWithError(respWriter, 400, "Email cannot be empty. Must provide old or new email")
		return
	}
	if reqBody.Password == "" {
		helpers.RespondWithError(respWriter, 400, "Password cannot be empty. Must provide old or new password.")
		return
	}
	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		usersHandler.Logger.Printf("Error trying to hash passowrd: %v", err)
		helpers.RespondWithError(respWriter, 500, "Something went wrong. Please try again.")
		return
	}
	updateUserArgs := database.UpdateUserParams{
		ID:             userId,
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	}
	user, err := usersHandler.DB.UpdateUser(req.Context(), updateUserArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 400, "Invalid token. No user associated with this token.")
			return
		}
		usersHandler.Logger.Printf("Error trying to get user from db. UserId: %v Error : %v", userId, err)
		helpers.RespondWithError(respWriter, 500, "Internal server error")
		return
	}
	helpers.RespondWithJson(respWriter, 200, user)
}

func generateRefreshTokenForUser(userId uuid.UUID, usersHandler *UsersHandler, req *http.Request) (string, error) {
	token := auth.MakeRefreshToken()
	refreshToken := database.CreateRefreshTokenParams{
		Token:     token,
		UserID:    userId,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}
	rt, err := usersHandler.DB.CreateRefreshToken(req.Context(), refreshToken)
	if err != nil {
		usersHandler.Logger.Printf("Error tryring to generate refresh token: %v", err)
		return "", err
	}
	return rt.Token, nil
}

func (usersHandler *UsersHandler) HandlerRefresh(respWriter http.ResponseWriter, req *http.Request) {
	authToken := req.Header.Get("Authorization")
	refreshToken, err := usersHandler.validateRefreshToken(authToken, &respWriter, req)
	if err != nil {
		return
	}
	user, err := usersHandler.DB.GetUserFromRefreshToken(req.Context(), refreshToken.Token)
	if err != nil {
		usersHandler.Logger.Printf("Error trying to get user from refresh_token: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal Server Error.")
		return
	}
	if user.ID == uuid.Nil {
		helpers.RespondWithError(respWriter, 401, "No User found for the given token. Please try again.")
		return
	}
	newToken, err := auth.MakeJWT(user.ID, usersHandler.JWTSecret, time.Duration(1)*time.Hour)
	if err != nil {
		usersHandler.Logger.Printf("Error trying to create new jwt token: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal Server Error.")
		return
	}
	resp := struct {
		Token string `json:"token"`
	}{Token: newToken}
	respJson, err := json.Marshal(resp)
	if err != nil {
		usersHandler.Logger.Printf("Error trying to generate json from response object: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal Server Error.")
		return
	}
	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.WriteHeader(http.StatusOK)
	respWriter.Write(respJson)
}
func (usersHandler *UsersHandler) HandlerRevoke(respWriter http.ResponseWriter, req *http.Request) {
	authToken := req.Header.Get("Authorization")
	refreshToken, err := usersHandler.validateRefreshToken(authToken, &respWriter, req)
	if err != nil {
		return
	}
	revokeArgs := database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		Token:     refreshToken.Token,
	}
	err = usersHandler.DB.RevokeRefreshToken(req.Context(), revokeArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 400, "Invalid token.")
			return
		}
		usersHandler.Logger.Printf("DB Error trying to update revoked_at in refresh_token: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal server error.")
		return
	}
	respWriter.WriteHeader(http.StatusNoContent)
}

func (usersHandler *UsersHandler) HandlerUpgradeUser(respWriter http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid Api Key.")
		return
	}
	if apiKey != usersHandler.PolkaKey {
		helpers.RespondWithError(respWriter, 400, "Invalid Api Key.")
		return
	}
	reqBody := struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}{}
	defer req.Body.Close()
	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid request.")
		return
	}
	if reqBody.Event == "" || reqBody.Data.UserId == "" {
		helpers.RespondWithError(respWriter, 400, "Invalid request body.")
		return

	}
	if reqBody.Event != "user.upgraded" {
		helpers.RespondWithJson(respWriter, 204, "")
		return
	}
	userId, err := uuid.Parse(reqBody.Data.UserId)
	if err != nil {
		helpers.RespondWithError(respWriter, 400, "Invalid user_id.")
		return
	}

	upgradedUser, err := usersHandler.DB.UpgradeUserToRed(req.Context(), userId)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(respWriter, 404, "Invalid request. No user found for given user_id.")
			return
		}
		usersHandler.Logger.Printf("Error upgrading the user to chirpy_red: %v", err)
		helpers.RespondWithError(respWriter, 500, "Internal server error. Please try again.")
		return
	}
	if upgradedUser.ID == uuid.Nil {
		helpers.RespondWithError(respWriter, 404, "Invalid request. No user found for given user_id.")
		return
	}
	helpers.RespondWithJson(respWriter, 204, "")
}

func (usersHandler *UsersHandler) validateRefreshToken(authToken string, respWriter *http.ResponseWriter, req *http.Request) (refreshToken database.RefreshToken, err error) {
	if authToken == "" {
		helpers.RespondWithError(*respWriter, http.StatusBadRequest, "No Authorization header passed in request.")
		return refreshToken, fmt.Errorf("Invalid Token")
	}
	authToken = strings.TrimSpace(strings.ReplaceAll(authToken, "Bearer ", ""))
	if authToken == "" {
		helpers.RespondWithError(*respWriter, http.StatusBadRequest, "No Authorization header passed in request.")
		return refreshToken, fmt.Errorf("Invalid Token")
	}
	refreshToken, err = usersHandler.DB.GetRefreshToken(req.Context(), authToken)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.RespondWithError(*respWriter, 401, "Invalid Auth Token.")
			return refreshToken, fmt.Errorf("Invalid Token")
		}
		usersHandler.Logger.Printf("Error trying to get refresh token: %v", err)
		helpers.RespondWithError(*respWriter, http.StatusInternalServerError, "Internal Server Error.")
		return refreshToken, fmt.Errorf("Invalid Token")
	}
	if refreshToken.Token == "" {
		helpers.RespondWithError(*respWriter, 401, "Invalid Auth Token. No record found for given token.")
		return refreshToken, fmt.Errorf("Invalid Token")
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		helpers.RespondWithError(*respWriter, 401, "Session expired. Please login again.")
		return refreshToken, fmt.Errorf("Invalid Token")
	}
	return refreshToken, nil
}
func (usersHandler *UsersHandler) validateAuthToken(authToken string, respWriter *http.ResponseWriter) (userId uuid.UUID, err error) {
	if authToken == "" {
		helpers.RespondWithError(*respWriter, 401, "No Authorization header passed in request.")
		return userId, fmt.Errorf("Invalid Token")
	}
	authToken = strings.TrimSpace(strings.ReplaceAll(authToken, "Bearer ", ""))
	if authToken == "" {
		helpers.RespondWithError(*respWriter, 401, "No Authorization header passed in request.")
		return userId, fmt.Errorf("Invalid Token")
	}
	userId, err = auth.ValidateJWT(authToken, usersHandler.JWTSecret)
	if err != nil {
		helpers.RespondWithError(*respWriter, 401, "Invalid Auth Token")
		return
	}
	return userId, nil
}
