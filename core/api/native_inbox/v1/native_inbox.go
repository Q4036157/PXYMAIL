package v1

import (
	inbox "billionmail-core/internal/service/native_inbox"
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type AccountsReq struct {
	g.Meta `path:"/native_inbox/accounts" method:"get" tags:"NativeInbox"`
}
type AccountsRes struct {
	api_v1.StandardRes
	Data []inbox.Account `json:"data"`
}
type ListReq struct {
	g.Meta    `path:"/native_inbox/list" method:"get" tags:"NativeInbox"`
	AccountID int    `json:"account_id" v:"min:0"`
	Kind      string `json:"kind" v:"in:all,customer_reply,bounce,other" default:"all"`
	Keyword   string `json:"keyword" v:"length:0,200"`
	Page      int    `json:"page" v:"min:1" default:"1"`
	PageSize  int    `json:"page_size" v:"min:1|max:100" default:"20"`
}
type ListRes struct {
	api_v1.StandardRes
	Data struct {
		List  []inbox.Mail `json:"list"`
		Total int          `json:"total"`
	} `json:"data"`
}
type DetailReq struct {
	g.Meta `path:"/native_inbox/detail" method:"get" tags:"NativeInbox"`
	ID     int64 `json:"id" v:"required|min:1"`
}
type DetailRes struct {
	api_v1.StandardRes
	Data *inbox.Mail `json:"data"`
}
type SyncReq struct {
	g.Meta    `path:"/native_inbox/sync" method:"post" tags:"NativeInbox"`
	AccountID int `json:"account_id" v:"min:0"`
}
type CheckNowReq struct {
	g.Meta    `path:"/native_inbox/check_now" method:"post" tags:"NativeInbox"`
	AccountID int `json:"account_id" v:"required|min:1"`
}
type CheckNowRes struct {
	api_v1.StandardRes
	Data *inbox.CheckResult `json:"data"`
}
type UpdateReq struct {
	g.Meta    `path:"/native_inbox/update" method:"post" tags:"NativeInbox"`
	AccountID int   `json:"account_id" v:"required|min:1"`
	Enabled   *bool `json:"enabled" v:"required"`
}
type MarkReq struct {
	g.Meta    `path:"/native_inbox/mark" method:"post" tags:"NativeInbox"`
	ID        int64 `json:"id" v:"required|min:1"`
	Processed *bool `json:"processed" v:"required"`
}
type ActionRes struct{ api_v1.StandardRes }
