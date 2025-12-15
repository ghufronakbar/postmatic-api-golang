// pkg/response/response.go

package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"postmatic-api/pkg/errs" // Import package error yang kita buat diatas
)

// Struktur sesuai request Anda
type MetaData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BaseResponse struct {
	MetaData        MetaData    `json:"metaData"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"` // 'any' di Go adalah interface{}
}

// Helper untuk mengirim response JSON
func JSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Mapping status code ke message standard HTTP (opsional, biar rapi)
	metaMsg := http.StatusText(status)

	resp := BaseResponse{
		MetaData: MetaData{
			Code:    status,
			Message: metaMsg,
		},
		ResponseMessage: message,
		Data:            data,
	}

	json.NewEncoder(w).Encode(resp)
}

func OK(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	// Mapping status code ke message standard HTTP (opsional, biar rapi)
	metaMsg := http.StatusText(200)

	resp := BaseResponse{
		MetaData: MetaData{
			Code:    200,
			Message: metaMsg,
		},
		ResponseMessage: message,
		Data:            data,
	}

	json.NewEncoder(w).Encode(resp)
}

// Helper khusus untuk Error
func Error(w http.ResponseWriter, err error) {
	// Default error (jika bukan AppError) -> 500
	code := http.StatusInternalServerError
	msg := "INTERNAL_SERVER_ERROR"

	// Cek apakah error tersebut adalah *errs.AppError
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		// Jika ya, pakai Code & Message dari error tersebut
		code = appErr.Code
		msg = appErr.Message

		if appErr.Code == http.StatusInternalServerError {
			fmt.Println(err)
		}
	}

	JSON(w, code, msg, nil)
}

func ValidationFailed(w http.ResponseWriter, errsMap map[string]string) {
	JSON(w, http.StatusBadRequest, "VALIDATION_FAILED", errsMap)
}

func InvalidJsonFormat(w http.ResponseWriter) {
	JSON(w, http.StatusBadRequest, "INVALID_JSON_FORMAT", nil)
}
