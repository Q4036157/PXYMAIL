// Import configuration, templates, and contact history without sending tasks.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/crypto/gaes"
	_ "github.com/lib/pq"
)

type profile struct {
	Type, Label, Host, Username, Password, From, Name, Security, Auth string
	Port                                                              int
	Enabled                                                           bool
}
type template struct {
	ID                        int `json:"id"`
	Name, Subject, HTML, Text string
	Published                 bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Import failed:", err)
		os.Exit(1)
	}
}

func run() error {
	root := "E:/pxy-runtime/Mail"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	data, err := os.ReadFile(root + "/.env")
	if err != nil {
		return err
	}
	cfg := map[string]string{}
	for _, row := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(strings.TrimSpace(row), "=", 2)
		if len(parts) == 2 {
			cfg[parts[0]] = parts[1]
		}
	}
	dsn := url.URL{Scheme: "postgres", Host: cfg["DBHOST"] + ":" + cfg["DBPORT"], Path: "/billionmail", User: url.UserPassword("billionmail", cfg["DBPASS"])}
	q := dsn.Query()
	q.Set("sslmode", "disable")
	dsn.RawQuery = q.Encode()
	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "ALTER TABLE bm_relay_config ADD COLUMN IF NOT EXISTS tls_protocol varchar(20) DEFAULT ''"); err != nil {
		return err
	}
	var keyHex string
	err = tx.QueryRowContext(ctx, "SELECT value FROM bm_options WHERE name='relay_encryption_key'").Scan(&keyHex)
	if err == sql.ErrNoRows {
		key := make([]byte, 16)
		if _, err = rand.Read(key); err != nil {
			return err
		}
		keyHex = hex.EncodeToString(key)
		_, err = tx.ExecContext(ctx, "INSERT INTO bm_options(name,value) VALUES('relay_encryption_key',$1)", keyHex)
	}
	if err != nil {
		return err
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) < 16 {
		return fmt.Errorf("invalid relay encryption key")
	}
	data, err = os.ReadFile("E:/pxy-runtime/YXHK/tenants/owner/mail-queue/profiles.json")
	if err != nil {
		return err
	}
	profiles := map[string]profile{}
	if err = json.Unmarshal(data, &profiles); err != nil {
		return err
	}
	ids := make([]string, 0, len(profiles))
	for id := range profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	activeCount := 0
	for _, id := range ids {
		p := profiles[id]
		p.From = strings.ToLower(strings.TrimSpace(p.From))
		if _, err = mail.ParseAddress(p.From); err != nil {
			return fmt.Errorf("invalid source account %s", id)
		}
		active := 0
		if p.Enabled && p.Password != "" && p.Auth == "password" && (p.Port == 465 || p.Port == 587) {
			active = 1
			activeCount++
		}
		password := ""
		if p.Password != "" {
			ciphertext, e := gaes.Encrypt([]byte(p.Password), key[:16])
			if e != nil {
				return e
			}
			password = hex.EncodeToString(ciphertext)
		}
		var relayID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM bm_relay_config WHERE auth_user=$1 ORDER BY id LIMIT 1", p.Username).Scan(&relayID)
		label := "YXHK " + p.From
		if err == sql.ErrNoRows {
			err = tx.QueryRowContext(ctx, `INSERT INTO bm_relay_config(remark,rtype,relay_host,relay_port,auth_user,auth_password,active,auth_method,tls_protocol,skip_tls_verify,smtp_name,create_time,update_time)
			VALUES($1,'custom',$2,$3,$4,$5,$6,'PLAIN',$7,0,$8,$9,$9) RETURNING id`, label, p.Host, fmt.Sprint(p.Port), p.Username, password, active, p.Security, "yxhk_"+id, time.Now().Unix()).Scan(&relayID)
		} else if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE bm_relay_config SET remark=$1,relay_host=$2,relay_port=$3,auth_password=$4,active=$5,auth_method='PLAIN',tls_protocol=$6,skip_tls_verify=0,update_time=$7 WHERE id=$8`, label, p.Host, fmt.Sprint(p.Port), password, active, p.Security, time.Now().Unix(), relayID)
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(p.From, "@", 2)
		_, err = tx.ExecContext(ctx, `INSERT INTO domain(domain,a_record,create_time) VALUES($1,$2,$3) ON CONFLICT(domain) DO NOTHING`, parts[1], "mail."+parts[1], time.Now().Unix())
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO mailbox(username,password,password_encode,full_name,maildir,local_part,domain,active,create_time,update_time)
		 VALUES($1,'EXTERNAL_SMTP','',$2,'',$3,$4,$5,$6,$6) ON CONFLICT(username) DO UPDATE SET full_name=EXCLUDED.full_name,active=EXCLUDED.active,update_time=EXCLUDED.update_time`, p.From, p.Name, parts[0], parts[1], active, time.Now().Unix())
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO bm_relay_domain_mapping(relay_id,sender_domain,create_time) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, relayID, "@"+parts[1], time.Now().Unix())
		if err != nil {
			return err
		}
	}
	data, err = os.ReadFile(root + "/migration-data/yxhk-templates.json")
	if err != nil {
		return err
	}
	var templates []template
	if err = json.Unmarshal(data, &templates); err != nil {
		return err
	}
	templateCount := 0
	metadata := map[string]any{}
	for _, t := range templates {
		content := t.HTML
		content = strings.ReplaceAll(content, "{unsubscribe_url}", "__UNSUBSCRIBE_URL__")
		if strings.TrimSpace(content) == "" && strings.TrimSpace(t.Text) != "" {
			content = "<div style=\"white-space:pre-wrap\">" + html.EscapeString(t.Text) + "</div>"
		}
		if strings.TrimSpace(content) == "" {
			continue
		}
		name := fmt.Sprintf("YXHK #%d %s", t.ID, t.Name)
		metadata[name] = map[string]any{"subject": t.Subject, "published": t.Published, "source_id": t.ID}
		_, err = tx.ExecContext(ctx, `INSERT INTO email_templates(temp_name,add_type,content,render,create_time,update_time) VALUES($1,0,$2,$2,$3,$3)
		ON CONFLICT(temp_name) DO UPDATE SET content=EXCLUDED.content,render=EXCLUDED.render,update_time=EXCLUDED.update_time`, name, content, time.Now().Unix())
		if err != nil {
			return err
		}
		templateCount++
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO bm_options(name,value) VALUES('yxhk_template_metadata',$1) ON CONFLICT(name) DO UPDATE SET value=EXCLUDED.value`, string(metadataJSON)); err != nil {
		return err
	}
	stateData, err := os.ReadFile("E:/pxy-runtime/YXHK/tenants/owner/mail-queue/state.json")
	if err != nil {
		return err
	}
	var state struct {
		Interval int `json:"interval"`
	}
	if err = json.Unmarshal(stateData, &state); err != nil {
		return err
	}
	if state.Interval <= 0 {
		state.Interval = 20
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO bm_options(name,value) VALUES('native_send_interval',$1) ON CONFLICT(name) DO UPDATE SET value=EXCLUDED.value`, fmt.Sprint(state.Interval)); err != nil {
		return err
	}
	contactCounts, err := importContacts(ctx, tx, root, stateData)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	report := map[string]any{"accounts": len(profiles), "enabled": activeCount, "disabled": len(profiles) - activeCount, "templates": templateCount, "sending_tasks_created": 0, "time": time.Now()}
	report["contact_groups"] = contactCounts
	output, _ := json.MarshalIndent(report, "", "  ")
	if err = os.WriteFile(root+"/migration-data/import-result.json", output, 0600); err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
