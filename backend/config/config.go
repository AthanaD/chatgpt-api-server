package config

import (
	"backend/utility"
	"math/rand"
	"time"

	baseservice "github.com/cool-team-official/cool-admin-go/modules/base/service"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// func CHATPROXY(ctx g.Ctx) string {
// 	return g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
// }

func AUTHKEY(ctx g.Ctx) string {
	// g.Log().Debug(ctx, "config.AUTHKEY", g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String())
	return g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String()
}

func USERTOKENLOCK(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "USERTOKENLOCK").Bool()
}

var (
	DefaultModel              = "auto"
	FreeModels                = garray.NewStrArray()
	PlusModels                = garray.NewStrArray()
	NormalSet                 = utility.NewSafeQueue("NormalSet")
	PlusSet                   = utility.NewSafeQueue("PlusSet")
	Gpt4oLiteSet              = utility.NewSafeQueue("Gpt4oLiteSet")
	NormalGptsSet             = utility.NewSafeQueue("NormalGptsSet")
	Gpt_4o_Set                = utility.NewSafeQueue("Gpt_4o_Set")
	O1previewSet              = utility.NewSafeQueue("O1previewSet")
	O1MiniSet                 = utility.NewSafeQueue("O1MiniSet")
	MAXTIME                   = 0
	TraceparentCache          = gcache.New()
	CHATPROXY                 = ""
	Redis                     = g.Redis("cool")
	MAX_REQUEST_PER_DAY int64 = 0
	MODERATION                = "https://gateway.ai.cloudflare.com/v1/a8cace244ffbc233655fefeaca37d515/xyhelper/openai/moderations"
	OAIKEY                    = ""
	OAIKEYLOG                 = "" // OAIKEYLOG 隐藏
	KEEP_CHAT_HISTORY         = false
)

func PORT(ctx g.Ctx) int {
	// g.Log().Debug(ctx, "config.PORT", g.Cfg().MustGetWithEnv(ctx, "PORT").Int())
	if g.Cfg().MustGetWithEnv(ctx, "PORT").Int() == 0 {
		return 8001
	}
	return g.Cfg().MustGetWithEnv(ctx, "PORT").Int()
}

func ISFREE(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "ISFREE").Bool()
}

func APIAUTH(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "APIAUTH").String()
}
func CRONINTERVAL(ctx g.Ctx) string {
	// 生成随机时间的每3天执行一次的表达式，格式为：秒 分 时 天 月 星期
	// 生成随机秒数 在0-59之间
	second := generateRandomNumber(59)
	secondStr := gconv.String(second)
	// 生成随机分钟数 在0-59之间
	minute := generateRandomNumber(59)
	minuteStr := gconv.String(minute)
	// 生成随机小时数 在0-23之间
	hour := generateRandomNumber(23)
	hourStr := gconv.String(hour)
	// 拼接cron表达式
	cronStr := secondStr + " " + minuteStr + " " + hourStr + " * * *"
	return cronStr

}

func generateRandomNumber(max int) int {
	rand.Seed(time.Now().UnixNano()) // 使用当前时间作为随机数生成器的种子
	return rand.Intn(max)            // 生成0到59之间的随机数
}

// continue
func CONTINUEMAX(ctx g.Ctx) int {
	if g.Cfg().MustGetWithEnv(ctx, "CONTINUEMAX").IsEmpty() {
		return 3
	}
	return g.Cfg().MustGetWithEnv(ctx, "CONTINUEMAX").Int()
}

type CacheSession struct {
	Email        string `json:"email"`
	AccessToken  string `json:"accessToken"`
	CooldownTime int64  `json:"cooldownTime"`
	RefreshToken string `json:"refreshToken"`
	PlanType     string `json:"planType"`
}

