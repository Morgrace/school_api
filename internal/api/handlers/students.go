package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"simpleapi/internal/models"
	"simpleapi/internal/repository"
	"simpleapi/pkg/utils"
)

type StudentHandler struct {
	Repo *repository.StudentRepositoty
}

func NewStudentHandler(repo *repository.StudentRepositoty) *StudentHandler {
	return &StudentHandler{Repo: repo}
}

func (h *StudentHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	filter := models.StudentFilter{
		FirstName: r.URL.Query().Get("first_name"),
		LastName:  r.URL.Query().Get("last_name"),
		Email:     r.URL.Query().Get("email"),
		Class:     r.URL.Query().Get("class"),
		SortBy:    r.URL.Query().Get("sortby"),
		SortOrder: r.URL.Query().Get("order"),
	}
	students, err := h.Repo.GetAll(r.Context(), filter)
	if err != nil {
		log.Printf("Error fetching students list: %v", err)
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

func (h *StudentHandler) CreateStudents(w http.ResponseWriter, r *http.Request) {
	var newStudents []models.Student

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&newStudents); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	studentValidationErrors := models.ValidateBatch(newStudents)

	if len(studentValidationErrors) > 0 {
		utils.WriteError(w, 400, "Validation failed", studentValidationErrors)
		return
	}

	added, err := h.Repo.CreateBulk(r.Context(), newStudents)
	if err != nil {
		log.Printf("Error creating students builk %v", err)
		utils.ResponseError(w, err, "")
		return
	}

	response := struct {
		Count int              `json:"count"`
		Data  []models.Student `json:"data"`
	}{
		Count: len(added),
		Data:  added,
	}

	utils.WriteJSON(w, 201, "Students created successfully", response)
}
