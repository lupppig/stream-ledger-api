package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/IBM/sarama/mocks"
	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/uptrace/bun"
)

func SetupTestDB(t *testing.T) (*postgres.PostgresDB, *mocks.AsyncProducer) {
	dsn := os.Getenv("DB_TEST_URL")
	if dsn == "" {
		dsn = "postgres://user:mypassword@localhost:5044/mydbtest?sslmode=disable"
	}
	var TestDB *bun.DB
	var pdb *postgres.PostgresDB
	var err error
	pdb, err = postgres.Connect(dsn)
	if err != nil {
		log.Fatal(err.Error())
	}

	TestDB = pdb.DB
	ctx := context.Background()
	// Drop & recreate tables before running tests
	_, _ = TestDB.NewDropTable().Model((*model.Transaction)(nil)).IfExists().Cascade().Exec(ctx)
	_, _ = TestDB.NewDropTable().Model((*model.Wallet)(nil)).IfExists().Cascade().Exec(ctx)
	_, _ = TestDB.NewDropTable().Model((*model.User)(nil)).IfExists().Cascade().Exec(ctx)
	_, _ = TestDB.ExecContext(ctx, `DROP TYPE IF EXISTS transaction_entry CASCADE;`)

	_, _ = TestDB.ExecContext(ctx, `CREATE TYPE  transaction_entry AS ENUM ('credit', 'debit');`)
	_, _ = TestDB.NewCreateTable().Model((*model.User)(nil)).IfNotExists().Exec(ctx)
	_, _ = TestDB.NewCreateTable().Model((*model.Wallet)(nil)).IfNotExists().Exec(ctx)
	_, err = TestDB.NewCreateTable().Model((*model.Transaction)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Println(err.Error())
	}
	mockProducer := mocks.NewAsyncProducer(t, nil)
	// defer mockProducer.Close()

	for i := 0; i < 100; i++ {
		mockProducer.ExpectInputAndSucceed()
	}

	log.Println("Test DB setup complete")

	return pdb, mockProducer
}
