// internal/repository/redis/invitation_limiter_repository/dto.go
package invitation_limiter_repository

type LimiterInvitationInput struct {
	Email          string
	BusinessRootID int64
}
