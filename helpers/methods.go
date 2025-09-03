package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func CleanString(s string, words []string, replacement string) string {
	wordsMap := map[string]bool{}
	for _, w := range words {
		wordsMap[w] = true
	}
	listOfWords := strings.Split(s, " ")
	for idx, w := range listOfWords {
		if wordsMap[strings.ToLower(w)] {
			listOfWords[idx] = replacement
		}
	}
	return strings.Join(listOfWords, " ")
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	errJson := struct {
		Error string `json:"error"`
	}{Error: msg}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsonStringBytes, _ := json.Marshal(errJson)
	w.Write(jsonStringBytes)

}
func RespondWithJson(w http.ResponseWriter, code int, payload any) {
	respJson := struct {
		Body any `json:"body"`
	}{Body: payload}
	respBytes, err := json.Marshal(respJson)
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Something went wrong: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(respBytes)
}
