package models

import (
	"qiyetalk-server-go/db"
	"time"
)

//User ...
type User struct {
	ID                string    `db:"id, primarykey" json:"id"`
	EncryptedPassword string    `db:"encryped_password" json:"encryped_password"`
	Data              Jsonb     `db:"data" json:"data"`
	Email             string    `db:"email" json:"email"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at" pg:",null"`
	CreatedAt         time.Time `db:"created_at" json:"created_at" pg:",null"`
}

// FindByEmail ...
func FindByEmail(email string) *User {
	var users []User
	_db := db.GetDB()
	_db.Model(&users).Where("email = ?", email).Select()
	if len(users) > 0 {
		return &users[0]
	}
	return nil
}
