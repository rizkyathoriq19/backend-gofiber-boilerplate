package shared

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaginationQuery struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PaginationMeta struct {
	Page        int `json:"page"`
	Limit       int `json:"limit"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

func GetPaginationFromQuery(c *fiber.Ctx) PaginationQuery {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	return PaginationQuery{
		Page:  page,
		Limit: limit,
	}
}

func (p PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

func NewPaginationMeta(page, limit, total int) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	return PaginationMeta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}