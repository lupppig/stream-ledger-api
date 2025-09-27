package utils

import (
	"fmt"
	"net/http"
)

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

func GetPagination(r *http.Request) Pagination {
	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
