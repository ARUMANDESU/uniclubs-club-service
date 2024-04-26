package domain

type Notification struct {
	UserID      int64  `json:"user_id"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Severity    string `json:"severity"`
	Source      string `json:"source"`
	DisplayType string `json:"display_type"`
	CreatedAt   string `json:"created_at"`
	ExpiryAt    string `json:"expiry_at"`
}
