package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/lupppig/stream-ledger-api/controller/middleware"
	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/utils"
)

func (ru *Router) CreateTransactions(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(middleware.ContextKeyUserID).(int64)
	if !ok {
		resp := utils.BuildResponse(http.StatusUnauthorized, "unauthorized user", nil, nil)
		resp.BadResponse(w)
		return
	}

	var transaction struct {
		Entry   string `json:"entry"`
		Amount  int    `json:"amount"`
		TransID string `json:"trans_id"`
	}
	if err := utils.ReadJSONRequest(r, &transaction); err != nil {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid request sent", nil, err)
		resp.BadResponse(w)
		return
	}

	var trx = &model.Transaction{
		Entry:   transaction.Entry,
		Amount:  int64(transaction.Amount),
		TransID: transaction.TransID,
	}

	if err := trx.CreateTransaction(ru.DB, id); err != nil {
		if errors.Is(err, model.ErrorInsuffcientBalance) {
			resp := utils.BuildResponse(http.StatusBadRequest, "cannot debit wallet: balance is too low", nil, err.Error())
			resp.BadResponse(w)
			return
		}
		if errors.Is(err, model.ErrorDuplicateTransaction) {
			resp := utils.BuildResponse(http.StatusConflict, "duplicate transaction entry", nil, err.Error())
			resp.BadResponse(w)
			return
		}

		log.Println(err.Error())
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil)
		resp.BadResponse(w)
		return
	}

	var resData = struct {
		TransactionID int64  `json:"transaction_id"`
		WalletID      int64  `json:"wallet_id"`
		Entry         string `json:"entry"`
		Amount        int64  `json:"amount"`
		Balance       int64  `json:"balance"`
	}{
		TransactionID: trx.ID,
		WalletID:      trx.WalletID,
		Entry:         trx.Entry,
		Amount:        trx.Amount,
		Balance:       trx.Wallet.Balance,
	}
	resp := utils.BuildResponse(http.StatusOK, "transaction successfully", resData, nil)
	resp.SuccessResponse(w)
}
