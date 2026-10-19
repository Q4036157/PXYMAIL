package public

import "os"

// NativeMode runs the application without the Linux mail-server stack.
func NativeMode() bool {
	return os.Getenv("BILLIONMAIL_NATIVE") == "1"
}

func EnvironmentFile() string {
	if NativeMode() && os.Getenv("BILLIONMAIL_ENV_FILE") != "" {
		return os.Getenv("BILLIONMAIL_ENV_FILE")
	}
	return AbsPath("../.env")
}
