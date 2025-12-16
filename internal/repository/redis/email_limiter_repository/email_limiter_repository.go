// internal/repository/redis/email_limiter_repository/email_limiter_repository.go
package email_limiter_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type LimiterEmailRepo struct {
	rdb *redis.Client
}

func NewLimiterEmailRepository(rdb *redis.Client) *LimiterEmailRepo {
	return &LimiterEmailRepo{rdb: rdb}
}

// 1. SAVE LIMITER (Set Key dengan Value simple & TTL)
func (r *LimiterEmailRepo) SaveLimiterEmail(ctx context.Context, email string, ttl time.Duration) error {
	key := r.constructKey(email)

	// Kita tidak perlu simpan JSON ribet, cukup string "1" atau emailnya saja.
	// Karena validasi kita hanya berdasarkan "Apakah Key ini ada?"
	if err := r.rdb.Set(ctx, key, email, ttl).Err(); err != nil {
		return err
	}
	return nil
}

// 2. GET LIMITER (Cek TTL)
func (r *LimiterEmailRepo) GetLimiterEmail(ctx context.Context, email string) (*LimiterEmailResponse, error) {
	key := r.constructKey(email)

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
	return &LimiterEmailResponse{
		Email:             email,
		RetryAfterSeconds: int64(duration.Seconds()),
	}, nil
}

func (r *LimiterEmailRepo) constructKey(email string) string {
	return fmt.Sprintf("limiter_email:%s", email)
}
