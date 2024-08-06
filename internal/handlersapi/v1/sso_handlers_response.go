package v1

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/mailru/easyjson"
	"net/http"
	"strings"
)

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

type Logout struct {
	Token string `json:"token"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const StatusError = "Error"

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func HealthOk(msg string) Response {
	return Response{
		Status: msg,
	}
}

func ValidationError(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return strings.Join(errMsgs, ", ")
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func responseError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	dataMarshal, _ := easyjson.Marshal(Error(message))
	w.Write(dataMarshal)
}

func responseHealth(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	dataMarshal, _ := json.Marshal(HealthOk(message))
	w.Write(dataMarshal)
}
