package middleware

import "github.com/gogf/gf/v2/frame/g"

func init() {
	g.Server().BindMiddleware("/adminapi/*", ChatgptAdminapiAuth)
	g.Server().BindMiddleware("/admin/base/comm/person", ArkoseCheck)

}
