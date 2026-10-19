package native_rotation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type TemplateChoice struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
}
type Config struct {
	Name            string            `json:"name"`
	Senders         []string          `json:"senders"`
	Templates       []TemplateChoice  `json:"templates"`
	GroupID         int               `json:"group_id"`
	RecipientSource string            `json:"recipient_source,omitempty"`
	RecipientText   string            `json:"recipient_text,omitempty"`
	IncludeSent     bool              `json:"include_sent,omitempty"`
	AcceptedHistory map[string]string `json:"accepted_history,omitempty"`
	PerSender       int               `json:"per_sender"`
	Interval        int               `json:"interval"`
	FullName        string            `json:"full_name"`
}
type Item struct {
	Sequence     int    `json:"sequence"`
	Recipient    string `json:"recipient"`
	ContactID    int    `json:"contact_id"`
	Sender       string `json:"sender"`
	TemplateID   int    `json:"template_id"`
	TemplateName string `json:"template_name"`
	Subject      string `json:"subject"`
	ContentHash  string `json:"content_hash"`
	TaskID       int    `json:"task_id"`
	State        string `json:"state"`
	IsSent       int    `json:"is_sent"`
	PreviousSent string `json:"previous_sent,omitempty"`
}
type Preview struct {
	Config   Config          `json:"config"`
	Items    []Item          `json:"items"`
	Skipped  map[string]int  `json:"skipped"`
	Hash     string          `json:"hash"`
	Input    *RecipientInput `json:"input,omitempty"`
	Repeated RepeatSummary   `json:"repeated"`
}
type Job struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Accepted  int    `json:"accepted"`
	Failed    int    `json:"failed"`
	CreatedAt int64  `json:"created_at"`
	Error     string `json:"error"`
	Config    string `json:"config"`
	NextSend  int64  `json:"next_send"`
}

var once sync.Once
var initErr error
var operations sync.Mutex

func Init(ctx context.Context) error {
	if !public.NativeMode() {
		return fmt.Errorf("轮换任务需要 Windows 原生运行模式")
	}
	once.Do(func() {
		_, initErr = g.DB().Exec(ctx, `CREATE TABLE IF NOT EXISTS bm_rotation_jobs (
   id SERIAL PRIMARY KEY, request_key VARCHAR(80) UNIQUE NOT NULL, name TEXT NOT NULL,
   status TEXT NOT NULL DEFAULT 'paused', config TEXT NOT NULL, total INTEGER NOT NULL,
   created_at BIGINT NOT NULL, next_send BIGINT NOT NULL DEFAULT 0, error TEXT NOT NULL DEFAULT '');
  CREATE TABLE IF NOT EXISTS bm_rotation_items (
   job_id INTEGER NOT NULL REFERENCES bm_rotation_jobs(id) ON DELETE CASCADE,
   sequence INTEGER NOT NULL, recipient TEXT NOT NULL, contact_id INTEGER NOT NULL,
   sender TEXT NOT NULL, template_id INTEGER NOT NULL, template_name TEXT NOT NULL,
   subject TEXT NOT NULL, content_hash TEXT NOT NULL, task_id INTEGER UNIQUE NOT NULL REFERENCES email_tasks(id) ON DELETE CASCADE,
   state TEXT NOT NULL DEFAULT 'pending', dispatched_at BIGINT NOT NULL DEFAULT 0,
   PRIMARY KEY(job_id,sequence));
  CREATE INDEX IF NOT EXISTS bm_rotation_items_recipient ON bm_rotation_items(lower(recipient));`)
	})
	return initErr
}

func Start(ctx context.Context) error {
	if !public.NativeMode() {
		return nil
	}
	if err := Init(ctx); err != nil {
		return err
	}
	// Never automatically replay a record interrupted around SMTP submission.
	if _, err := g.DB().Exec(ctx, `UPDATE bm_rotation_jobs SET status='paused',error='程序重启，任务已暂停；请核对发送状态后继续' WHERE status='running';
 UPDATE email_tasks SET pause=1,task_process=3 WHERE id IN (SELECT i.task_id FROM bm_rotation_items i JOIN recipient_info r ON r.task_id=i.task_id WHERE r.is_sent IN(0,2)) AND task_process IN(0,1);
 UPDATE bm_rotation_items SET state='pending',dispatched_at=0 WHERE state='dispatched' AND task_id IN(SELECT task_id FROM recipient_info WHERE is_sent=0);`); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := tick(context.Background()); err != nil {
					g.Log().Error(context.Background(), "轮换任务调度失败: ", err)
				}
			}
		}
	}()
	return nil
}

