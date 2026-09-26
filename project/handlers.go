package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

// ---------- общие помощники ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseID достаёт {id} из пути и превращает в int64
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive number")
		return 0, false
	}
	return id, true
}

// mapStoreError единообразно превращает ошибки store в HTTP-ответы
func mapStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "database error")
	}
}

// ================= USERS =================

type userRequest struct {
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Phone *string `json:"phone"`
}

func validateUser(req *userRequest) string {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" {
		return "name is required"
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return "valid email is required"
	}
	return ""
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers(r.Context())
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	u, err := h.store.GetUser(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateUser(&req); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	u, err := h.store.CreateUser(r.Context(), req.Name, req.Email, req.Phone)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateUser(&req); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	u, err := h.store.UpdateUser(r.Context(), id, req.Name, req.Email, req.Phone)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteUser(r.Context(), id); err != nil {
		mapStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ================= CATEGORIES =================

type categoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func validateCategory(req *categoryRequest) string {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Name == "" {
		return "name is required"
	}
	if req.Slug == "" {
		return "slug is required"
	}
	return ""
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListCategories(r.Context())
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	c, err := h.store.GetCategory(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateCategory(&req); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	c, err := h.store.CreateCategory(r.Context(), req.Name, req.Slug)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateCategory(&req); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	c, err := h.store.UpdateCategory(r.Context(), id, req.Name, req.Slug)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteCategory(r.Context(), id); err != nil {
		mapStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ================= LISTINGS =================

type listingRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
	UserID      int64   `json:"user_id"`
	CategoryID  int64   `json:"category_id"`
}

func validateListing(req *listingRequest, isCreate bool) string {
	req.Title = strings.TrimSpace(req.Title)
	req.Status = strings.TrimSpace(req.Status)
	if isCreate {
		if req.UserID <= 0 {
			return "user_id is required"
		}
		if req.CategoryID <= 0 {
			return "category_id is required"
		}
	}
	if req.Title == "" {
		return "title is required"
	}
	if req.Price < 0 {
		return "price must be >= 0"
	}
	if req.Status != "" && req.Status != "active" && req.Status != "sold" && req.Status != "archived" {
		return "status must be active, sold or archived"
	}
	return ""
}

func (h *Handler) listListings(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListListings(r.Context())
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getListing(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	l, err := h.store.GetListing(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (h *Handler) createListing(w http.ResponseWriter, r *http.Request) {
	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateListing(&req, true); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	l, err := h.store.CreateListing(r.Context(), Listing{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

func (h *Handler) updateListing(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	// Для PATCH читаем текущее состояние, чтобы можно было обновить частично
	existing, err := h.store.GetListing(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}

	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if msg := validateListing(&req, false); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}

	// Пустые поля берём из существующей записи
	if req.Title == "" {
		req.Title = existing.Title
	}
	if req.Status == "" {
		req.Status = existing.Status
	}
	if req.UserID == 0 {
		req.UserID = existing.UserID
	}
	if req.CategoryID == 0 {
		req.CategoryID = existing.CategoryID
	}

	l, err := h.store.UpdateListing(r.Context(), id, Listing{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Status:      req.Status,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (h *Handler) deleteListing(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteListing(r.Context(), id); err != nil {
		mapStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ================= MESSAGES =================

type messageCreateRequest struct {
	Body      string `json:"body"`
	ListingID int64  `json:"listing_id"`
	SenderID  int64  `json:"sender_id"`
}

type messageUpdateRequest struct {
	Body   string `json:"body"`
	IsRead bool   `json:"is_read"`
}

func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListMessages(r.Context())
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	m, err := h.store.GetMessage(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request) {
	var req messageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		writeError(w, http.StatusUnprocessableEntity, "body is required")
		return
	}
	if req.ListingID <= 0 || req.SenderID <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "listing_id and sender_id are required")
		return
	}
	m, err := h.store.CreateMessage(r.Context(), req.Body, req.ListingID, req.SenderID)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *Handler) updateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	existing, err := h.store.GetMessage(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	var req messageUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		req.Body = existing.Body
	}
	m, err := h.store.UpdateMessage(r.Context(), id, req.Body, req.IsRead)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) deleteMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteMessage(r.Context(), id); err != nil {
		mapStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ================= FAVORITES =================

type favoriteCreateRequest struct {
	UserID    int64 `json:"user_id"`
	ListingID int64 `json:"listing_id"`
}

type favoriteUpdateRequest struct {
	UserID    int64 `json:"user_id"`
	ListingID int64 `json:"listing_id"`
}

func (h *Handler) listFavorites(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListFavorites(r.Context())
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	f, err := h.store.GetFavorite(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) createFavorite(w http.ResponseWriter, r *http.Request) {
	var req favoriteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UserID <= 0 || req.ListingID <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "user_id and listing_id are required")
		return
	}
	f, err := h.store.CreateFavorite(r.Context(), req.UserID, req.ListingID)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *Handler) updateFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	existing, err := h.store.GetFavorite(r.Context(), id)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	var req favoriteUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UserID == 0 {
		req.UserID = existing.UserID
	}
	if req.ListingID == 0 {
		req.ListingID = existing.ListingID
	}
	f, err := h.store.UpdateFavorite(r.Context(), id, req.UserID, req.ListingID)
	if err != nil {
		mapStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) deleteFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteFavorite(r.Context(), id); err != nil {
		mapStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
