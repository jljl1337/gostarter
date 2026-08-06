package repository

type Note struct {
	ID         string `json:"id" db:"id"`
	AccountID  string `json:"accountID" db:"account_id"`
	Body       string `json:"body" db:"body"`
	Positivity int    `json:"positivity" db:"positivity"`
	CreatedAt  string `json:"createdAt" db:"created_at"`
	UpdatedAt  string `json:"updatedAt" db:"updated_at"`
}
