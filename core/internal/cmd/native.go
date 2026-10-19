package cmd

import "strings"

func nativeUnavailable(path string) bool {
	for _, endpoint := range []string{"/api/settings/set_system_config", "/api/settings/set_system_config_key",
		"/api/settings/set_ssl_config", "/api/settings/delete_ip_whitelist", "/api/settings/set_api_doc_swagger"} {
		if path == endpoint {
			return true
		}
	}
	for _, prefix := range []string{"/api/docker", "/api/services/", "/api/postfix_queue/",
		"/api/mail_bcc/", "/roundcube", "/rspamd"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
