package mail_service

import (
	"testing"

	"billionmail-core/internal/model/entity"
)

func TestNativeSenderSeparatesLoginAndFrom(t *testing.T) {
	config := &entity.BmRelayConfig{
		AuthUser: "smtp-login-123", SenderEmail: "sales@example.com",
		RelayHost: "smtp.example.com", RelayPort: "587",
	}
	sender, err := nativeSenderWithConfig("sales@example.com", config, "test-smtp-password")
	if err != nil {
		t.Fatal(err)
	}
	if sender.Email != config.SenderEmail || sender.UserName != config.AuthUser || sender.Password != "test-smtp-password" {
		t.Fatal("SMTP 登录凭据和实际发件地址发生混用")
	}
	if _, err := nativeSenderWithConfig(config.AuthUser, config, "test-smtp-password"); err == nil {
		t.Fatal("配置独立发件邮箱后不能将登录名用作发件地址")
	}
	if _, err := nativeSenderWithConfig("other@example.com", config, "test-smtp-password"); err == nil {
		t.Fatal("不能冒用此配置之外的发件邮箱")
	}
}

func TestNativeSenderLegacyConfiguration(t *testing.T) {
	config := &entity.BmRelayConfig{
		AuthUser: "legacy@126.com", RelayHost: "smtp.126.com", RelayPort: "465",
	}
	sender, err := nativeSenderWithConfig("legacy@126.com", config, "test-authorization-code")
	if err != nil {
		t.Fatal(err)
	}
	if sender.Email != config.AuthUser || sender.UserName != config.AuthUser {
		t.Fatal("旧配置没有沿用登录邮箱发件")
	}
}
