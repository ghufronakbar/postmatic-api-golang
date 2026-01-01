// internal/http/middleware/owned_business.go
package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"postmatic-api/internal/repository/entity"
	ownedBusinessRepo "postmatic-api/internal/repository/redis/owned_business_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type contextOwnedBusinessKey string

const OwnedBusinessContextKey contextOwnedBusinessKey = "ownedBusinessContextKey"

type OwnedBusiness struct {
	store entity.Store
	repo  *ownedBusinessRepo.OwnedBusinessRepository
	ttl   time.Duration
}

func NewOwnedBusiness(store entity.Store, repo *ownedBusinessRepo.OwnedBusinessRepository) *OwnedBusiness {
	return &OwnedBusiness{store: store, repo: repo, ttl: time.Hour}
}

func (o *OwnedBusiness) OwnedBusinessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		businessId := chi.URLParam(r, "businessId")
		if businessId == "" {
			next.ServeHTTP(w, r)
			return
		}

		// pastikan auth middleware sudah jalan
		prof, err := GetUserFromContext(r.Context())
		if err != nil || prof == nil {
			response.Error(w, r, errs.NewUnauthorized("UNAUTHORIZED"), nil)
			return
		}

		intBusinessId, err := strconv.ParseInt(businessId, 10, 64)
		if err != nil {
			response.Error(w, r, errs.NewValidationFailed(map[string]string{
				"businessId": "businessId must be an integer64",
			}), nil)
			return
		}

		// 1) cek redis dulu
		list, err := o.repo.GetOwnedBusinessByProfileID(r.Context(), prof.ID)
		if err != nil {
			// redis error => boleh fallback DB (recommended) atau return 500
			// aku fallback DB biar lebih resilient:
			list = nil
		} else {
			for _, v := range list {
				if v.BusinessRootID == intBusinessId {
					ctx := context.WithValue(r.Context(), OwnedBusinessContextKey, &OwnedBusinessContext{
						MemberID:       v.MemberID,
						BusinessRootID: v.BusinessRootID,
						Role:           v.Role,
					})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// 2) kalau redis kosong / tidak ada businessId tsb => cek DB
		dbMember, err := o.store.GetMemberByProfileIdAndBusinessRootId(r.Context(),
			entity.GetMemberByProfileIdAndBusinessRootIdParams{
				ProfileID:      prof.ID,
				BusinessRootID: intBusinessId,
			},
		)

		if err == sql.ErrNoRows {
			fmt.Println("err == sql.ErrNoRows")
			// kamu bilang: kalau tidak ditemukan => forbidden karena bukan member
			response.Error(w, r, errs.NewForbidden("FORBIDDEN"), nil)
			return
		}

		if dbMember.Status != entity.BusinessMemberStatusAccepted {
			fmt.Println("dbMember.Status != entity.BusinessMemberStatusAccepted")
			response.Error(w, r, errs.NewForbidden("FORBIDDEN"), nil)
			return
		}

		if err != nil {
			response.Error(w, r, errs.NewInternalServerError(err), nil)
			return
		}

		// 3) upsert ke redis (best-effort; kalau gagal jangan block request)
		_ = o.repo.UpsertOneBusiness(r.Context(), prof.ID, ownedBusinessRepo.RedisBusinessSub{
			MemberID:       dbMember.ID,
			BusinessRootID: dbMember.BusinessRootID,
			Role:           dbMember.Role,
		}, o.ttl)

		// 4) set context dan lanjut
		ctx := context.WithValue(r.Context(), OwnedBusinessContextKey, &OwnedBusinessContext{
			MemberID:       dbMember.ID,
			BusinessRootID: dbMember.BusinessRootID,
			Role:           dbMember.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type OwnedBusinessContext struct {
	MemberID       int64                     `json:"memberId"`
	BusinessRootID int64                     `json:"businessRootId"`
	Role           entity.BusinessMemberRole `json:"role"`
}

func OwnedBusinessFromContext(ctx context.Context) (*OwnedBusinessContext, error) {
	v := ctx.Value(OwnedBusinessContextKey)
	if v == nil {
		fmt.Println("v == nil")
		return nil, errs.NewForbidden("FORBIDDEN")
	}
	ob, ok := v.(*OwnedBusinessContext)
	if !ok || ob == nil {
		fmt.Println("!ok || ob == nil")
		return nil, errs.NewForbidden("FORBIDDEN")
	}
	return ob, nil
}
