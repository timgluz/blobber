package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/timgluz/blobber/blob"
	"github.com/timgluz/blobber/health"
	"github.com/timgluz/blobber/home"
	"github.com/timgluz/blobber/pkg/blobstore"
	"github.com/timgluz/blobber/pkg/cors"
	"github.com/timgluz/blobber/pkg/secret"
	"gopkg.in/yaml.v2"
)

var appName = "blobber"
var appVersion = "0.0.1"

// TODO: move to separate config.go file, add validation
type appConfig struct {
	Port int `yaml:"port"`

	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`

	Store struct {
		Provider string                   `yaml:"provider"`
		S3       blobstore.S3Config       `yaml:"s3,omitempty"`
		GCP      blobstore.GCPConfig      `yaml:"gcp,omitempty"`
		Azure    blobstore.AzureConfig    `yaml:"azure,omitempty"`
		Alicloud blobstore.AlicloudConfig `yaml:"alicloud,omitempty"`
	} `yaml:"store"`

	Auth struct {
		Provider       string `yaml:"provider"`
		APITokenEnvVar string `yaml:"api_token_env_var"`
	} `yaml:"auth"`
}

func main() {
	var configPath string
	var port int
	flag.StringVar(&configPath, "config", "configs/dev.yaml", "Path to configuration file")
	flag.IntVar(&port, "port", 8000, "Port to run the server on")
	flag.Parse()

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	if port != 0 {
		config.Port = port
	}
	logger := initAppLogger(config)

	ctx := context.Background()
	otelShutdown, err := setupOpenTelemetry(ctx)
	if err != nil {
		logger.Error("Failed to setup OpenTelemetry", slog.String("error", err.Error()))
		return
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			logger.Error("Error during OpenTelemetry shutdown", slog.String("error", err.Error()))
		}
		ctx.Done()
	}()

	logger.Debug("Initializing backend store", slog.String("provider", config.Store.Provider))
	store, err := initStore(config, logger)
	if err != nil {
		fmt.Println("Error initializing blob store:", err)
		return
	}

	authMiddleware, err := initAuthMiddleware(config, logger)
	if err != nil {
		fmt.Println("Error initializing auth middleware:", err)
		return
	}

	homeData := home.HandlerData{
		Title:        "Blobber - Blob Storage Service",
		Version:      "0.0.1",
		BlobProvider: config.Store.Provider,
	}

	homeHandler := home.NewHandler(homeData, logger)
	blobHandler := blob.NewHandler(store, logger)
	healthHandler := health.NewHandler(store, logger)

	// Public routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler.Handle)
	mux.HandleFunc("/healthz", healthHandler.Healthz)
	mux.HandleFunc("/readyz", healthHandler.Readyz)

	// Protected routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/blobs", blobHandler.HandleList)
	apiMux.HandleFunc("/blobs/{key}", blobHandler.Handle)

	// Wrap protected routes with auth middleware
	mux.Handle("/blobs", authMiddleware.Handler(apiMux))
	mux.Handle("/blobs/", authMiddleware.Handler(apiMux))

	// add static file server for /static/
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: cors.CORSMiddleware(otelhttp.NewHandler(mux, "/")),
	}

	logger.Info("Running server", slog.Int("port", config.Port))
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}

func loadConfig(path string) (appConfig, error) {
	var cfg appConfig

	content, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func initAppLogger(config appConfig) *slog.Logger {
	logLevel := logLevelFromString(config.Log.Level)

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func initAuthMiddleware(config appConfig, logger *slog.Logger) (*secret.APITokenMiddleware, error) {
	if config.Auth.Provider == "" {
		return nil, fmt.Errorf("auth store type is not configured")
	}

	var store secret.SecretStore
	switch config.Auth.Provider {
	case string(secret.AuthProviderEnv):
		store = secret.NewEnvSecretStore(config.Auth.APITokenEnvVar)
	default:
		logger.Warn("Unknown auth store type, defaulting to env", slog.String("store_type", config.Auth.Provider))
		return nil, fmt.Errorf("unknown auth store type: %s", config.Auth.Provider)
	}

	return secret.NewAPITokenMiddleware(store, logger), nil
}

func initStore(config appConfig, logger *slog.Logger) (blobstore.BlobStore, error) {
	switch strings.ToLower(config.Store.Provider) {
	case string(blobstore.BlobStoreTypeS3):
		return initS3Store(config, logger)
	case string(blobstore.BlobStoreTypeGCP):
		return initGCPStore(config, logger)
	case string(blobstore.BlobStoreTypeAzure):
		return initAzureStore(config, logger)
	case string(blobstore.BlobStoreTypeAlicloud):
		return initAlicloudStore(config, logger)
	default:
		return nil, fmt.Errorf("unsupported blob provider: %s", config.Store.Provider)
	}
}

func initS3Store(config appConfig, logger *slog.Logger) (blobstore.BlobStore, error) {
	credsProvider := blobstore.NewEnvS3Credentials()
	if _, err := credsProvider.Retrieve(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to retrieve S3 credentials, stopping initialization: %w", err)
	}

	s3Client, err := blobstore.NewS3Client(config.Store.S3, credsProvider, logger)
	if err != nil {
		fmt.Println("Error creating S3 client:", err)
		return nil, err
	}

	store, err := blobstore.NewS3BlobStore(config.Store.S3.Bucket, s3Client, logger)
	if err != nil {
		fmt.Println("Error creating S3 blob store:", err)
		return nil, err
	}

	return store, nil
}

func initGCPStore(config appConfig, logger *slog.Logger) (blobstore.BlobStore, error) {
	credsProvider := blobstore.NewJSONFileGCPCredentials(config.Store.GCP.CredentialsPath)
	if _, err := credsProvider.Retrieve(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to retrieve GCP credentials, stopping initialization: %w", err)
	}

	gcpClient, err := blobstore.NewGCPClient(config.Store.GCP, credsProvider, logger)
	if err != nil {
		fmt.Println("Error creating GCP client:", err)
		return nil, err
	}

	store, err := blobstore.NewGCPBlobStore(config.Store.GCP.Bucket, gcpClient, logger)
	if err != nil {
		fmt.Println("Error creating GCP blob store:", err)
		return nil, err
	}

	return store, nil
}

func initAzureStore(config appConfig, logger *slog.Logger) (blobstore.BlobStore, error) {
	credsProvider := blobstore.NewEnvAzureCredentials()
	if _, err := credsProvider.Retrieve(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to retrieve Azure credentials, stopping initialization: %w", err)
	}

	azureClient, err := blobstore.NewAzureClient(config.Store.Azure, credsProvider, logger)
	if err != nil {
		fmt.Println("Error creating Azure client:", err)
		return nil, err
	}

	store, err := blobstore.NewAzureBlobStore(config.Store.Azure.Container, azureClient, logger)
	if err != nil {
		fmt.Println("Error creating Azure blob store:", err)
		return nil, err
	}

	return store, nil
}

func initAlicloudStore(config appConfig, logger *slog.Logger) (blobstore.BlobStore, error) {
	credsProvider := blobstore.NewEnvAlicloudCredentials()
	if _, err := credsProvider.Retrieve(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to retrieve Alicloud credentials, stopping initialization: %w", err)

	}

	alicloudClient, err := blobstore.NewAlicloudClient(config.Store.Alicloud, credsProvider, logger)
	if err != nil {
		fmt.Println("Error creating Alicloud client:", err)
		return nil, err
	}

	store, err := blobstore.NewAlicloudBlobStore(config.Store.Alicloud, alicloudClient, logger)
	if err != nil {
		fmt.Println("Error creating Alicloud blob store:", err)
		return nil, err
	}

	return store, nil
}

func logLevelFromString(level string) slog.Level {
	normalizedLevel := strings.TrimSpace(strings.ToLower(level))

	switch normalizedLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
