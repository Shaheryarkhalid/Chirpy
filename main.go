package main

import (
	"Chirpy/handlers"
	"Chirpy/internal/config"
	"Chirpy/internal/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func handlerHealth(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Content-Type", "text/plain")
	respWriter.Header().Set("Cache-Control", "no-cache")
	respWriter.WriteHeader(200)
	respWriter.Write([]byte("OK"))
}
func createLogger() (*log.Logger, *os.File) {
	file, err := os.OpenFile("./app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("Opening the log file failed: %w", err))
	}
	newLogger := log.New(file, "Chirpy Logger ", log.Ldate|log.Ltime|log.Llongfile)
	return newLogger, file
}

func openDbConnection(dbUrl string) (*database.Queries, *sql.DB) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("Error connecting with database: %v\n", err)
		os.Exit(1)
	}
	return database.New(db), db
}

func addHandlers(chirpyMux *http.ServeMux, apiCfg *config.ApiConfig) {
	metricsHandler := handlers.MetricsHandler{apiCfg}
	usersHandler := handlers.UsersHandler{apiCfg}
	chirpHanlder := handlers.ChirpHandler{apiCfg}

	chirpyMux.Handle("/app/", metricsHandler.MiddlewareMatricInc(http.StripPrefix("/app", http.FileServer(http.Dir("./static/")))))
	chirpyMux.Handle("/app/logo.png", metricsHandler.MiddlewareMatricInc(http.StripPrefix("/app", http.FileServer(http.Dir("./static//assets")))))

	chirpyMux.HandleFunc("POST /api/users", usersHandler.HandleCreateUser)
	chirpyMux.HandleFunc("PUT /api/users", usersHandler.HandlerUpdateUser)
	chirpyMux.HandleFunc("POST /api/login", usersHandler.HandlerLogin)
	chirpyMux.HandleFunc("POST /api/refresh", usersHandler.HandlerRefresh)
	chirpyMux.HandleFunc("POST /api/revoke", usersHandler.HandlerRevoke)

	chirpyMux.HandleFunc("POST /api/chirps", chirpHanlder.HandlerCreateChirp)
	chirpyMux.HandleFunc("GET /api/chirps", chirpHanlder.HandlerGetAllCirps)
	chirpyMux.HandleFunc("GET /api/chirps/{chirpID}", chirpHanlder.HandlerGetOneCirps)
	chirpyMux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpHanlder.HandlerDeleteCirp)

	chirpyMux.HandleFunc("POST /api/polka/webhooks", usersHandler.HandlerUpgradeUser)

	chirpyMux.HandleFunc("GET /api/healthz", handlerHealth)

	chirpyMux.HandleFunc("GET /admin/metrics", metricsHandler.HandlerMetrics)
	chirpyMux.HandleFunc("POST /admin/reset", metricsHandler.HandlerReset)
}

func getEnv() (string, string, string, string) {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWTSECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	return dbUrl, platform, jwtSecret, polkaKey
}

func main() {
	dbUrl, platform, jwtSecret, polkaKey := getEnv()
	dbQueries, db := openDbConnection(dbUrl)
	newLogger, file := createLogger()
	defer func() {
		db.Close()
		file.Close()
	}()
	port := "8080"
	chirpyMux := http.NewServeMux()
	apiCfg := config.ApiConfig{
		Logger:    newLogger,
		DB:        dbQueries,
		Platform:  platform,
		JWTSecret: jwtSecret,
		PolkaKey:  polkaKey,
	}
	addHandlers(chirpyMux, &apiCfg)
	server := http.Server{
		Handler: chirpyMux,
		Addr:    ":" + port,
	}
	apiCfg.Logger.Printf("Chirpy running on localhost:%v\n", port)
	apiCfg.Logger.Fatalf("Error while Running the Server: %v\n", server.ListenAndServe())
}
