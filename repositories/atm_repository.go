package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sebasegovia01/base-template-go-gin/models"
)

type ATMRepository struct {
	db *sql.DB
}

func NewATMRepository(db *sql.DB) *ATMRepository {
	return &ATMRepository{db: db}
}

func (r *ATMRepository) Create(atm models.ATM) (*models.ATM, error) {
	query := `
		INSERT INTO presential_service_channels.automated_teller_machines (
			atmidentifier, atmaddress_streetname, atmaddress_buildingnumber,
			atmtownname, atmdistrictname, atmcountrysubdivisionmajorname,
			atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour,
			atmservicetype, atmaccesstype
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	err := r.db.QueryRow(
		query,
		atm.ATMIdentifier, atm.ATMAddressStreetName, atm.ATMAddressBuildingNumber,
		atm.ATMTownName, atm.ATMDistrictName, atm.ATMCountrySubdivisionMajorName,
		atm.ATMFromDateTime, atm.ATMToDateTime, atm.ATMTimeType, atm.ATMAttentionHour,
		atm.ATMServiceType, atm.ATMAccessType,
	).Scan(&atm.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating ATM: %w", err)
	}

	return &atm, nil
}

func (r *ATMRepository) GetAll() ([]models.ATM, error) {
	query := `SELECT * FROM presential_service_channels.automated_teller_machines`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all ATMs: %w", err)
	}
	defer rows.Close()

	var atms []models.ATM
	for rows.Next() {
		var atm models.ATM
		err := rows.Scan(
			&atm.ID, &atm.ATMIdentifier, &atm.ATMAddressStreetName, &atm.ATMAddressBuildingNumber,
			&atm.ATMTownName, &atm.ATMDistrictName, &atm.ATMCountrySubdivisionMajorName,
			&atm.ATMFromDateTime, &atm.ATMToDateTime, &atm.ATMTimeType, &atm.ATMAttentionHour,
			&atm.ATMServiceType, &atm.ATMAccessType,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ATM: %w", err)
		}
		atms = append(atms, atm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ATMs: %w", err)
	}

	return atms, nil
}

func (r *ATMRepository) GetByID(id int) (*models.ATM, error) {
	query := `SELECT * FROM presential_service_channels.automated_teller_machines WHERE id = $1`
	var atm models.ATM
	err := r.db.QueryRow(query, id).Scan(
		&atm.ID, &atm.ATMIdentifier, &atm.ATMAddressStreetName, &atm.ATMAddressBuildingNumber,
		&atm.ATMTownName, &atm.ATMDistrictName, &atm.ATMCountrySubdivisionMajorName,
		&atm.ATMFromDateTime, &atm.ATMToDateTime, &atm.ATMTimeType, &atm.ATMAttentionHour,
		&atm.ATMServiceType, &atm.ATMAccessType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ATM not found")
		}
		return nil, fmt.Errorf("error getting ATM by ID: %w", err)
	}
	return &atm, nil
}

func (r *ATMRepository) Update(atm models.ATM) (*models.ATM, error) {
	query := `
		UPDATE presential_service_channels.automated_teller_machines SET
			atmidentifier = $2, atmaddress_streetname = $3, atmaddress_buildingnumber = $4,
			atmtownname = $5, atmdistrictname = $6, atmcountrysubdivisionmajorname = $7,
			atmfromdatetime = $8, atmtodatetime = $9, atmtimetype = $10, atmattentionhour = $11,
			atmservicetype = $12, atmaccesstype = $13
		WHERE id = $1
		RETURNING *`

	err := r.db.QueryRow(
		query,
		atm.ID, atm.ATMIdentifier, atm.ATMAddressStreetName, atm.ATMAddressBuildingNumber,
		atm.ATMTownName, atm.ATMDistrictName, atm.ATMCountrySubdivisionMajorName,
		atm.ATMFromDateTime, atm.ATMToDateTime, atm.ATMTimeType, atm.ATMAttentionHour,
		atm.ATMServiceType, atm.ATMAccessType,
	).Scan(
		&atm.ID, &atm.ATMIdentifier, &atm.ATMAddressStreetName, &atm.ATMAddressBuildingNumber,
		&atm.ATMTownName, &atm.ATMDistrictName, &atm.ATMCountrySubdivisionMajorName,
		&atm.ATMFromDateTime, &atm.ATMToDateTime, &atm.ATMTimeType, &atm.ATMAttentionHour,
		&atm.ATMServiceType, &atm.ATMAccessType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ATM not found")
		}
		return nil, fmt.Errorf("error updating ATM: %w", err)
	}

	return &atm, nil
}

func (r *ATMRepository) Delete(id int) error {
	query := `DELETE FROM presential_service_channels.automated_teller_machines WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting ATM: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ATM not found")
	}

	return nil
}
