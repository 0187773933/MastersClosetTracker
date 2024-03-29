package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"path/filepath"
	server "github.com/0187773933/MastersClosetTracker/v1/server"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	log "github.com/0187773933/MastersClosetTracker/v1/log"
)

var s server.Server

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		log.PrintlnConsole( "\r- Ctrl+C pressed in Terminal" )
		log.PrintlnConsole( "Shutting Down Master's Closet Tracking Server" )
		s.FiberApp.Shutdown()
		os.Exit( 0 )
	}()
}

func main() {
	// utils.GenerateNewKeys()
	SetupCloseHandler()
	config_file_path , _ := filepath.Abs( os.Args[ 1 ] )
	fmt.Printf( "Loaded Config File From : %s\n" , config_file_path )
	config := utils.ParseConfig( config_file_path )
	config.FingerPrint = utils.FingerPrint( &config )
	fmt.Println( config )
	log.Init( config )
	s = server.New( config )
	s.Start()
}