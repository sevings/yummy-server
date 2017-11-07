package yummy

import (
	"github.com/sevings/yummy-server/gen/models"
)

// NewError returns error object with some message
func NewError(msg string) *models.Error {
	return &models.Error{Message: msg}
}
