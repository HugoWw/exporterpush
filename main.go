package main

import (
	_ "github.com/exporterpush/config"
	"github.com/exporterpush/global"
	"github.com/exporterpush/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	server.Run()

	// wait syscall signal
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sigInfo := <-quit

	global.LogObj.Errorf("Shutting down server get signal info: %v", sigInfo)

}
