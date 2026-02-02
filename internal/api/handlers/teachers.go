package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"simpleapi/internal/models"
	"simpleapi/internal/repository"
	"simpleapi/pkg/utils"
	"strconv"
)

// TeacherHandler holds the dependencies for these HTTP endpoints
type TeacherHandler struct {
	Repo *repository.TeacherRepository
}

// NewTeacherHandler is the constructor
func NewTeacherHandler(repo *repository.TeacherRepository) *TeacherHandler {
	return &TeacherHandler{Repo: repo}
}

// --- HANDLERS ---

func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	filter := models.TeacherFilter{
		FirstName: r.URL.Query().Get("first_name"),
		LastName:  r.URL.Query().Get("last_name"),
		Email:     r.URL.Query().Get("email"),
		Class:     r.URL.Query().Get("class"),
		Subject:   r.URL.Query().Get("subject"),
		SortBy:    r.URL.Query().Get("sortby"),
		SortOrder: r.URL.Query().Get("order"),
	}

	teachers, err := h.Repo.GetAll(r.Context(), filter)
	if err != nil {
		// Log the internal error details for the developer
		log.Printf("Error fetching teachers list: %v", err)
		utils.ResponseError(w, err, "")
		return
	}

	response := struct {
		Count int              `json:"count"`
		Data  []models.Teacher `json:"data"`
	}{
		Count: len(teachers),
		Data:  teachers,
	}

	// util automatically adds "status": "success"
	utils.WriteJSON(w, http.StatusOK, "Teachers fetched successfully", response)
}

func (h *TeacherHandler) GetTeacherByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid teacher ID")
		return
	}

	teacher, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		// Log the error (could be "Not Found" or "DB Connection Failed")
		log.Printf("Error fetching teacher %d: %v", id, err)

		// If 404, util sends 404. If DB crash, util sends 500.
		utils.ResponseError(w, err, fmt.Sprintf("Teacher with ID %d not found", id))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Teacher fetched successfully", teacher)
}

func (h *TeacherHandler) CreateTeachers(w http.ResponseWriter, r *http.Request) {
	var newTeachers []models.Teacher

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&newTeachers); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	teacherValidationErrors := models.ValidateBatch(newTeachers)

	if len(teacherValidationErrors) > 0 {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", teacherValidationErrors)
		return
	}

	added, err := h.Repo.CreateBulk(r.Context(), newTeachers)
	if err != nil {
		log.Printf("Error creating teachers bulk: %v", err)
		utils.ResponseError(w, err, "")
		return
	}

	response := struct {
		Count int              `json:"count"`
		Data  []models.Teacher `json:"data"`
	}{
		Count: len(added),
		Data:  added,
	}

	utils.WriteJSON(w, http.StatusCreated, "Teachers created successfully", response)
}

func (h *TeacherHandler) UpdateTeacherFull(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid teacher ID")
		return
	}

	var updatedTeacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&updatedTeacher); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	result, err := h.Repo.UpdateFull(r.Context(), id, updatedTeacher)
	if err != nil {
		log.Printf("Error updating teacher %d: %v", id, err)
		utils.ResponseError(w, err, fmt.Sprintf("Teacher with ID %d not found", id))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Teacher updated successfully", result)
}

func (h *TeacherHandler) PatchTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid teacher ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	result, err := h.Repo.Patch(r.Context(), id, updates)
	if err != nil {
		// Log error (includes validation errors from Repo or DB errors)
		log.Printf("Error patching teacher %d: %v", id, err)

		// This handles both "Not Found" AND "Invalid Input"
		utils.ResponseError(w, err, "")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Teacher updated successfully", result)
}

func (h *TeacherHandler) BulkPatchTeachers(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	updatedIds, err := h.Repo.BulkPatch(r.Context(), updates)
	if err != nil {
		log.Printf("Error during bulk patch: %v", err)
		utils.ResponseError(w, err, "Bulk patch failed")
		return
	}

	response := map[string]interface{}{
		"message":     fmt.Sprintf("Successfully updated %d teachers", len(updatedIds)),
		"updated_ids": updatedIds,
	}

	utils.WriteJSON(w, http.StatusOK, "Teachers updated successfully", response)
}

func (h *TeacherHandler) DeleteTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid teacher ID")
		return
	}

	deleted, err := h.Repo.Delete(r.Context(), id)
	if err != nil {
		log.Printf("Error deleting teacher %d: %v", id, err)
		utils.ResponseError(w, err, "")
		return
	}

	// If Delete returns false, it means 0 rows affected (Not Found)
	if !deleted {
		// No need to log here (it's just a user mistake), but you can if you want
		utils.WriteError(w, http.StatusNotFound, "Teacher not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeacherHandler) BulkDeleteTeachers(w http.ResponseWriter, r *http.Request) {
	var ids []int
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	if len(ids) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "No IDs provided")
		return
	}

	validIds, err := h.Repo.BulkDelete(r.Context(), ids)
	if err != nil {
		log.Printf("Error during bulk delete: %v", err)
		utils.ResponseError(w, err, "Bulk delete failed")
		return
	}

	if len(validIds) == 0 {
		utils.WriteError(w, http.StatusNotFound, "None of the provided IDs exist")
		return
	}

	response := struct {
		DeletedIDs []int `json:"deleted_ids"`
	}{
		DeletedIDs: validIds,
	}

	utils.WriteJSON(w, http.StatusOK, "Teachers deleted successfully", response)
}

func (h *TeacherHandler) GetStudentsByTeacherId(w http.ResponseWriter, r *http.Request) {
	teacherId := r.PathValue("id")

	var students = make([]models.Student, 0)

	query := `SELECT s.id, s.first_name, s.last_name, s.email, s.class 
			  FROM teachers t 
			  INNER JOIN students s ON t.class = s.class
			  WHERE t.id = ?`

	rows, err := h.Repo.DB.QueryContext(r.Context(), query, teacherId)
	if err != nil {
		log.Println(err)
		utils.ResponseError(w, err, "")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student

		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			log.Println(err)
			utils.ResponseError(w, err, "")
			return
		}
		students = append(students, student)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		utils.ResponseError(w, err, "")
		return
	}
	response := struct {
		Count int              `json:"count"`
		Data  []models.Student `json:"data"`
	}{
		Count: len(students),
		Data:  students,
	}

	utils.WriteJSON(w, 200, "Students fetched successfully", response)

}