func Options(ctx context.Context) (g.Map, error) {
	if err := Init(ctx); err != nil {
		return nil, err
	}
	accounts, err := g.DB().GetAll(ctx, `SELECT MIN(id) AS id,lower(COALESCE(NULLIF(sender_email,''),auth_user)) AS email FROM bm_relay_config WHERE active=1 AND COALESCE(NULLIF(sender_email,''),auth_user)<>'' GROUP BY lower(COALESCE(NULLIF(sender_email,''),auth_user)) ORDER BY MIN(id)`)
	if err != nil {
		return nil, err
	}
	templates, err := g.DB().GetAll(ctx, `SELECT id,temp_name FROM email_templates ORDER BY id`)
	if err != nil {
		return nil, err
	}
	groups, err := g.DB().GetAll(ctx, `SELECT g.id,g.name,COUNT(c.id) AS count FROM bm_contact_groups g LEFT JOIN bm_contacts c ON c.group_id=g.id GROUP BY g.id ORDER BY g.id`)
	if err != nil {
		return nil, err
	}
	return g.Map{"accounts": accounts, "templates": templates, "groups": groups}, nil
}

func digest(value interface{}) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalize(cfg *Config) error {
	// History acknowledgements are generated from current records, never client input.
	cfg.AcceptedHistory = nil
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.FullName = strings.TrimSpace(cfg.FullName)
	if cfg.Name == "" || len(cfg.Name) > 200 || len(cfg.FullName) > 200 || strings.ContainsAny(cfg.FullName, "\r\n") {
		return fmt.Errorf("请填写任务名称（不超过 200 字节）和有效的发件人显示名")
	}
	if len(cfg.Senders) < 1 || len(cfg.Senders) > 1000 || len(cfg.Templates) < 1 || len(cfg.Templates) > 20 || cfg.PerSender < 1 || cfg.PerSender > 100 || len(cfg.Senders)*cfg.PerSender > 1000 || cfg.Interval < 20 || cfg.Interval > 3600 {
		return fmt.Errorf("请选择账号和模板；每账号上限 1–100 封，总上限不超过 1000，每封间隔 20–3600 秒")
	}
	if cfg.RecipientSource == "" {
		cfg.RecipientSource = "group"
	}
	switch cfg.RecipientSource {
	case "group":
		cfg.RecipientText = ""
		if cfg.GroupID < 1 {
			return fmt.Errorf("请选择客户分组，或切换到直接粘贴邮箱")
		}
	case "paste":
		cfg.GroupID = 0
		cfg.RecipientText = strings.TrimSpace(cfg.RecipientText)
		if cfg.RecipientText == "" || len(cfg.RecipientText) > 100000 {
			return fmt.Errorf("请粘贴客户邮箱，内容不超过 100 KB")
		}
	default:
		return fmt.Errorf("客户来源无效，请选择客户分组或直接粘贴邮箱")
	}
	seen := map[string]bool{}
	for i, email := range cfg.Senders {
		email = strings.ToLower(strings.TrimSpace(email))
		a, err := mail.ParseAddress(email)
		if err != nil || a.Address != email || seen[email] {
			return fmt.Errorf("发件邮箱无效或重复")
		}
		cfg.Senders[i] = email
		seen[email] = true
	}
	ids := map[int]bool{}
	for i, t := range cfg.Templates {
		t.Subject = strings.TrimSpace(t.Subject)
		if t.ID < 1 || ids[t.ID] || t.Subject == "" || len(t.Subject) > 500 || strings.ContainsAny(t.Subject, "\r\n") {
			return fmt.Errorf("每个模板必须填写邮件主题，且模板不能重复")
		}
		cfg.Templates[i] = t
		ids[t.ID] = true
	}
	return nil
}

