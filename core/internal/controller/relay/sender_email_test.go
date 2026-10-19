package relay

import (
	"encoding/json"
	"strings"
	"testing"

	v1 "billionmail-core/api/relay/v1"
	"github.com/gogf/gf/v2/util/gconv"
)

func TestValidatedSenderEmail(t *testing.T) {
	for _, test := range []struct {
		name    string
		value   string
		domains []string
		want    string
	}{
		{"旧配置回退登录名", "", nil, ""},
		{"独立发件邮箱", "sales@example.com", []string{"example.com"}, "sales@example.com"},
		{"兼容已有域名前缀", "Sales@EXAMPLE.COM", []string{"@example.com"}, "sales@example.com"},
		{"多个域名精确匹配", "sales@sub.example.com", []string{"other.example.com", "@SUB.EXAMPLE.COM"}, "sales@sub.example.com"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := validatedSenderEmail(test.value, test.domains)
			if err != nil || got != test.want {
				t.Fatalf("发件地址校验失败: %q, %v", got, err)
			}
		})
	}
	for name, value := range map[string]string{
		"显示名":       "Sales <sales@example.com>",
		"多个邮箱":      "one@example.com,two@example.com",
		"换行注入":      "sales@example.com\r\nBcc:other@example.com",
		"前后空格":      " sales@example.com ",
		"登录标识不作为邮箱": "smtp-login-123",
		"缺少域名":      "sales@",
		"无效域名":      "sales@-example.com",
		"本地主机名":     "sales@localhost",
		"域名空标签":     "sales@example..com",
		"IP字面值":     "sales@[192.0.2.1]",
		"超长邮箱":      strings.Repeat("a", 250) + "@example.com",
		"超长本地部分":    strings.Repeat("a", 65) + "@example.com",
		"其他域名":      "sales@other.example.com",
		"域名后缀冒充":    "sales@example.com.attacker.test",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validatedSenderEmail(value, []string{"example.com"}); err == nil {
				t.Fatal("未拒绝无效邮箱或域名不匹配")
			}
		})
	}
	if _, err := validatedSenderEmail("sales@example.com", nil); err == nil {
		t.Fatal("非空发件邮箱没有绑定域名时不应通过")
	}
}

func TestSenderEmailUpdatePresence(t *testing.T) {
	for _, test := range []struct {
		body     string
		provided bool
		value    string
	}{
		{`{"id":1}`, false, ""},
		{`{"id":1,"sender_email":null}`, false, ""},
		{`{"id":1,"sender_email":""}`, true, ""},
		{`{"id":1,"sender_email":"sales@example.com"}`, true, "sales@example.com"},
	} {
		var req v1.UpdateRelayConfigReq
		if err := json.Unmarshal([]byte(test.body), &req); err != nil {
			t.Fatal(err)
		}
		if (req.SenderEmail != nil) != test.provided || req.SenderEmail != nil && *req.SenderEmail != test.value {
			t.Fatalf("JSON 未传、清空与更新语义错误: %s", test.body)
		}
		var fields map[string]any
		if err := json.Unmarshal([]byte(test.body), &fields); err != nil {
			t.Fatal(err)
		}
		req = v1.UpdateRelayConfigReq{}
		if err := gconv.Scan(fields, &req); err != nil {
			t.Fatal(err)
		}
		if (req.SenderEmail != nil) != test.provided || req.SenderEmail != nil && *req.SenderEmail != test.value {
			t.Fatalf("GoFrame 绑定未传、清空与更新语义错误: %s", test.body)
		}
	}
}
