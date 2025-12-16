// internal/repository/redis/email_limiter_repository/viewmodel.go
package email_limiter_repository

type LimiterEmailResponse struct {
	Email             string
	RetryAfterSeconds int64
}
