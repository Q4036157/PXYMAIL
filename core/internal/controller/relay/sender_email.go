package relay

import (
	"fmt"
	"net/mail"
	"strings"
)

func validatedSenderEmail(value string, senderDomains []string) (string, error) {
	if value == "" {
		return "", nil
	}
	address, err := mail.ParseAddress(value)
	if len(value) > 254 || err != nil || address.Address != value || strings.Count(value, "@") != 1 || strings.IndexByte(value, '@') > 64 || strings.ContainsAny(value, "\r\n\t <>") {
		return "", fmt.Errorf("实际发件地址必须是单个纯邮箱，不能包含显示名、空格或换行")
	}
	value = strings.ToLower(value)
	domain := strings.SplitN(value, "@", 2)[1]
	if !strings.Contains(domain, ".") || len(domain) > 253 {
		return "", fmt.Errorf("实际发件地址的域名无效")
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", fmt.Errorf("实际发件地址的域名无效")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", fmt.Errorf("实际发件地址的域名无效")
			}
		}
	}
	for _, allowed := range senderDomains {
		if strings.ToLower(strings.TrimPrefix(strings.TrimSpace(allowed), "@")) == domain {
			return value, nil
		}
	}
	return "", fmt.Errorf("实际发件地址必须属于此 SMTP 配置的发件域名")
}
