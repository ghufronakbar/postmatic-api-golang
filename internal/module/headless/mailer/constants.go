// internal/module/headless/mailer/constants.go
package mailer

type EmailTemplate string

const (
	// Template Names
	ResetPasswordTemplate EmailTemplate = "reset_password.html"
	VerificationTemplate  EmailTemplate = "verification.html"
	WelcomeTemplate       EmailTemplate = "welcome.html"

	// Member
	MemberInvitationTemplate   EmailTemplate = "member_invitation.html"
	MemberAnnounceKickTemplate EmailTemplate = "member_announce_kick.html"
	MemberAnnounceRoleTemplate EmailTemplate = "member_announce_role.html"

	// Layout
	LayoutTemplate EmailTemplate = "layout.html"
)

// Validasi Helper
func (e EmailTemplate) IsValid() bool {
	switch e {
	case MemberInvitationTemplate, MemberAnnounceKickTemplate, MemberAnnounceRoleTemplate, ResetPasswordTemplate, VerificationTemplate, WelcomeTemplate:
		return true
	}
	return false
}

// Stringer Interface agar saat diprint muncul string aslinya
func (e EmailTemplate) String() string {
	return string(e)
}
