package repository

import (
	"context"
	"database/sql"
	"fmt"
	"simpleapi/internal/models"
	"strings"
)

// TeacherRepository holds the dependency (the DB connection)
type TeacherRepository struct {
	DB *sql.DB
}

// NewTeacherRepository is the constructor
func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{DB: db}
}

// --- READ ---

func (r *TeacherRepository) GetAll(ctx context.Context, filter models.TeacherFilter) ([]models.Teacher, error) {
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}

	query, args = r.addFilter(filter, query, args)
	query = r.addSorts(filter, query)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to query teachers: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice (returns [] in JSON instead of null)
	teachers := make([]models.Teacher, 0)
	for rows.Next() {
		var t models.Teacher
		if err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Class, &t.Subject); err != nil {
			return nil, fmt.Errorf("repo: failed to scan teacher row: %w", err)
		}
		teachers = append(teachers, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: error iterating rows: %w", err)
	}
	return teachers, nil
}

func (r *TeacherRepository) GetByID(ctx context.Context, id int) (*models.Teacher, error) {
	var t models.Teacher
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?"

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Class, &t.Subject,
	)

	// 1. Translation: DB "No Rows" -> Domain "Not Found"
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repo: teacher %d not found: %w", id, models.ErrNotFound)
	}
	// 2. System Error
	if err != nil {
		return nil, fmt.Errorf("repo: failed to get teacher %d: %w", id, err)
	}
	return &t, nil
}

// --- CREATE ---

func (r *TeacherRepository) CreateBulk(ctx context.Context, teachers []models.Teacher) ([]models.Teacher, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, fmt.Errorf("repo: failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	result := make([]models.Teacher, len(teachers))
	for i, t := range teachers {
		res, err := stmt.ExecContext(ctx, t.FirstName, t.LastName, t.Email, t.Class, t.Subject)
		if err != nil {
			// Pro Tip: Check for MySQL duplicate entry error (Error 1062)
			if strings.Contains(err.Error(), "Duplicate entry") {
				return nil, fmt.Errorf("repo: duplicate email %s: %w", t.Email, models.ErrConflict)
			}
			return nil, fmt.Errorf("repo: failed to insert teacher: %w", err)
		}
		id, _ := res.LastInsertId()
		t.ID = int(id)
		result[i] = t
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repo: failed to commit transaction: %w", err)
	}
	return result, nil
}

// --- UPDATE & PATCH ---

func (r *TeacherRepository) UpdateFull(ctx context.Context, id int, update models.Teacher) (*models.Teacher, error) {
	query := "UPDATE teachers SET first_name=?, last_name=?, email=?, class=?, subject=? WHERE id=?"
	res, err := r.DB.ExecContext(ctx, query, update.FirstName, update.LastName, update.Email, update.Class, update.Subject, id)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to update teacher: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repo: failed to get rows affected: %w", err)
	}

	// Translation: 0 rows affected -> Domain "Not Found"
	if rows == 0 {
		return nil, fmt.Errorf("repo: teacher %d not found: %w", id, models.ErrNotFound)
	}

	update.ID = id
	return &update, nil
}

func (r *TeacherRepository) Patch(ctx context.Context, id int, updates map[string]interface{}) (*models.Teacher, error) {
	// Re-use GetByID (it handles Not Found logic for us!)
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query := "UPDATE teachers SET "
	var args []interface{}
	var columns []string
	allowedCols := map[string]bool{"first_name": true, "last_name": true, "email": true, "class": true, "subject": true}

	for k, v := range updates {
		if !allowedCols[k] {
			continue
		}

		strVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("repo: field %s expected string: %w", k, models.ErrInvalidInput)
		}

		columns = append(columns, fmt.Sprintf("%s = ?", k))
		args = append(args, strVal)

		// Update in-memory struct
		switch k {
		case "first_name":
			current.FirstName = strVal
		case "last_name":
			current.LastName = strVal
		case "email":
			current.Email = strVal
		case "class":
			current.Class = strVal
		case "subject":
			current.Subject = strVal
		}
	}

	if len(columns) > 0 {
		query += strings.Join(columns, ", ") + " WHERE id = ?"
		args = append(args, id)
		if _, err := r.DB.ExecContext(ctx, query, args...); err != nil {
			return nil, fmt.Errorf("repo: failed to patch teacher: %w", err)
		}
	}
	return current, nil
}

