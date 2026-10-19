package mail_service

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"billionmail-core/internal/model/entity"
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/frame/g"
)

// Match the sender account, rather than sending all mail through a domain-wide relay.
func newNativeEmailSender(email string) (*EmailSender, error) {
	ctx := context.Background()
	var config *entity.BmRelayConfig
	err := g.DB().Model("bm_relay_config").Ctx(ctx).Where("active", 1).
		Where("LOWER(COALESCE(NULLIF(sender_email,''),auth_user)) = ?", strings.ToLower(strings.TrimSpace(email))).Order("id").Scan(&config)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("no enabled external SMTP configuration for %s", email)
	}
	if config.RelayPort != "465" && config.RelayPort != "587" {
		return nil, fmt.Errorf("native SMTP requires TLS on port 465 or 587")
	}
	key, err := g.DB().Model("bm_options").Where("name", "relay_encryption_key").Value("value")
	if err != nil || key.IsEmpty() {
		return nil, fmt.Errorf("SMTP encryption key is unavailable")
	}
	keyBytes, err := hex.DecodeString(key.String())
	if err != nil || len(keyBytes) < 16 {
		return nil, fmt.Errorf("invalid SMTP encryption key")
	}
	ciphertext, err := hex.DecodeString(config.AuthPassword)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted SMTP password")
	}
	password, err := gaes.Decrypt(ciphertext, keyBytes[:16])
	if err != nil {
		return nil, fmt.Errorf("decrypt SMTP password: %w", err)
	}
	return nativeSenderWithConfig(email, config, string(password))
}

func nativeSenderWithConfig(email string, config *entity.BmRelayConfig, password string) (*EmailSender, error) {
	configuredEmail := config.SenderEmail
	if configuredEmail == "" {
		configuredEmail = config.AuthUser
	}
	if !strings.EqualFold(strings.TrimSpace(email), configuredEmail) {
		return nil, fmt.Errorf("实际发件邮箱与 SMTP 配置不匹配")
	}
	sender := NewEmailSender()
	sender.Email, sender.UserName = email, config.AuthUser
	sender.Host, sender.Port, sender.SNI = config.RelayHost, config.RelayPort, config.RelayHost
	sender.Password = password
	if !sender.IsConfigured() {
		return nil, fmt.Errorf("external SMTP configuration is incomplete for %s", email)
	}
	return sender, nil
}
