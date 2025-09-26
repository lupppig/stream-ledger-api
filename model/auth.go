package model

import (
	"fmt"
	"time"

	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/lupppig/stream-ledger-api/utils"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`
	ID            int64 `bun:",pk,autoincrement"`
	FirstName     string
	LastName      string
	Email         string `bun:",unique"`
	Password      string
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:",nullzero,default:current_timestamp,updated_at"`
}

func (u *User) CreateUser(db *postgres.PostgresDB) error {
	u.Password = utils.HashPassword(u.Password)
	if err := db.CreateModel(u); err != nil {
		return err
	}
	return nil
}

func (u *User) GetUser(db *postgres.PostgresDB, email string, password string) error {
	query := `where email = ?`
	if err := db.SelectSingleEntity(query, email, u); err != nil {
		return fmt.Errorf("invalid email or password provided")
	}

	if !utils.ComparePassword(password, u.Password) {
		return fmt.Errorf("invalid password or email provided")
	}
	return nil
}
