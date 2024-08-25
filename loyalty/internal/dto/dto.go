package dto

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mailru/easyjson"
)

// Input

type UserLoyalty struct {
	UUID  string `json:"uuid" validate:"uuid"`
	Value int    `json:"value"`
}

// Output

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	UUID   string `json:"uuid,omitempty"`
	Value  int    `json:"value,omitempty"`
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

func ResponseOKLoyalty(
	w http.ResponseWriter,
	uuid string,
	value int,
) {
	dataMarshal, _ := easyjson.Marshal(
		Response{
			Status: StatusSuccess,
			UUID:   uuid,
			Value:  value,
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
