package main

import (
	_ "backend/internal/packed"
	_ "backend/websocket"

	_ "backend/websocket"

	_ "github.com/cool-team-official/cool-admin-go/contrib/drivers/mysql"
	_ "github.com/cool-team-official/cool-admin-go/contrib/drivers/sqlite"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	// _ "backend/arkose"
	_ "backend/backend-api"
	_ "backend/modules"
	_ "backend/v1"

	"github.com/gogf/gf/v2/os/gctx"

	"backend/internal/cmd"
)

func main() {
	// gres.Dump()
	// go ghttp.StartPProfServer(8299)
	ctx := gctx.New()

	cmd.Main.Run(ctx)
}
