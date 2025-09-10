package rest

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/fuzail/08-todosvc/internal/todo"
)

type apiHandler struct {
	svc todo.Service
}

func RegisterHandlers(mux *http.ServeMux, svc todo.Service) {
	h := &apiHandler{svc: svc}
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/tasks", h.tasks)     // POST create, GET list
	mux.HandleFunc("/tasks/", h.taskByID) // GET, PUT, DELETE, PATCH mark complete
}

func (h *apiHandler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *apiHandler) tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodGet:
		h.listTasks(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *apiHandler) taskByID(w http.ResponseWriter, r *http.Request) {
	// path: /tasks/{id} or /tasks/{id}/complete
	id := r.URL.Path[len("/tasks/"):]
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	case http.MethodPatch:
		// support mark complete via PATCH with JSON {"completed": true}
		h.markComplete(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type createReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *apiHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	t, err := h.svc.CreateTask(ctx, req.Title, req.Description)
	if err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *apiHandler) getTask(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()
	t, err := h.svc.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *apiHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize == 0 {
		pageSize = 10
	}
	ctx := r.Context()
	tasks, total, err := h.svc.ListTasks(ctx, page, pageSize)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"tasks":     tasks,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}
	writeJSON(w, http.StatusOK, resp)
}

type updateReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *apiHandler) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	t, err := h.svc.UpdateTask(ctx, id, req.Title, req.Description)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type markReq struct {
	Completed bool `json:"completed"`
}

func (h *apiHandler) markComplete(w http.ResponseWriter, r *http.Request, id string) {
	var req markReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	t, err := h.svc.MarkComplete(ctx, id, req.Completed)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *apiHandler) deleteTask(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()
	if err := h.svc.DeleteTask(ctx, id); err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