func init() {
	ctx := gctx.GetInitCtx()
	FreeModels.Append("text-davinci-002-render-sha")
	FreeModels.Append("text-davinci-002-render-sha-mobile")
	FreeModels.Append("auto")
	FreeModels.Append("gpt-4o-mini")
	PlusModels.Append("gpt-4")
	PlusModels.Append("gpt-4o")
	PlusModels.Append("gpt-4-browsing")
	PlusModels.Append("gpt-4-plugins")
	PlusModels.Append("gpt-4-mobile")
	PlusModels.Append("gpt-4-dalle")
	PlusModels.Append("gpt-4-code-interpreter")
	PlusModels.Append("gpt-4-gizmo")
	PlusModels.Append("gpt-4-gizmo-lite")
	PlusModels.Append("o1-preview")
	PlusModels.Append("o1-mini")

	chatproxy := g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
	if chatproxy != "" {
		CHATPROXY = chatproxy
	} else {
		panic("CHATPROXY is empty")
	}
	g.Log().Info(ctx, "CHATPROXY:", CHATPROXY)
	maxRequestPerDay := g.Cfg().MustGetWithEnv(ctx, "MAX_REQUEST_PER_DAY").Int64()
	if maxRequestPerDay > 0 {
		MAX_REQUEST_PER_DAY = maxRequestPerDay
	}
	g.Log().Info(ctx, "MAX_REQUEST_PER_DAY:", MAX_REQUEST_PER_DAY)

	oaikey := g.Cfg().MustGetWithEnv(ctx, "OAIKEY").String()
	// oaikey 不为空
	if oaikey != "" {
		OAIKEY = oaikey
		// 日志隐藏 oaikey，有 * 代表有值
		OAIKEYLOG = "******"
	}
	g.Log().Info(ctx, "OAIKEY:", OAIKEYLOG)
	moderation := g.Cfg().MustGetWithEnv(ctx, "MODERATION").String()
	if moderation != "" {
		MODERATION = moderation
	}
	g.Log().Info(ctx, "MODERATION:", MODERATION)
	keepChatHistory := g.Cfg().MustGetWithEnv(ctx, "KEEP_CHAT_HISTORY").Bool()
	if keepChatHistory {
		KEEP_CHAT_HISTORY = keepChatHistory
	}
	g.Log().Info(ctx, "KEEP_CHAT_HISTORY:", KEEP_CHAT_HISTORY)
	modelmapStr, err := baseservice.NewBaseSysParamService().DataByKey(ctx, "modelmap")
	if err != nil {
		panic(err)
	}
	modelmap := gconv.MapStrStr(modelmapStr)
	g.Dump(modelmap)

}

func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// rand.Seed(time.Now().UnixNano())

	id := "chatcmpl-"
	for i := 0; i < length; i++ {
		id += string(charset[rand.Intn(len(charset))])
	}
	return id
}

func GetModel(ctx g.Ctx, model string, isPlusUser bool) string {
	// g.Log().Debug(ctx, "GetModel", model)
	modelMapStr, err := baseservice.NewBaseSysParamService().DataByKey(ctx, "modelmap")
	if err != nil {
		g.Log().Error(ctx, "GetModel", err)
		return DefaultModel
	}
	// g.Dump(modelMapStr)
	modelMap := gconv.MapStrStr(gjson.New(modelMapStr))
	// g.Dump(modelMap)
	if v, ok := modelMap[model]; ok {
		return v
	}
	if gstr.HasPrefix(model, "gpt-4-gizmo-") {
		if v, ok := modelMap["gpt-4-gizmo-*"]; ok {
			return v
		} else {
			if isPlusUser {
				return "gpt-4o"
			} else {
				return "gpt-4o-lite"
			}

		}
	}
	return DefaultModel
}

func DayCountAdd(ctx g.Ctx, key string) (res int64, err error) {
	// g.Log().Debug(ctx, "CountAdd", key, value)
	redisKey := "daycount:" + utility.GetEsyncValue(ctx) + ":" + key

	res, err = Redis.Incr(ctx, redisKey)
	if err != nil {
		g.Log().Error(ctx, "CountAdd", err)
		return
	}
	// 设置key的过期时间
	Redis.Expire(ctx, redisKey, 86400)
	return
}

// GetTodayLefeSecond 获取今天剩余秒数
func GetTodayLefeSecond(ctx g.Ctx) int64 {
	now := time.Now()
	// 获取当前时间的年月日
	year, month, day := now.Date()
	// 获取明天的时间
	tomorrow := time.Date(year, month, day+1, 0, 0, 0, 0, now.Location())
	// 获取当前时间到明天的时间差
	return tomorrow.Unix() - now.Unix()
}

// RefreshQueueSession 刷新队列中的session
func RefreshQueueSession(ctx g.Ctx, session *utility.Session) {
	email := session.Email
	if session.PlanType == "plus" || session.PlanType == "team" {
		PlusSet.Add(email)
		O1MiniSet.Add(email)
		O1previewSet.Add(email)
		NormalSet.Remove(email)
		Gpt4oLiteSet.Remove(email)
		NormalGptsSet.Remove(email)
		for _, v := range session.TeamIds {
			PlusSet.Add(email + "|" + v)
			O1MiniSet.Add(email + "|" + v)
			O1previewSet.Add(email + "|" + v)
		}
	}
	if session.PlanType == "free" {
		NormalSet.Add(email)
		Gpt4oLiteSet.Add(email)
		PlusSet.Remove(email)
		O1MiniSet.Remove(email)
		O1previewSet.Remove(email)
		if session.FreeWithGpts {
			NormalGptsSet.Add(email)
		}
	}
	g.Log().Info(ctx, "RefreshQueueSession", email, "PlusSet", PlusSet.Size(), "NormalSet", NormalSet.Size(), "Gpt4oLiteSet", Gpt4oLiteSet.Size(), "NormalGptsSet", NormalGptsSet.Size(), "O1MiniSet", O1MiniSet.Size(), "O1previewSet", O1previewSet.Size())
}
