package batch_mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
)

type nativeUnsubscribeKey struct{}

// Reuse the existing unsubscribe registry during the native migration.
func prepareNativeUnsubscribe(ctx context.Context, email, sender string) (string, error) {
	var configPath string
	if err := public.OptionsMgrInstance.GetOption(ctx, "native_legacy_unsubscribe_config", &configPath); err != nil || configPath == "" {
		base, err := url.Parse(domains.GetBaseURLBySender(sender))
		if err != nil || base.Scheme != "https" || base.Host == "" {
			return "", fmt.Errorf("native sending requires a public unsubscribe address")
		}
		return "", nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("unsubscribe configuration unavailable")
	}
	var config struct {
		URL string `json:"internal_url"`
		Key string `json:"api_key"`
	}
	if err = json.Unmarshal(data, &config); err != nil || len(config.Key) < 32 {
		return "", fmt.Errorf("invalid unsubscribe configuration")
	}
	body, _ := json.Marshal(map[string]string{"email": email})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(config.URL, "/")+"/internal/prepare", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+config.Key)
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 8 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("unsubscribe service unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unsubscribe service returned HTTP %d", response.StatusCode)
	}
	var result struct {
		Blocked *bool  `json:"blocked"`
		URL     string `json:"url"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&result); err != nil || result.Blocked == nil {
		return "", fmt.Errorf("invalid unsubscribe service response")
	}
	if *result.Blocked {
		return "", fmt.Errorf("recipient is unsubscribed; sending blocked")
	}
	parsed, err := url.Parse(result.URL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("unsubscribe service did not provide a public HTTPS address")
	}
	return result.URL, nil
}
