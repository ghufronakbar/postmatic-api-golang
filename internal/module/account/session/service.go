package session

import (
	"context"
	"errors"

	"postmatic-api/internal/module/headless/token"
	sessRepo "postmatic-api/internal/repository/redis/session_repository"
	"postmatic-api/pkg/errs"

	"github.com/google/uuid"
)

type SessionService struct {
	sessionRepo *sessRepo.SessionRepository
	tm          token.TokenMaker
}

func NewService(sessionRepo *sessRepo.SessionRepository, tm token.TokenMaker) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		tm:          tm,
	}
}

func (s *SessionService) Logout(ctx context.Context, input LogoutInput, profileId uuid.UUID) (*sessRepo.RedisSession, error) {
	sessionId := input.SessionID
	sess, err := s.sessionRepo.GetSessionsByProfileID(ctx, profileId)
	if err != nil {
		return nil, errs.NewInternalServerError(errors.New("FAILED_TO_LOG_OUT"))
	}

	var check *sessRepo.RedisSession
	for _, session := range sess {
		if session.ID == sessionId {
			check = &session
			break
		}
	}

	if check == nil {
		return nil, errs.NewUnauthorized("INVALID_SESSION_ID")
	}

	err = s.sessionRepo.DeleteSessionByID(ctx, profileId, sessionId)
	if err != nil {
		return nil, errs.NewInternalServerError(errors.New("FAILED_TO_LOG_OUT"))
	}
	// TODO: web socket to notify client and force logout
	return check, nil
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

	decoded, err := s.tm.DecodeTokenWithoutVerify(refreshToken)

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

	session := sessRepo.RedisSession{
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

func (s *SessionService) GetAllSession(ctx context.Context, profileId uuid.UUID) ([]*SessionListResponse, error) {
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