func BuildPreview(ctx context.Context, cfg Config) (*Preview, error) {
	if err := Init(ctx); err != nil {
		return nil, err
	}
	if err := normalize(&cfg); err != nil {
		return nil, err
	}
	p := &Preview{Config: cfg, Items: []Item{}, Skipped: map[string]int{}}
	for _, sender := range cfg.Senders {
		n, err := g.DB().Model("bm_relay_config").Ctx(ctx).Where("active", 1).Where("lower(COALESCE(NULLIF(sender_email,''),auth_user))=?", sender).Count()
		if err != nil {
			return nil, err
		}
		if n != 1 {
			return nil, fmt.Errorf("%s 未配置、已停用或配置重复，请先在 SMTP 页面处理", sender)
		}
	}
	type tpl struct {
		Name string
		Hash string
	}
	tpls := map[int]tpl{}
	for _, t := range cfg.Templates {
		r, err := g.DB().GetOne(ctx, "SELECT temp_name,content FROM email_templates WHERE id=?", t.ID)
		if err != nil {
			return nil, err
		}
		if r.IsEmpty() || r["content"].String() == "" {
			return nil, fmt.Errorf("模板 %d 不存在或正文为空", t.ID)
		}
		tpls[t.ID] = tpl{r["temp_name"].String(), digest(r["content"].String())}
	}
	blocked, err := blockedRecipients(ctx)
	if err != nil {
		return nil, err
	}
	history, err := sentHistory(ctx)
	if err != nil {
		return nil, err
	}
	var rows gdb.Result
	if cfg.RecipientSource == "paste" {
		rows, p.Input, err = pastedRecipients(ctx, cfg.RecipientText)
		if p.Input != nil {
			if p.Input.Duplicates > 0 {
				p.Skipped["名单重复"] = p.Input.Duplicates
			}
			if len(p.Input.Invalid) > 0 {
				p.Skipped["邮箱无效"] = len(p.Input.Invalid)
			}
		}
	} else {
		rows, err = g.DB().GetAll(ctx, `SELECT c.id,c.email,c.attribs,c.active,c.status,cg.double_optin FROM bm_contacts c JOIN bm_contact_groups cg ON cg.id=c.group_id WHERE c.group_id=? ORDER BY CASE WHEN c.attribs->>'csv_position' ~ '^[0-9]{1,12}$' THEN (c.attribs->>'csv_position')::bigint ELSE c.id END,c.id`, cfg.GroupID)
	}
	if err != nil {
		return nil, err
	}
	registry, err := loadRegistry(ctx)
	if err != nil {
		return nil, err
	}
	defer registry.client.CloseIdleConnections()
	seen := map[string]bool{}
	target := len(cfg.Senders) * cfg.PerSender
	for _, r := range rows {
		email := strings.ToLower(strings.TrimSpace(r["email"].String()))
		reason := ""
		if contactRoleBlocked(r) || blocked[email] == "管理员或发件账号" {
			reason = "管理员或发件账号"
		} else if !eligibleContact(r) {
			reason = "退订或未确认"
		} else if blocked[email] != "" {
			reason = blocked[email]
		} else if seen[email] {
			reason = "名单重复"
		} else if a, e := mail.ParseAddress(email); e != nil || a.Address != email {
			reason = "邮箱无效"
		}
		if reason != "" {
			p.Skipped[reason]++
			continue
		}
		seen[email] = true
		allowed, e := registry.allowed(ctx, email)
		if e != nil {
			return nil, e
		}
		if !allowed {
			p.Skipped["退订或拒发"]++
			continue
		}
		entries := history[email]
		if p.Repeated.PlannedTotal < target {
			choice := cfg.Templates[p.Repeated.PlannedTotal%len(cfg.Templates)]
			p.Repeated.PlannedTotal++
			if len(entries) > 0 {
				p.Repeated.Add(email, historyKind(entries, choice.Subject, tpls[choice.ID].Hash))
			}
		}
		if len(entries) > 0 && !cfg.IncludeSent {
			p.Skipped["已发送（本次跳过）"]++
			continue
		}
		if len(p.Items) >= target {
			p.Skipped["超出本次发送上限"]++
			continue
		}
		i := len(p.Items)
		choice := cfg.Templates[i%len(cfg.Templates)]
		t := tpls[choice.ID]
		item := Item{Sequence: i + 1, Recipient: email, ContactID: r["id"].Int(), Sender: cfg.Senders[i%len(cfg.Senders)], TemplateID: choice.ID, TemplateName: t.Name, Subject: choice.Subject, ContentHash: t.Hash, State: "pending"}
		if len(entries) > 0 {
			item.PreviousSent = historyKind(entries, choice.Subject, t.Hash)
			if p.Config.AcceptedHistory == nil {
				p.Config.AcceptedHistory = map[string]string{}
			}
			p.Config.AcceptedHistory[email] = digest(entries)
		}
		p.Items = append(p.Items, item)
		if len(p.Items) == target && cfg.RecipientSource != "paste" {
			break
		}
	}
	if len(p.Items) == 0 && p.Repeated.Count == 0 {
		return nil, fmt.Errorf("本次名单没有可发送客户；请检查邮箱格式、已发送、退订、退信及其他任务占用记录")
	}
	p.Hash = digest(g.Map{"config": p.Config, "items": p.Items})
	return p, nil
}

