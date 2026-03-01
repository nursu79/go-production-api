package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/config"
	deliveryHttp "github.com/nursu79/go-production-api/internal/delivery/http"
	"github.com/nursu79/go-production-api/internal/delivery/http/handler"
	"github.com/nursu79/go-production-api/internal/repository"
	"github.com/nursu79/go-production-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)


func prepareTestApp(t *testing.T) (*postgres.PostgresContainer, *httptest.Server) {
	ctx := context.Background()

	// Spin up PostgreSQL Container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Wait explicitly for PostgreSQL port mappings allowing network hook cycles.
	time.Sleep(1 * time.Second)

	// Run Migrations
	migrationPath := filepath.Join("..", "..", "..", "..", "migrations")
	err = repository.RunMigrations(connStr, "file://"+migrationPath)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Initialize pgxpool
	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	// Ping database validating live states
	err = dbPool.Ping(ctx)
	if err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	// Set up dependencies
	cfg := &config.Config{
		AppEnv:           "development",
		JwtSecret:        "test-secret",
		JwtRefreshSecret: "test-refresh",
		CorsOrigins:      []string{"*"},
	}

	userRepo := repository.NewUserRepository(dbPool)
	userUc := usecase.NewUserUsecase(userRepo, nil, cfg.JwtSecret, cfg.JwtRefreshSecret)
	userHandler := handler.NewUserHandler(userUc)
	adminHandler := handler.NewAdminHandler(userUc)

	router := deliveryHttp.NewRouter(dbPool, nil, userHandler, adminHandler, cfg)
	server := httptest.NewServer(router)

	return pgContainer, server
}

func TestE2E_UserRegistrationAndLoginFlow(t *testing.T) {
	pgContainer, server := prepareTestApp(t)
	defer pgContainer.Terminate(context.Background())
	defer server.Close()

	// 1. Register User
	registerBody := map[string]string{
		"email":    "john.doe@example.com",
		"password": "securepassword123",
	}
	body, _ := json.Marshal(registerBody)
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. Login User
	loginBody := map[string]string{
		"email":    "john.doe@example.com",
		"password": "securepassword123",
	}
	lBody, _ := json.Marshal(loginBody)
	lResp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(lBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, lResp.StatusCode)

	var loginData handler.LoginResponse
	json.NewDecoder(lResp.Body).Decode(&loginData)
	assert.NotEmpty(t, loginData.AccessToken)
	lResp.Body.Close()

	// 3. Fetch User Profile (GET /me)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginData.AccessToken)
	client := &http.Client{}
	meResp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, meResp.StatusCode)
	
	var meData handler.RegisterResponse
	json.NewDecoder(meResp.Body).Decode(&meData)
	assert.Equal(t, "john.doe@example.com", meData.Email)
	assert.Equal(t, "user", meData.Role)
	assert.NotEmpty(t, meData.ID)
	meResp.Body.Close()
}
