package service

import (
	"backend/config"
	"backend/modules/chatgpt/model"
	"backend/utility"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

type ChatgptSessionService struct {
	*cool.Service
}

func NewChatgptSessionService() *ChatgptSessionService {
	return &ChatgptSessionService{
		&cool.Service{
			Model: model.NewChatgptSession(),
			UniqueKey: g.MapStrStr{
				"email": "邮箱不能重复",
			},
			NotNullKey: g.MapStrStr{
				"email":    "邮箱不能为空",
				"password": "密码不能为空",
			},
			PageQueryOp: &cool.QueryOp{
				FieldEQ:      []string{"email", "password", "officialSession", "remark"},
				KeyWordField: []string{"email", "password", "officialSession", "remark"},
			},
		},
	}
}

// MofifyBefore 新增/删除/修改之前的操作
func (s *ChatgptSessionService) ModifyBefore(ctx g.Ctx, method string, param map[string]interface{}) (err error) {
	g.Log().Debug(ctx, "ChatgptSessionService.ModifyBefore", method, param)

	// g.Dump(idsJson)
	// 如果是删除，就删除缓存及set
	if method == "Delete" {
		ids := gjson.New(param["ids"]).Array()
		for _, id := range ids {
			record, err := cool.DBM(s.Model).Where("id=?", id).One()
			if err != nil {
				return err
			}
			email := record["email"].String()
			isPlus := record["isPlus"].Int()

			// 删除缓存
			cool.CacheManager.Remove(ctx, "session:"+email)
			// 删除set
			if isPlus == 1 {
				config.PlusSet.Remove(email)
			} else {
				config.NormalSet.Remove(email)
			}
		}
	}

	return
}

// ModifyAfter 新增/删除/修改之后的操作
func (s *ChatgptSessionService) ModifyAfter(ctx g.Ctx, method string, param map[string]interface{}) (err error) {
	g.Log().Debug(ctx, "ChatgptSessionService.ModifyAfter", method, param)
	// 新增/修改 之后，更新session
	if method != "Add" && method != "Update" {
		return
	}
	officialSession := gjson.New(param["officialSession"])
	refreshToken := officialSession.Get("refresh_token").String()

	// 如果没有officialSession，就去获取
	s.GetSessionAndUpdateStatus(ctx, param, refreshToken)
	return
}

// AddSession 获取session并更新状态
func (s *ChatgptSessionService) AddSession(ctx g.Ctx, username, password, refreshToken string) error {
	ctxid := gctx.CtxId(ctx)
	// 先检查是否已经存在
	count, err := cool.DBM(s.Model).Where("email=?", username).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("账号已存在")
	}
	// g.Log().Info(ctx, "AddSession", "username", username, "password", password)
	loginurl := config.CHATPROXY + "/applelogin"
	if refreshToken != "" {
		loginurl = config.CHATPROXY + "/auth/refresh"
	}
	g.Log().Debug(ctx, "ChatgptSessionService.AddSession", loginurl)
	sessionVar := g.Client().PostVar(ctx, loginurl, g.Map{
		"username":      username,
		"password":      password,
		"refresh_token": refreshToken,
	})
	sessionJson := gjson.New(sessionVar)
	// g.Dump(sessionVar)
	if sessionJson.Get("accessToken").String() == "" {
		g.Log().Error(ctx, "ChatgptSessionService.ModifyAfter", "get session error", sessionVar)
		detail := sessionJson.Get("detail").String()
		if detail != "" {
			err := gerror.New(detail)
			cool.DBM(s.Model).Data(g.Map{
				"createTime":      gtime.Now().String(),
				"updateTime":      gtime.Now().String(),
				"email":           username,
				"password":        password,
				"officialSession": sessionJson.String(),
				"status":          0,
				"isPlus":          0,
				"remark":          ctxid + "|批量添加",
			}).Insert()
			return err
		} else {
			cool.DBM(s.Model).Data(g.Map{
				"createTime":      gtime.Now().String(),
				"updateTime":      gtime.Now().String(),
				"email":           username,
				"password":        password,
				"officialSession": "get session error",
				"status":          0,
				"isPlus":          0,
				"remark":          ctxid + "|批量添加",
			}).Insert()
			return gerror.New("get session error")
		}
	}
	var isPlus int
	plan_type := sessionJson.Get("accountCheckInfo.plan_type").String()
	if plan_type == "plus" || plan_type == "team" {
		isPlus = 1
	} else {
		isPlus = 0

	}
	_, err = cool.DBM(s.Model).Data(g.Map{
		"createTime":      gtime.Now().String(),
		"updateTime":      gtime.Now().String(),
		"email":           username,
		"password":        password,
		"officialSession": sessionJson.String(),
		"isPlus":          isPlus,
		"status":          1,
		"remark":          ctxid + "|批量添加",
	}).Insert()
	if err != nil {
		return err
	}
	session, err := utility.ParseSession(sessionJson.String())
	if err != nil {
		g.Log().Error(ctx, "AddSession", username, err)
		return err

	}
	// g.Dump(session)

	// 添加到缓存
	email := session.Email

	cacheSession := &config.CacheSession{
		Email:        email,
		AccessToken:  session.AccessToken,
		CooldownTime: 0,
		RefreshToken: session.RefreshToken,
		PlanType:     session.PlanType,
	}
	cool.CacheManager.Set(ctx, "session:"+email, cacheSession, time.Hour*24*10)
	config.RefreshQueueSession(ctx, session)

	g.Log().Info(ctx, "AddSession finish", "plusSet", config.PlusSet.Size(), "normalSet", config.NormalSet.Size())
	return nil
}

