package native_rotation

import (
	v1 "billionmail-core/api/native_rotation/v1"
	rotation "billionmail-core/internal/service/native_rotation"
	"context"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }
func (c *ControllerV1) Options(ctx context.Context, req *v1.OptionsReq) (res *v1.OptionsRes, err error) {
	res = &v1.OptionsRes{}
	res.Data, err = rotation.Options(ctx)
	if err == nil {
		res.SetSuccess("轮换选项")
	}
	return
}
func (c *ControllerV1) Preview(ctx context.Context, req *v1.PreviewReq) (res *v1.PreviewRes, err error) {
	res = &v1.PreviewRes{}
	res.Data, err = rotation.BuildPreview(ctx, req.Config)
	if err == nil {
		res.SetSuccess("分配预览，不会发送邮件")
	}
	return
}
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	res = &v1.CreateRes{}
	res.Data.ID, err = rotation.Create(ctx, req.Config, req.PreviewHash, req.RequestKey)
	if err == nil {
		res.SetSuccess("已创建待发送轮换任务，请检查后开始")
	}
	return
}
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	res = &v1.ListRes{}
	res.Data, err = rotation.List(ctx)
	if err == nil {
		res.SetSuccess("轮换任务列表")
	}
	return
}
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	res = &v1.DetailRes{}
	res.Data.Job, res.Data.Items, err = rotation.Detail(ctx, req.ID)
	if err == nil {
		res.SetSuccess("轮换任务详情")
	}
	return
}
func (c *ControllerV1) Action(ctx context.Context, req *v1.ActionReq) (res *v1.ActionRes, err error) {
	res = &v1.ActionRes{}
	err = rotation.Action(ctx, req.ID, req.Action)
	if err == nil {
		res.SetSuccess("任务状态已更新；已经提交的邮件可能仍会完成")
	}
	return
}
