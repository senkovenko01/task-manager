package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"task-manager/internal/models"
	"task-manager/internal/repository"
	"task-manager/internal/service"
)

const (
	ErrMsgUnhealthy        = "Server is not available"
	ErrMsgMethodNotAllowed = "Method not allowed. please use appropriate method for this operation"
	ErrMsgInvalidJSON      = "Invalid JSON! can't parse incoming model please check the input"
	ErrMsgInvalidStatus    = "Invalid status! Status can be only: `new`, `in_progress` or `done`"
	ErrMsgInvalidLimit     = "Invalid limit! Limit value must be greater than zero"
	ErrMsgInvalidOffset    = "Invalid offset! Offset value must be greater than zero"
	ErrMsgInvalidID        = "Invalid id! Id must be a valid uuid"
	ErrMsgNotFound         = "Not found!"
	ErrMsgFailedToList     = "Failed to list tasks due to an internal server error"
	ErrMsgFailedToGet      = "Failed to get task due to an internal server error"
	ErrMsgFailedToDelete   = "Failed to delete task due to an internal server error"
	ErrMsgTitleTooShort    = "Title must be at least 3 characters. Please check the input and try again"
)
const (
	DefaultLimit  = 50
	DefaultOffset = 0
)

type TaskHandler struct {
	service service.TaskService
}

func NewTaskHandler(service service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("POST /tasks", h.handleCreateTask)
	mux.HandleFunc("GET /tasks", h.handleListTasks)
	mux.HandleFunc("GET /tasks/{id}", h.handleGetTask)
	mux.HandleFunc("PUT /tasks/{id}", h.handleUpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.handleDeleteTask)
}

func (h *TaskHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		http.Error(w, ErrMsgUnhealthy, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *TaskHandler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	var createInput models.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&createInput); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}
	task, err := h.service.CreateTask(r.Context(), createInput)
	if err != nil {
		errMsg := err.Error()
		if errMsg == ErrMsgTitleTooShort {
			http.Error(w, ErrMsgTitleTooShort, http.StatusBadRequest)
		} else {
			http.Error(w, errMsg, http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, ErrMsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	queryParams := r.URL.Query()
	var taskStatus *models.TaskStatus
	if statusStr := queryParams.Get("status"); statusStr != "" {
		parsedStatus := models.TaskStatus(statusStr)
		switch parsedStatus {
		case models.TaskStatusNew, models.TaskStatusInProgress, models.TaskStatusDone:
			taskStatus = &parsedStatus
		default:
			http.Error(w, ErrMsgInvalidStatus, http.StatusBadRequest)
			return
		}
	}
	limit := DefaultLimit
	offset := DefaultOffset
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if limitValue, err := strconv.Atoi(limitStr); err == nil && limitValue > 0 {
			limit = limitValue
		} else {
			http.Error(w, ErrMsgInvalidLimit, http.StatusBadRequest)
			return
		}
	}
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		if offsetValue, err := strconv.Atoi(offsetStr); err == nil && offsetValue >= 0 {
			offset = offsetValue
		} else {
			http.Error(w, ErrMsgInvalidOffset, http.StatusBadRequest)
			return
		}
	}

	tasks, err := h.service.ListTasks(r.Context(), taskStatus, limit, offset)
	if err != nil {
		http.Error(w, ErrMsgFailedToList, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) handleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, ErrMsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	taskIDStr := r.PathValue("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, ErrMsgInvalidID, http.StatusBadRequest)
		return
	}
	task, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			http.Error(w, ErrMsgNotFound, http.StatusNotFound)
		} else {
			http.Error(w, ErrMsgFailedToGet, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, ErrMsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	taskIDStr := r.PathValue("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, ErrMsgInvalidID, http.StatusBadRequest)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	var updateInput models.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&updateInput); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}
	task, err := h.service.UpdateTask(r.Context(), taskID, updateInput)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			http.Error(w, ErrMsgNotFound, http.StatusNotFound)
		} else {
			errMsg := err.Error()
			if errMsg == ErrMsgTitleTooShort {
				http.Error(w, ErrMsgTitleTooShort, http.StatusBadRequest)
			} else if errMsg == ErrMsgInvalidStatus {
				http.Error(w, ErrMsgInvalidStatus, http.StatusBadRequest)
			} else {
				http.Error(w, errMsg, http.StatusBadRequest)
			}
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, ErrMsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	taskIDStr := r.PathValue("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, ErrMsgInvalidID, http.StatusBadRequest)
		return
	}
	err = h.service.DeleteTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			http.Error(w, ErrMsgNotFound, http.StatusNotFound)
		} else {
			http.Error(w, ErrMsgFailedToDelete, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
