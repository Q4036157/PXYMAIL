package native_rotation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type RecipientInput struct {
	Total      int      `json:"total"`
	Emails     []string `json:"emails"`
	Duplicates int      `json:"duplicates"`
	Invalid    []string `json:"invalid"`
}

func parseRecipientText(text string) *RecipientInput {
	input := &RecipientInput{Emails: []string{}, Invalid: []string{}}
	seen := map[string]bool{}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(",;，；、\ufeff", r)
	})
	input.Total = len(parts)
	for _, part := range parts {
		email := strings.ToLower(part)
		a, err := mail.ParseAddress(email)
		at := strings.LastIndex(email, "@")
		if err != nil || a.Address != email || len(email) > 254 || at < 1 || !strings.Contains(email[at+1:], ".") || strings.ContainsAny(email, "<>\"") {
			input.Invalid = append(input.Invalid, part)
			continue
		}
		if seen[email] {
			input.Duplicates++
			continue
		}
		seen[email] = true
		input.Emails = append(input.Emails, email)
	}
	return input
}

func contactRoleBlocked(r gdb.Record) bool {
	var attrs map[string]interface{}
	_ = json.Unmarshal([]byte(r["attribs"].String()), &attrs)
	role := fmt.Sprint(attrs["role"])
	return role != "<nil>" && role != "" && role != "成员" && role != "member"
}

// Preview is read-only; contacts are saved only with the confirmed job.
func pastedRecipients(ctx context.Context, text string) (gdb.Result, *RecipientInput, error) {
	input := parseRecipientText(text)
	if len(input.Emails) == 0 || len(input.Emails) > 1000 {
		return nil, nil, fmt.Errorf("请粘贴 1–1000 个有效邮箱；当前识别 %d 个有效邮箱、%d 个无效项", len(input.Emails), len(input.Invalid))
	}
	known, err := g.DB().Model("bm_contacts c").Ctx(ctx).
		InnerJoin("bm_contact_groups cg", "cg.id=c.group_id").
		Fields("c.id,c.email,c.attribs,c.active,c.status,cg.double_optin").
		WhereIn("lower(c.email)", input.Emails).Order("c.id").All()
	if err != nil {
		return nil, nil, err
	}
	byEmail := map[string]gdb.Record{}
	for _, r := range known {
		email := strings.ToLower(strings.TrimSpace(r["email"].String()))
		prior := byEmail[email]
		if prior == nil || contactRoleBlocked(r) || (!contactRoleBlocked(prior) && !eligibleContact(prior) && eligibleContact(r)) {
			byEmail[email] = r
		}
	}
	rows := gdb.Result{}
	for _, email := range input.Emails {
		r := byEmail[email]
		if r == nil {
			r = gdb.Record{"id": g.NewVar(0), "email": g.NewVar(email), "attribs": g.NewVar(`{"role":"member"}`), "active": g.NewVar(1), "status": g.NewVar(1), "double_optin": g.NewVar(0)}
		}
		rows = append(rows, r)
	}
	return rows, input, nil
}
