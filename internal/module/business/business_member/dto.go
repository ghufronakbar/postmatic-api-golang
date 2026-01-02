// internal/module/business/business_member/dto.go
package business_member

import "github.com/google/uuid"

type InviteBusinessMemberInput struct {
	Email          string `json:"email" validate:"required,email"`
	Role           string `json:"role" validate:"required,oneof=admin member"`
	BusinessRootID int64  `json:"businessRootId" validate:"required"`
}

type UpdateBusinessMemberInput struct {
	MemberID int64  `json:"memberId" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=admin member"`
}

type ResendEmailInvitationInput struct {
	MemberID       int64     `json:"memberId" validate:"required"`
	ProfileID      uuid.UUID `json:"profileId" validate:"required"`
	BusinessRootID int64     `json:"businessRootId" validate:"required"`
}

type VerifyMemberInvitationInput struct {
	MemberInvitationToken string `json:"memberInvitationToken" validate:"required"`
}

type AnswerMemberInvitationInput struct {
	MemberInvitationToken string `json:"memberInvitationToken" validate:"required"`
	Answer                string `json:"answer" validate:"required,oneof=accept reject"`
}

type RemoveBusinessMemberInput struct {
	MemberID       int64     `json:"memberId" validate:"required"`
	BusinessRootID int64     `json:"businessRootId" validate:"required"`
	ProfileID      uuid.UUID `json:"profileId" validate:"required"`
}

type EditMemberInput struct {
	MemberID       int64     `json:"memberId" validate:"required"`
	BusinessRootID int64     `json:"businessRootId" validate:"required"`
	ProfileID      uuid.UUID `json:"profileId" validate:"required"`
	Role           string    `json:"role" validate:"required,oneof=admin member"`
}
