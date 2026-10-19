package main

import (
	_ "billionmail-core/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"billionmail-core/internal/cmd"
	"billionmail-core/internal/service/platform_gateway"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--platform-gateway" {
		if err := platform_gateway.Run(); err != nil {
			log.Fatal(err)
		}
		return
	}
	cmd.Main.Run(gctx.GetInitCtx())
}
