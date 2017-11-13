package main

import (
	"github.com/eleme/nex"

	"github.com/eleme/purchaseMeiTuan/handler"
)

func main() {
	nex.Init()
	processorFactory := handler.NewPlayerServiceProcessorFactory()
	nex.Serve(processorFactory)
}
