package repository

import (
	"context"
	"database/sql"
	"fmt"
	"simpleapi/internal/models"
	"strings"
)

type StudentRepositoty struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepositoty {
	return &StudentRepositoty{DB: db}
}

func (r *StudentRepositoty) GetAll(ctx context.Context, filter models.StudentFilter) ([]models.Student, error) {
	query := "SELECT id, first_name, last_name,email, class FROM students WHERE 1=1"
	var args []interface{}

	query, args = r.addFilter(filter, query, args)
	query = r.addSorts(filter, query)

	// the context ctx serves as a kill switch for operations; if user closes the browser kill the request; or you can manually set a timeout for the context- this is purely server side kill switch for DB operations;
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to query students: %w", err)
	}

	defer rows.Close()

	// this initializes to [] when we don't append anything
	students := make([]models.Student, 0)

	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class); err != nil {
			return nil, fmt.Errorf("Failed to scan teacher row: %w", err)
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating rows: %w", err)
	}
	return students, nil

}

func (r *StudentRepositoty) GetByID(ctx context.Context, id int) (*models.Student, error) {
	var s models.Student
	query := "SELECT id, first_name, last_name, email, class FROM students WHERE id = ?"

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Class,
	)

	// 1. Translation: DB "No Rows" -> Domain "Not Found"
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Student %d not found: %w", id, models.ErrNotFound)
	}
	// 2. System Error
	if err != nil {
		return nil, fmt.Errorf("Failed to get student %d: %w", id, err)
	}
	return &s, nil
}

func (r *StudentRepositoty) CreateBulk(ctx context.Context, students []models.Student) ([]models.Student, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO students (first_name, last_name, email, class) VALUES(?,?,?,?)")

	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	result := make([]models.Student, len(students))
	for i, s := range students {
		res, err := stmt.ExecContext(ctx, s.FirstName, s.LastName, s.Email, s.Class)
		if err != nil {
			// Pro Tip: Check for MySQL duplicate entry error (Error 1062)
			if strings.Contains(err.Error(), "Duplicate entry") {
				return nil, fmt.Errorf("Duplicate email %s: %w", s.Email, models.ErrConflict)
			}
			// Check for Foreign Key Constraint Failure (Error 1452)
			if strings.Contains(err.Error(), "1452") {
				// We map this to ErrConflict or ErrInvalidInput depending on your preference
				return nil, fmt.Errorf("cannot assign student to class '%s' (class does not exist): %w", s.Class, err)
			}

			return nil, fmt.Errorf("Failed to insert student: %w", err)
		}

		id, _ := res.LastInsertId()
		s.ID = int(id)
		result[i] = s
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return result, nil
}

func (r *StudentRepositoty) addSorts(filter models.StudentFilter, query string) string {
	validSorts := map[string]bool{"first_name": true, "last_name": true, "email": true, "class": true}
	if validSorts[filter.SortBy] {
		order := "ASC"
		if strings.ToUpper(filter.SortOrder) == "DESC" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	}
	return query
}

func (r *StudentRepositoty) addFilter(filter models.StudentFilter, query string, args []interface{}) (string, []interface{}) {
	if filter.FirstName != "" {
		query += " AND first_name = ?"
		args = append(args, filter.FirstName)
	}
	if filter.LastName != "" {
		query += " AND last_name = ?"
		args = append(args, filter.LastName)
	}
	if filter.Email != "" {
		query += " AND email = ?"
		args = append(args, filter.Email)
	}
	if filter.Class != "" {
		query += " AND class = ?"
		args = append(args, filter.Class)
	}

	return query, args
}
