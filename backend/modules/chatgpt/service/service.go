package service

import (
	"backend/config"
	"backend/modules/chatgpt/model"
	"backend/utility"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
)

func init() {
	ctx := gctx.GetInitCtx()
	go AddAllSession(ctx)
	go RefreshAllSession(ctx)
	corn, err := gcron.AddSingleton(ctx, config.CRONINTERVAL(ctx), RefreshAllSession, "RefreshSession")
	if err != nil {
		panic(err)
	}
	g.Log().Info(ctx, "RefreshAllSession", "corn", corn, "cornInterval", config.CRONINTERVAL(ctx), "注册成功")
}

// 启动时添加所有账号的session到缓存及set
func AddAllSession(ctx g.Ctx) {

	record, err := cool.DBM(model.NewChatgptSession()).OrderAsc("updateTime").Where("status=1").All()
	if err != nil {
		g.Log().Error(ctx, "AddAllSession", err)
		return
	}
	for _, v := range record {
		email := v["email"].String()
		officialSession := gjson.New(v["officialSession"])
		session, err := utility.ParseSession(officialSession.String())
		if err != nil {
			g.Log().Error(ctx, "AddAllSession", email, err)
			continue

		}
		// g.Dump(session)

		// 添加到缓存
		cacheSession := &config.CacheSession{
			Email:        email,
			AccessToken:  session.AccessToken,
			CooldownTime: 0,
			RefreshToken: session.RefreshToken,
			PlanType:     session.PlanType,
		}
		err = cool.CacheManager.Set(ctx, "session:"+email, cacheSession, time.Hour*24*10)

		if err != nil {
			g.Log().Error(ctx, "AddAllSession to cache ", email, err)
			continue
		}
		g.Log().Info(ctx, "AddAllSession to cache", email, "success")
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

	}

	g.Log().Info(ctx, "AddSession finish", "plusSet", config.PlusSet.Size(), "normalSet", config.NormalSet.Size(), "Gpt4oLiteSet", config.Gpt4oLiteSet.Size(), "NormalGptsSet", config.NormalGptsSet.Size())

}

// RefreshAllSession 刷新所有session
func RefreshAllSession(ctx g.Ctx) {
	// if !gmlock.TryLock("RefreshAllSession") {
	// 	g.Log().Info(ctx, "RefreshAllSession", "已有任务在执行")
	// 	return
	// }
	// defer gmlock.Unlock("RefreshAllSession")
	record, err := cool.DBM(model.NewChatgptSession()).OrderAsc("status").All()
	if err != nil {
		g.Log().Error(ctx, "RefreshAllSession", err)
		return
	}
	for _, v := range record {

		email := v["email"].String()
		password := v["password"].String()
		g.Log().Info(ctx, "RefreshAllSession", email, len(record))
		officialSession := gjson.New(v["officialSession"])
		session, err := utility.ParseSession(officialSession.String())
		if err != nil {
			g.Log().Error(ctx, "RefreshAllSession", email, err)
			if session.Disabled {
				g.Log().Error(ctx, "RefreshAllSession", email, "跳过刷新")
				continue
			}
		}
		getSessionUrl := config.CHATPROXY + "/auth/refresh"
		if session.RefreshToken == "" {
			getSessionUrl = config.CHATPROXY + "/applelogin"
		}
		sessionVar := g.Client().PostVar(ctx, getSessionUrl, g.Map{
			"username":      email,
			"password":      password,
			"refresh_token": session.RefreshToken,
		})
		detail := gjson.New(sessionVar).Get("detail").String()
		if detail != "" {
			g.Log().Error(ctx, "RefreshAllSession", email, detail)
			cool.DBM(model.NewChatgptSession()).Where("email=?", email).Update(g.Map{
				"officialSession": sessionVar,
				"status":          0,
			})
			continue
		}
		session, err = utility.ParseSession(sessionVar.String())
		if err != nil {
			g.Log().Error(ctx, "RefreshAllSession", email, err)
			continue
		}
		// 添加到缓存
		cacheSession := &config.CacheSession{
			Email:        email,
			AccessToken:  session.AccessToken,
			CooldownTime: 0,
			RefreshToken: session.RefreshToken,
			PlanType:     session.PlanType,
		}
		err = cool.CacheManager.Set(ctx, "session:"+email, cacheSession, time.Hour*24*10)
		if err != nil {
			g.Log().Error(ctx, "AddAllSession to cache ", email, err)
			continue
		}
		g.Log().Info(ctx, "AddAllSession to cache", email, "success")
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
		// 关闭个人区记忆
		g.Client().SetHeaderMap(g.MapStrStr{
			"Authorization": "Bearer " + session.AccessToken,
			"Content-Type":  "application/json",
		}).PatchVar(ctx, config.CHATPROXY+"/backend-api/settings/account_user_setting?feature=sunshine&value=false", g.Map{})
		for _, v := range session.TeamIds {
			config.PlusSet.Add(email + "|" + v)
			// 关闭团队区记忆
			g.Client().SetHeaderMap(g.MapStrStr{
				"Authorization":      "Bearer " + session.AccessToken,
				"Content-Type":       "application/json",
				"Chatgpt-Account-Id": v,
			}).PatchVar(ctx, config.CHATPROXY+"/backend-api/settings/account_user_setting?feature=sunshine&value=false", g.Map{})
		}
	}

	g.Log().Info(ctx, "RefreshAllSession finish", "plusSet", config.PlusSet.Size(), "normalSet", config.NormalSet.Size(), "Gpt4oLiteSet", config.Gpt4oLiteSet.Size(), "NormalGptsSet", config.NormalGptsSet.Size())
}
