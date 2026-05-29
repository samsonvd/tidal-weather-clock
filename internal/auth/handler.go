package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/mailer"
)

const (
	magicLinkExpiry = 15 * time.Minute
	sessionExpiry   = 30 * 24 * time.Hour
	sessionCookie   = "session"
)

type Handler struct {
	db     *db.Queries
	mailer mailer.Mailer
}

func NewHandler(queries *db.Queries, m mailer.Mailer) *Handler {
	return &Handler{db: queries, mailer: m}
}

func (h *Handler) RequestLink(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.String(http.StatusBadRequest, "email is required")
		return
	}

	user, err := h.db.GetOrCreateUser(c.Request.Context(), email)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	token, err := generateToken()
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	if _, err = h.db.CreateMagicLink(c.Request.Context(), db.CreateMagicLinkParams{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(magicLinkExpiry),
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.mailer.SendMagicLink(email, token); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	c.Redirect(http.StatusSeeOther, "/login?sent=true")
}

func (h *Handler) VerifyToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.Redirect(http.StatusSeeOther, "/login?error=invalid")
		return
	}

	link, err := h.db.GetMagicLink(c.Request.Context(), token)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login?error=invalid")
		return
	}

	if err := h.db.MarkMagicLinkUsed(c.Request.Context(), link.Token); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	sessionToken, err := generateToken()
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	if _, err = h.db.CreateSession(c.Request.Context(), db.CreateSessionParams{
		Token:     sessionToken,
		UserID:    link.UserID,
		ExpiresAt: time.Now().Add(sessionExpiry),
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.db.EnsureLocation(c.Request.Context(), link.UserID); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	secure := os.Getenv("APP_ENV") != "development"
	c.SetCookie(sessionCookie, sessionToken, int(sessionExpiry.Seconds()), "/", "", secure, true)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) Logout(c *gin.Context) {
	token, err := c.Cookie(sessionCookie)
	if err == nil {
		_ = h.db.DeleteSession(c.Request.Context(), token)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
