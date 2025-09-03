package auth

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthGenerateAndCompare(t *testing.T) {
	hash, err := HashPassword("This is just a string....")
	if err != nil {
		t.Errorf("Unable to hash password: %v", err)
		t.FailNow()
	}
	err = CheckPasswordHash("This is just a string....", hash)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

var TokenSecret string = "TestSecret123"
var TokenValidityDuration time.Duration = time.Duration(5 * time.Second)

func TestCreateJWTToken(t *testing.T) {
	newUUID := uuid.New()
	token, err := MakeJWT(newUUID, TokenSecret, TokenValidityDuration)
	if err != nil {
		t.Errorf("Couldn't create jwt Token: %v", err)
		t.FailNow()
	}
	if token == "" {
		t.Errorf("Invalid Token returned by MakeJWT")
		t.FailNow()
	}
}
func TestValidateJWTToken(t *testing.T) {
	newUUID := uuid.New()
	token, err := MakeJWT(newUUID, TokenSecret, TokenValidityDuration)
	if err != nil {
		t.Errorf("Couldn't create jwt Token: %v", err)
		t.FailNow()
	}
	if token == "" {
		t.Errorf("Invalid Token returned by MakeJWT")
		t.FailNow()
	}
	userId, err := ValidateJWT(token, TokenSecret)
	if err != nil {
		t.Errorf("Couldn't validate the token just generated: %v", err)
		t.FailNow()
	}
	if userId != newUUID {
		t.Errorf("Invalid userId returned by ValidateJWT. Exprected: %v, Actual: %v", newUUID, userId)
		t.FailNow()
	}

}
func TestValidateJWTTokenDuration(t *testing.T) {

	newUUID := uuid.New()
	token, err := MakeJWT(newUUID, TokenSecret, TokenValidityDuration)
	if err != nil {
		t.Errorf("Couldn't create jwt Token: %v", err)
		t.FailNow()
	}
	if token == "" {
		t.Error("Invalid Token returned by MakeJWT")
		t.FailNow()
	}
	time.Sleep(TokenValidityDuration)
	_, err = ValidateJWT(token, TokenSecret)
	if err == nil {
		t.Error("Token still valid even after set timeout time has been passed.")
		t.FailNow()
	}
	if !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("Token was not expired Other error happened: Error : %v", err)
		t.FailNow()
	}

}

func TestGetBearerToken(t *testing.T) {
	newUUID := uuid.New()
	token, err := MakeJWT(newUUID, TokenSecret, TokenValidityDuration)
	if err != nil {
		t.Errorf("Couldn't create jwt Token: %v", err)
		t.FailNow()
	}
	if token == "" {
		t.Error("Invalid Token returned by MakeJWT")
		t.FailNow()
	}

	request, _ := http.NewRequest("GET", "/api/healthz", bytes.NewReader([]byte("")))
	request.Header.Add("Authorization", "Bearer "+token)
	bearerToken, err := GetBearerToken(request.Header)
	if err != nil {
		t.Errorf("Failed to parse auth token from http headers: %v", err)
		t.FailNow()
	}

	userId, err := ValidateJWT(bearerToken, TokenSecret)
	if err != nil {
		t.Errorf("Couldn't validate the token returned by GetBearerToken: %v", err)
		t.FailNow()
	}
	if userId != newUUID {
		t.Errorf("Invalid userId returned by ValidateJWT. Exprected: %v, Actual: %v", newUUID, userId)
		t.FailNow()
	}

}
