// internal/module/app/timezone/service.go
package timezone

import (
	"context"
	"embed"
	"encoding/json"
	"sync"
	"time"

	"postmatic-api/pkg/errs"
)

//go:embed timezone.json
var timezoneFS embed.FS

type TimezoneService struct {
	once    sync.Once
	list    []TimezoneResponse
	byName  map[string]TimezoneResponse
	loadErr error
}

func NewTimezoneService() *TimezoneService {
	return &TimezoneService{}
}

func (s *TimezoneService) load() {
	s.once.Do(func() {
		b, err := timezoneFS.ReadFile("timezone.json")
		if err != nil {
			s.loadErr = errs.NewInternalServerError(err)
			return
		}

		var parsed []TimezoneResponse
		if err := json.Unmarshal(b, &parsed); err != nil {
			s.loadErr = errs.NewInternalServerError(err)
			return
		}

		m := make(map[string]TimezoneResponse, len(parsed))
		for _, tz := range parsed {
			// skip item invalid di JSON (mis. name kosong)
			if tz.Name == "" {
				continue
			}
			m[tz.Name] = tz
		}

		s.list = parsed
		s.byName = m
	})
}

func (s *TimezoneService) GetAllTimezone(ctx context.Context) ([]TimezoneResponse, error) {
	s.load()
	if s.loadErr != nil {
		return nil, s.loadErr
	}

	// return copy (biar caller tidak bisa mutate cache internal)
	out := make([]TimezoneResponse, 0, len(s.list))
	out = append(out, s.list...)
	return out, nil
}

func (s *TimezoneService) ValidateTimezone(ctx context.Context, name string) (TimezoneResponse, error) {
	if name == "" {
		return TimezoneResponse{}, errs.NewBadRequest("TIMEZONE_NOT_VALID")
	}

	s.load()
	if s.loadErr != nil {
		return TimezoneResponse{}, s.loadErr
	}

	tz, ok := s.byName[name]
	if !ok {
		return TimezoneResponse{}, errs.NewBadRequest("TIMEZONE_NOT_VALID")
	}

	// Validasi runtime: memastikan timezone ini bisa di-load oleh Go pada environment sekarang.
	// Ini penting kalau besok mau dipakai cron (agar tidak gagal saat scheduler parse).
	if _, err := time.LoadLocation(name); err != nil {
		return TimezoneResponse{}, errs.NewBadRequest("TIMEZONE_NOT_VALID")
	}

	return tz, nil
}
