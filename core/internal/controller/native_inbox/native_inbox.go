package native_inbox

import (
	v1 "billionmail-core/api/native_inbox/v1"
	inbox "billionmail-core/internal/service/native_inbox"
	"context"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }
func (c *ControllerV1) Accounts(ctx context.Context, req *v1.AccountsReq) (res *v1.AccountsRes, err error) {
	res = &v1.AccountsRes{}
	res.Data, err = inbox.Accounts(ctx)
	res.SetSuccess("收件账号列表")
	return
}
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{}
	res.Data.List, res.Data.Total, err = inbox.List(ctx, req.AccountID, req.Kind, req.Keyword, req.Page, req.PageSize)
	res.SetSuccess("邮件列表")
	return
}
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	res = &v1.DetailRes{}
	res.Data, err = inbox.Detail(ctx, req.ID)
	res.SetSuccess("邮件内容")
	return
}
func (c *ControllerV1) Sync(ctx context.Context, req *v1.SyncReq) (res *v1.ActionRes, err error) {
	res = &v1.ActionRes{}
	err = inbox.Queue(ctx, req.AccountID, nil)
	res.SetSuccess("已排队；系统会按账号轮询，并遵守 15 分钟登录间隔")
	return
}
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.ActionRes, err error) {
	res = &v1.ActionRes{}
	err = inbox.Queue(ctx, req.AccountID, req.Enabled)
	res.SetSuccess("收件设置已更新")
	return
}
func (c *ControllerV1) CheckNow(ctx context.Context, req *v1.CheckNowReq) (res *v1.CheckNowRes, err error) {
	res = &v1.CheckNowRes{}
	res.Data, err = inbox.CheckNow(ctx, req.AccountID)
	if err == nil {
		res.SetSuccess("收件检查完成")
	}
	return
}
func (c *ControllerV1) Mark(ctx context.Context, req *v1.MarkReq) (res *v1.ActionRes, err error) {
	res = &v1.ActionRes{}
	err = inbox.Mark(ctx, req.ID, *req.Processed)
	res.SetSuccess("处理标记已保存；原邮箱未修改")
	return
}
