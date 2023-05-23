package main

import (
	"context"
	"time"

	"github.com/ecodeclub/ecron/internal/scheduler"
	"github.com/ecodeclub/ecron/internal/storage/mysql"
)

func main() {
	ctx := context.TODO()
	storeIns, err := mysql.NewMysqlStorage("root:@tcp(localhost:3306)/ecron",
		mysql.WithPreemptInterval(1*time.Second))
	if err != nil {
		return
	}
	go storeIns.RunPreempt(ctx)
	go storeIns.AutoRefresh(ctx)

	sche := scheduler.NewScheduler(storeIns)

	if err = sche.Start(ctx); err != nil {
		panic(err)
	}
}
