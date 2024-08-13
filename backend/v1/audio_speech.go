package v1

import (
	backendapi "backend/backend-api"
	"backend/config"
	"backend/modules/chatgpt/model"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	Proxy       *httputil.ReverseProxy
	UpStream, _ = url.Parse(config.CHATPROXY)
)

func init() {
	Proxy = httputil.NewSingleHostReverseProxy(UpStream)
	Proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		writer.WriteHeader(http.StatusBadGateway)
	}
}
func AudioSpeech(r *ghttp.Request) {
	ctx := r.GetCtx()
	userToken := r.Header.Get("authorization")
	if gstr.HasPrefix(userToken, "Bearer ") {
		userToken = strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	}
	// 如果 Authorization 为空，返回 401
	if userToken == "" {
		r.Response.WriteStatusExit(401)
	}
	userRecord, err := cool.DBM(model.NewChatgptUser()).Where("userToken", userToken).Where("expireTime>now()").Cache(gdb.CacheOption{
		Duration: 10 * time.Minute,
		Name:     "userToken:" + userToken,
		Force:    true,
	}).One()
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"detail": err.Error(),
		})
		return
	}
	if userRecord.IsEmpty() {
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"detail": "userToken not found",
		})
		return
	}
	email, ok := config.NormalSet.Pop()
	if !ok {
		g.Log().Error(ctx, "Get email from set error")
		r.Response.Status = 429
		r.Response.WriteJson(g.Map{
			"error": g.Map{
				"message": "Server is busy, please try again later",
				"type":    "invalid_request_error",
				"param":   "normalset",
				"code":    "server_busy",
			},
		})
		return
	}
	clears_in := 0
	isReturn := true

	defer func() {
		go func() {
			if email != "" && isReturn {
				ctx := gctx.New()
				count, err := config.DayCountAdd(ctx, email)
				if err != nil {
					g.Log().Error(ctx, "DayCountAdd", err)
				}
				if config.MAX_REQUEST_PER_DAY > 0 && count > config.MAX_REQUEST_PER_DAY {
					// 如果超过每日限制，今日不再归还
					g.Log().Info(ctx, email, "超过每日限制", count)
					clears_in = int(config.GetTodayLefeSecond(ctx))
				}

				if clears_in > 0 {
					// 延迟归还
					g.Log().Info(ctx, "延迟"+gconv.String(clears_in)+"秒归还", email, "到NormalSet", count)
					time.Sleep(time.Duration(clears_in) * time.Second)
				}
				config.NormalSet.Add(email)
				g.Log().Info(ctx, "归还", email, "到NormalSet", count)
			}
		}()
	}()
	Proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == 401 || resp.StatusCode == 402 {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				g.Log().Error(ctx, err)
				return err
			}
			g.Log().Error(ctx, "token过期,需要重新获取token", email, "返回", string(respBody))
			isReturn = false
			cool.DBM(model.NewChatgptSession()).Where("email", email).Update(g.Map{
				"status":          0, // token过期
				"officialSession": "token过期,需要重新获取token",
			})
			go backendapi.RefreshSession(email)
			r.Response.Status = 500
			r.Response.WriteJson(g.Map{
				"error": g.Map{
					"message": "Server is busy, please try again later|" + gconv.String(resp.StatusCode),
					"type":    "invalid_request_error",
					"param":   nil,
					"code":    "server_busy",
				},
			})
			return err
		}
		if resp.StatusCode == 429 {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				g.Log().Error(ctx, err)
				return err
			}
			resStr := string(respBody)
			clears_in = gjson.New(resStr).Get("detail.clears_in").Int()
			detail := gjson.New(resStr).Get("detail").String()
			if detail == "You've reached our limit of messages per hour. Please try again later." {
				clears_in = 3600
			}
			if detail == "You've hit your monthly limit. Please try again later." {
				clears_in = 3600 * 24
			}
			g.Log().Error(ctx, email, "resp.StatusCode==429", resStr)
			r.Response.Status = 500
			r.Response.WriteJson(g.Map{
				"error": g.Map{
					"message": "Server is busy, please try again later|429",
					"type":    "invalid_request_error",
					"param":   nil,
					"code":    "server_busy",
				},
			})
			return err
		}

		return nil
	}
	// 使用email获取 accessToken
	sessionCache := &config.CacheSession{}
	cool.CacheManager.MustGet(ctx, "session:"+email).Scan(&sessionCache)
	newreq := r.Request.Clone(ctx)
	newreq.URL.Host = UpStream.Host
	newreq.URL.Scheme = UpStream.Scheme
	newreq.Host = UpStream.Host
	newreq.Header.Set("Authorization", "Bearer "+sessionCache.AccessToken)
	newreq.Header.Del("Accept-Encoding")
	g.Log().Info(ctx, userToken, "使用", email, "-->", r.Request.URL.Path)

	Proxy.ServeHTTP(r.Response.RawWriter(), newreq)

}
