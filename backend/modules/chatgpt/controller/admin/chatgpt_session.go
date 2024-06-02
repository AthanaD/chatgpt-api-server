package admin

import (
	"backend/modules/chatgpt/service"
	"context"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/frame/g"
)

type ChatgptSessionController struct {
	*cool.Controller
}
type AddBulkReq struct {
	g.Meta        `path:"/addbulk" method:"POST"`
	Authorization string `json:"Authorization" in:"header"`
	Accouts       string `json:"accouts" in:"body"`
}

func (c *ChatgptSessionController) Move(ctx context.Context, req *AddBulkReq) (res *cool.BaseRes, err error) {
	// err = service.NewBaseSysUserService().Move(ctx)
	g.Dump(req)
	res = cool.Ok(nil)
	return
}
func init() {
	var chatgpt_session_controller = &ChatgptSessionController{
		&cool.Controller{
			Prefix:  "/admin/chatgpt/session",
			Api:     []string{"Add", "Delete", "Update", "Info", "List", "Page"},
			Service: service.NewChatgptSessionService(),
		},
	}
	// 注册路由
	cool.RegisterController(chatgpt_session_controller)
}
