package response

import (
	"encoding/json"
	"net/http"
)

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

func RenderPaginatedJSON[T any](w http.ResponseWriter, items []T, pagination Pagination) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := PaginatedResponse[T]{
		Items:      items,
		Pagination: pagination,
	}

	return json.NewEncoder(w).Encode(response)
}

func RenderJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(data)
}
