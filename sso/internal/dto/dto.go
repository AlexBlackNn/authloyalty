package dto

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AlexBlackNn/authloyalty/sso/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/mailru/easyjson"
)

// DTO http and grpc structures.

type Login struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password"`
}

type UserInfo struct {
	FileName string `json:"file_name"`
}

type Register struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
	Avatar   string `json:"avatar"`
}

type Refresh struct {
	Token string `json:"token" validate:"jwt"`
}

type Logout struct {
	Token string `json:"token" validate:"jwt"`
}

// Output http structures.

type Response struct {
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type UserResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Birth  string `json:"birthday,omitempty"`
	Email  string `json:"email,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

const StatusError = "Error"
const StatusSuccess = "Success"

func sendJSON(w http.ResponseWriter, statusCode int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

// Errors.

func ResponseErrorNotFound(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := easyjson.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusNotFound, dataMarshal)
}

func ResponseErrorInternal(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := easyjson.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusInternalServerError, dataMarshal)
}

func ResponseErrorNowAllowed(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := easyjson.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusMethodNotAllowed, dataMarshal)
}

func ResponseErrorStatusConflict(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := easyjson.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusMethodNotAllowed, dataMarshal)
}

func ResponseErrorBadRequest(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := easyjson.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusBadRequest, dataMarshal)
}

// OK.

func ResponseOK(w http.ResponseWriter) {
	dataMarshal, _ := easyjson.Marshal(
		Response{
			Status: StatusSuccess,
		},
	)
	sendJSON(w, http.StatusOK, dataMarshal)
}

func ResponseOKAccessRefresh(
	w http.ResponseWriter,
	userWithTokens *domain.UserWithTokens,
) {
	dataMarshal, _ := easyjson.Marshal(
		Response{
			Status:       StatusSuccess,
			UserID:       userWithTokens.ID,
			AccessToken:  userWithTokens.AccessToken,
			RefreshToken: userWithTokens.RefreshToken,
		},
	)
	sendJSON(w, http.StatusCreated, dataMarshal)
}

func UserResponseOk(
	w http.ResponseWriter,
	user *domain.User,
) {
	dataMarshal, _ := easyjson.Marshal(
		UserResponse{
			Status: StatusSuccess,
			UserID: user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Birth:  user.Birthday,
			Avatar: user.Avatar,
		},
	)
	sendJSON(w, http.StatusOK, dataMarshal)
}

func ValidationError(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(
				errMsgs, fmt.Sprintf("field %s is a required field", err.Field()),
			)
		default:
			errMsgs = append(
				errMsgs, fmt.Sprintf("field %s is not valid", err.Field()),
			)
		}
	}
	return strings.Join(errMsgs, ", ")
}
