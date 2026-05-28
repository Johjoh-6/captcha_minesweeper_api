package handlers

import (
	"captcha_sweeper/internal/captcha"
	"captcha_sweeper/internal/contexts"
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/request"
	"captcha_sweeper/internal/response"
	"captcha_sweeper/internal/servers"
	"captcha_sweeper/internal/validator"
	"log/slog"
	"net/http"
)

type CaptchasHandlers struct {
	queries   *db.Queries
	logger    *slog.Logger
	errHelper *servers.Errors
}

func (h *CaptchasHandlers) respondCaptcha(w http.ResponseWriter, r *http.Request, model db.Captcha, status int) {
	resp := captcha.BuildCaptchaResponse(model)
	if err := response.JSON(w, status, resp); err != nil {
		h.errHelper.ServerError(w, r, err)
	}
}

func NewCaptchasHandlers(
	queries *db.Queries,
	logger *slog.Logger,
	errHelper *servers.Errors,
) *CaptchasHandlers {
	return &CaptchasHandlers{
		queries:   queries,
		logger:    logger,
		errHelper: errHelper,
	}
}

func (h *CaptchasHandlers) GetCaptcha(w http.ResponseWriter, r *http.Request) {
	// get captcha from session
	session, exist := contexts.ContextGetAuthenticatedSession(r)
	if !exist {
		h.errHelper.NotFound(w, r)
		return
	}

	captchaModel, err := h.queries.GetCaptchaBySessionID(r.Context(), session.SessionID)
	if err != nil {
		h.errHelper.NotFound(w, r)
		return
	}

	// Always return the current captcha state (even if solved/failed)
	h.respondCaptcha(w, r, captchaModel, http.StatusOK)
}

func (h *CaptchasHandlers) NewCaptcha(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, exist := contexts.ContextGetAuthenticatedSession(r)
	if !exist {
		h.errHelper.NotFound(w, r)
		return
	}

	var input struct {
		DifficultyLevel int32 `json:"difficulty_level"`
	}
	if err := request.DecodeJSON(w, r, &input); err != nil {
		h.errHelper.ServerError(w, r, err)
		return
	}

	// Validate difficulty input
	v := validator.Validator{}
	validator.ValidateDifficultyInput(&v, input.DifficultyLevel)
	if v.HasErrors() {
		h.errHelper.FailedValidation(w, r, v)
		return
	}

	// Generate new captcha
	mineSweeperCaptcha := captcha.NewMineSweeper(input.DifficultyLevel)
	// init grid
	mineSweeperCaptcha.Initialize()
	captchaModel := mineSweeperCaptcha.ToCaptcha()
	captchaDB, err := h.queries.CreateCaptcha(r.Context(), db.CreateCaptchaParams{
		SessionID:       session.SessionID,
		GridSize:        captchaModel.GridSize,
		MineCount:       captchaModel.MineCount,
		Grid:            captchaModel.Grid,
		Revealed:        captchaModel.Revealed,
		DifficultyLevel: captchaModel.DifficultyLevel,
	})
	if err != nil {
		h.errHelper.ServerError(w, r, err)
		return
	}

	h.respondCaptcha(w, r, captchaDB, http.StatusOK)
}

func (h *CaptchasHandlers) PlayCaptcha(w http.ResponseWriter, r *http.Request) {
	// get captcha from session
	session, exist := contexts.ContextGetAuthenticatedSession(r)
	if !exist {
		h.errHelper.NotFound(w, r)
		return
	}

	var input struct {
		Row int32 `json:"row"`
		Col int32 `json:"col"`
	}
	if err := request.DecodeJSON(w, r, &input); err != nil {
		h.errHelper.ServerError(w, r, err)
		return
	}

	captchaModel, err := h.queries.GetCaptchaBySessionID(r.Context(), session.SessionID)
	if err != nil {
		h.errHelper.NotFound(w, r)
		return
	}

	// Validate captcha movement input
	v := validator.Validator{}
	validator.ValidateCaptchaMovementInput(&v, input.Row, input.Col, captchaModel.GridSize)
	if v.HasErrors() {
		h.errHelper.FailedValidation(w, r, v)
		return
	}

	if (captchaModel.Solved.Valid && captchaModel.Solved.Bool) || (captchaModel.Failed.Valid && captchaModel.Failed.Bool) {
		h.respondCaptcha(w, r, captchaModel, http.StatusOK)
		return
	}

	mineSweeper := captcha.NewMineSweeperFromCaptcha(captchaModel)
	safe := mineSweeper.Reveal(input.Row, input.Col)

	updated := mineSweeper.ToCaptcha()
	updatedCaptcha, err := h.queries.UpdateCaptcha(r.Context(), db.UpdateCaptchaParams{
		CaptchaID: captchaModel.CaptchaID,
		Revealed:  updated.Revealed,
		Solved:    updated.Solved,
		Failed:    updated.Failed,
	})
	if err != nil {
		h.errHelper.ServerError(w, r, err)
		return
	}

	status := http.StatusOK
	if !safe {
		status = http.StatusAccepted
	}
	h.respondCaptcha(w, r, updatedCaptcha, status)
	return
}
