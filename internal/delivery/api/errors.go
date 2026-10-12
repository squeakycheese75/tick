package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/squeakycheese75/tick/internal/domain"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// fail answers a failed request. Problems with the request itself keep their own status; use
// case errors are classified by domain.KindOf, so the API never inspects individual errors.
func (s *server) fail(w http.ResponseWriter, r *http.Request, err error) {
	var reqErr *requestError
	if errors.As(err, &reqErr) {
		writeError(w, reqErr.status, reqErr.code, reqErr.Error())
		return
	}

	kind := domain.KindOf(err)
	status, code := statusFor(kind)
	message := err.Error()

	// What a provider or the store says can carry URLs with API keys in them, or internals.
	// The caller gets a fixed message; the detail goes to the log with secrets removed.
	if kind == domain.KindInternal || kind == domain.KindUnavailable {
		s.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "kind", kind.String(), "error", redact(err.Error()))
		message = publicMessage(kind)
	}

	writeError(w, status, code, message)
}

func statusFor(kind domain.ErrorKind) (int, string) {
	switch kind {
	case domain.KindInvalid:
		return http.StatusUnprocessableEntity, "invalid"
	case domain.KindNotFound:
		return http.StatusNotFound, "not_found"
	case domain.KindConflict:
		return http.StatusConflict, "conflict"
	case domain.KindUnavailable:
		return http.StatusBadGateway, "unavailable"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

func publicMessage(kind domain.ErrorKind) string {
	if kind == domain.KindUnavailable {
		return "a price or exchange rate provider is unavailable; try again shortly"
	}

	return "internal error"
}

var secretParam = regexp.MustCompile(`(?i)(token|apikey|api_key|key)=[^&\s"]+`)

// redact hides values of query parameters that hold credentials.
func redact(s string) string {
	return secretParam.ReplaceAllString(s, "${1}=REDACTED")
}
