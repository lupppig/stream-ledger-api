package model

import (
	"context"
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
	Wallet        *Wallet   `bun:"rel:has-one,join:id=user_id"`
}

type Wallet struct {
	ID        int64     `bun:",pk,autoincrement"`
	UserID    int64     `bun:",unique,notnull"`    // each wallet belongs to a single user
	Balance   int64     `bun:",notnull,default:0"` // balance in kobo
	Currency  string    `bun:",notnull,default:NGN"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

func (u *User) CreateUser(db *postgres.PostgresDB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		u.Password = utils.HashPassword(u.Password)
		if err := db.CreateModel(u); err != nil {
			return err
		}
		wallet := &Wallet{
			ID:      u.ID,
			Balance: 0,
		}
		if err := db.CreateModel(wallet); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (u *User) GetUser(db *postgres.PostgresDB, email string, password string) error {
	query := ` email = ?`
	if err := db.SelectSingleEntity(query, u, email); err != nil {
		log.Println(err.Error())
		return fmt.Errorf("invalid email or password provided")
	}

	if !utils.ComparePassword(password, u.Password) {
		return fmt.Errorf("invalid password or email provided")
	}
	return nil
}
