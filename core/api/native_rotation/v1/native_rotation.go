package v1

import (
	rotation "billionmail-core/internal/service/native_rotation"
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type OptionsReq struct {
	g.Meta `path:"/native_rotation/options" method:"get" tags:"NativeRotation"`
}
type OptionsRes struct {
	api_v1.StandardRes
	Data g.Map `json:"data"`
}
type PreviewReq struct {
	g.Meta `path:"/native_rotation/preview" method:"post" tags:"NativeRotation"`
	Config rotation.Config `json:"config"`
}
type PreviewRes struct {
	api_v1.StandardRes
	Data *rotation.Preview `json:"data"`
}
type CreateReq struct {
	g.Meta      `path:"/native_rotation/create" method:"post" tags:"NativeRotation"`
	Config      rotation.Config `json:"config"`
	PreviewHash string          `json:"preview_hash" v:"required|length:64,64"`
	RequestKey  string          `json:"request_key" v:"required|length:16,80"`
}
type CreateRes struct {
	api_v1.StandardRes
	Data struct {
		ID int `json:"id"`
	} `json:"data"`
}
type ListReq struct {
	g.Meta `path:"/native_rotation/list" method:"get" tags:"NativeRotation"`
}
type ListRes struct {
	api_v1.StandardRes
	Data []rotation.Job `json:"data"`
}
type DetailReq struct {
	g.Meta `path:"/native_rotation/detail" method:"get" tags:"NativeRotation"`
	ID     int `json:"id" v:"required|min:1"`
}
type DetailRes struct {
	api_v1.StandardRes
	Data struct {
		Job   *rotation.Job   `json:"job"`
		Items []rotation.Item `json:"items"`
	} `json:"data"`
}
type ActionReq struct {
	g.Meta `path:"/native_rotation/action" method:"post" tags:"NativeRotation"`
	ID     int    `json:"id" v:"required|min:1"`
	Action string `json:"action" v:"required|in:start,pause,stop"`
}
type ActionRes struct{ api_v1.StandardRes }