func Create(ctx context.Context, cfg Config, hash, key string) (int, error) {
	if err := Init(ctx); err != nil {
		return 0, err
	}
	if len(key) < 16 || len(key) > 80 {
		return 0, fmt.Errorf("请求标识无效，请重新预览")
	}
	operations.Lock()
	defer operations.Unlock()
	prior, err := g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("request_key", key).Value("id")
	if err != nil {
		return 0, err
	}
	if !prior.IsEmpty() {
		return prior.Int(), nil
	}
	p, err := BuildPreview(ctx, cfg)
	if err != nil {
		return 0, err
	}
	if p.Hash != hash {
		return 0, fmt.Errorf("名单、配置或模板已变化，请重新预览")
	}
	if len(p.Items) == 0 {
		return 0, fmt.Errorf("没有未发送客户；如需再次发送，请重新预览并确认包含已发送客户")
	}
	if p.Skipped["超出本次发送上限"] > 0 {
		return 0, fmt.Errorf("粘贴名单还有 %d 位可发送客户未分配，请提高每账号上限后重新预览", p.Skipped["超出本次发送上限"])
	}
	id := 0
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, e := tx.Exec("SELECT pg_advisory_xact_lock(714091)"); e != nil {
			return e
		}
		legacy, e := legacySentHistory()
		if e != nil {
			return e
		}
		for _, item := range p.Items {
			records, e := tx.GetAll(recipientRecordsSQL+" AND lower(r.recipient)=?", item.Recipient)
			if e != nil {
				return e
			}
			if e = checkRecipientRecords(records, legacy[item.Recipient], p.Config, item.Recipient); e != nil {
				return fmt.Errorf("客户发送状态已变化，请重新预览：%s", e)
			}
		}
		configJSON, _ := json.Marshal(p.Config)
		now := time.Now().Unix()
		v, e := tx.Model("bm_rotation_jobs").Data(g.Map{"request_key": key, "name": p.Config.Name, "config": string(configJSON), "total": len(p.Items), "created_at": now}).InsertAndGetId()
		if e != nil {
			return e
		}
		id = int(v)
		if p.Config.RecipientSource == "paste" {
			groupID, e := tx.Model("bm_contact_groups").Data(g.Map{"name": fmt.Sprintf("手动粘贴 · 轮换任务 %d", id), "description": "直接粘贴名单，在确认任务时保存；排除记录未导入。", "create_time": now, "update_time": now, "double_optin": 0}).InsertAndGetId()
			if e != nil {
				return e
			}
			p.Config.GroupID = int(groupID)
			configJSON, _ = json.Marshal(p.Config)
			if _, e = tx.Model("bm_rotation_jobs").Where("id", id).Data("config", string(configJSON)).Update(); e != nil {
				return e
			}
		}
		for _, item := range p.Items {
			if p.Config.RecipientSource == "paste" {
				attrs, _ := json.Marshal(g.Map{"role": "member", "source": "manual_paste", "csv_position": item.Sequence, "rotation_job_id": id})
				contactID, e := tx.Model("bm_contacts").Data(g.Map{"email": item.Recipient, "group_id": p.Config.GroupID, "active": 1, "status": 1, "attribs": string(attrs), "create_time": now}).InsertAndGetId()
				if e != nil {
					return e
				}
				item.ContactID = int(contactID)
			}
			taskID, e := tx.Model("email_tasks").Data(g.Map{"task_name": fmt.Sprintf("rotation_%d_%d", id, item.Sequence), "addresser": item.Sender, "subject": item.Subject, "full_name": p.Config.FullName, "recipient_count": 1, "task_process": 3, "pause": 1, "template_id": item.TemplateID, "is_record": 1, "unsubscribe": 1, "threads": 1, "track_open": 0, "track_click": 0, "start_time": 2147483647, "create_time": now, "update_time": now, "active": 1, "remark": fmt.Sprintf("轮换任务 %s · %d/%d", p.Config.Name, item.Sequence, len(p.Items)), "group_id": p.Config.GroupID}).InsertAndGetId()
			if e != nil {
				return e
			}
			if _, e = tx.Model("recipient_info").Data(g.Map{"task_id": taskID, "recipient": item.Recipient, "message_id": "", "is_sent": 0}).Insert(); e != nil {
				return e
			}
			if _, e = tx.Model("bm_rotation_items").Data(g.Map{"job_id": id, "sequence": item.Sequence, "recipient": item.Recipient, "contact_id": item.ContactID, "sender": item.Sender, "template_id": item.TemplateID, "template_name": item.TemplateName, "subject": item.Subject, "content_hash": item.ContentHash, "task_id": taskID}).Insert(); e != nil {
				return e
			}
		}
		return nil
	})
	return id, err
}

