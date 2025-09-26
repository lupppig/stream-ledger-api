package model

import (
	"fmt"
	"log"
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
}

func (u *User) CreateUser(db *postgres.PostgresDB) error {
	u.Password = utils.HashPassword(u.Password)
	if err := db.CreateModel(u); err != nil {
		return err
	}
	return nil
}

func (u *User) GetUser(db *postgres.PostgresDB, email string, password string) error {
	query := ` email = ?`
	if err := db.SelectSingleEntity(query, u, email); err != nil {
		log.Println(err.Error())
		return fmt.Errorf("invalid email or password provided")
	}

	fmt.Println(password, u.Password)
	if !utils.ComparePassword(password, u.Password) {
		return fmt.Errorf("invalid password or email provided")
	}
	return nil
}
