package config

import (
	"os"

	"github.com/te0tl/go-clean-arch-base/core/pkg/domain/commons"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	errorsWrapper "github.com/pkg/errors"
)

// Base holds the fields every service needs regardless of what it does.
// It intentionally does NOT require Mongo, JWT, Email or a frontend URL —
// a stateless backend-only service must be able to boot with just these.
//
// Compose optional fragments (Mongo, JWT, Email, Frontend) into your
// project Config only when the service actually uses them:
//
//	type Config struct {
//	    config.Base
//	    config.JWT    // only if you issue tokens
//	    config.Mongo  // only if you persist
//	}
//
// caarlos0/env recurses into embedded structs, so a single LoadEnv(cfg)
// parses Base plus every embedded fragment.
type Base struct {
	Env                  commons.ENV       `env:"ENV,required"`
	Port                 int               `env:"PORT,required"`
	LogLevel             commons.LOG_LEVEL `env:"LOG_LEVEL,required"`
	GoogleCloudProjectID string            `env:"GOOGLE_CLOUD_PROJECT_ID"`
	GoogleCloudRevision  string            `env:"K_REVISION"`
}

// Mongo is an optional config fragment: embed it when the service persists
// to MongoDB (i.e. it imports the mongo module).
type Mongo struct {
	MongoURI string `env:"MONGO_URI,required"`
}

// JWT is an optional config fragment: embed it when the service issues or
// verifies JWTs.
type JWT struct {
	JwtSecretKey string `env:"JWT_SECRET_KEY,required"`
}

// Email is an optional config fragment: embed it when the service sends email.
type Email struct {
	EmailAPIKey    string `env:"EMAIL_API_KEY,required"`
	EmailFromEmail string `env:"EMAIL_FROM_EMAIL,required"`
	EmailFromName  string `env:"EMAIL_FROM_NAME,required"`
}

// Frontend is an optional config fragment: embed it when the service needs to
// build links back to a user-facing frontend (password reset, invites, etc.).
type Frontend struct {
	FrontendURL string `env:"FRONTEND_URL,required"`
}

func (c *Base) IsRunningInCloudRun() bool {
	return c.GoogleCloudRevision != ""
}

func (c *Base) Validate() {
	c.Env.Validate()
	c.LogLevel.Validate()
}

// LoadEnv loads .env (skipped when running in Cloud Run) and parses env vars
// into cfg via caarlos0/env. It panics on any error because a misconfigured
// app should fail to start.
func LoadEnv(cfg any) {
	// Load .env in local dev when it exists. A missing file is fine — env vars
	// can come straight from the environment (containers, CI, `VAR=… go run`).
	// Skipped entirely in Cloud Run (K_REVISION is set there).
	if os.Getenv("K_REVISION") == "" {
		if _, statErr := os.Stat(".env"); statErr == nil {
			if err := godotenv.Load(".env"); err != nil {
				panic(errorsWrapper.Wrap(err, "error loading .env file"))
			}
		}
	}

	if err := env.Parse(cfg); err != nil {
		panic(errorsWrapper.Wrap(err, "error parsing environment variables"))
	}
}
