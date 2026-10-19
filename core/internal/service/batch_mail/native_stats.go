package batch_mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// NativeTaskStats reports SMTP outcomes without treating acceptance as delivery.
func NativeTaskStats(ctx context.Context, taskID int64, domain string, start, end int64) (map[string]interface{}, error) {
	if start < 0 || end < start {
		return nil, fmt.Errorf("invalid statistics time range")
	}
	if end-start > 31622400 {
		start = end - 31622400
	}
	query := g.DB().Model("recipient_info r").Ctx(ctx).
		LeftJoin("email_tasks t", "t.id = r.task_id").
		WhereIn("r.is_sent", []int{1, 3}).
		WhereGTE("r.sent_time", start).WhereLTE("r.sent_time", end)
	if taskID > 0 {
		query = query.Where("r.task_id", taskID)
	}
	if domain != "" {
		query = query.Where("LOWER(t.addresser) LIKE ?", "%@"+strings.ToLower(domain))
	}
	rows, err := query.Fields("r.is_sent, r.sent_time, t.addresser").Order("r.sent_time ASC").All()
	if err != nil {
		return nil, err
	}
	accepted, failed := 0, 0
	type day struct {
		Date     string `json:"date"`
		Accepted int    `json:"accepted"`
		Failed   int    `json:"failed"`
	}
	days := make([]*day, 0)
	byDate := make(map[string]*day)
	type account struct {
		Email    string `json:"email"`
		Accepted int    `json:"accepted"`
		Failed   int    `json:"failed"`
	}
	accounts := make([]*account, 0)
	byAccount := make(map[string]*account)
	for _, row := range rows {
		date := time.Unix(row["sent_time"].Int64(), 0).In(time.Local).Format("2006-01-02")
		bucket := byDate[date]
		if bucket == nil {
			bucket = &day{Date: date}
			byDate[date] = bucket
			days = append(days, bucket)
		}
		email := row["addresser"].String()
		accountBucket := byAccount[email]
		if accountBucket == nil {
			accountBucket = &account{Email: email}
			byAccount[email] = accountBucket
			accounts = append(accounts, accountBucket)
		}
		if row["is_sent"].Int() == 1 {
			accepted++
			bucket.Accepted++
			accountBucket.Accepted++
		} else {
			failed++
			bucket.Failed++
			accountBucket.Failed++
		}
	}
	trackingState := func(enabled bool) string {
		if !enabled {
			return "disabled"
		}
		return "unavailable"
	}
	openState, clickState := "unavailable", "unavailable"
	if taskID > 0 {
		task, queryErr := g.DB().Model("email_tasks").Ctx(ctx).Where("id", taskID).Fields("track_open, track_click").One()
		if queryErr != nil {
			return nil, queryErr
		}
		openState = trackingState(task["track_open"].Int() == 1)
		clickState = trackingState(task["track_click"].Int() == 1)
	}
	return map[string]interface{}{
		"source": "native_smtp", "accepted": accepted, "failed": failed,
		"delivered": nil, "opened": nil, "clicked": nil, "bounced": nil,
		"delivery_state": "unavailable", "bounce_state": "unavailable",
		"open_state":  openState,
		"click_state": clickState,
		"days":        days,
		"accounts":    accounts,
	}, nil
}
