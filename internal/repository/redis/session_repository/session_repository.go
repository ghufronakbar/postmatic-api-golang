// internal/repository/redis/redis_session/session_repository.go
package session_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

// 1. SAVE SESSION (Create)
func (r *SessionRepository) SaveSession(ctx context.Context, session RedisSession, ttl time.Duration) error {
	key := r.constructKey(session.ProfileID, session.ID)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err := r.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}
	return nil
}

// 2. GET ALL SESSIONS BY PROFILE ID
func (r *SessionRepository) GetSessionsByProfileID(ctx context.Context, profileID string) ([]RedisSession, error) {
	var sessions []RedisSession

	// Gunakan pattern matching untuk mencari semua key milik profile ini
	// session:profile_id:*
	pattern := fmt.Sprintf("session:%s:*", profileID)

	// Gunakan SCAN bukan KEYS agar tidak memblokir server Redis jika data jutaan
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	// Jika tidak ada session, return kosong
	if len(keys) == 0 {
		return sessions, nil
	}

	// Gunakan MGET (Multi Get) untuk mengambil value sekaligus (lebih cepat dari loop GET)
	values, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON ke Struct
	for _, val := range values {
		// MGet bisa return nil jika key expired saat proses berjalan
		if val == nil {
			continue
		}

		var sess RedisSession
		// Value dari Redis berupa string/interface{}, perlu assert ke string
		if strVal, ok := val.(string); ok {
			if err := json.Unmarshal([]byte(strVal), &sess); err == nil {
				sessions = append(sessions, sess)
			}
		}
	}

	return sessions, nil
}

// 3. DELETE SESSION BY ID
func (r *SessionRepository) DeleteSessionByID(ctx context.Context, profileID, sessionID string) error {
	key := r.constructKey(profileID, sessionID)
	return r.rdb.Del(ctx, key).Err()
}

// 4. DELETE ALL SESSIONS BY PROFILE ID (Logout All Devices)
func (r *SessionRepository) DeleteAllSessions(ctx context.Context, profileID string) error {
	pattern := fmt.Sprintf("session:%s:*", profileID)

	// Cari semua key
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) == 0 {
		return nil
	}

	// Hapus sekaligus (Bulk Delete)
	return r.rdb.Del(ctx, keys...).Err()
}

// 5. DELETE BY REFRESH TOKEN (Logic: Search -> Delete)
// Kita butuh profileID karena JWT Refresh Token mengandung profileID.
// Tanpa profileID, kita harus scan SELURUH Redis (sangat berat).
func (r *SessionRepository) DeleteByRefreshToken(ctx context.Context, profileID, refreshToken string) error {
	// 1. Ambil semua session user ini
	sessions, err := r.GetSessionsByProfileID(ctx, profileID)
	if err != nil {
		return err
	}

	// 2. Cari session yang refresh token-nya cocok
	var targetSessionID string
	for _, sess := range sessions {
		if sess.RefreshToken == refreshToken {
			targetSessionID = sess.ID
			break
		}
	}

	// 3. Jika ketemu, hapus by ID
	if targetSessionID != "" {
		return r.DeleteSessionByID(ctx, profileID, targetSessionID)
	}

	// Jika tidak ketemu, anggap sukses (idempotent) atau return error
	return nil
}

// 6. GET SESSION BY REFRESH TOKEN
// Kita WAJIB meminta profileID agar pencarian efisien (hanya scan milik user tsb).
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, profileID, refreshToken string) (*RedisSession, error) {
	// 1. Ambil semua session milik user ini (Reuse function yg sudah ada)
	sessions, err := r.GetSessionsByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// 2. Loop di memory Go (cepat karena session per user biasanya sedikit, misal < 10)
	for _, sess := range sessions {
		if sess.RefreshToken == refreshToken {
			// Ketemu! Return pointer session
			return &sess, nil
		}
	}

	// 3. Tidak ketemu
	return nil, nil
}

// Helper Private
func (r *SessionRepository) constructKey(profileID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", profileID, sessionID)
}
