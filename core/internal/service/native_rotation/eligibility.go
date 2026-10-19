package native_rotation

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Migrated single-opt-in candidates are reviewed in the explicit assignment preview.
func eligibleContact(r gdb.Record) bool {
	if r["active"].Int() != 1 {
		return false
	}
	if r["status"].Int() == 1 {
		return true
	}
	var attrs map[string]interface{}
	if json.Unmarshal([]byte(r["attribs"].String()), &attrs) != nil {
		return false
	}
	return r["status"].Int() == 0 && r["double_optin"].Int() == 0 && attrs["migration_status"] == "未发送待核对" && attrs["role"] == "成员"
}

func readOptionalJSON(path string, value interface{}) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取历史筛选记录失败")
	}
	if json.Unmarshal(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), value) != nil {
		return fmt.Errorf("历史筛选记录格式错误")
	}
	return nil
}

func blockedRecipients(ctx context.Context) (map[string]string, error) {
	blocked := map[string]string{}
	rows, err := g.DB().GetAll(ctx, `SELECT lower(r.recipient) AS email,CASE WHEN r.is_sent=2 THEN '发送状态待核对' WHEN r.is_sent=3 THEN '发送失败或退信' ELSE '已在其他任务中' END AS reason FROM recipient_info r JOIN email_tasks t ON t.id=r.task_id LEFT JOIN bm_rotation_items i ON i.task_id=t.id LEFT JOIN bm_rotation_jobs j ON j.id=i.job_id WHERE r.is_sent IN(2,3) OR (r.is_sent=0 AND t.active=1 AND t.task_process IN(0,1,3) AND (j.id IS NULL OR j.status<>'cancelled' OR i.state='dispatched'))`)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		blocked[r["email"].String()] = r["reason"].String()
	}
	rows, err = g.DB().GetAll(ctx, `SELECT lower(email) AS email FROM unsubscribe_records UNION SELECT lower(recipient) AS email FROM abnormal_recipient UNION SELECT lower(email) AS email FROM bm_contacts WHERE active<>1`)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		blocked[r["email"].String()] = "退订或拒发"
	}
	rows, err = g.DB().GetAll(ctx, `SELECT lower(COALESCE(NULLIF(sender_email,''),auth_user)) AS email FROM bm_relay_config WHERE COALESCE(NULLIF(sender_email,''),auth_user)<>''`)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		blocked[r["email"].String()] = "管理员或发件账号"
	}
	root := filepath.Dir(public.EnvironmentFile())
	configRoot := filepath.Join(root, "configuration")
	var state struct {
		Suppressed map[string]json.RawMessage `json:"suppressed"`
		Jobs       map[string]struct {
			Status     string `json:"status"`
			Recipients []struct {
				Email  string `json:"email"`
				Status string `json:"status"`
			} `json:"recipients"`
		} `json:"jobs"`
	}
	if err = readOptionalJSON(filepath.Join(configRoot, "state.json"), &state); err != nil {
		return nil, err
	}
	for email := range state.Suppressed {
		blocked[strings.ToLower(email)] = "退信或退订"
	}
	for _, job := range state.Jobs {
		for _, r := range job.Recipients {
			if r.Status != "sent" && (r.Status != "pending" || job.Status != "paused") {
				blocked[strings.ToLower(r.Email)] = "旧项目已有发送记录"
			}
		}
	}
	paths, err := filepath.Glob(filepath.Join(root, "recipient-lists", "*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		var list struct {
			Status     string            `json:"status"`
			Recipients []json.RawMessage `json:"recipients"`
		}
		if err = readOptionalJSON(path, &list); err != nil {
			return nil, err
		}
		if list.Status == "cancelled" || list.Status == "released" {
			continue
		}
		for _, entry := range list.Recipients {
			var email string
			if json.Unmarshal(entry, &email) != nil {
				var recipient struct {
					Email string `json:"email"`
				}
				if json.Unmarshal(entry, &recipient) != nil {
					return nil, fmt.Errorf("保留名单格式错误")
				}
				email = recipient.Email
			}
			blocked[strings.ToLower(email)] = "已保留或已发送名单"
		}
	}
	file, err := os.Open(filepath.Join(configRoot, "original-customers.csv"))
	if os.IsNotExist(err) {
		return blocked, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取原名单角色失败")
	}
	defer file.Close()
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("原名单 CSV 格式错误")
	}
	emailCol, roleCol := -1, -1
	for i, s := range header {
		switch strings.TrimPrefix(s, "\ufeff") {
		case "email":
			emailCol = i
		case "role":
			roleCol = i
		}
	}
	if emailCol < 0 || roleCol < 0 {
		return blocked, nil
	}
	for {
		row, e := reader.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("原名单 CSV 格式错误")
		}
		role := strings.TrimSpace(row[roleCol])
		if role != "" && role != "成员" && role != "member" {
			blocked[strings.ToLower(strings.TrimSpace(row[emailCol]))] = "管理员或发件账号"
		}
	}
	return blocked, nil
}

type registryClient struct {
	URL    string `json:"internal_url"`
	Key    string `json:"api_key"`
	client *http.Client
}

func loadRegistry(ctx context.Context) (*registryClient, error) {
	var path string
	if err := public.OptionsMgrInstance.GetOption(ctx, "native_legacy_unsubscribe_config", &path); err != nil {
		return nil, fmt.Errorf("读取退订设置失败")
	}
	r := &registryClient{client: &http.Client{Timeout: 8 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	if path == "" {
		return r, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("退订配置不可用")
	}
	if json.Unmarshal(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), r) != nil || len(r.Key) < 32 || r.URL == "" {
		return nil, fmt.Errorf("退订配置无效")
	}
	return r, nil
}
func (r *registryClient) allowed(ctx context.Context, email string) (bool, error) {
	if r.URL == "" {
		return true, nil
	}
	body, _ := json.Marshal(map[string]string{"email": email})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.URL, "/")+"/internal/check", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("退订服务地址无效")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.Key)
	res, err := r.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("退订服务暂不可用，请稍后重试")
	}
	defer res.Body.Close()
	var result struct {
		Blocked *bool `json:"blocked"`
	}
	if res.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(res.Body, 65536)).Decode(&result) != nil || result.Blocked == nil {
		return false, fmt.Errorf("退订服务未返回有效检查结果")
	}
	return !*result.Blocked, nil
}
