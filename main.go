package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"path/filepath"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/MastersClosetTracker/v1/server"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
)

var s *fiber.App

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		fmt.Println( "\r- Ctrl+C pressed in Terminal" )
		fmt.Println( "Shutting Down Master's Closet Tracking Server" )
		s.Shutdown()
		os.Exit( 0 )
	}()
}

func main() {
	SetupCloseHandler()
	config_file_path , _ := filepath.Abs( os.Args[ 1 ] )
	fmt.Println( "Loaded Config File From : %s" , config_file_path )
	config := utils.ParseConfig( config_file_path )
	fmt.Println( config )
	s = server.New( config )
	fmt.Printf( "Listening on %s\n" , config.ServerPort )
	s.Listen( fmt.Sprintf( ":%s" , config.ServerPort ) )
}