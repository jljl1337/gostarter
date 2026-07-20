package repository

type Migration struct {
	ID            string `json:"id" db:"id"`
	UpStatement   string `json:"upStatement" db:"up_statement"`
	DownStatement string `json:"downStatement" db:"down_statement"`
	ExecutedAt    string `json:"executedAt" db:"executed_at"`
}

type Account struct {
	ID           string `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:"passwordHash" db:"password_hash"`
	Role         string `json:"role" db:"role"`
	LanguageCode string `json:"languageCode" db:"language_code"`
	CreatedAt    string `json:"createdAt" db:"created_at"`
	UpdatedAt    string `json:"updatedAt" db:"updated_at"`
}

type Session struct {
	ID        string  `json:"id" db:"id"`
	AccountID *string `json:"accountID" db:"account_id"`
	Token     string  `json:"token" db:"token"`
	CsrfToken string  `json:"csrfToken" db:"csrf_token"`
	ExpiresAt string  `json:"expiresAt" db:"expires_at"`
	CreatedAt string  `json:"createdAt" db:"created_at"`
	UpdatedAt string  `json:"updatedAt" db:"updated_at"`
}
