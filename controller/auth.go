package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/lupppig/stream-ledger-api/repository/kafka"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/lupppig/stream-ledger-api/utils"
)

type Router struct {
	DB   *postgres.PostgresDB
	Prod *kafka.Producer
}

func (ru *Router) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Password  string `json:"password"`
		Email     string `json:"email"`
	}
	if err := utils.ReadJSONRequest(r, &user); err != nil {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid request sent", nil, err.Error(), nil)
		resp.BadResponse(w)
		return
	}

	if !utils.CheckValidEmail(user.Email) {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid email provided", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	if !utils.CheckValidPassword(user.Password) {
		resp := utils.BuildResponse(http.StatusBadRequest, "password field cannot be empty", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	usre := &model.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  user.Password,
		Email:     user.Email,
	}

	err := usre.CreateUser(ru.DB)
	if err != nil {
		if errors.Is(err, postgres.ErrorDuplicateEmail) {
			resp := utils.BuildResponse(http.StatusBadRequest, "email already exists", nil, nil, nil)
			resp.BadResponse(w)
			return
		}
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	resp, err := authResponse(user.FirstName, user.LastName, user.Email, usre.ID)
	if err != nil {
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	rsp := utils.BuildResponse(http.StatusCreated, "user account created successfully", resp, nil, nil)
	rsp.SuccessResponse(w)
}

func (ru *Router) SignIn(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := utils.ReadJSONRequest(r, &user); err != nil {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid request sent", nil, err, nil)
		resp.BadResponse(w)
		return
	}

	if !utils.CheckValidEmail(user.Email) {
		resp := utils.BuildResponse(http.StatusBadRequest, "invalid email provided", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	if !utils.CheckValidPassword(user.Password) {
		resp := utils.BuildResponse(http.StatusBadRequest, "password field cannot be empty", nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	usr := &model.User{}

	err := usr.GetUser(ru.DB, user.Email, user.Password)

	if err != nil {
		resp := utils.BuildResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil, err.Error(), nil)
		resp.BadResponse(w)
		return
	}

	rsp, err := authResponse(usr.FirstName, usr.LastName, usr.Email, usr.ID)
	if err != nil {
		resp := utils.BuildResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil, nil, nil)
		resp.BadResponse(w)
		return
	}

	resp := utils.BuildResponse(http.StatusOK, "login successful", rsp, nil, nil)
	resp.BadResponse(w)
}

func authResponse(firstName, lastName, email string, id int64) (interface{}, error) {
	duration := time.Minute * 30
	token, err := utils.CreateToken(id, duration)
	if err != nil {
		return nil, err
	}

	resp := struct {
		ID          int64  `json:"id"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		AccessToken struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expires_at"`
		} `json:"access_token"`
	}{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		AccessToken: struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expires_at"`
		}{
			Token:     token,
			ExpiresAt: time.Now().Add(duration).UnixNano(),
		},
	}

	return resp, nil
}
