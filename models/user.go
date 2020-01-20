package models

import "time"

//User ...
type User struct {
	ID                string    `db:"id, primarykey" json:"id"`
	EncryptedPassword string    `db:"encryped_password" json:"encryped_password"`
	Data              Jsonb     `db:"data" json:"data"`
	Email             string    `db:"email" json:"email"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at" pg:",null"`
	CreatedAt         time.Time `db:"created_at" json:"created_at" pg:",null"`
}
