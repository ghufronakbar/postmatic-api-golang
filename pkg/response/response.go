// pkg/response/response.go

package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"postmatic-api/pkg/errs" // Import package error yang kita buat diatas
	"postmatic-api/pkg/filter"
	"postmatic-api/pkg/pagination"
	"reflect"
)

// Struktur sesuai request Anda
type MetaData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BaseResponse struct {
	MetaData         MetaData               `json:"metaData"`
	ResponseMessage  string                 `json:"responseMessage"`
	Data             interface{}            `json:"data"` // 'any' di Go adalah interface{}
	ValidationErrors map[string]string      `json:"validationErrors"`
	FilterQuery      *filter.ReqFilter      `json:"filterQuery"`
	Pagination       *pagination.Pagination `json:"pagination"`
}

// Helper untuk mengirim response JSON
func JSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	metaMsg := http.StatusText(status)

	var validationErrors map[string]string
	var returnData interface{}

	if message == "VALIDATION_FAILED" {
		validationErrors = data.(map[string]string)
		returnData = nil
	} else {
		// ✅ normalize nil slice
		returnData = normalizeNilSlice(data)
	}

	resp := BaseResponse{
		MetaData: MetaData{
			Code:    status,
			Message: metaMsg,
		},
		ResponseMessage:  message,
		Data:             returnData,
		ValidationErrors: validationErrors,
		FilterQuery:      nil,
	}

	_ = json.NewEncoder(w).Encode(resp)
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

func LIST(w http.ResponseWriter, message string, data interface{}, filterQuery *filter.ReqFilter, pagination *pagination.Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	metaMsg := http.StatusText(200)

	resp := BaseResponse{
		MetaData: MetaData{
			Code:    200,
			Message: metaMsg,
		},
		ResponseMessage: message,
		// ✅ normalize nil slice
		Data:        normalizeNilSlice(data),
		FilterQuery: filterQuery,
		Pagination:  pagination,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Helper khusus untuk Error
func Error(w http.ResponseWriter, err error, data interface{}) {
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

		// Cek Validation Error
		if appErr.ValidationErrors != nil {
			JSON(w, http.StatusBadRequest, "VALIDATION_FAILED", appErr.ValidationErrors)
			return
		}
	}

	JSON(w, code, msg, data)
}

func ValidationFailed(w http.ResponseWriter, errsMap map[string]string) {
	JSON(w, http.StatusBadRequest, "VALIDATION_FAILED", errsMap)
}

func InvalidJsonFormat(w http.ResponseWriter) {
	JSON(w, http.StatusBadRequest, "INVALID_JSON_FORMAT", nil)
}

// normalizeNilSlice mengubah nil slice (mis. []T(nil)) menjadi []T{} agar JSON jadi [] bukan null.
func normalizeNilSlice(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	rv := reflect.ValueOf(data)
	rt := rv.Type()

	// interface berisi nil slice → Kind slice, IsNil true
	if rt.Kind() == reflect.Slice && rv.IsNil() {
		empty := reflect.MakeSlice(rt, 0, 0)
		return empty.Interface()
	}

	return data
}
