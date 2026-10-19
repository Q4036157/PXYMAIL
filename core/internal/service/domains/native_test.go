package domains

import (
	"context"
	"strings"
	"testing"

	v1 "billionmail-core/api/domains/v1"
)

func TestNativeDomainRejectsLocalCatchall(t *testing.T) {
	t.Setenv("BILLIONMAIL_NATIVE", "1")
	// 在访问数据库或邮件服务器前拒绝不支持的本地收件配置。
	err := Add(context.Background(), &v1.Domain{Domain: "example.com", Catchall: "support@example.com"})
	if err == nil || !strings.Contains(err.Error(), "不支持本地全收邮箱") {
		t.Fatalf("应拒绝本地全收邮箱，实际：%v", err)
	}
}
