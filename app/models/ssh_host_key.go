package models

import "time"

// SshHostKey is an approved SSH server identity shared by all render-engine nodes.
type SshHostKey struct {
	ID          string     `json:"id"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	KeyType     string     `json:"key_type"`
	PublicKey   string     `json:"public_key"`
	Fingerprint string     `json:"fingerprint"`
	ApprovedBy  *string    `json:"approved_by"`
	RevokedBy   *string    `json:"revoked_by"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (SshHostKey) TableName() string {
	return "ssh_host_keys"
}
