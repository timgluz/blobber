# Blobber

WIP: This project is a work in progress and not yet ready for production use.

Blobber is a simple gateway service that allows users to upload and download small files (<= 1MB)
to and from various cloud storage providers. It provides a unified API for interacting
with different storage backends, making it easy to manage files across multiple platforms.

My current use case for Blobber is to serve as a file storage api for SpinCloud apps,
which has no direct access to cloud storage providers and usually relies only web requests.

## Features

- Upload and download files to/from multiple cloud storage providers
- Unified API for different storage backends
- Easy to extend with new storage providers
- Simple configuration using environment variables or config files
- Monitoring and logging support

## Supported Storage Providers

- [x] Amazon S3 and compatible services (e.g., MinIO, DigitalOcean Spaces, Cloudflare R2)
- [x] Google Cloud Storage
- [x] Microsoft Azure Blob Storage
- [x] Alicloud Object Storage Service (OSS)

## Getting Started

To get started with Blobber, follow these steps:

1. Clone the repository:

```bash
  git clone github.com/timgluz/blobber
  cd blobber
```

2. Install the dependencies:

```bash
  go mod tidy
```

3. Configure the storage providers by setting the appropriate environment variables.
   Refer to the documentation for each provider for the required configuration.

```bash
cp .env.example .env
vim .env

source .env
```

4. Configure and run the Blobber server:

```bash
  nano config.yaml

  go run main.go
```

## Testing

### E2E Tests

- Run the server in one terminal:

```bash
  task run
```

- Run the tests in another terminal:

```bash
  task test:e2e
```
