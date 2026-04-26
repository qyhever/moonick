package response

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UserProfile struct {
	ID            int64  `json:"id"`
	Phone         string `json:"phone"`
	Nickname      string `json:"nickname"`
	AvatarURL     string `json:"avatarUrl"`
	Status        string `json:"status"`
	DefaultWechat string `json:"defaultWechat"`
	DefaultPhone  string `json:"defaultPhone"`
}

type AdminProfile struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type AuthPayload struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *UserProfile  `json:"user,omitempty"`
	Admin        *AdminProfile `json:"admin,omitempty"`
}
