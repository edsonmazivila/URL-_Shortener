package domain

import "time"

// URL represents a shortened URL entity in the system.
type URL struct {
	ID           int64      `json:"id"`
	ShortCode    string     `json:"short_code"`
	OriginalURL  string     `json:"original_url"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	AccessCount  int64      `json:"access_count"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
}

// IsExpired checks if the URL has expired.
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// IncrementAccessCount increments the access counter.
func (u *URL) IncrementAccessCount() {
	u.AccessCount++
	now := time.Now()
	u.LastAccessed = &now
}
