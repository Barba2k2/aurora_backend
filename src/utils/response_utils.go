package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponseDetails contém os detalhes do erro
type ErrorResponseDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse representa o formato padrão de resposta de erro
type ErrorReponse struct {
	Success bool                 `json:"success"`
	Error   ErrorResponseDetails `json:"error"`
}

// SuccessResponse representa o formato padrão de resposta de sucesso
type SuccessResponse struct {
	Success bool                   `json:"success"`
	Data    interface{}            `json:"data,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// PaginationMeta contem metadados de paginacao
type PaginationMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

// SendErrorResponse envia uma resposta de erro padronizada
func SendErrorResponse(ctx *gin.Context, status int, code string, message string, details map[string]interface{}) {
	response := ErrorReponse{
		Success: false,
		Error: ErrorResponseDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	ctx.JSON(status, response)
}

// SendSuccessResponse envia uma resposta de sucesso padronizada
func SendSuccessResponse(ctx *gin.Context, status int, data interface{}, meta map[string]interface{}) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	ctx.JSON(status, response)
}

// SendSuccessResponseWithPagination envia uma resposta de sucesso com paginação
func SendSuccessResponseWithPagination(ctx *gin.Context, data interface{}, total, page, limit int) {
	// Calculamos o numero total de paginas
	pages := total / limit
	if total%limit > 0 {
		pages++
	}

	// Criamos os metadados de paginação
	pagination := PaginationMeta{
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	}

	meta := map[string]interface{}{
		"pagination": pagination,
	}

	response := SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	ctx.JSON(http.StatusOK, response)
}

// SendNoContentResponse envia uma resposta sem conteúdo
func SendNoContentResponse(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}
