package models

import (
	"database/sql"
	"fmt"
	dbserver "literary-lions/internal/db"
)

type Subcategory struct {
	ID       int
	Title    string
	IconPath string
}

type CategoryModel struct {
	DB *sql.DB
}

func NewCategoryModel() (*CategoryModel, error) {
	db, err := sql.Open("sqlite3", DatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &CategoryModel{DB: db}, nil
}

// Get retrieves categories and their associated subcategories from the database
func (cm *CategoryModel) Get() (map[string][]Subcategory, error) {
	// Execute the query to retrieve category and subcategory data
	rows, err := cm.DB.Query(dbserver.CategoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make(map[string][]Subcategory)

	// Iterate over the rows returned by the query
	for rows.Next() {
		var categoryTitle, subcategoryTitle, iconPath string
		var subcategoryId int

		// Scan the row into variables
		if err := rows.Scan(&categoryTitle, &subcategoryId, &subcategoryTitle, &iconPath); err != nil {
			return nil, fmt.Errorf("failed row scan: %w", err)
		}

		// Create a new subcategory from the scanned data
		subcategory := Subcategory{
			ID:       subcategoryId,
			Title:    subcategoryTitle,
			IconPath: iconPath,
		}

		// Append the subcategory to the corresponding category in the map
		categories[categoryTitle] = append(categories[categoryTitle], subcategory)
	}

	return categories, nil
}
