package services

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type AlloyDbService struct {
	db *sql.DB
}

func NewAlloyDB(db *sql.DB) *AlloyDbService {
	return &AlloyDbService{db: db}
}

// Execute a single SQL query
func (adb *AlloyDbService) ExecuteQuery(query string, params []interface{}, fetch bool) ([]map[string]interface{}, error) {
	rows, err := adb.db.Query(query, params...)
	if err != nil {
		log.Printf("Database operation failed: %v", err)
		return nil, fmt.Errorf("database operation failed: %w", err)
	}
	defer rows.Close()

	// If fetch is true, return the results
	if fetch {
		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("error getting columns: %w", err)
		}

		var results []map[string]interface{}
		for rows.Next() {
			// Create a slice of interface{} to hold each column's value
			values := make([]interface{}, len(columns))
			// Create a slice of pointers to the values
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// Scan the row into the value pointers
			err := rows.Scan(valuePtrs...)
			if err != nil {
				return nil, fmt.Errorf("error scanning row: %w", err)
			}

			// Create a map for the row with column names as keys
			rowMap := make(map[string]interface{})
			for i, col := range columns {
				rowMap[col] = values[i]
			}

			results = append(results, rowMap)
		}
		return results, nil
	}

	return nil, nil
}

// Execute a bulk query (multiple statements)
func (adb *AlloyDbService) ExecuteBulkQuery(query string, data [][]interface{}) (sql.Result, error) {
	tx, err := adb.db.Begin()
	if err != nil {
		log.Printf("Bulk query execution failed: %v", err)
		return nil, fmt.Errorf("bulk query execution failed: %w", err)
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error preparing bulk query: %w", err)
	}
	defer stmt.Close()

	for _, params := range data {
		_, err := stmt.Exec(params...)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error executing bulk query: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	log.Println("Bulk query executed successfully.")
	return nil, nil
}
