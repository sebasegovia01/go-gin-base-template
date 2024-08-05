// storage.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"google.golang.org/api/option"
)

type StorageService struct {
	client     *storage.Client
	atmService *ATMService
}

func NewStorageService(cfg *config.Config, atmService *ATMService) (*StorageService, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.GCPCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &StorageService{
		client:     client,
		atmService: atmService,
	}, nil
}

func (s *StorageService) ProcessFile(bucketName, objectName string) error {
	ctx := context.Background()

	bucket := s.client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	payload, ok := jsonData["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("payload not found in JSON")
	}

	atm, err := convertToATM(payload)
	if err != nil {
		return fmt.Errorf("failed to convert payload to ATM: %w", err)
	}

	_, err = s.atmService.Create(*atm)
	if err != nil {
		return fmt.Errorf("failed to create ATM in database: %w", err)
	}

	fmt.Printf("Successfully processed and saved ATM with ID: %d\n", atm.ID)

	return nil
}

func convertToATM(payload map[string]interface{}) (*models.ATM, error) {
	atm := &models.ATM{}

	if id, ok := payload["id"].(string); ok {
		atm.ID, _ = strconv.Atoi(id)
	}
	atm.ATMIdentifier = payload["atmidentifier"].(string)
	atm.ATMAddressStreetName = payload["atmaddress_streetname"].(string)
	atm.ATMAddressBuildingNumber = payload["atmaddress_buildingnumber"].(string)
	atm.ATMTownName = payload["atmtownname"].(string)
	atm.ATMDistrictName = payload["atmdistrictname"].(string)
	atm.ATMCountrySubdivisionMajorName = payload["atmcountrysubdivisionmajorname"].(string)
	atm.ATMTimeType = payload["atmtimetype"].(string)
	atm.ATMAttentionHour = payload["atmattentionhour"].(string)
	atm.ATMServiceType = payload["atmservicetype"].(string)
	atm.ATMAccessType = payload["atmaccesstype"].(string)

	fromDateTime, err := time.Parse("2006-01-02 15:04:05.000", payload["atmfromdatetime"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse atmfromdatetime: %w", err)
	}
	atm.ATMFromDateTime = fromDateTime

	toDateTime, err := time.Parse("2006-01-02 15:04:05.000", payload["atmtodatetime"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse atmtodatetime: %w", err)
	}
	atm.ATMToDateTime = toDateTime

	return atm, nil
}

func (s *StorageService) Close() error {
	return s.client.Close()
}
