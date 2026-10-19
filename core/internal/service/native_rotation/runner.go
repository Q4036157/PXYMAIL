package native_rotation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func tick(ctx context.Context) error {
	operations.Lock()
	defer operations.Unlock()
	jobs, err := g.DB().GetAll(ctx, `SELECT DISTINCT j.id FROM bm_rotation_jobs j LEFT JOIN bm_rotation_items i ON i.job_id=j.id WHERE j.status='running' OR i.state='dispatched'`)
	if err != nil {
		return err
	}
	for _, row := range jobs {
		if err = advance(ctx, row["id"].Int()); err != nil {
			return err
		}
	}
	return nil
}
func pauseError(ctx context.Context, id int, message string) error {
	_, err := g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Where("status", "running").Data(g.Map{"status": "paused", "error": message}).Update()
	return err
}
func advance(ctx context.Context, id int) error {
	job, items, err := Detail(ctx, id)
	if err != nil {
		return err
	}
	var cfg Config
	if json.Unmarshal([]byte(job.Config), &cfg) != nil {
		return pauseError(ctx, id, "任务配置损坏，请停止并重新创建")
	}
	now := time.Now().Unix()
	inflight := false
	changed := false
	for _, item := range items {
		if item.State != "dispatched" {
			continue
		}
		switch item.IsSent {
		case 1, 3:
			state := "accepted"
			if item.IsSent == 3 {
				state = "failed"
			}
			if _, err = g.DB().Model("bm_rotation_items").Ctx(ctx).Where("job_id", id).Where("sequence", item.Sequence).Data("state", state).Update(); err != nil {
				return err
			}
			if _, err = g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Data("next_send", now+int64(cfg.Interval)).Update(); err != nil {
				return err
			}
			changed = true
			if item.IsSent == 3 {
				if err = pauseError(ctx, id, fmt.Sprintf("第 %d 封（%s）发送失败，任务已暂停；失败记录不会自动重发", item.Sequence, item.Sender)); err != nil {
					return err
				}
			}
		default:
			inflight = true
			at, e := g.DB().Model("bm_rotation_items").Ctx(ctx).Where("job_id", id).Where("sequence", item.Sequence).Value("dispatched_at")
			if e != nil {
				return e
			}
			if now-at.Int64() > 300 {
				if err = pauseError(ctx, id, fmt.Sprintf("第 %d 封发送状态待核对；请先检查 SMTP 结果", item.Sequence)); err != nil {
					return err
				}
			}
		}
	}
	if inflight || changed || job.Status != "running" || job.NextSend > now {
		return nil
	}
	var next *Item
	for i := range items {
		if items[i].State == "pending" && items[i].IsSent == 0 {
			next = &items[i]
			break
		}
	}
	if next == nil {
		_, err = g.DB().Model("bm_rotation_jobs").Ctx(ctx).Where("id", id).Where("status", "running").Data("status", "completed").Update()
		return err
	}
	if err = validatePending(ctx, *next, cfg); err != nil {
		return pauseError(ctx, id, fmt.Sprintf("第 %d 封未发送：%s", next.Sequence, err))
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, e := tx.Exec("SELECT pg_advisory_xact_lock(714091)"); e != nil {
			return e
		}
		_, e := tx.Exec(`UPDATE email_tasks SET pause=0,task_process=0,start_time=?,update_time=? WHERE id=? AND EXISTS(SELECT 1 FROM bm_rotation_jobs j JOIN bm_rotation_items i ON i.job_id=j.id JOIN recipient_info r ON r.task_id=i.task_id WHERE j.id=? AND j.status='running' AND i.task_id=email_tasks.id AND i.state='pending' AND r.is_sent=0)`, now, now, next.TaskID, id)
		if e != nil {
			return e
		}
		_, e = tx.Exec(`UPDATE bm_rotation_items SET state='dispatched',dispatched_at=? WHERE job_id=? AND sequence=? AND state='pending' AND EXISTS(SELECT 1 FROM email_tasks t WHERE t.id=task_id AND t.pause=0 AND t.task_process=0)`, now, id, next.Sequence)
		return e
	})
}
func validatePending(ctx context.Context, item Item, cfg Config) error {
	n, err := g.DB().Model("bm_relay_config").Ctx(ctx).Where("active", 1).Where("lower(COALESCE(NULLIF(sender_email,''),auth_user))=?", item.Sender).Count()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("发件账号已停用或配置重复")
	}
	content, err := g.DB().Model("email_templates").Ctx(ctx).Where("id", item.TemplateID).Value("content")
	if err != nil {
		return err
	}
	if digest(content.String()) != item.ContentHash {
		return fmt.Errorf("模板内容已修改，请停止任务后重新预览创建")
	}
	contact, err := g.DB().GetOne(ctx, `SELECT c.active,c.status,c.attribs,cg.double_optin FROM bm_contacts c JOIN bm_contact_groups cg ON cg.id=c.group_id WHERE c.id=? AND lower(c.email)=?`, item.ContactID, item.Recipient)
	if err != nil {
		return err
	}
	if contact.IsEmpty() || !eligibleContact(contact) || contactRoleBlocked(contact) {
		return fmt.Errorf("客户记录已修改、退订或停用")
	}
	blocked, err := blockedRecipients(ctx)
	if err != nil {
		return err
	}
	// The current pending item is expected; other pending records are checked below.
	if reason := blocked[item.Recipient]; reason != "" && reason != "已在其他任务中" {
		return fmt.Errorf("客户已被排除：%s", reason)
	}
	records, err := g.DB().GetAll(ctx, recipientRecordsSQL+" AND lower(r.recipient)=? AND r.task_id<>?", item.Recipient, item.TaskID)
	if err != nil {
		return err
	}
	legacy, err := legacySentHistory()
	if err != nil {
		return err
	}
	if err = checkRecipientRecords(records, legacy[item.Recipient], cfg, item.Recipient); err != nil {
		return err
	}
	n, err = g.DB().GetCount(ctx, `SELECT COUNT(*) FROM (SELECT email FROM unsubscribe_records UNION SELECT recipient AS email FROM abnormal_recipient UNION SELECT email FROM bm_contacts WHERE active<>1) b WHERE lower(email)=?`, item.Recipient)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("客户已退订或被拒发")
	}
	registry, err := loadRegistry(ctx)
	if err != nil {
		return err
	}
	defer registry.client.CloseIdleConnections()
	allowed, err := registry.allowed(ctx, item.Recipient)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("客户已退订或被拒发")
	}
	return nil
}
