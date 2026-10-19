package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func importContacts(ctx context.Context, tx *sql.Tx, root string, stateData []byte) (map[string]int, error) {
	type customer struct {
		Email                               string
		Attr                                map[string]any
		Sent, Bounce, Unsubscribe, Excluded bool
	}
	customers := map[string]*customer{}
	get := func(email string) *customer {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" || !strings.Contains(email, "@") {
			return nil
		}
		if customers[email] == nil {
			customers[email] = &customer{Email: email, Attr: map[string]any{}}
		}
		return customers[email]
	}
	data, err := os.ReadFile(root + "/migration-data/yxhk-contacts.json")
	if err != nil {
		return nil, err
	}
	var source struct {
		Contacts []struct {
			ID                         int
			Email, Firstname, Lastname string
		}
		DNC []struct {
			LeadID   int `json:"lead_id"`
			Reason   int
			Comments string
		} `json:"do_not_contact"`
	}
	if err = json.Unmarshal(data, &source); err != nil {
		return nil, err
	}
	ids := map[int]*customer{}
	for _, c := range source.Contacts {
		if p := get(c.Email); p != nil {
			ids[c.ID] = p
			p.Attr["yxhk_id"] = c.ID
			p.Attr["name"] = strings.TrimSpace(c.Firstname + " " + c.Lastname)
		}
	}
	for _, d := range source.DNC {
		if p := ids[d.LeadID]; p != nil {
			p.Attr["do_not_contact_reason"] = d.Reason
			p.Attr["do_not_contact_comments"] = d.Comments
			if d.Reason == 2 {
				p.Bounce = true
			} else {
				p.Unsubscribe = true
			}
		}
	}
	file, err := os.Open("C:/Users/Work/Downloads/4036157松鼠球 AI技术交流群_20260928_171227 (1).csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	headers := map[string]int{}
	for i, h := range rows[0] {
		headers[strings.TrimPrefix(h, "\ufeff")] = i
	}
	for pos, row := range rows[1:] {
		p := get(row[headers["email"]])
		if p == nil {
			continue
		}
		p.Attr["csv_position"] = pos + 1
		for _, key := range []string{"nickname", "last_sent_time", "role", "user_id", "join_time"} {
			if col, ok := headers[key]; ok {
				p.Attr[key] = row[col]
			}
		}
		role := row[headers["role"]]
		if role == "owner" || role == "admin" || role == "群主" || role == "管理员" {
			p.Excluded = true
		}
	}
	var state struct {
		Jobs map[string]struct {
			Name       string
			Recipients []map[string]any
		}
		Suppressed map[string]any
	}
	if err = json.Unmarshal(stateData, &state); err != nil {
		return nil, err
	}
	for id, j := range state.Jobs {
		for _, r := range j.Recipients {
			email, _ := r["email"].(string)
			p := get(email)
			if p == nil {
				continue
			}
			status, _ := r["status"].(string)
			if status == "sent" {
				p.Sent = true
			}
			if status == "bounced" {
				p.Bounce = true
			}
			history, _ := p.Attr["yxhk_history"].([]any)
			p.Attr["yxhk_history"] = append(history, map[string]any{"batch_id": id, "batch_name": j.Name, "record": r})
		}
	}
	for email, reason := range state.Suppressed {
		if p := get(email); p != nil {
			p.Bounce = true
			p.Attr["suppression"] = reason
		}
	}
	counts := map[string]int{}
	groups := map[string]int64{}
	now := time.Now().Unix()
	for _, p := range customers {
		category := "未发送待核对"
		if p.Sent {
			category = "已发送历史"
		}
		if p.Bounce {
			category = "退信拒发"
		}
		if p.Unsubscribe {
			category = "已退订拒发"
		}
		if p.Excluded {
			category = "管理员排除"
		}
		p.Attr["migration_status"] = category
		p.Attr["previously_sent"] = p.Sent
		p.Attr["bounced"] = p.Bounce
		p.Attr["unsubscribed"] = p.Unsubscribe
		name := "YXHK · " + category
		group, ok := groups[name]
		if !ok {
			err = tx.QueryRowContext(ctx, `INSERT INTO bm_contact_groups(name,description,create_time,update_time) VALUES($1,$2,$3,$3) ON CONFLICT(name) DO UPDATE SET update_time=EXCLUDED.update_time RETURNING id`, name, "YXHK迁入：历史标记见联系人属性。未发送记录需核对后再创建发送任务。", now).Scan(&group)
			if err != nil {
				return nil, err
			}
			groups[name] = group
		}
		attrs, _ := json.Marshal(p.Attr)
		// Remove only prior migration copies when a historical category changes.
		_, err = tx.ExecContext(ctx, `DELETE FROM bm_contacts c USING bm_contact_groups g WHERE c.group_id=g.id AND g.name LIKE 'YXHK · %' AND c.email=$1 AND c.group_id<>$2 AND c.attribs ? 'migration_status'`, p.Email, group)
		if err != nil {
			return nil, err
		}
		// Import does not infer subscription consent or create runnable tasks.
		active := 0
		if category == "未发送待核对" {
			active = 1
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO bm_contacts(email,group_id,active,status,attribs,create_time) VALUES($1,$2,$5,0,$3,$4) ON CONFLICT(group_id,email) DO UPDATE SET active=EXCLUDED.active,status=0,attribs=EXCLUDED.attribs`, p.Email, group, string(attrs), now, active)
		if err != nil {
			return nil, fmt.Errorf("contact import: %w", err)
		}
		counts[category]++
	}
	return counts, nil
}
