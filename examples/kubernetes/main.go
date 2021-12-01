package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ksrichard/easyraft"
	"github.com/ksrichard/easyraft/discovery"
	"github.com/ksrichard/easyraft/fsm"
	"github.com/ksrichard/easyraft/serializer"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	// raft
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dataDir := os.Getenv("DATA_DIR")

	node, err := easyraft.NewNode(
		port, port+1,
		dataDir,
		[]fsm.FSMService{fsm.NewInMemoryMapService()},
		serializer.NewMsgPackSerializer(),
		discovery.NewMDNSDiscovery(),
		false,
	)

	if err != nil {
		panic(err)
	}
	stoppedCh, err := node.Start()
	if err != nil {
		panic(err)
	}
	defer node.Stop()

	// http server
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/put", func(ctx *gin.Context) {
		mapName, _ := ctx.GetQuery("map")
		key, _ := ctx.GetQuery("key")
		value, _ := ctx.GetQuery("value")
		result, err := node.RaftApply(fsm.MapPutRequest{
			MapName: mapName,
			Key:     key,
			Value:   value,
		}, 0)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	})
	r.GET("/get", func(ctx *gin.Context) {
		mapName, _ := ctx.GetQuery("map")
		key, _ := ctx.GetQuery("key")
		result, err := node.RaftApply(
			fsm.MapGetRequest{
				MapName: mapName,
				Key:     key,
			}, 0)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	})

	go func() {
		if err := r.Run(fmt.Sprintf(":%d", port+2)); err != nil {
			log.Fatal(err)
		}
	}()

	// wait for interruption/termination
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()
	<-done
	<-stoppedCh
}
