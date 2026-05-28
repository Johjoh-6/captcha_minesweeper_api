package handlers

import (
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/response"
	"captcha_sweeper/internal/servers"
	"log/slog"
	"net/http"
)

type Handlers struct {
	*db.Queries     // Embedded for direct access (e.g., h.GetUserByID())
	*slog.Logger    // Embedded for direct access (e.g., h.Info(), h.Error())
	*servers.Errors // Embedded for direct access (e.g., h.ServerError())
	Captcha         *CaptchasHandlers
	// Users                    *UsersHandlers
	// ... Add more handlers as needed
}

func NewHandlers(queries *db.Queries, logger *slog.Logger, errHelper *servers.Errors, cfg *servers.Config) *Handlers {
	return &Handlers{
		Queries: queries,
		Logger:  logger,
		Errors:  errHelper,
		Captcha: NewCaptchasHandlers(queries, logger, errHelper),
		// Users:         NewUsersHandlers(queries, logger, errHelper),
	}
}

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		h.ServerError(w, r, err)
	}
}

func (h *Handlers) Restricted(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Message": "This is a restricted handler",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		h.ServerError(w, r, err)
	}
}
