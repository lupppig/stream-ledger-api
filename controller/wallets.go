package controller

import (
	"net/http"

	"github.com/lupppig/stream-ledger-api/controller/middleware"
	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/utils"
)

func (ru *Router) GetWallet(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(middleware.ContextKeyUserID).(int64)
	if !ok {
		resp := utils.BuildResponse(http.StatusUnauthorized, "unauthorized user", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	var user = &model.User{}

	if err := user.GetWallet(ru.DB, id); err != nil {
		resp := utils.BuildResponse(http.StatusNotFound, "user wallet not found", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	resp := utils.BuildResponse(http.StatusOK, "user info", user, nil, nil)
	resp.SuccessResponse(w)
}
