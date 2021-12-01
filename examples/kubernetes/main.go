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
	"time"
)

func main() {
	// raft details
	raftPort, _ := strconv.Atoi(os.Getenv("EASYRAFT_PORT"))
	discoveryPort, _ := strconv.Atoi(os.Getenv("DISCOVERY_PORT"))
	httpPort, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))
	dataDir := os.Getenv("DATA_DIR")

	// EasyRaft Node
	node, err := easyraft.NewNode(
		raftPort,
		discoveryPort,
		dataDir,
		[]fsm.FSMService{fsm.NewInMemoryMapService()},
		serializer.NewMsgPackSerializer(),
		discovery.NewKubernetesDiscovery("", nil, ""),
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
	r.POST("/put", func(ctx *gin.Context) {
		mapName, _ := ctx.GetQuery("map")
		key, _ := ctx.GetQuery("key")
		value, _ := ctx.GetQuery("value")
		result, err := node.RaftApply(fsm.MapPutRequest{
			MapName: mapName,
			Key:     key,
			Value:   value,
		}, time.Second)
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
			}, time.Second)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	})

	go func() {
		if err := r.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
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
