package native_inbox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/public"
	"billionmail-core/internal/service/relay"
	"github.com/emersion/go-imap"
	imapid "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	messageMail "github.com/emersion/go-message/mail"
	"github.com/gogf/gf/v2/frame/g"
)

const bodyLimit = 1024 * 1024

var workerOnce sync.Once
var connectionLock = make(chan struct{}, 1)
var lastConnection time.Time
var manualLock sync.Mutex
var manualWaiting atomic.Bool
var errManualNotReady = errors.New("暂时无法手动检查")

type CheckResult struct {
	AccountID   int   `json:"account_id"`
	CheckedAt   int64 `json:"checked_at"`
	NewCount    int   `json:"new_count"`
	CachedCount int   `json:"cached_count"`
	ReplyCount  int   `json:"reply_count"`
}

func manualReady(account *Account) error {
	if account == nil || !account.Enabled {
		return fmt.Errorf("%w: 请先启用此账号的收件功能", errManualNotReady)
	}
	if remaining := account.LastAttempt + 60 - time.Now().Unix(); remaining > 0 {
		return fmt.Errorf("%w: 此邮箱刚刚检查过，请在 %d 秒后重试", errManualNotReady, remaining)
	}
	if (account.Status == "正在同步" || account.Status == "正在手动检查") && time.Now().Unix()-account.LastAttempt < 120 {
		return fmt.Errorf("%w: 此邮箱正在同步，请等待当前检查完成", errManualNotReady)
	}
	return nil
}

func CheckNow(ctx context.Context, accountID int) (*CheckResult, error) {
	if !manualLock.TryLock() {
		return nil, fmt.Errorf("已有手动检查正在进行，请等待完成后再检查其他邮箱")
	}
	defer manualLock.Unlock()
	manualWaiting.Store(true)
	defer manualWaiting.Store(false)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := reconcile(ctx); err != nil {
		return nil, err
	}
	var account *Account
	if err := g.DB().Model("bm_native_inbox_accounts").Ctx(ctx).Where("relay_id", accountID).Scan(&account); err != nil {
		return nil, err
	}
	if err := manualReady(account); err != nil {
		return nil, err
	}
	before, err := g.DB().Model("bm_native_inbox_messages").Ctx(ctx).Where("account_id", accountID).Count()
	if err != nil {
		return nil, err
	}
	if err = syncAccountMode(ctx, accountID, true); err != nil {
		if !errors.Is(err, errManualNotReady) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			failAccount(ctx, accountID, err)
		}
		return nil, err
	}
	if err = g.DB().Model("bm_native_inbox_accounts").Ctx(ctx).Where("relay_id", accountID).Scan(&account); err != nil {
		return nil, err
	}
	after, err := g.DB().Model("bm_native_inbox_messages").Ctx(ctx).Where("account_id", accountID).Count()
	if err != nil {
		return nil, err
	}
	replies, err := g.DB().Model("bm_native_inbox_messages").Ctx(ctx).Where("account_id", accountID).Where("kind", "customer_reply").Count()
	if err != nil {
		return nil, err
	}
	return &CheckResult{AccountID: accountID, CheckedAt: account.LastSuccess, NewCount: max(0, after-before), CachedCount: after, ReplyCount: replies}, nil
}