func List(ctx context.Context) ([]Job, error) {
	if err := Init(ctx); err != nil {
		return nil, err
	}
	var jobs = []Job{}
	err := g.DB().Raw(`SELECT j.*,COALESCE(SUM(CASE WHEN r.is_sent=1 THEN 1 ELSE 0 END),0) AS accepted,COALESCE(SUM(CASE WHEN r.is_sent=3 THEN 1 ELSE 0 END),0) AS failed FROM bm_rotation_jobs j LEFT JOIN bm_rotation_items i ON i.job_id=j.id LEFT JOIN recipient_info r ON r.task_id=i.task_id GROUP BY j.id ORDER BY j.id DESC LIMIT 100`).Ctx(ctx).Scan(&jobs)
	return jobs, err
}
func Detail(ctx context.Context, id int) (*Job, []Item, error) {
	if err := Init(ctx); err != nil {
		return nil, nil, err
	}
	var job *Job
	err := g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Scan(&job)
	if err != nil {
		return nil, nil, err
	}
	if job == nil {
		return nil, nil, fmt.Errorf("轮换任务不存在")
	}
	items := []Item{}
	err = g.DB().Raw(`SELECT i.*,r.is_sent FROM bm_rotation_items i JOIN recipient_info r ON r.task_id=i.task_id WHERE i.job_id=? ORDER BY i.sequence`, id).Ctx(ctx).Scan(&items)
	var cfg Config
	if json.Unmarshal([]byte(job.Config), &cfg) == nil && cfg.IncludeSent {
		for i := range items {
			if cfg.AcceptedHistory[items[i].Recipient] != "" {
				items[i].PreviousSent = "confirmed"
			}
		}
	}
	return job, items, err
}

func Action(ctx context.Context, id int, action string) error {
	operations.Lock()
	defer operations.Unlock()
	job, items, err := Detail(ctx, id)
	if err != nil {
		return err
	}
	if action == "pause" {
		_, err = g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Where("status", "running").Data("status", "paused").Update()
		return err
	}
	if action == "stop" {
		_, err = g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).WhereIn("status", []string{"running", "paused"}).Data("status", "cancelled").Update()
		return err
	}
	if action != "start" {
		return fmt.Errorf("操作无效")
	}
	if job.Status != "paused" {
		return fmt.Errorf("仅待发送或已暂停任务可以开始")
	}
	var cfg Config
	if json.Unmarshal([]byte(job.Config), &cfg) != nil {
		return fmt.Errorf("任务配置损坏")
	}
	pending := map[string]bool{}
	for _, i := range items {
		if i.IsSent == 2 {
			return fmt.Errorf("存在待核对的 SMTP 记录，为防重复发送，请先核对后处理")
		}
		if i.IsSent == 0 {
			pending[i.Sender] = true
		}
	}
	if len(pending) == 0 {
		return fmt.Errorf("没有待发送记录，失败记录不会自动重发")
	}
	// Validate account authentication without submitting any message.
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []string
	slots := make(chan struct{}, 3)
	for _, email := range cfg.Senders {
		if !pending[email] {
			continue
		}
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			s, e := mail_service.NewEmailSenderWithLocal(email)
			if e == nil {
				e = s.Connect()
				s.Close()
			}
			if e != nil {
				mu.Lock()
				failures = append(failures, email+"：SMTP 验证失败，请检查授权码或安全验证")
				mu.Unlock()
			}
		}(email)
	}
	wg.Wait()
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "；"))
	}
	_, err = g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Where("status", "paused").Data(g.Map{"status": "running", "error": ""}).Update()
	return err
}
