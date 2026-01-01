// internal/repository/redis/invitation_limiter_repository/invitation_limiter_repository.go
package invitation_limiter_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type LimiterInvitationRepo struct {
	rdb *redis.Client
}

func NewLimiterInvitationRepository(rdb *redis.Client) *LimiterInvitationRepo {
	return &LimiterInvitationRepo{rdb: rdb}
}

// 1. SAVE LIMITER (Set Key dengan Value simple & TTL)
func (r *LimiterInvitationRepo) SaveLimiterInvitation(ctx context.Context, input LimiterInvitationInput, ttl time.Duration) error {
	key := r.constructKey(input)

	// Save key with value simple & TTL
	if err := r.rdb.Set(ctx, key, input.Email, ttl).Err(); err != nil {
		return err
	}
	return nil
}

// 2. GET LIMITER (Cek TTL)
func (r *LimiterInvitationRepo) GetLimiterInvitation(ctx context.Context, input LimiterInvitationInput) (*LimiterInvitationResponse, error) {
	key := r.constructKey(input)

	// 1. Cek apakah key ada? (Gunakan TTL command sekalian)
	// TTL returns:
	// -2: Key tidak ada (Aman, boleh kirim email)
	// -1: Key ada tapi tidak punya expiry (Tidak mungkin terjadi di logic kita)
	// >0: Sisa waktu detik (Rate limit aktif)
	duration, err := r.rdb.TTL(ctx, key).Result()

	if err != nil {
		return nil, err
	}

	// Jika -2, berarti key expired atau tidak ada -> User lolos rate limit
	if duration == -2 {
		return nil, nil
	}

	// Jika > 0, berarti Rate Limit AKTIF
	return &LimiterInvitationResponse{
		Email:             input.Email,
		RetryAfterSeconds: int64(duration.Seconds()),
		BusinessRootID:    input.BusinessRootID,
	}, nil
}

func (r *LimiterInvitationRepo) constructKey(input LimiterInvitationInput) string {
	return fmt.Sprintf("limiter_invitation:%s:%d", input.Email, input.BusinessRootID)
}
