package config

import (
	"Chirpy/internal/database"
	"log"
	"sync/atomic"
)

type ApiConfig struct {
	Logger         *log.Logger
	DB             *database.Queries
	Platform       string
	JWTSecret      string
	PolkaKey       string
	FileServerHits atomic.Int32
}
