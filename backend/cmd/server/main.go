package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/solobueno/erp/internal/auth"
	"github.com/solobueno/erp/internal/shared/database"
	"github.com/solobueno/erp/pkg/jwt"
)

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	db := database.MustConnect(database.DefaultConfig())

	km := jwt.NewKeyManager()
	if err := km.LoadKeysFromEnv(); err != nil {
		log.Fatalf("failed to load JWT keys: %v", err)
	}
	if !km.HasPrivateKey() || !km.HasPublicKey() {
		log.Println("WARNING: JWT_PRIVATE_KEY/JWT_PUBLIC_KEY not set - generating an ephemeral RSA keypair for this run (dev only, tokens will not survive a restart)")
		generateEphemeralKeys(km)
	}

	authModule, err := auth.NewModule(auth.ModuleConfig{
		DB:         db,
		KeyManager: km,
		JWTConfig:  jwt.DefaultTokenGeneratorConfig(),
	})
	if err != nil {
		log.Fatalf("failed to initialize auth module: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	authModule.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Solobueno ERP Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}

// generateEphemeralKeys creates an in-memory RSA keypair for local development
// when JWT_PRIVATE_KEY/JWT_PUBLIC_KEY aren't configured in the environment.
func generateEphemeralKeys(km *jwt.KeyManager) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate ephemeral JWT keypair: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("failed to marshal ephemeral JWT public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	if err := km.LoadPrivateKeyFromPEM(privPEM); err != nil {
		log.Fatalf("failed to load ephemeral JWT private key: %v", err)
	}
	if err := km.LoadPublicKeyFromPEM(pubPEM); err != nil {
		log.Fatalf("failed to load ephemeral JWT public key: %v", err)
	}
}
