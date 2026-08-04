
package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// ================================================================================
// ENGINE LOGIC CORE STRUCTS
// ================================================================================

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Plan         string    `json:"plan"`
	CreatedAt    time.Time `json:"created_at"`
}

type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimiter struct {
	ips sync.Map
	rate   float64
	cap    float64
}

type AppEnv struct {
	DB      *sql.DB
	Limiter *RateLimiter
	JWTKey  []byte
}

// ================================================================================
// SYSTEM ENGINE CORNERSTONE INITIALIZATION
// ================================================================================

func NewRateLimiter(rate float64, cap float64) *RateLimiter {
	return &RateLimiter{rate: rate, cap: cap}
}

func (rl *RateLimiter) Allow(ip string) bool {
	v, _ := rl.ips.LoadOrStore(ip, &TokenBucket{tokens: rl.cap, lastRefill: time.Now()})
	bucket := v.(*TokenBucket)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.lastRefill = now

	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > rl.cap {
		bucket.tokens = rl.cap
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

func InitDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Optimize engine hardware constraints via WAL mode configuration
	pragmaQuery := `
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA busy_timeout = 5000;
		PRAGMA foreign_keys = ON;
	`
	if _, err := db.Exec(pragmaQuery); err != nil {
		return nil, err
	}

	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			plan TEXT NOT NULL DEFAULT 'free',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}

// ================================================================================
// CRYPTOGRAPHIC DATA TRANSFORMS & MIDDLEWARE
// ================================================================================

func generateSecureID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("usr_%x", b)
}

func (env *AppEnv) SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = strings.Split(xff, ",")[0]
		}

		if !env.Limiter.Allow(ip) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"Rate threshold reached. Backoff immediately."}`))
			return
		}
		next(w, r)
	}
}

// ================================================================================
// ENDPOINT CONTROLLERS
// ================================================================================

func (env *AppEnv) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"Invalid payload parameter matrix"}`))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := User{
		ID:           generateSecureID(),
		Email:        strings.ToLower(req.Email),
		PasswordHash: string(hash),
		Plan:         "free",
		CreatedAt:    time.Now(),
	}

	_, err = env.DB.Exec(
		"INSERT INTO users (id, email, password_hash, plan, created_at) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Email, user.PasswordHash, user.Plan, user.CreatedAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"Identity conflict token duplicate"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func (env *AppEnv) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				CustomerEmail string `json:"customer_email"`
				AmountPaid    int64  `json:"amount_paid"`
			} `json:"object"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Type == "invoice.payment_succeeded" {
		_, err := env.DB.Exec(
			"UPDATE users SET plan = 'premium' WHERE email = ?",
			strings.ToLower(event.Data.Object.CustomerEmail),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"received":true}`))
}

// ================================================================================
// SERVER EXECUTION MATRIX ENTRYPOINT
// ================================================================================

func main() {
	log.Println("[SYS] Ignition matrix tracking active. Launching Micro-SaaS Core...")

	db, err := InitDatabase("./saas_production.db")
	if err != nil {
		log.Fatalf("Core database failed to bind: %v", err)
	}
	defer db.Close()

	env := &AppEnv{
		DB:      db,
		Limiter: NewRateLimiter(5.0, 10.0), // Allows 5 requests/sec with burst to 10
		JWTKey:  []byte("SYSTEM_SECRET_ROUTING_KEY_REPLACE_THIS_PRODUCTION"),
	}

	http.HandleFunc("/api/v1/auth/register", env.SecurityMiddleware(env.HandleRegister))
	http.HandleFunc("/api/v1/webhooks/stripe", env.SecurityMiddleware(env.HandleStripeWebhook))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER] Live runtime accepting network bindings on interface port: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Socket collision aborting service: %v", err)
	}
}
