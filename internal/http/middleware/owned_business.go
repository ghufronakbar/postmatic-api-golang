// internal/http/middleware/owned_business.go
package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"postmatic-api/internal/repository/entity"
	ownedBusinessRepo "postmatic-api/internal/repository/redis/owned_business_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

		// parse UUID
		profUUID, err := uuid.Parse(prof.ID)
		if err != nil {
			response.Error(w, r, errs.NewBadRequest("INVALID_PROFILE_ID"), nil)
			return
		}
		businessUUID, err := uuid.Parse(businessId)
		if err != nil {
			response.Error(w, r, errs.NewBadRequest("INVALID_BUSINESS_ID"), nil)
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
				if v.BusinessRootID == businessId {
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
				ProfileID:      profUUID,
				BusinessRootID: businessUUID,
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
			MemberID:       dbMember.ID.String(),
			BusinessRootID: dbMember.BusinessRootID.String(),
			Role:           dbMember.Role,
		}, o.ttl)

		// 4) set context dan lanjut
		ctx := context.WithValue(r.Context(), OwnedBusinessContextKey, &OwnedBusinessContext{
			MemberID:       dbMember.ID.String(),
			BusinessRootID: dbMember.BusinessRootID.String(),
			Role:           dbMember.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type OwnedBusinessContext struct {
	MemberID       string                    `json:"memberId"`
	BusinessRootID string                    `json:"businessRootId"`
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
