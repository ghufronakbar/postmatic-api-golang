// sql.go
package utils

import (
	"database/sql"
	"strings"
	"time"
)

// StringToNullString mengonversi string biasa/pointer ke sql.NullString
func StringToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullStringToString mengonversi sql.NullString ke *string untuk JSON response
func NullStringToString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// NullStringToStringVal mengonversi sql.NullString ke string kosong jika null (untuk field non-pointer)
func NullStringToStringVal(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func NullStringToNullTime(s *string) sql.NullTime {
	if s == nil {
		return sql.NullTime{Valid: false}
	}

	val := strings.TrimSpace(*s)
	if val == "" {
		return sql.NullTime{Valid: false}
	}

	// parse date YYYY-MM-DD
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: t, Valid: true}
}
