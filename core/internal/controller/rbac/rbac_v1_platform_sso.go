package rbac

import (
	"billionmail-core/api/rbac/v1"
	"billionmail-core/internal/model"
	platform "billionmail-core/internal/service/platform_session"
	"billionmail-core/internal/service/public"
	service "billionmail-core/internal/service/rbac"
	"context"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func (c *ControllerV1) PlatformSSO(ctx context.Context, req *v1.PlatformSSOReq) (*v1.LoginRes, error) {
	identity, err := platform.VerifyWorkspace(req.PlatformToken)
	if err != nil {
		return nil, err
	}
	expected := public.MustGetDockerEnv("MAIL_SPACE_EMAIL", "")
	if expected == "" || identity.Email != expected {
		return nil, errors.New("this workspace belongs to another platform account")
	}
	var account model.Account
	err = g.DB().Model("account").Where("username", public.MustGetDockerEnv("ADMIN_USERNAME", "")).Scan(&account)
	if err != nil {
		return nil, err
	}
	if account.AccountId == 0 || account.Status != 1 {
		return nil, errors.New("workspace account unavailable")
	}
	// UI language comes from workspace options, independently of the account field.
	languageValue, err := g.DB().Model("bm_options").Where("name", "CurrentLanguage").Value("value")
	if err != nil {
		return nil, err
	}
	language := languageValue.String()
	public.SetCache("language", language, 3600)
	expiry := identity.ExpiresAt.Time
	if localExpiry := time.Now().Add(service.JWT().AccessExpiry); localExpiry.Before(expiry) {
		expiry = localExpiry
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &service.JWTCustomClaims{
		AccountId: account.AccountId, Username: account.Username, Roles: []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiry), IssuedAt: jwt.NewNumericDate(time.Now()), Subject: "access_token", ID: identity.ID},
	}).SignedString([]byte(service.JWT().Secret))
	if err != nil {
		return nil, err
	}
	_, err = g.DB().Model("account").Where("account_id", account.AccountId).Data(g.Map{"email": identity.Email, "last_login_time": time.Now().Unix(), "language": language}).Update()
	if err != nil {
		return nil, err
	}
	res := &v1.LoginRes{}
	res.Success = true
	res.Code = 0
	res.Msg = "Platform login successful"
	res.Data.Token = token
	res.Data.TTL = int64(time.Until(expiry).Seconds())
	res.Data.AccountInfo.Id = account.AccountId
	res.Data.AccountInfo.Username = identity.Email
	res.Data.AccountInfo.Email = identity.Email
	res.Data.AccountInfo.Status = 1
	res.Data.AccountInfo.Lang = language
	r := g.RequestFromCtx(ctx)
	_ = r.Session.Set("SignedToken", token)
	_ = r.Session.Set("safe_path_pass", true)
	_ = r.Session.Set("language", language)
	return res, nil
}
