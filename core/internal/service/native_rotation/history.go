package native_rotation

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Include accepted records and all records that must still block a new submission.
const recipientRecordsSQL = `SELECT lower(r.recipient) AS email,r.task_id,r.is_sent,t.subject,COALESCE(i.content_hash,'') AS content_hash FROM recipient_info r JOIN email_tasks t ON t.id=r.task_id LEFT JOIN bm_rotation_items i ON i.task_id=t.id LEFT JOIN bm_rotation_jobs j ON j.id=i.job_id WHERE (r.is_sent IN(1,2,3) OR (r.is_sent=0 AND t.active=1 AND t.task_process IN(0,1,3) AND (j.id IS NULL OR j.status<>'cancelled' OR i.state='dispatched')))`

type sentEntry struct {
	TaskID      int    `json:"task_id,omitempty"`
	LegacyKey   string `json:"legacy_key,omitempty"`
	Subject     string `json:"subject"`
	ContentHash string `json:"content_hash,omitempty"`
}

type RepeatRecipient struct {
	Email string `json:"email"`
	Kind  string `json:"kind"`
}

type RepeatSummary struct {
	Count        int               `json:"count"`
	Same         int               `json:"same"`
	Other        int               `json:"other"`
	Unknown      int               `json:"unknown"`
	PlannedTotal int               `json:"planned_total"`
	Recipients   []RepeatRecipient `json:"recipients"`
}

func (s *RepeatSummary) Add(email, kind string) {
	s.Count++
	s.Recipients = append(s.Recipients, RepeatRecipient{Email: email, Kind: kind})
	switch kind {
	case "same":
		s.Same++
	case "other":
		s.Other++
	default:
		s.Unknown++
	}
}

func historyKind(entries []sentEntry, subject, hash string) string {
	unknown := false
	for _, entry := range entries {
		if entry.Subject == subject && entry.ContentHash == hash {
			return "same"
		}
		if entry.Subject == "" || (entry.Subject == subject && entry.ContentHash == "") {
			unknown = true
		}
	}
	if unknown {
		return "unknown"
	}
	return "other"
}

func sortHistory(entries []sentEntry) {
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.TaskID != b.TaskID {
			return a.TaskID < b.TaskID
		}
		if a.LegacyKey != b.LegacyKey {
			return a.LegacyKey < b.LegacyKey
		}
		if a.Subject != b.Subject {
			return a.Subject < b.Subject
		}
		return a.ContentHash < b.ContentHash
	})
}

func legacySentHistory() (map[string][]sentEntry, error) {
	var state struct {
		Jobs map[string]struct {
			Subject    string `json:"subject"`
			Recipients []struct {
				Email  string `json:"email"`
				Status string `json:"status"`
			} `json:"recipients"`
		} `json:"jobs"`
	}
	path := filepath.Join(filepath.Dir(public.EnvironmentFile()), "configuration", "state.json")
	if err := readOptionalJSON(path, &state); err != nil {
		return nil, err
	}
	history := map[string][]sentEntry{}
	for key, job := range state.Jobs {
		for index, recipient := range job.Recipients {
			if recipient.Status == "sent" {
				email := strings.ToLower(strings.TrimSpace(recipient.Email))
				history[email] = append(history[email], sentEntry{LegacyKey: fmt.Sprintf("%s/%d", key, index), Subject: job.Subject})
			}
		}
	}
	return history, nil
}

func sentHistory(ctx context.Context) (map[string][]sentEntry, error) {
	history, err := legacySentHistory()
	if err != nil {
		return nil, err
	}
	rows, err := g.DB().GetAll(ctx, recipientRecordsSQL+" AND r.is_sent=1")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		email := row["email"].String()
		history[email] = append(history[email], sentEntry{TaskID: row["task_id"].Int(), Subject: row["subject"].String(), ContentHash: row["content_hash"].String()})
	}
	for _, entries := range history {
		sortHistory(entries)
	}
	return history, nil
}

func checkRecipientRecords(rows gdb.Result, legacy []sentEntry, cfg Config, email string) error {
	entries := append([]sentEntry{}, legacy...)
	for _, row := range rows {
		if row["is_sent"].Int() != 1 {
			return fmt.Errorf("客户存在待核对、失败或其他任务占用记录")
		}
		entries = append(entries, sentEntry{TaskID: row["task_id"].Int(), Subject: row["subject"].String(), ContentHash: row["content_hash"].String()})
	}
	sortHistory(entries)
	if len(entries) == 0 && cfg.AcceptedHistory[email] == "" {
		return nil
	}
	if !cfg.IncludeSent || cfg.AcceptedHistory[email] == "" || digest(entries) != cfg.AcceptedHistory[email] {
		return fmt.Errorf("发送历史已变化或尚未确认再次发送，请重新预览")
	}
	return nil
}
