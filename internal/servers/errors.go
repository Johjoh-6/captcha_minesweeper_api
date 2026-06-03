package servers

import (
	"captcha_sweeper/internal/response"
	"captcha_sweeper/internal/validator"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// Errors provides consistent JSON error responses and server-side error reporting.
// It is intentionally separated from Application to avoid passing the whole app
// into handlers/middlewares.
type Errors struct {
	Logger *slog.Logger
}

func NewErrors(logger *slog.Logger) *Errors {
	return &Errors{Logger: logger}
}

func (e *Errors) ReportServerError(r *http.Request, err error) {
	if e == nil || e.Logger == nil {
		return
	}

	var (
		message = err.Error()
		method  = r.Method
		url     = r.URL.String()
		trace   = string(debug.Stack())
	)

	requestAttrs := slog.Group("request", "method", method, "url", url)
	e.Logger.Error(message, requestAttrs, "trace", trace)
}

func (e *Errors) ErrorMessage(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	if message != "" {
		message = strings.ToUpper(message[:1]) + message[1:]
	}

	if err := response.JSONWithHeaders(w, status, map[string]string{"Error": message}, headers); err != nil {
		// Best effort: log it, then fall back to status code only.
		e.ReportServerError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (e *Errors) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	e.ReportServerError(r, err)

	message := "The server encountered a problem and could not process your request"
	e.ErrorMessage(w, r, http.StatusInternalServerError, message, nil)
}

func (e *Errors) NotFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	e.ErrorMessage(w, r, http.StatusNotFound, message, nil)
}

func (e *Errors) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	e.ErrorMessage(w, r, http.StatusMethodNotAllowed, message, nil)
}

func (e *Errors) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	e.ErrorMessage(w, r, http.StatusBadRequest, err.Error(), nil)
}

func (e *Errors) FailedValidation(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	if err := response.JSON(w, http.StatusUnprocessableEntity, v); err != nil {
		e.ServerError(w, r, err)
	}
}

func (e *Errors) Unauthorized(w http.ResponseWriter, r *http.Request, err error, target string) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", target)
	e.ErrorMessage(w, r, http.StatusUnauthorized, fmt.Sprintf("%s: %s", target, err.Error()), headers)
}

func (e *Errors) InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")
	e.ErrorMessage(w, r, http.StatusUnauthorized, "Invalid authentication token", headers)
}

func (e *Errors) AuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")
	e.ErrorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource", headers)
}

func (e *Errors) BasicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	e.ErrorMessage(w, r, http.StatusUnauthorized, message, headers)
}
