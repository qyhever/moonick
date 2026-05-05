package request

const (
	VerificationCodeTypeRegister      = "register"
	VerificationCodeTypeResetPassword = "reset_password"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required"`
	Type  string `json:"type" binding:"required"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
}

type UpdateContactRequest struct {
	DefaultWechat string `json:"defaultWechat"`
	DefaultPhone  string `json:"defaultPhone"`
}
