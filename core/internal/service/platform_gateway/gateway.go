package platform_gateway

import (
	platform "billionmail-core/internal/service/platform_session"
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type space struct {
	Email      string `json:"email"`
	HTTPPort   int    `json:"http_port"`
	RedisPort  int    `json:"redis_port"`
	Root       string `json:"root"`
	env        map[string]string
	app, redis *exec.Cmd
	done       chan struct{}
}
type gateway struct {
	root, exe, owner string
	settings         map[string]string
	db               *sql.DB
	jobs             *processGroup
	mu               sync.Mutex
	spaces           map[string]*space
	client           *http.Client
}

func readEnv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
		if len(parts) == 2 && !strings.HasPrefix(parts[0], "#") {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}
func writeEnv(path string, values map[string]string) error {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, k := range keys {
		out.WriteString(k + "=" + values[k] + "\n")
	}
	return os.WriteFile(path, []byte(out.String()), 0600)
}
func secret() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
func spaceKey(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:16])
}
func pgURL(values map[string]string, dbname string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(values["DBHOST"], values["DBPORT"]), Path: "/" + dbname, User: url.UserPassword(values["DBUSER"], values["DBPASS"])}
	q := u.Query()
	q.Set("sslmode", "disable")
	q.Set("connect_timeout", "5")
	u.RawQuery = q.Encode()
	return u.String()
}
func Run() error {
	settings, err := readEnv(public.EnvironmentFile())
	if err != nil {
		return err
	}
	if len(settings["MAIL_APP_SESSION_SECRET"]) < 32 || settings["MAIL_SPACE_EMAIL"] == "" || len(settings["MAIL_BOOTSTRAP_SECRET"]) < 32 {
		return errors.New("Mail platform login configuration is incomplete")
	}
	root := filepath.Dir(public.EnvironmentFile())
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	jobs, err := newProcessGroup()
	if err != nil {
		return err
	}
	defer jobs.close()
	db, err := sql.Open("postgres", pgURL(settings, "postgres"))
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(2)
	// Tenant roles must not inherit access to the legacy owner's database.
	if _, err = db.Exec(`REVOKE CONNECT ON DATABASE billionmail FROM PUBLIC`); err != nil {
		return errors.New("failed to secure owner database")
	}
	g := &gateway{root: root, exe: exe, owner: settings["MAIL_SPACE_EMAIL"], settings: settings, db: db, jobs: jobs, spaces: map[string]*space{}, client: &http.Client{Timeout: 70 * time.Second}}
	go g.restore()
	server := &http.Server{Addr: "127.0.0.1:3041", Handler: g, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 32768}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()
	log.Print("Mail platform gateway listening on loopback port 3041")
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func (g *gateway) restore() {
	entries, _ := os.ReadDir(filepath.Join(g.root, "users"))
	for _, entry := range entries {
		if !entry.IsDir() || len(entry.Name()) != 32 {
			continue
		}
		var saved space
		data, err := os.ReadFile(filepath.Join(g.root, "users", entry.Name(), "space.json"))
		if err != nil || json.Unmarshal(data, &saved) != nil || spaceKey(saved.Email) != entry.Name() {
			continue
		}
		if _, err = g.ensure(saved.Email); err != nil {
			log.Printf("workspace restore failed: key=%s", entry.Name())
		}
	}
}
func (g *gateway) identity(r *http.Request) (*platform.Claims, string, error) {
	raw := ""
	if cookie, err := r.Cookie("pxy_mail_session"); err == nil {
		raw = cookie.Value
	}
	// Only the bootstrap endpoint accepts a token from the URL fragment via POST.
	if r.URL.Path == "/api/platform/session" && r.Method == http.MethodPost {
		var body struct {
			PlatformToken string `json:"platformToken"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 32768)).Decode(&body); err != nil {
			return nil, "", err
		}
		if body.PlatformToken != "" {
			raw = body.PlatformToken
		}
	}
	claims, err := platform.Verify(raw)
	return claims, raw, err
}
func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path == "/health/native" {
		respond(w, 200, map[string]any{"status": "ok", "application": "BillionMail", "mode": "native", "auth": "platform_isolated"})
		return
	}
	identity, raw, err := g.identity(r)
	if err != nil {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			http.Redirect(w, r, "/apps/launch/mail", http.StatusFound)
		} else {
			respond(w, 401, map[string]any{"code": 401, "msg": "请从平台左侧的PXYMail 入口重新进入"})
		}
		return
	}
	// Require the ingress to retain its HttpOnly session cookie as the routing identity.
	cookie, cookieErr := r.Cookie("pxy_mail_session")
	if cookieErr != nil || cookie.Value != raw {
		respond(w, 401, map[string]any{"code": 401, "msg": "平台会话未生效，请重新进入 PXYMail"})
		return
	}
	s, err := g.ensure(identity.Email)
	if err != nil {
		log.Printf("workspace unavailable: key=%s", spaceKey(identity.Email))
		respond(w, 503, map[string]any{"code": 503, "msg": "个人邮箱空间暂不可用，请稍后重试"})
		return
	}
	if r.URL.Path == "/api/platform/session" {
		if r.Method != http.MethodPost {
			respond(w, 405, map[string]any{"code": 405})
			return
		}
		g.bootstrap(w, r, s, identity)
		return
	}
	if r.URL.Path == "/api/platform/sso" || r.URL.Path == "/api/aapanel/sso" || r.URL.Path == "/api/login" || r.URL.Path == "/api/get_validate_code" {
		respond(w, 403, map[string]any{"code": 403, "msg": "请使用平台统一登录"})
		return
	}
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", s.HTTPPort))
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Header.Del("X-PXY-User-ID")
		req.Header.Del("X-PXY-Is-Admin")
		req.Header.Del("X-PXY-User-Email")
		req.Header.Del("Cookie")
		req.Header.Set("X-Forwarded-Prefix", "/apps/mail")
	}
	proxy.ModifyResponse = func(res *http.Response) error {
		res.Header.Del("Set-Cookie")
		if strings.HasPrefix(r.URL.Path, "/static/") && res.StatusCode == http.StatusOK {
			res.Header.Set("Cache-Control", "private, max-age=31536000, immutable")
		} else {
			res.Header.Set("Cache-Control", "no-store")
		}
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		respond(w, 502, map[string]any{"code": 502, "msg": "个人邮箱进程正在恢复，请稍后重试"})
	}
	// The upstream response chooses the policy; do not append a conflicting no-store.
	w.Header().Del("Cache-Control")
	proxy.ServeHTTP(w, r)
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (g *gateway) bootstrap(w http.ResponseWriter, r *http.Request, s *space, identity *platform.Claims) {
	claims := *identity
	claims.SessionExpires = identity.ExpiresAt.Unix()
	claims.Issuer = "pxy-mail-gateway"
	claims.TokenUse = "workspace_bootstrap"
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(2 * time.Minute))
	ticket, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims).SignedString([]byte(s.env["MAIL_BOOTSTRAP_SECRET"]))
	if err != nil {
		respond(w, 503, map[string]any{"code": 503})
		return
	}
	payload, _ := json.Marshal(map[string]string{"platformToken": ticket})
	request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/api/platform/sso", s.HTTPPort), bytes.NewReader(payload))
	if err != nil {
		respond(w, 503, map[string]any{"code": 503})
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := g.client.Do(request)
	if err != nil {
		respond(w, 503, map[string]any{"code": 503, "msg": "个人邮箱登录失败"})
		return
	}
	defer response.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(response.Body, 32768))
}
func healthy(port int) bool {
	client := http.Client{Timeout: time.Second}
	response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health/native", port))
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == 200
}
func (g *gateway) ensure(email string) (*space, error) {
	if email == g.owner {
		return &space{Email: email, HTTPPort: 3040, Root: g.root, env: g.settings}, nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	key := spaceKey(email)
	if existing := g.spaces[key]; existing != nil {
		select {
		case <-existing.done:
			delete(g.spaces, key)
		default:
			if healthy(existing.HTTPPort) {
				return existing, nil
			}
			return nil, errors.New("workspace process unhealthy")
		}
	}
	root := filepath.Join(g.root, "users", key)
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	s := &space{Email: email, Root: root, done: make(chan struct{})}
	if data, err := os.ReadFile(filepath.Join(root, "space.json")); err == nil {
		if json.Unmarshal(data, s) != nil || s.Email != email || s.Root != root {
			return nil, errors.New("workspace identity mismatch")
		}
	} else if !os.IsNotExist(err) {
		return nil, errors.New("cannot read workspace metadata")
	} else {
		port, err := freePort()
		if err != nil {
			return nil, err
		}
		redisPort, err := freePort()
		if err != nil {
			return nil, err
		}
		for redisPort == port {
			redisPort, err = freePort()
			if err != nil {
				return nil, err
			}
		}
		s.HTTPPort = port
		s.RedisPort = redisPort
		values := map[string]string{"ADMIN_USERNAME": "workspace", "ADMIN_PASSWORD": secret(), "SafePath": "", "DBHOST": g.settings["DBHOST"], "DBPORT": g.settings["DBPORT"], "DBUSER": "bm_u_" + key, "DBNAME": "bm_u_" + key, "DBPASS": secret(), "REDIS_ADDRESS": fmt.Sprintf("127.0.0.1:%d", redisPort), "REDISDB": "1", "REDISPASS": secret(), "HTTP_PORT": strconv.Itoa(port), "BILLIONMAIL_HOSTNAME": "localhost", "TZ": "Asia/Shanghai", "FAIL2BAN_INIT": "n", "MAIL_SPACE_EMAIL": email, "MAIL_BOOTSTRAP_SECRET": secret()}
		if err = writeEnv(filepath.Join(root, ".env"), values); err != nil {
			return nil, err
		}
		data, _ := json.MarshalIndent(s, "", "  ")
		if err = os.WriteFile(filepath.Join(root, "space.json"), data, 0600); err != nil {
			return nil, err
		}
	}
	var err error
	s.env, err = readEnv(filepath.Join(root, ".env"))
	if err != nil {
		return nil, err
	}
	if s.env["MAIL_SPACE_EMAIL"] != email || s.env["DBNAME"] != "bm_u_"+key || s.env["DBUSER"] != "bm_u_"+key {
		return nil, errors.New("invalid workspace settings")
	}
	if err = g.prepareDatabase(s); err != nil {
		return nil, err
	}
	if err = g.copyAssets(s); err != nil {
		return nil, err
	}
	if err = g.startSpace(s); err != nil {
		return nil, err
	}
	g.spaces[key] = s
	return s, nil
}
func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
func (g *gateway) prepareDatabase(s *space) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	name := s.env["DBNAME"]
	var exists bool
	if err := g.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname=$1)", name).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := g.db.ExecContext(ctx, "CREATE ROLE "+pq.QuoteIdentifier(name)+" LOGIN PASSWORD "+pq.QuoteLiteral(s.env["DBPASS"])+" NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT"); err != nil {
			return errors.New("workspace database role creation failed")
		}
	}
	if err := g.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := g.db.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(name)+" OWNER "+pq.QuoteIdentifier(name)+" ENCODING 'UTF8'"); err != nil {
			return errors.New("workspace database creation failed")
		}
	}
	_, err := g.db.ExecContext(ctx, "REVOKE ALL ON DATABASE "+pq.QuoteIdentifier(name)+" FROM PUBLIC")
	return err
}
func (g *gateway) copyAssets(s *space) error {
	// Copy application assets only: no legacy uploads, sessions, caches or credentials.
	source := filepath.Join(g.root, "core")
	dest := filepath.Join(s.Root, "core")
	for _, folder := range []string{"manifest", "languages", "template", "resource", "public/dist", "public/html"} {
		src := filepath.Join(source, folder)
		dst := filepath.Join(dest, folder)
		err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			target := filepath.Join(dst, rel)
			if d.Type()&os.ModeSymlink != 0 {
				return errors.New("unexpected link in application assets")
			}
			if d.IsDir() {
				return os.MkdirAll(target, 0700)
			}
			input, err := os.Open(path)
			if err != nil {
				return err
			}
			defer input.Close()
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, input)
			closeErr := out.Close()
			if err != nil {
				return err
			}
			return closeErr
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *gateway) startSpace(s *space) error {
	redisRoot := filepath.Join(s.Root, "redis")
	if err := os.MkdirAll(redisRoot, 0700); err != nil {
		return err
	}
	conf := fmt.Sprintf("bind 127.0.0.1\nport %d\nrequirepass %s\nappendonly yes\nmaxmemory 16mb\nmaxmemory-policy noeviction\ndir %s\n", s.RedisPort, s.env["REDISPASS"], strconv.Quote(filepath.ToSlash(redisRoot)))
	confPath := filepath.Join(redisRoot, "redis.conf")
	if err := os.WriteFile(confPath, []byte(conf), 0600); err != nil {
		return err
	}
	redisExe := g.settings["MAIL_REDIS_EXE"]
	if redisExe == "" {
		redisExe = `C:\Program Files\Redis\redis-server.exe`
	}
	s.redis = exec.Command(redisExe, confPath)
	s.redis.Dir = redisRoot
	if err := g.launch(s.redis, filepath.Join(s.Root, "redis-process.log")); err != nil {
		return err
	}
	redisDone := make(chan struct{})
	go func() { _ = s.redis.Wait(); close(redisDone) }()
	s.app = exec.Command(g.exe)
	s.app.Dir = filepath.Join(s.Root, "core")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "BILLIONMAIL_ENV_FILE=") && !strings.HasPrefix(entry, "BILLIONMAIL_NATIVE=") && !strings.HasPrefix(entry, "GF_GCFG_FILE=") {
			s.app.Env = append(s.app.Env, entry)
		}
	}
	s.app.Env = append(s.app.Env, "BILLIONMAIL_NATIVE=1", "BILLIONMAIL_ENV_FILE="+filepath.Join(s.Root, ".env"))
	if err := g.launch(s.app, filepath.Join(s.Root, "application-process.log")); err != nil {
		_ = s.redis.Process.Kill()
		return err
	}
	go func() { _ = s.app.Wait(); _ = s.redis.Process.Kill(); close(s.done) }()
	for i := 0; i < 60; i++ {
		if healthy(s.HTTPPort) {
			if err := g.seedProviderDomains(s); err != nil {
				_ = s.app.Process.Kill()
				return err
			}
			return nil
		}
		select {
		case <-s.done:
			return errors.New("workspace exited during startup")
		case <-redisDone:
			_ = s.app.Process.Kill()
			return errors.New("workspace cache exited during startup")
		case <-time.After(time.Second):
		}
	}
	_ = s.app.Process.Kill()
	return errors.New("workspace startup timed out")
}
func (g *gateway) launch(cmd *exec.Cmd, path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	cmd.Stdout = file
	cmd.Stderr = file
	hideProcess(cmd)
	if err = cmd.Start(); err != nil {
		return err
	}
	if err = g.jobs.attach(cmd); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	return nil
}
func (g *gateway) seedProviderDomains(s *space) error {
	db, err := sql.Open("postgres", pgURL(s.env, s.env["DBNAME"]))
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, domain := range []string{"126.com", "163.com", "qq.com", "outlook.com"} {
		if _, err = db.ExecContext(ctx, `INSERT INTO domain(domain,active,create_time) VALUES($1,1,$2) ON CONFLICT(domain) DO NOTHING`, domain, time.Now().Unix()); err != nil {
			return err
		}
	}
	return nil
}
