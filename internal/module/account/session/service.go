package session

import (
	"context"
	"errors"

	sessionRepository "postmatic-api/internal/repository/redis"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/token"
)

type SessionService struct {
	sessionRepo *sessionRepository.SessionRepository
}

func NewService(sessionRepo *sessionRepository.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *SessionService) Logout(ctx context.Context, input LogoutInput, profileId string) error {
	sessionId := input.SessionID
	err := s.sessionRepo.DeleteSessionByID(ctx, sessionId, profileId)
	if err != nil {
		return errs.NewInternalServerError(errors.New("FAILED_TO_LOG_OUT_OTHER_DEVICE"))
	}
	// TODO: web socket to notify client and force logout
	return nil
}

func (s *SessionService) LogoutAll(ctx context.Context, input LogoutAllInput) error {
	profileId := input.ProfileID
	err := s.sessionRepo.DeleteAllSessions(ctx, profileId)
	if err != nil {
		return errs.NewInternalServerError(errors.New("FAILED_TO_LOG_OUT_ALL_DEVICE"))
	}
	// TODO: web socket to notify client and force logout
	return nil
}

func (s *SessionService) GetSession(ctx context.Context, claims *token.Claims, refreshToken string) (*SessionResponse, error) {

	profileId := claims.ID
	sess, err := s.sessionRepo.GetSessionByRefreshToken(ctx, profileId, refreshToken)
	if err != nil {
		return nil, errs.NewInternalServerError(errors.New("FAILED_TO_GET_SESSION"))
	}
	if sess == nil {
		return nil, errs.NewUnauthorized("")
	}

	decoded, err := token.DecodeTokenWithoutVerify(refreshToken)

	if decoded == nil {
		return nil, errs.NewUnauthorized("INVALID_REFRESH_TOKEN")
	}

	if err != nil {
		return nil, errs.NewUnauthorized("INVALID_REFRESH_TOKEN")
	}

	var imageUrl *string
	if claims.ImageUrl != nil {
		imageUrl = claims.ImageUrl
	}

	session := sessionRepository.RedisSession{
		ID:           sess.ID,
		ProfileID:    profileId,
		RefreshToken: refreshToken,
		Browser:      sess.Browser,
		Device:       sess.Device,
		Platform:     sess.Platform,
		ClientIP:     sess.ClientIP,
		CreatedAt:    sess.CreatedAt,
		ExpiredAt:    decoded.ExpiresAt.Time,
	}

	return &SessionResponse{
		ID:       claims.ID,
		Email:    claims.Email,
		Name:     claims.Name,
		ImageUrl: imageUrl,
		Exp:      claims.ExpiresAt.Unix(),
		Session:  session,
	}, nil
}

func (s *SessionService) GetAllSession(ctx context.Context, profileId string) ([]*SessionListResponse, error) {
	sess, err := s.sessionRepo.GetSessionsByProfileID(ctx, profileId)
	if err != nil {
		return nil, errs.NewInternalServerError(errors.New("FAILED_TO_GET_SESSION_LIST"))
	}

	var sessionList []*SessionListResponse
	for _, session := range sess {
		sessionList = append(sessionList, &SessionListResponse{
			ID:        session.ID,
			ProfileID: session.ProfileID,
			Browser:   session.Browser,
			Device:    session.Device,
			Platform:  session.Platform,
			ClientIP:  session.ClientIP,
			CreatedAt: session.CreatedAt,
			ExpiredAt: session.ExpiredAt,
		})
	}

	return sessionList, nil
}