func (r *TeacherRepository) BulkPatch(ctx context.Context, updates []map[string]interface{}) ([]int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	var updatedIds []int
	for _, update := range updates {
		idFloat, ok := update["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("repo: missing or invalid 'id' in patch data: %w", models.ErrInvalidInput)
		}
		id := int(idFloat)

		rows, err := r.updateTeacherTx(ctx, tx, id, update)
		if err != nil {
			return nil, fmt.Errorf("repo: patch failed for id %d: %w", id, err)
		}
		if rows == 0 {
			// In bulk ops, if one ID is missing, we fail the batch (common practice)
			return nil, fmt.Errorf("repo: teacher %d not found: %w", id, models.ErrNotFound)
		}
		updatedIds = append(updatedIds, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repo: failed to commit tx: %w", err)
	}
	return updatedIds, nil
}

func (r *TeacherRepository) updateTeacherTx(ctx context.Context, tx *sql.Tx, id int, updates map[string]interface{}) (int64, error) {
	query := "UPDATE teachers SET "
	var args []interface{}
	var columns []string
	allowedCols := map[string]bool{"first_name": true, "last_name": true, "email": true, "class": true, "subject": true}

	for k, v := range updates {
		if k == "id" || !allowedCols[k] {
			continue
		}

		strVal, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("field %s expected string", k)
		}
		columns = append(columns, fmt.Sprintf("%s=?", k))
		args = append(args, strVal)
	}

	if len(columns) == 0 {
		return 1, nil // No fields to update, but ID exists conceptually
	}

	query += strings.Join(columns, ", ") + " WHERE id=?"
	args = append(args, id)

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- DELETE ---

func (r *TeacherRepository) Delete(ctx context.Context, id int) (bool, error) {
	res, err := r.DB.ExecContext(ctx, "DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("repo: delete failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repo: failed to check rows: %w", err)
	}

	// We return 'false' if 0 rows deleted, Handler converts this to 404
	return rows > 0, nil
}

func (r *TeacherRepository) BulkDelete(ctx context.Context, ids []int) ([]int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if len(ids) == 0 {
		return nil, nil
	}

	// 1. Verify existence using FOR UPDATE (Locks rows)
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	querySelect := fmt.Sprintf("SELECT id FROM teachers WHERE id IN (%s) FOR UPDATE", strings.Join(placeholders, ","))
	rows, err := tx.QueryContext(ctx, querySelect, args...)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to check bulk IDs: %w", err)
	}

	var validIds []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("repo: scan failed: %w", err)
		}
		validIds = append(validIds, id)
	}
	rows.Close()

	if len(validIds) == 0 {
		return nil, nil // Nothing to delete
	}

	// 2. Delete valid
	validPlaceholders := make([]string, len(validIds))
	validArgs := make([]interface{}, len(validIds))
	for i, id := range validIds {
		validPlaceholders[i] = "?"
		validArgs[i] = id
	}

	queryDelete := fmt.Sprintf("DELETE FROM teachers WHERE id IN (%s)", strings.Join(validPlaceholders, ","))
	if _, err := tx.ExecContext(ctx, queryDelete, validArgs...); err != nil {
		return nil, fmt.Errorf("repo: bulk delete failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repo: commit failed: %w", err)
	}
	return validIds, nil
}

// func (r *TeacherRepository)

// --- HELPERS ---
func (r *TeacherRepository) addSorts(filter models.TeacherFilter, query string) string {
	validSorts := map[string]bool{"first_name": true, "last_name": true, "email": true, "class": true, "subject": true}
	if validSorts[filter.SortBy] {
		order := "ASC"
		if strings.ToUpper(filter.SortOrder) == "DESC" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	}
	return query
}

func (r *TeacherRepository) addFilter(filter models.TeacherFilter, query string, args []interface{}) (string, []interface{}) {
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
	if filter.Subject != "" {
		query += " AND subject = ?"
		args = append(args, filter.Subject)
	}
	return query, args
}