func (s *ChatgptSessionService) GetSessionAndUpdateStatus(ctx g.Ctx, param g.Map, refreshToken string) error {
	getSessionUrl := config.CHATPROXY + "/applelogin"
	if refreshToken != "" {
		getSessionUrl = config.CHATPROXY + "/auth/refresh"
	}
	g.Log().Debug(ctx, "ChatgptSessionService.GetSessionAndUpdateStatus", getSessionUrl, param, refreshToken)
	sessionVar := g.Client().SetHeader("authkey", config.AUTHKEY(ctx)).SetCookie("arkoseToken", gconv.String(param["arkoseToken"])).PostVar(ctx, getSessionUrl, g.Map{
		"username":      param["email"],
		"password":      param["password"],
		"authkey":       config.AUTHKEY(ctx),
		"refresh_token": refreshToken,
	})
	sessionJson := gjson.New(sessionVar)
	// g.Dump(sessionJson)
	if sessionJson.Get("accessToken").String() == "" {
		g.Log().Error(ctx, "ChatgptSessionService.ModifyAfter", "get session error", sessionJson)
		detail := sessionJson.Get("detail").String()
		if detail != "" {
			err := gerror.New(detail)
			cool.DBM(s.Model).Where("email=?", param["email"]).Update(g.Map{
				"officialSession": sessionJson.String(),
				"status":          0,
			})
			return err
		} else {
			return gerror.New("get session error")
		}
	}
	var isPlus int

	models := sessionJson.Get("models").Array()
	if len(models) > 1 {
		isPlus = 1
	} else {
		isPlus = 0
	}
	plan_type := sessionJson.Get("accountCheckInfo.plan_type").String()
	if plan_type == "plus" || plan_type == "team" {
		isPlus = 1
	} else {
		isPlus = 0

	}
	_, err := cool.DBM(s.Model).Where("email=?", param["email"]).Update(g.Map{
		"officialSession": sessionJson.String(),
		"isPlus":          isPlus,
		"status":          1,
	})
	if err != nil {
		g.Log().Error(ctx, "ChatgptSessionService.ModifyAfter", "update session error", err)
		return err
	}
	// 写入缓存及set
	session, err := utility.ParseSession(sessionJson.String())
	if err != nil {
		g.Log().Error(ctx, "AddSession", param["email"], err)
		return err

	}
	// g.Dump(session)

	// 添加到缓存
	email := session.Email

	cacheSession := &config.CacheSession{
		Email:        email,
		AccessToken:  session.AccessToken,
		CooldownTime: 0,
		RefreshToken: session.RefreshToken,
		PlanType:     session.PlanType,
	}
	cool.CacheManager.Set(ctx, "session:"+email, cacheSession, time.Hour*24*10)
	if session.PlanType == "plus" || session.PlanType == "team" {
		config.PlusSet.Add(email)
		config.NormalSet.Remove(email)
		config.Gpt4oLiteSet.Remove(email)
		config.NormalGptsSet.Remove(email)
		for _, v := range session.TeamIds {
			config.PlusSet.Add(email + "|" + v)
		}
	}
	if session.PlanType == "free" {
		config.NormalSet.Add(email)
		config.Gpt4oLiteSet.Add(email)
		config.PlusSet.Remove(email)
		if session.FreeWithGpts {
			config.NormalGptsSet.Add(email)
		}
	}
	g.Log().Info(ctx, "AddSession finish", "plusSet", config.PlusSet.Size(), "normalSet", config.NormalSet.Size(), "Gpt4oLiteSet", config.Gpt4oLiteSet.Size(), "NormalGptsSet", config.NormalGptsSet.Size())

	return err
}
