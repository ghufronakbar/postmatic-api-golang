// internal/module/headless/mailer/dto_member.go
package mailer

// MEMBER
// MEMBER INVITATION EMAIL
type MemberInvitationInputDTO struct {
	Email        string `json:"Email"`
	ConfirmUrl   string `json:"ConfirmUrl"`
	BusinessName string `json:"BusinessName"`
}

// MEMBER ANNOUNCE ROLE EMAIL
type MemberAnnounceRoleInputDTO struct {
	Email        string `json:"Email"`
	BusinessName string `json:"BusinessName"`
	NewRole      string `json:"NewRole"`
}

// MEMBER ANNOUNCE KICK EMAIL
type MemberAnnounceKickInputDTO struct {
	Email        string `json:"Email"`
	BusinessName string `json:"BusinessName"`
}

// MEMBER WELCOME BUSINESS EMAIL
type MemberWelcomeBusinessInputDTO struct {
	// recipient
	Email string `json:"Email"`

	// profile info (nilable)
	ProfileImage *string `json:"ProfileImage,omitempty"`

	// business info
	BusinessName string  `json:"BusinessName"`
	BusinessLogo *string `json:"BusinessLogo,omitempty"`

	// member info
	Role     string  `json:"Role"`
	JoinedAt *string `json:"JoinedAt,omitempty"` // bebas: string RFC3339 / formatted
}
