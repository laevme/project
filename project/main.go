package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func mustEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	ctx := context.Background()

	dsn := mustEnv("DATABASE_URL",
		"postgres://app:secret@localhost:5433/app?sslmode=disable")

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("cannot create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot reach database: %v", err)
	}

	h := &Handler{store: NewStore(pool)}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Users
	mux.HandleFunc("GET /users", h.listUsers)
	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("GET /users/{id}", h.getUser)
	mux.HandleFunc("PATCH /users/{id}", h.updateUser)
	mux.HandleFunc("DELETE /users/{id}", h.deleteUser)

	// Categories
	mux.HandleFunc("GET /categories", h.listCategories)
	mux.HandleFunc("POST /categories", h.createCategory)
	mux.HandleFunc("GET /categories/{id}", h.getCategory)
	mux.HandleFunc("PATCH /categories/{id}", h.updateCategory)
	mux.HandleFunc("DELETE /categories/{id}", h.deleteCategory)

	// Listings
	mux.HandleFunc("GET /listings", h.listListings)
	mux.HandleFunc("POST /listings", h.createListing)
	mux.HandleFunc("GET /listings/{id}", h.getListing)
	mux.HandleFunc("PATCH /listings/{id}", h.updateListing)
	mux.HandleFunc("DELETE /listings/{id}", h.deleteListing)

	// Messages
	mux.HandleFunc("GET /messages", h.listMessages)
	mux.HandleFunc("POST /messages", h.createMessage)
	mux.HandleFunc("GET /messages/{id}", h.getMessage)
	mux.HandleFunc("PATCH /messages/{id}", h.updateMessage)
	mux.HandleFunc("DELETE /messages/{id}", h.deleteMessage)

	// Favorites
	mux.HandleFunc("GET /favorites", h.listFavorites)
	mux.HandleFunc("POST /favorites", h.createFavorite)
	mux.HandleFunc("GET /favorites/{id}", h.getFavorite)
	mux.HandleFunc("PATCH /favorites/{id}", h.updateFavorite)
	mux.HandleFunc("DELETE /favorites/{id}", h.deleteFavorite)

	addr := ":" + mustEnv("PORT", "8080")
	log.Printf("listening on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      withLogging(withCORS(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
