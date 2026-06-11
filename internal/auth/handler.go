package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
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
	appUrl string
}

func NewHandler(queries *db.Queries, m mailer.Mailer, appUrl string) *Handler {
	return &Handler{db: queries, mailer: m, appUrl: appUrl}
}

func (h *Handler) Me(c *gin.Context) {
	user := GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, domain.User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Handler) RequestLink(c *gin.Context) {
	var body struct {
		Email string `json:"email" form:"email"`
	}
	if err := c.ShouldBind(&body); err != nil || body.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	user, err := h.db.GetOrCreateUser(c.Request.Context(), body.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if _, err = h.db.CreateMagicLink(c.Request.Context(), db.CreateMagicLinkParams{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(magicLinkExpiry),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := h.mailer.SendMagicLink(body.Email, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Check your email for a sign-in link."})
}

func (h *Handler) VerifyToken(c *gin.Context) {
	fail := func() { c.Redirect(http.StatusSeeOther, h.appUrl+"/login?error=invalid") }

	token := c.Query("token")
	if token == "" {
		fail()
		return
	}

	link, err := h.db.GetMagicLink(c.Request.Context(), token)
	if err != nil {
		fail()
		return
	}

	if err := h.db.MarkMagicLinkUsed(c.Request.Context(), link.Token); err != nil {
		fail()
		return
	}

	sessionToken, err := generateToken()
	if err != nil {
		fail()
		return
	}

	if _, err = h.db.CreateSession(c.Request.Context(), db.CreateSessionParams{
		Token:     sessionToken,
		UserID:    link.UserID,
		ExpiresAt: time.Now().Add(sessionExpiry),
	}); err != nil {
		fail()
		return
	}

	secure := os.Getenv("APP_ENV") != "development"
	c.SetCookie(sessionCookie, sessionToken, int(sessionExpiry.Seconds()), "/", "", secure, true)

	c.Redirect(http.StatusSeeOther, h.appUrl+"/")
}

func (h *Handler) Logout(c *gin.Context) {
	token, err := c.Cookie(sessionCookie)
	if err == nil {
		_ = h.db.DeleteSession(c.Request.Context(), token)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
