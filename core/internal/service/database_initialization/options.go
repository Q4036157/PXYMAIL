package database_initialization

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	registerHandler(func() {
		sqlList := []string{
			`-- BillionMail Options
			CREATE TABLE IF NOT EXISTS bm_options (
				name VARCHAR(255) NOT NULL PRIMARY KEY,
				value TEXT NOT NULL
			)`,
		}
		if public.NativeMode() && public.MustGetDockerEnv("MAIL_SPACE_EMAIL", "") != "" {
			// Initialize missing workspace locale while retaining a saved user choice.
			sqlList = append(sqlList, `INSERT INTO bm_options(name,value) VALUES('CurrentLanguage','zh') ON CONFLICT(name) DO UPDATE SET value=EXCLUDED.value WHERE bm_options.value=''`)
		}

		for _, sql := range sqlList {
			_, err := g.DB().Exec(context.Background(), sql)
			if err != nil {
				g.Log().Error(context.Background(), "Failed to create billion-mail options table:", err)
				return
			}
		}
	})
}
