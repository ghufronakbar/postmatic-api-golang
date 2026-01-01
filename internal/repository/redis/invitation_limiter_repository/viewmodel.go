// internal/repository/redis/invitation_limiter_repository/viewmodel.go
package invitation_limiter_repository

type LimiterInvitationResponse struct {
	Email             string
	RetryAfterSeconds int64
	BusinessRootID    int64
}
