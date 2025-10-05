package blobstore

import (
	"context"
	"os"

	"gopkg.in/yaml.v2"
)

type S3Config struct {
	Endpoint string `yaml:"endpoint"`
	Region   string `yaml:"region"`
	Bucket   string `yaml:"bucket"`

	UsePathStyle bool `yaml:"use_path_style"`
}

type S3ConfigProvider interface {
	Retrieve(ctx context.Context) (*S3Config, error)
}

type YamlS3Config struct {
	path string
}

func NewYamlS3Config(path string) *YamlS3Config {
	return &YamlS3Config{path: path}
}

func (y *YamlS3Config) Retrieve(ctx context.Context) (*S3Config, error) {
	reader, err := os.ReadFile(y.path)
	if err != nil {
		return nil, err
	}

	var config S3Config
	if err := yaml.Unmarshal(reader, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
