// sql.go
package utils

import "database/sql"

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
