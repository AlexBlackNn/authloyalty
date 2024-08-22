package domain

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/mailru/easyjson"
	"net/http"
	"strings"
)

// DTO http and grpc structures.

type Login struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password"`
}

type Register struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
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
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
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
	accessToken string,
	refreshToken string,
) {
	dataMarshal, _ := easyjson.Marshal(
		Response{
			Status:       StatusSuccess,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	)
	sendJSON(w, http.StatusCreated, dataMarshal)
}

// Validation error.

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
