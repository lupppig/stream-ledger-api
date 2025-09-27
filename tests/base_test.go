package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/uptrace/bun"
)

var pdb *postgres.PostgresDB

func SetupTestDB(m *testing.M) {
	dsn := os.Getenv("DB_TEST_URL")
	if dsn == "" {
		dsn = "postgres://test:secret@localhost:5433/streamledger_test?sslmode=disable"
	}
	var TestDB *bun.DB
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

	_, _ = TestDB.NewCreateTable().Model((*model.User)(nil)).IfNotExists().Exec(ctx)
	_, _ = TestDB.NewCreateTable().Model((*model.Wallet)(nil)).IfNotExists().Exec(ctx)
	_, _ = TestDB.NewCreateTable().Model((*model.Transaction)(nil)).IfNotExists().Exec(ctx)
	log.Println("✅ Test DB setup complete")

	code := m.Run()
	os.Exit(code)
}
