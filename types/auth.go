package types

type FirebaseLoginResponse struct {
	IDToken      string `json:"idToken"`
	UID          string `json:"localId"`
	RefreshToken string `json:"refreshToken"`
}

type FirebaseRegisterResponse struct {
	IDToken      string `json:"iDToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	UID          string `json:"localId"`
}
