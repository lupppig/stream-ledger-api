package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	services "github.com/lupppig/stream-ledger-api/service"
	"github.com/riverqueue/river"
)

type ExportTransactionsArgs struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

func (ExportTransactionsArgs) Kind() string {
	return "export_transactions"
}

type ExportTransactionsWorker struct {
	river.WorkerDefaults[ExportTransactionsArgs]
	DB *postgres.PostgresDB
}

func (w *ExportTransactionsWorker) Work(ctx context.Context, job *river.Job[ExportTransactionsArgs]) error {
	args := job.Args

	log.Printf("Starting transaction export for user %d", args.UserID)

	transactions, err := model.FetchUserTransactions(context.Background(), w.DB, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	filePath, err := services.GenerateExcel(transactions, args.UserID)
	if err != nil {
		return fmt.Errorf("failed to generate Excel report: %w", err)
	}

	log.Printf("Transaction export completed for user %d: %s", args.UserID, filePath)
	return nil
}
