// storage.go
package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"google.golang.org/api/option"
)

type StorageServiceInterface interface {
	ProcessFile(filename string) ([]map[string]interface{}, error)
}

type StorageService struct {
	client     *storage.Client
	bucketName string
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	log.Printf("Initializing storage service for project: %s", cfg.GCPCredentials.ProjectID)
	ctx := context.Background()

	// Usar las credenciales directamente como JSON
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(cfg.GCPCredentials.JSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &StorageService{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *StorageService) ProcessFile(objectName string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	log.Printf("Attempting to process file: %s from bucket: %s", objectName, s.bucketName)

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)

	// Verificar si el objeto existe
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, fmt.Errorf("object does not exist: %s", objectName)
		}
		return nil, fmt.Errorf("error checking object attributes: %w", err)
	}

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()

	var jsonData []map[string]interface{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var lineData map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &lineData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON line: %w", err)
		}
		jsonData = append(jsonData, lineData)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	log.Printf("Successfully processed file: %s, found %d JSON objects", objectName, len(jsonData))

	return jsonData, nil
}

func (s *StorageService) Close() error {
	return s.client.Close()
}
