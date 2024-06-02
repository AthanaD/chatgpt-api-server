package admin

import (
	"backend/modules/chatgpt/service"
	"backend/websocket"
	"context"
	"fmt"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gmlock"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type ChatgptSessionController struct {
	*cool.Controller
}
type AddBulkReq struct {
	g.Meta        `path:"/addbulk" method:"POST"`
	Authorization string `json:"Authorization" in:"header"`
	Accounts      string `jv:"required#账号密码不能为空" json:"accounts" in:"body"`
}
type Account struct {
	Username string `json:"username" v:"email"`
	Password string `json:"password" v:"required"`
}

func (c *ChatgptSessionController) AddBulk(ctx context.Context, req *AddBulkReq) (res *cool.BaseRes, err error) {

	// err = service.NewBaseSysUserService().Move(ctx)
	accounts := gstr.SplitAndTrim(req.Accounts, "\n")
	// g.Dump(accounts)
	Accounts := make([]Account, 0)
	for _, v := range accounts {
		av := gstr.SplitAndTrim(v, ",")
		if len(av) >= 2 {
			account := Account{Username: gstr.ToLower(av[0]), Password: av[1]}
			if err := g.Validator().Data(account).Run(ctx); err != nil {
				fmt.Print(gstr.Join(err.Strings(), "\n"))
				// res = cool.Fail(gstr.Join(err.Strings(), "\n"))

				return res, err
			}

			Accounts = append(Accounts, account)
		} else {
			g.Log().Error(ctx, v+"|格式错误")
			res = cool.Fail(v + "|格式错误")
			return
		}
	}
	if len(Accounts) == 0 {
		res = cool.Fail("无有效账号")
		return
	}
	if !gmlock.TryLock("chatgpt_session_addbulk") {
		res = cool.Fail("已经有批量任务在执行中,请等待任务完成后再开始新的任务")
		return
	}
	// g.Dump(Accounts)
	go func() {

		countALL := len(Accounts)
		countSuccess := 0
		countFail := 0
		time.Sleep(5 * time.Second)
		ctx := gctx.New()
		ctxid := gctx.CtxId(ctx)

		for _, account := range Accounts {
			err = service.NewChatgptSessionService().AddSession(ctx, account.Username, account.Password)
			if err != nil {
				countFail++
				g.Log().Error(ctx, account.Username, "添加失败", err)
				websocket.SendToAll(&websocket.WResponse{Event: "error", Data: ctxid + " " + gtime.Now().String() + ":" + account.Username + " " + err.Error() + ",总数:" + gconv.String(countALL) + ",成功:" + gconv.String(countSuccess) + ",失败:" + gconv.String(countFail)})
			} else {
				countSuccess++
				websocket.SendToAll(&websocket.WResponse{Event: "success", Data: ctxid + " " + gtime.Now().String() + ":" + account.Username + " 添加成功,总数:" + gconv.String(countALL) + ",成功:" + gconv.String(countSuccess) + ",失败:" + gconv.String(countFail)})
			}
		}
		g.Log().Info(ctx, "批量任务完成,总数:", countALL, "成功:", countSuccess, "失败:", countFail)
		websocket.SendToAll(&websocket.WResponse{Event: "finish", Data: ctxid + " " + gtime.Now().String() + ":批量任务完成,总数:" + gconv.String(countALL) + ",成功:" + gconv.String(countSuccess) + ",失败:" + gconv.String(countFail)})
		gmlock.Unlock("chatgpt_session_addbulk")
	}()

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
