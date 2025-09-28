package controller

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/lupppig/stream-ledger-api/controller/middleware"
	"github.com/lupppig/stream-ledger-api/jobs"
	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/repository/kafka"
	"github.com/lupppig/stream-ledger-api/utils"
)

func (ru *Router) CreateTransactions(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(middleware.ContextKeyUserID).(int64)
	if !ok {
		resp := utils.BuildResponse(http.StatusUnauthorized, "unauthorized user", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	var transaction struct {
		Entry   string `json:"entry"`
		Amount  int    `json:"amount"`
		TransID string `json:"trans_id"`
	}
	if err := utils.ReadJSONRequest(r, &transaction); err != nil {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid request sent", nil, err.Error(), nil)
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
			resp := utils.BuildResponse(http.StatusBadRequest, "cannot debit wallet: balance is too low", nil, err.Error(), nil)
			resp.BadResponse(w)
			return
		}
		if errors.Is(err, model.ErrorDuplicateTransaction) {
			resp := utils.BuildResponse(http.StatusConflict, "duplicate transaction entry", nil, err.Error(), nil)
			resp.BadResponse(w)
			return
		}

		log.Println(err.Error())
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	trans := kafka.TransactionEvent{
		UserID:  id,
		Entry:   trx.Entry,
		Amount:  trx.Amount,
		Balance: trx.Wallet.Balance,
	}

	go ru.Prod.PublishTransaction(trans)

	var resData = struct {
		TransactionID int64  `json:"transaction_id"`
		WalletID      int64  `json:"wallet_id"`
		Entry         string `json:"entry"`
		Amount        int64  `json:"amount"`
		TransID       string `json:"trans_id"`
	}{
		TransactionID: trx.ID,
		WalletID:      trx.WalletID,
		Entry:         trx.Entry,
		Amount:        trx.Amount,
		TransID:       trx.TransID,
	}
	resp := utils.BuildResponse(http.StatusOK, "transaction successfully", resData, nil, nil)
	resp.SuccessResponse(w)
}

func (ru *Router) ListUserTransactions(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(middleware.ContextKeyUserID).(int64)
	if !ok {
		resp := utils.BuildResponse(http.StatusUnauthorized, "unauthorized user", nil, nil, nil)
		resp.BadResponse(w)
		return
	}
	pagination := utils.GetPagination(r)
	trx := model.Transaction{}
	transactions, total, err := trx.GetUserTransaction(ru.DB, id, pagination)

	if err != nil {
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	q := r.URL.Query()
	nextPage := ""
	prevPage := ""

	if pagination.Offset+pagination.Limit < total {
		q.Set("page", fmt.Sprintf("%d", pagination.Page+1))
		nextPage = fmt.Sprintf("%s?%s", r.URL.Path, q.Encode())
	}

	if pagination.Page > 1 {
		q.Set("page", fmt.Sprintf("%d", pagination.Page-1))
		prevPage = fmt.Sprintf("%s?%s", r.URL.Path, q.Encode())
	}

	var pagin = struct {
		Page     int    `json:"page"`
		Limit    int    `limit:"limit"`
		Total    int    `total:"total"`
		NextPage string `json:"next_page"`
		PrevPage string `json:"prev_page"`
	}{
		Page:     pagination.Page,
		Limit:    pagination.Limit,
		Total:    total,
		NextPage: nextPage,
		PrevPage: prevPage,
	}

	resp := utils.BuildResponse(http.StatusOK, "user transactions list", transactions, nil, pagin)
	resp.SuccessResponse(w)
}

func (ru *Router) ExportTransaction(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(middleware.ContextKeyUserID).(int64)
	if !ok {
		resp := utils.BuildResponse(http.StatusUnauthorized, "unauthorized user", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	user, err := model.GetUser(ru.DB, id)
	if err != nil {
		log.Println(err.Error())
		resp := utils.BuildResponse(http.StatusNotFound, "user not found", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	jbArg := jobs.ExportTransactionsArgs{
		UserID: user.ID,
		Email:  user.Email,
	}

	_, err = ru.DB.River.Insert(r.Context(), jbArg, nil)

	if err != nil {
		log.Println(err.Error())
		resp := utils.BuildResponse(http.StatusInternalServerError, "failed to queue export job", nil, nil, nil)
		resp.BadResponse(w)
	}

	rsp := utils.BuildResponse(http.StatusOK, "your transaction data has been successfully exported to Excel", nil, nil, nil)
	rsp.SuccessResponse(w)
}
