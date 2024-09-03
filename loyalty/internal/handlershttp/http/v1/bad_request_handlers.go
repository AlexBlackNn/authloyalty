package v1

import (
	"errors"
	"io"
	"net/http"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

func handleAddLoyaltyBadRequest(w http.ResponseWriter, r *http.Request, reqData *dto.UserLoyalty) (*dto.UserLoyalty, error) {
	if r.Method != http.MethodPost {
		dto.ResponseErrorNowAllowed(w, "only POST method allowed")
		return nil, errors.New("method not allowed")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ResponseErrorBadRequest(w, "failed to read body")
		return nil, errors.New("failed to read body")
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err = json.Unmarshal(body, reqData)
	if err != nil {
		dto.ResponseErrorBadRequest(w, "failed to decode body")
		return nil, errors.New("failed to decode request")
	}

	if err = validator.New().Struct(reqData); err != nil {
		var validateErr validator.ValidationErrors
		if errors.As(err, &validateErr) {
			dto.ResponseErrorBadRequest(w, dto.ValidationError(validateErr))
			return nil, errors.New("validation error")
		}
		dto.ResponseErrorBadRequest(w, "bad request")
		return nil, errors.New("bad request")
	}
	return reqData, nil
}

func handleGetLoyaltyBadRequest(w http.ResponseWriter, r *http.Request) (*domain.UserLoyalty, error) {
	if r.Method != http.MethodGet {
		dto.ResponseErrorNowAllowed(w, "only Get method allowed")
		return nil, errors.New("method not allowed")
	}
	currentUUID := chi.URLParam(r, "uuid")
	_, err := uuid.Parse(currentUUID)
	if err != nil {
		dto.ResponseErrorNowAllowed(w, "invalid uuid")
		return nil, errors.New("invalid uuid")
	}
	return &domain.UserLoyalty{UUID: currentUUID}, nil
}
