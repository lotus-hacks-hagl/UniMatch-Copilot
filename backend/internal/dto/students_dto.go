package dto

import (
	"unimatch-be/internal/model"
	"unimatch-be/pkg/response"
)

type ListStudentsResponse struct {
	Data []model.Student `json:"data"`
	Meta response.Meta   `json:"meta"`
}
