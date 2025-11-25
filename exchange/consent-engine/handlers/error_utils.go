package handlers

import (
	"net/http"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// handleConsentError standardizes error responses for consent handlers.
func handleConsentError(w http.ResponseWriter, status int, code, message string) {
	utils.RespondWithJSON(w, status, models.ErrorResponseWithCode{
		Code:  code,
		Error: message,
	})
}
