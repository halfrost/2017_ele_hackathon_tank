package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eleme/purchaseMeiTuan/rpc/player"

	"github.com/eleme/nex"
)

func main() {
	nex.Init()

	ctx := context.Background()
	client, err := player.GetThriftPlayerServiceClient()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		if _, err := client.Ping(ctx); err != nil {
			fmt.Printf("Ping failed: %v\n", err)
		}

		time.Sleep(time.Second)
	}
}
