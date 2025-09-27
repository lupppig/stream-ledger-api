package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/xuri/excelize/v2"
)

func GenerateExcel(transactions []model.Transaction, userId int64) (string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "user_transaction_report"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headers := []string{"ID", "Entry", "Amount", "CreatedAt"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)

		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		f.SetCellStyle(sheet, cell, cell, style)
	}

	for row, t := range transactions {
		rowNum := row + 2

		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), t.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), t.Entry)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), t.Amount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), t.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	cols := []string{"A", "B", "C", "D"}
	for _, col := range cols {
		if err := f.SetColWidth(sheet, col, col, 15); err != nil {
			// Log but don't fail on width setting
			log.Printf("Warning: failed to set column width for %s: %v", col, err)
		}
	}

	exportDir := "tmp"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(exportDir, fmt.Sprintf("transactions_user_%d_%s.xlsx", userId, timestamp))

	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filePath, nil
}