// Serialize logins across the worker and on-demand previews.
func acquire(ctx context.Context) (func(), error) {
	select {
	case connectionLock <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	wait := time.Until(lastConnection.Add(20 * time.Second))
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			<-connectionLock
			return nil, ctx.Err()
		}
	}
	lastConnection = time.Now()
	return func() { <-connectionLock }, nil
}
func openInbox(ctx context.Context, accountID int) (*client.Client, *imap.MailboxStatus, error) {
	var config *entity.BmRelayConfig
	if err := g.DB().Model("bm_relay_config").Where("id", accountID).Scan(&config); err != nil {
		return nil, nil, err
	}
	if config == nil {
		return nil, nil, fmt.Errorf("发件账号配置已移除，请重新配置 SMTP 账号")
	}
	password, err := relay.DecryptPassword(ctx, config.AuthPassword)
	if err != nil {
		return nil, nil, fmt.Errorf("无法读取已保存的授权码，请在 SMTP 页面重新保存")
	}
	host := ""
	address := strings.ToLower(config.AuthUser)
	if strings.HasSuffix(address, "@126.com") {
		host = "imap.126.com"
	} else if strings.HasSuffix(address, "@163.com") {
		host = "imap.163.com"
	}
	if host == "" {
		return nil, nil, fmt.Errorf("当前收件支持 126 和 163 邮箱")
	}
	c, err := client.DialWithDialerTLS(&net.Dialer{Timeout: 10 * time.Second}, host+":993", &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接 %s:993，请检查网络连接", host)
	}
	c.Timeout = 15 * time.Second
	if err = c.Login(config.AuthUser, password); err != nil {
		c.Terminate()
		return nil, nil, fmt.Errorf("IMAP 登录失败，请检查邮箱授权码、IMAP 服务和网易安全验证")
	}
	// NetEase requires a client identity before opening the mailbox.
	if _, err = imapid.NewClient(c).ID(imapid.ID{"name": "PXYMail", "version": "1.0", "vendor": "PXYMail"}); err != nil {
		c.Terminate()
		return nil, nil, fmt.Errorf("网易未接受客户端身份，请检查邮箱安全设置")
	}
	box, err := c.Select("INBOX", true)
	if err != nil {
		c.Terminate()
		return nil, nil, fmt.Errorf("网易拒绝读取收件箱，请确认 IMAP 已开启并完成客户端安全验证")
	}
	return c, box, nil
}
func Start(ctx context.Context) error {
	if !public.NativeMode() {
		return nil
	}
	if err := reconcile(ctx); err != nil {
		return err
	}
	workerOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if manualWaiting.Load() {
						continue
					}
					if err := reconcile(ctx); err != nil {
						g.Log().Error(ctx, "inbox account reconciliation failed", err)
						continue
					}
					var account *Account
					err := g.DB().Model("bm_native_inbox_accounts").Where("enabled", true).Where("next_poll<=?", time.Now().Unix()).Order("pending DESC,next_poll,relay_id").Limit(1).Scan(&account)
					if err != nil || account == nil {
						continue
					}
					pollCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
					err = syncAccount(pollCtx, account.ID)
					cancel()
					if err != nil {
						failAccount(ctx, account.ID, err)
					}
				}
			}
		}()
	})
	return nil
}
func decodeHeader(value string) string {
	decoded, err := (&mime.WordDecoder{CharsetReader: charset.Reader}).DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}
