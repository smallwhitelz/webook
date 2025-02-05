package web

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginJWTReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EditReq struct {
	Nickname string `json:"nickname"`
	// YYYY-MM-DD
	Birthday    string `json:"birthday"`
	Description string `json:"description"`
}
