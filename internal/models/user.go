package models

import "time"

type User struct {
	ID        uint64     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Username  string     `db:"username" json:"username"`
	Password  string     `db:"password" json:"-"`
	LastLogin *time.Time `db:"last_login" json:"last_login"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}
