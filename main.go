package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/timgluz/blobber/blob"
	"github.com/timgluz/blobber/home"
	"github.com/timgluz/blobber/pkg/blobstore"
	"gopkg.in/yaml.v2"
)

type appConfig struct {
	LogLevel string `yaml:"log_level"`
	Port     int    `yaml:"port"`

	BlobProvider string             `yaml:"blob_provider"`
	S3Config     blobstore.S3Config `yaml:"s3_config"`
}

func main() {
	var configPath string
	var port int
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
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

	logger.Debug("Initializing backend store", slog.String("provider", config.BlobProvider))
	store, err := initStore(config, logger)
	if err != nil {
		fmt.Println("Error initializing blob store:", err)
		return
	}

	homeHandler := home.NewHandler(logger)
	blobHandler := blob.NewHandler(store, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler.Handle)
	mux.HandleFunc("/blobs", blobHandler.HandleList)
	mux.HandleFunc("/blobs/{key}", blobHandler.Handle)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
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
	logLevel := logLevelFromString(config.LogLevel)

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func initStore(config appConfig, logger *slog.Logger) (*blobstore.S3BlobStore, error) {
	credsProvider := blobstore.NewEnvS3Credentials()
	if _, err := credsProvider.Retrieve(nil); err != nil {
		return nil, fmt.Errorf("Failed to retrieve S3 credentials, stopping initialization: %w", err)
	}

	s3Client, err := blobstore.NewS3Client(config.S3Config, credsProvider, logger)
	if err != nil {
		fmt.Println("Error creating S3 client:", err)
		return nil, err
	}

	store, err := blobstore.NewS3BlobStore(config.S3Config.Bucket, s3Client, logger)
	if err != nil {
		fmt.Println("Error creating S3 blob store:", err)
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
