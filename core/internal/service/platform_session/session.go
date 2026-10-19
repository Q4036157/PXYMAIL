package platform_session

import (
	"billionmail-core/internal/service/public"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/mail"
	"strings"
	"time"
)

type Claims struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	AppID          string `json:"app_id"`
	Type           string `json:"type"`
	TokenUse       string `json:"token_use"`
	PlatformNode   string `json:"platform_node"`
	SessionExpires int64  `json:"session_expires,omitempty"`
	jwt.RegisteredClaims
}

func Verify(raw string) (*Claims, error) {
	secret := public.MustGetDockerEnv("MAIL_APP_SESSION_SECRET", "")
	if len(secret) < 32 {
		return nil, errors.New("platform login is not configured")
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (interface{}, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("pxylh"), jwt.WithAudience("mail"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return nil, errors.New("platform session invalid or expired")
	}
	claims.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	address, e := mail.ParseAddress(claims.Email)
	if e != nil || address.Address != claims.Email || claims.Subject == "" || claims.ID == "" || claims.IssuedAt == nil || claims.AppID != "mail" || claims.Type != "app_session" || claims.TokenUse != "proxy_session" || (claims.PlatformNode != "edge-109" && claims.PlatformNode != "app-204") {
		return nil, errors.New("invalid platform identity")
	}
	return claims, nil
}

// Each workspace receives only its own bootstrap key, never the platform signing key.
func VerifyWorkspace(raw string) (*Claims, error) {
	secret := public.MustGetDockerEnv("MAIL_BOOTSTRAP_SECRET", "")
	if len(secret) < 32 {
		return nil, errors.New("workspace login is not configured")
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (interface{}, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("pxy-mail-gateway"), jwt.WithAudience("mail"), jwt.WithExpirationRequired())
	if err != nil || claims.Email == "" || claims.Subject == "" || claims.TokenUse != "workspace_bootstrap" || claims.SessionExpires <= time.Now().Unix() || claims.SessionExpires > time.Now().Add(4*time.Hour+time.Minute).Unix() {
		return nil, errors.New("invalid workspace login ticket")
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Unix(claims.SessionExpires, 0))
	return claims, nil
}