func syncAccount(ctx context.Context, accountID int) error {
	return syncAccountMode(ctx, accountID, false)
}
func syncAccountMode(ctx context.Context, accountID int, manual bool) error {
	release, err := acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	var current *Account
	if err = g.DB().Model("bm_native_inbox_accounts").Where("relay_id", accountID).Scan(&current); err != nil {
		return err
	}
	if manual {
		if err := manualReady(current); err != nil {
			return err
		}
	}
	if current == nil || !current.Enabled {
		return nil
	}
	now := time.Now().Unix()
	if !manual && current.NextPoll > now {
		return nil
	}
	status := "正在同步"
	if manual {
		status = "正在手动检查"
	}
	if _, err = g.DB().Model("bm_native_inbox_accounts").Where("relay_id", accountID).Data(g.Map{"status": status, "last_attempt": now, "next_poll": now + pollInterval, "pending": false}).Update(); err != nil {
		return err
	}
	c, box, err := openInbox(ctx, accountID)
	if err != nil {
		return err
	}
	defer func() { _ = c.Logout(); _ = c.Terminate() }()
	var previewError error
	if box.Messages > 0 {
		start := uint32(1)
		if box.Messages > 100 {
			start = box.Messages - 99
		}
		set := new(imap.SeqSet)
		set.AddRange(start, box.Messages)
		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate}, messages)
		}()
		uids := make([]uint32, 0, 100)
		var saveError error
		for msg := range messages {
			if msg.Envelope == nil || msg.Uid == 0 {
				continue
			}
			env := msg.Envelope
			address, name := "", ""
			if len(env.From) > 0 {
				address = strings.ToLower(env.From[0].Address())
				name = decodeHeader(env.From[0].PersonalName)
			}
			date := msg.InternalDate
			if date.IsZero() {
				date = env.Date
			}
			seen := false
			for _, flag := range msg.Flags {
				if flag == imap.SeenFlag {
					seen = true
				}
			}
			kind := classify(ctx, current.Email, address, decodeHeader(env.Subject), env.InReplyTo)
			_, e := g.DB().Exec(ctx, `INSERT INTO bm_native_inbox_messages(account_id,uid_validity,uid,from_address,from_name,subject,received_at,message_id,in_reply_to,kind,seen)
    VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(account_id,uid_validity,uid) DO UPDATE SET seen=EXCLUDED.seen,kind=EXCLUDED.kind`, accountID, box.UidValidity, msg.Uid, address, name, decodeHeader(env.Subject), date.Unix(), env.MessageId, env.InReplyTo, kind, seen)
			if e != nil && saveError == nil {
				saveError = e
			}
			uids = append(uids, msg.Uid)
		}
		if err = <-done; err != nil {
			return fmt.Errorf("读取邮件列表失败，请稍后重试")
		}
		if saveError != nil {
			return saveError
		}
		// Keep recent previews available without opening another IMAP connection.
		previews := make([]*Mail, 0, 10)
		for i := len(uids) - 1; i >= 0 && i >= len(uids)-10; i-- {
			var mail *Mail
			if err = g.DB().Model("bm_native_inbox_messages").Where("account_id=? AND uid_validity=? AND uid=?", accountID, box.UidValidity, uids[i]).Scan(&mail); err != nil {
				return err
			}
			if mail != nil && !mail.BodyCached {
				previews = append(previews, mail)
			}
		}
		previewError = cacheBodies(ctx, c, previews)
	}
	status = "连接正常"
	lastError := ""
	if previewError != nil {
		status, lastError = "列表已同步，部分正文待读取", previewError.Error()
	}
	_, err = g.DB().Model("bm_native_inbox_accounts").Where("relay_id", accountID).Data(g.Map{"status": status, "last_error": lastError, "last_success": time.Now().Unix(), "mailbox_total": box.Messages, "next_poll": time.Now().Unix() + pollInterval}).Update()
	return err
}
func cacheBody(ctx context.Context, c *client.Client, mail *Mail) error {
	return cacheBodies(ctx, c, []*Mail{mail})
}
func cacheBodies(ctx context.Context, c *client.Client, mails []*Mail) error {
	if len(mails) == 0 {
		return nil
	}
	set := new(imap.SeqSet)
	byUID := make(map[uint32]*Mail, len(mails))
	for _, mail := range mails {
		set.AddNum(uint32(mail.UID))
		byUID[uint32(mail.UID)] = mail
	}
	section := &imap.BodySectionName{Peek: true, Partial: []int{0, bodyLimit}}
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchRFC822Size, section.FetchItem()}, messages)
	}()
	var cacheError error
	fetched := 0
	for msg := range messages {
		mail := byUID[msg.Uid]
		if mail == nil {
			continue
		}
		reader := msg.GetBody(section)
		if reader == nil {
			continue
		}
		raw, readError := io.ReadAll(io.LimitReader(reader, bodyLimit))
		if readError != nil || len(raw) == 0 {
			cacheError = fmt.Errorf("部分正文为空或原邮件已移走")
			continue
		}
		textBody, htmlBody, attachments := parseBody(raw)
		names, _ := json.Marshal(attachments)
		_, saveError := g.DB().Model("bm_native_inbox_messages").Where("id", mail.ID).Data(g.Map{"body_cached": true, "body_text": textBody, "body_html": htmlBody, "attachments": string(names), "truncated": int(msg.Size) > bodyLimit}).Update()
		if saveError != nil {
			cacheError = fmt.Errorf("正文缓存保存失败，请查看数据库日志")
		} else {
			fetched++
		}
	}
	if err := <-done; err != nil {
		return fmt.Errorf("正文读取失败：%s", providerError(err))
	}
	if cacheError != nil {
		return cacheError
	}
	if fetched != len(mails) {
		return fmt.Errorf("部分正文未返回，原邮件可能已移走，请稍后重试")
	}
	return nil
}
func providerError(err error) string {
	value := strings.ReplaceAll(strings.ReplaceAll(err.Error(), "\r", " "), "\n", " ")
	if len(value) > 240 {
		value = value[:240]
	}
	return value
}
func parseBody(raw []byte) (string, string, []string) {
	reader, _ := messageMail.CreateReader(strings.NewReader(string(raw)))
	names := make([]string, 0)
	if reader == nil {
		return "无法解析此邮件，请打开原邮箱查看。", "", names
	}
	textParts, htmlParts := []string{}, []string{}
	for {
		part, err := reader.NextPart()
		if part == nil {
			break
		}
		switch header := part.Header.(type) {
		case *messageMail.InlineHeader:
			kind, _, _ := header.ContentType()
			body, _ := io.ReadAll(io.LimitReader(part.Body, bodyLimit))
			if kind == "text/plain" {
				textParts = append(textParts, string(body))
			}
			if kind == "text/html" {
				htmlParts = append(htmlParts, string(body))
			}
		case *messageMail.AttachmentHeader:
			name, _ := header.Filename()
			if name == "" {
				name = "未命名附件"
			}
			names = append(names, name)
		}
		if err != nil {
			break
		}
	}
	return strings.Join(textParts, "\n"), strings.Join(htmlParts, "\n"), names
}
func Detail(ctx context.Context, id int64) (*Mail, error) {
	mail, err := getMail(ctx, id)
	if err != nil || mail.BodyCached {
		return mail, err
	}
	// Pausing an account also prevents on-demand network reads.
	enabled, err := g.DB().Model("bm_native_inbox_accounts").Where("relay_id=? AND enabled=TRUE", mail.AccountID).Exist()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, fmt.Errorf("此账号已暂停收件，请先启用再读取未缓存的正文")
	}
	release, err := acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	c, box, err := openInbox(ctx, mail.AccountID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = c.Logout(); _ = c.Terminate() }()
	if int64(box.UidValidity) != mail.UIDValidity {
		return nil, fmt.Errorf("原收件箱已重建，请同步后打开新的邮件记录")
	}
	var recent []Mail
	if err = g.DB().Model("bm_native_inbox_messages").Where("account_id=? AND uid_validity=?", mail.AccountID, box.UidValidity).Order("uid DESC").Limit(10).Scan(&recent); err != nil {
		return nil, err
	}
	previews := []*Mail{mail}
	for i := range recent {
		if !recent[i].BodyCached && recent[i].ID != mail.ID {
			previews = append(previews, &recent[i])
		}
	}
	if err = cacheBodies(ctx, c, previews); err != nil {
		return nil, err
	}
	_, err = g.DB().Model("bm_native_inbox_accounts").Where("relay_id", mail.AccountID).Data(g.Map{"status": "连接正常", "last_error": "", "last_success": time.Now().Unix(), "mailbox_total": box.Messages, "last_attempt": time.Now().Unix(), "next_poll": time.Now().Unix() + pollInterval}).Update()
	if err != nil {
		return nil, err
	}
	return getMail(ctx, id)
}
