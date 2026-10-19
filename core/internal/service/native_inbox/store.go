package native_inbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type Account struct {
	ID           int    `json:"id" orm:"relay_id"`
	Email        string `json:"email"`
	Remark       string `json:"remark"`
	Enabled      bool   `json:"enabled"`
	Status       string `json:"status"`
	LastError    string `json:"last_error"`
	LastAttempt  int64  `json:"last_attempt"`
	LastSuccess  int64  `json:"last_success"`
	NextPoll     int64  `json:"next_poll"`
	MailboxTotal int64  `json:"mailbox_total"`
	Count        int    `json:"cached_count" orm:"cached_count"`
	Pending      bool   `json:"pending"`
}
type Mail struct {
	ID          int64  `json:"id"`
	AccountID   int    `json:"account_id"`
	Account     string `json:"account"`
	UIDValidity int64  `json:"uid_validity"`
	UID         int64  `json:"uid"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	Subject     string `json:"subject"`
	ReceivedAt  int64  `json:"received_at"`
	MessageID   string `json:"message_id"`
	InReplyTo   string `json:"in_reply_to"`
	Kind        string `json:"kind"`
	Processed   bool   `json:"processed"`
	Seen        bool   `json:"seen"`
	BodyCached  bool   `json:"body_cached"`
	BodyText    string `json:"body_text,omitempty"`
	BodyHTML    string `json:"body_html,omitempty"`
	Attachments string `json:"attachments,omitempty"`
	Truncated   bool   `json:"truncated"`
}

var initialize sync.Once
var initError error
var legacySent = map[string]map[string]bool{}

const pollInterval = int64(15 * 60)

func Init(ctx context.Context) error {
	initialize.Do(func() {
		if !public.NativeMode() {
			initError = fmt.Errorf("收件箱需要 Windows 原生运行模式")
			return
		}
		_, initError = g.DB().Exec(ctx, `CREATE TABLE IF NOT EXISTS bm_native_inbox_accounts (
   relay_id INTEGER PRIMARY KEY, email TEXT NOT NULL, remark TEXT NOT NULL DEFAULT '',
   enabled BOOLEAN NOT NULL DEFAULT TRUE, pending BOOLEAN NOT NULL DEFAULT FALSE,
   status TEXT NOT NULL DEFAULT '等待首次同步', last_error TEXT NOT NULL DEFAULT '',
   last_attempt BIGINT NOT NULL DEFAULT 0, last_success BIGINT NOT NULL DEFAULT 0,
   next_poll BIGINT NOT NULL DEFAULT 0, mailbox_total BIGINT NOT NULL DEFAULT 0);
  CREATE TABLE IF NOT EXISTS bm_native_inbox_messages (
   id BIGSERIAL PRIMARY KEY, account_id INTEGER NOT NULL, uid_validity BIGINT NOT NULL, uid BIGINT NOT NULL,
   from_address TEXT NOT NULL DEFAULT '', from_name TEXT NOT NULL DEFAULT '', subject TEXT NOT NULL DEFAULT '',
   received_at BIGINT NOT NULL DEFAULT 0, message_id TEXT NOT NULL DEFAULT '', in_reply_to TEXT NOT NULL DEFAULT '',
   kind TEXT NOT NULL DEFAULT 'other', seen BOOLEAN NOT NULL DEFAULT FALSE, processed BOOLEAN NOT NULL DEFAULT FALSE,
   body_cached BOOLEAN NOT NULL DEFAULT FALSE, body_text TEXT NOT NULL DEFAULT '', body_html TEXT NOT NULL DEFAULT '',
   attachments TEXT NOT NULL DEFAULT '[]', truncated BOOLEAN NOT NULL DEFAULT FALSE,
   UNIQUE(account_id,uid_validity,uid));
  CREATE INDEX IF NOT EXISTS bm_native_inbox_date_idx ON bm_native_inbox_messages(received_at DESC,id DESC);
  CREATE INDEX IF NOT EXISTS bm_native_inbox_account_idx ON bm_native_inbox_messages(account_id,kind);`)
		if initError == nil {
			loadLegacySent()
		}
	})
	return initError
}
func reconcile(ctx context.Context) error {
	if err := Init(ctx); err != nil {
		return err
	}
	_, err := g.DB().Exec(ctx, `INSERT INTO bm_native_inbox_accounts(relay_id,email,remark,enabled)
  SELECT id,lower(auth_user),remark,active=1 FROM bm_relay_config
  WHERE (lower(auth_user) LIKE '%@126.com' OR lower(auth_user) LIKE '%@163.com') AND auth_password<>''
  ON CONFLICT(relay_id) DO UPDATE SET email=EXCLUDED.email,remark=EXCLUDED.remark`)
	return err
}
func Accounts(ctx context.Context) ([]Account, error) {
	if err := reconcile(ctx); err != nil {
		return nil, err
	}
	result := make([]Account, 0)
	err := g.DB().Model("bm_native_inbox_accounts a").Fields("a.*, (SELECT count(*) FROM bm_native_inbox_messages m WHERE m.account_id=a.relay_id) AS cached_count").Order("a.relay_id").Scan(&result)
	return result, err
}
func Queue(ctx context.Context, accountID int, enabled *bool) error {
	if err := reconcile(ctx); err != nil {
		return err
	}
	if accountID != 0 {
		exists, err := g.DB().Model("bm_native_inbox_accounts").Where("relay_id", accountID).Exist()
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("收件账号不存在")
		}
	}
	if enabled != nil {
		if accountID == 0 {
			return fmt.Errorf("请指定一个收件账号")
		}
		_, err := g.DB().Model("bm_native_inbox_accounts").Where("relay_id", accountID).Data(g.Map{"enabled": *enabled, "pending": *enabled, "status": map[bool]string{true: "已排队，等待同步", false: "已暂停收件"}[*enabled]}).Update()
		return err
	}
	query := g.DB().Model("bm_native_inbox_accounts").Where("enabled", true)
	if accountID != 0 {
		query = query.Where("relay_id", accountID)
	}
	_, err := query.Data(g.Map{"pending": true, "next_poll": gdb.Raw("GREATEST(last_attempt+900,0)"), "status": "已排队，等待同步"}).Update()
	return err
}
func List(ctx context.Context, accountID int, kind, keyword string, page, pageSize int) ([]Mail, int, error) {
	if err := Init(ctx); err != nil {
		return nil, 0, err
	}
	model := g.DB().Model("bm_native_inbox_messages m").LeftJoin("bm_native_inbox_accounts a", "a.relay_id=m.account_id")
	if accountID > 0 {
		model = model.Where("m.account_id", accountID)
	}
	if kind != "" && kind != "all" {
		model = model.Where("m.kind", kind)
	}
	if keyword != "" {
		model = model.Where("(m.subject ILIKE ? OR m.from_address ILIKE ? OR m.from_name ILIKE ?)", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	count, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	rows := make([]Mail, 0)
	err = model.Fields("m.id,m.account_id,a.email AS account,m.uid_validity,m.uid,m.from_address,m.from_name,m.subject,m.received_at,m.message_id,m.in_reply_to,m.kind,m.processed,m.seen,m.body_cached,m.truncated").Order("m.received_at DESC,m.id DESC").Page(page, pageSize).Scan(&rows)
	return rows, count, err
}
func getMail(ctx context.Context, id int64) (*Mail, error) {
	if err := Init(ctx); err != nil {
		return nil, err
	}
	var result *Mail
	err := g.DB().Model("bm_native_inbox_messages m").LeftJoin("bm_native_inbox_accounts a", "a.relay_id=m.account_id").Fields("m.*,a.email AS account").Where("m.id", id).Scan(&result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("邮件不存在")
	}
	return result, nil
}
func Mark(ctx context.Context, id int64, processed bool) error {
	if _, err := getMail(ctx, id); err != nil {
		return err
	}
	_, err := g.DB().Model("bm_native_inbox_messages").Where("id", id).Data("processed", processed).Update()
	return err
}
func classify(ctx context.Context, account, address, subject, inReplyTo string) string {
	lower := strings.ToLower(address + " " + subject)
	for _, word := range []string{"mailer-daemon", "postmaster", "undeliver", "delivery failure", "delivery status notification", "退信", "邮件投递失败"} {
		if strings.Contains(lower, word) {
			return "bounce"
		}
	}
	if legacySent[account][address] {
		return "customer_reply"
	}
	known, err := g.DB().Model("recipient_info r").InnerJoin("email_tasks t", "t.id=r.task_id").Where("lower(r.recipient)=? AND lower(t.addresser)=? AND r.is_sent=1", address, account).Exist()
	if err == nil && known {
		return "customer_reply"
	}
	if inReplyTo != "" {
		known, err = g.DB().Model("bm_contacts").Where("lower(email)", address).Exist()
		if err == nil && known {
			return "customer_reply"
		}
	}
	return "other"
}
func loadLegacySent() {
	root := filepath.Join(filepath.Dir(public.EnvironmentFile()), "configuration")
	var profiles map[string]struct {
		From string `json:"from"`
	}
	var state struct {
		Jobs map[string]struct {
			Recipients []struct {
				Email  string `json:"email"`
				Sender string `json:"sender"`
				Status string `json:"status"`
			} `json:"recipients"`
		} `json:"jobs"`
	}
	profileData, e1 := os.ReadFile(filepath.Join(root, "profiles.json"))
	stateData, e2 := os.ReadFile(filepath.Join(root, "state.json"))
	if e1 != nil || e2 != nil {
		return
	}
	if json.Unmarshal(profileData, &profiles) != nil || json.Unmarshal(stateData, &state) != nil {
		return
	}
	for _, job := range state.Jobs {
		for _, r := range job.Recipients {
			if r.Status != "sent" && r.Status != "bounced" {
				continue
			}
			account := strings.ToLower(profiles[r.Sender].From)
			if account == "" {
				continue
			}
			if legacySent[account] == nil {
				legacySent[account] = map[string]bool{}
			}
			legacySent[account][strings.ToLower(r.Email)] = true
		}
	}
}
func failAccount(ctx context.Context, id int, err error) {
	retry := time.Now().Unix() + 1800
	_, dbErr := g.DB().Model("bm_native_inbox_accounts").Where("relay_id", id).Data(g.Map{"status": "连接失败，稍后重试", "last_error": err.Error(), "next_poll": retry, "pending": false}).Update()
	if dbErr != nil {
		g.Log().Error(ctx, "inbox status update failed", dbErr)
	}
}
