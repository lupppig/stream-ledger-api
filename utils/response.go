package utils

import (
	"net/http"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

func BuildResponse(statusCode int, message string, data interface{}, err interface{}, pagination interface{}) Response {
	return Response{StatusCode: statusCode, Message: message, Data: data, Error: err, Pagination: pagination}
}

func (r Response) BadResponse(w http.ResponseWriter) {
	SendJSONResponse(w, r, r.StatusCode)
}

func (r Response) SuccessResponse(w http.ResponseWriter) {
	SendJSONResponse(w, r, r.StatusCode)
}
