package main

import (
	"testing"

	"github.com/ruilisi/qiyetalk-server-go/db"
)

// TestUser ...
func TestUser(m *testing.T) {
	Init()
	var users []User
	err := db.Model(&users).Select()
	if err != nil {
		panic(err)
	}
	fmt.Printn(users)
}

// TestMain
func TestMain(m *testing.T) {
	println("shit")
}
