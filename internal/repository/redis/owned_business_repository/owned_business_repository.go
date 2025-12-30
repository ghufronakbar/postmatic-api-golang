// internal/repository/redis/owned_business_repository/owned_business_repository.go
package owned_business_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OwnedBusinessRepository struct {
	rdb *redis.Client
}

func NewOwnedBusinessRepository(rdb *redis.Client) *OwnedBusinessRepository {
	return &OwnedBusinessRepository{rdb: rdb}
}

// key memiliki banyak RedisOwnedBusinessResponse[]

// 1) Save all (overwrite)
func (r *OwnedBusinessRepository) SaveOwnedBusiness(ctx context.Context, profileID string, businessSubs []RedisBusinessSub, ttl time.Duration) error {
	key := r.constructKey(profileID)

	data, err := json.Marshal(businessSubs)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, data, ttl).Err()
}

// 2) Get all
func (r *OwnedBusinessRepository) GetOwnedBusinessByProfileID(ctx context.Context, profileID string) ([]RedisOwnedBusinessResponse, error) {
	key := r.constructKey(profileID)

	res, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return []RedisOwnedBusinessResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	var ownedBusiness []RedisOwnedBusinessResponse
	if err := json.Unmarshal([]byte(res), &ownedBusiness); err != nil {
		return nil, err
	}

	// normalize biar gak nil
	if ownedBusiness == nil {
		ownedBusiness = []RedisOwnedBusinessResponse{}
	}
	return ownedBusiness, nil
}

// 3) Append / Upsert 1 business (update kalau sudah ada, append kalau belum)
// - Tidak replace seluruh list secara “semantik”, tapi tetap SET seluruh JSON (karena value string JSON).
func (r *OwnedBusinessRepository) UpsertOneBusiness(ctx context.Context, profileID string, b RedisBusinessSub, defaultTTL time.Duration) error {
	key := r.constructKey(profileID)

	list, err := r.GetOwnedBusinessByProfileID(ctx, profileID)
	if err != nil {
		return err
	}

	updated := false
	for i := range list {
		if list[i].BusinessRootID == b.BusinessRootID {
			list[i].MemberID = b.MemberID
			list[i].Role = b.Role
			updated = true
			break
		}
	}

	if !updated {
		// ✅ staticcheck S1016 friendly
		list = append(list, RedisOwnedBusinessResponse(b))
	}

	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	ttl := r.getExistingTTL(ctx, key, defaultTTL)
	return r.rdb.Set(ctx, key, data, ttl).Err()
}

// 4) Delete 1 business tertentu dari list cache
func (r *OwnedBusinessRepository) DeleteOneBusiness(ctx context.Context, profileID string, businessRootID int64, defaultTTL time.Duration) error {
	key := r.constructKey(profileID)

	list, err := r.GetOwnedBusinessByProfileID(ctx, profileID)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}

	// filter out
	newList := make([]RedisOwnedBusinessResponse, 0, len(list))
	found := false
	for _, it := range list {
		if it.BusinessRootID == businessRootID {
			found = true
			continue
		}
		newList = append(newList, it)
	}

	if !found {
		return nil
	}

	// kalau habis, delete key
	if len(newList) == 0 {
		return r.rdb.Del(ctx, key).Err()
	}

	data, err := json.Marshal(newList)
	if err != nil {
		return err
	}

	ttl := r.getExistingTTL(ctx, key, defaultTTL)
	return r.rdb.Set(ctx, key, data, ttl).Err()
}

// Helper: construct key
func (r *OwnedBusinessRepository) constructKey(profileID string) string {
	return fmt.Sprintf("owned_business:%s", profileID)
}

// Helper: preserve TTL.
// - kalau key ada dan punya TTL > 0 => pakai itu
// - kalau key persistent (-1) => TTL=0 (no expiration)
// - kalau key missing (-2) => pakai defaultTTL
func (r *OwnedBusinessRepository) getExistingTTL(ctx context.Context, key string, defaultTTL time.Duration) time.Duration {
	ttl, err := r.rdb.TTL(ctx, key).Result()
	if err != nil {
		return defaultTTL
	}

	// go-redis:
	// -2 = key doesn't exist
	// -1 = key exists but has no associated expire
	if ttl == -2*time.Second {
		return defaultTTL
	}
	if ttl == -1*time.Second {
		return 0
	}
	return ttl
}
