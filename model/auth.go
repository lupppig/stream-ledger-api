package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/lupppig/stream-ledger-api/utils"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`
	ID            int64     `bun:",pk,autoincrement" json:"user_id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `bun:",unique" json:"email"`
	Password      string    `json:"-"`
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	Wallet        *Wallet   `bun:"rel:has-one,join:id=user_id" json:"wallet"`
}

type Wallet struct {
	ID        int64     `bun:",pk,autoincrement" json:"wallet_id"`
	UserID    int64     `bun:",unique,notnull" json:"-"`          // each wallet belongs to a single user
	Balance   int64     `bun:",notnull,default:0" json:"balance"` // balance in kobo
	Currency  string    `bun:",notnull,default:'NGN'" json:"currency"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:",notnull,default:current_timestamp" json:"updated_at"`

	User         *User          `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Transactions []*Transaction `bun:"rel:has-many" json:"-"`
}

func (u *User) CreateUser(db *postgres.PostgresDB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := db.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		u.Password = utils.HashPassword(u.Password)
		if _, err := tx.NewInsert().Model(u).Returning("*").Exec(ctx); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					return postgres.ErrorDuplicateEmail
				}
			}
			return err
		}
		wallet := &Wallet{
			UserID:  u.ID,
			Balance: 0,
		}
		if _, err := tx.NewInsert().Model(wallet).Exec(ctx); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUser(db *postgres.PostgresDB, email string, password string) error {
	query := ` email = ?`
	if err := db.SelectSingleEntity(query, u, email); err != nil {
		return fmt.Errorf("invalid email or password provided")
	}

	if !utils.ComparePassword(password, u.Password) {
		return fmt.Errorf("invalid password or email provided")
	}
	return nil
}

func (u *User) GetWallet(db *postgres.PostgresDB, id int64) error {
	query := `"user".id = ?`
	if err := db.SelectOneWithRelation(query, "Wallet", u, id); err != nil {
		return err
	}
	return nil
}

func (w *Wallet) getWallet(tx bun.Tx, userId int64) error {
	query := "user_id = ?"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return tx.NewSelect().Model(w).Where(query, userId).Scan(ctx)
}

func GetUser(db *postgres.PostgresDB, id int64) (*User, error) {
	query := `id = ?`
	var user = &User{}
	if err := db.SelectSingleEntity(query, user, id); err != nil {
		return nil, err
	}
	return user, nil
}
