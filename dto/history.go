package dto

import (
	"errors"
	"ulascan-be/entity"

	"github.com/google/uuid"
)

const (
	// Failed
	MESSAGE_FAILED_CREATE_HISTORY = "failed create history"
	MESSAGE_FAILED_GET_HISTORIES  = "failed get histories"
	MESSAGE_FAILED_GET_HISTORY    = "failed get history"

	// Success
	MESSAGE_SUCCESS_CREATE_HISTORY = "success create history"
	MESSAGE_SUCCESS_GET_HISTORIES  = "success get histories"
	MESSAGE_SUCCESS_GET_HISTORY    = "success get history"
)

var (
	ErrCreateHistory = errors.New("failed to create history")
	ErrDeleteHistory = errors.New("failed to delete history")
	ErrGetHistories  = errors.New("failed to get histories")
	ErrGetHistory    = errors.New("failed to get history")
)

type (
	HistoriesGetRequest struct {
		Page        int    `json:"page"`
		Limit       int    `json:"limit"`
		ProductName string `json:"product_name"`
	}

	HistoriesResponse struct {
		Histories []entity.History `json:"histories"`
		Page      int              `json:"page"`
		Pages     int              `json:"pages"`
		Limit     int              `json:"limit"`
		Total     int64            `json:"total"`
	}

	HistoryCreateRequest struct {
		UserID           uuid.UUID `json:"user_id" form:"user_id" binding:"required"`
		ProductID        string    `json:"product_id" form:"product_id" binding:"required"`
		URL              string    `json:"url" form:"url" binding:"required"`
		Rating           int       `json:"rating" form:"rating" binding:"required"`
		Ulasan           int       `json:"ulasan" form:"ulasan" binding:"required"`
		Bintang          float64   `json:"bintang" form:"bintang" binding:"required"`
		ProductName      string    `json:"product_name" form:"product_name" binding:"required"`
		CountPositive    int       `json:"count_positive" form:"count_positive" binding:"required"`
		CountNegative    int       `json:"count_negative" form:"count_negative" binding:"required"`
		Packaging        float32   `json:"packaging"  form:"packaging" binding:"required"`
		Delivery         float32   `json:"delivery" form:"delivery" binding:"required"`
		AdminResponse    float32   `json:"admin_response" form:"admin_response" binding:"required"`
		ProductCondition float32   `json:"product_condition" form:"product_condition" binding:"required"`
		Summary          string    `json:"summary" form:"content"`
	}

	HistoryResponse struct {
		ID               uuid.UUID `json:"id"`
		UserID           uuid.UUID `json:"user_id"`
		ProductID        string    `json:"product_id"`
		URL              string    `json:"url"`
		Rating           int       `json:"rating"`
		Ulasan           int       `json:"ulasan"`
		Bintang          float64   `json:"bintang"`
		ProductName      string    `json:"product_name"`
		CountPositive    int       `json:"count_positive" `
		CountNegative    int       `json:"count_negative" `
		Packaging        float32   `json:"packaging"`
		Delivery         float32   `json:"delivery"`
		AdminResponse    float32   `json:"admin_response"`
		ProductCondition float32   `json:"product_condition"`
		Summary          string    `json:"summary"`
	}
)
