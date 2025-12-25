// internal/module/app/timezone/viewmodel.go
package timezone

type TimezoneResponse struct {
	Name   string `json:"name"`
	Offset string `json:"offset"`
	Label  string `json:"label"`
}
