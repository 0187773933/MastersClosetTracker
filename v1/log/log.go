package log

import (
	// "io"
	// "os"
	"time"
	"fmt"
	"log"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	// logrus "github.com/sirupsen/logrus"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
)
// var log Logger
// type Logger struct {
// 	Logrus *logrus.Logger
// }
func Init( config types.ConfigFile ) {
	prepended_timestamp := time.Now().Format( "20060102" )
	// log.SetFlags( 0 )
	rotator := &lumberjack.Logger{
		Filename: fmt.Sprintf( "./logs/%s-%s.log" , prepended_timestamp , config.FingerPrint ) ,
		MaxSize: 100 , // megabytes
		// MaxBackups: 3 ,   // number of backups
		MaxAge: 1 , // days
		Compress: true , // compress the rotated log files
	}
	// mw := io.MultiWriter( os.Stdout , rotator )
	// log = Logger{}
	// log.Logrus = logrus.New()
	// // log.Logrus.SetFormatter( &logrus.JSONFormatter{
	// log.Logrus.SetFormatter( &logrus.TextFormatter{
	//    DisableTimestamp: true ,
	// })
	// log.Logrus.SetOutput( mw )
	// log.Logrus.SetOutput( rotator )
	log.SetFlags( 0 )
	log.SetOutput( rotator )
}

func Println( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ time_string , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Println( args... )
	// log.Logrus.Println( args... )
	log.Println( args... )
}

func PrintlnConsole(format string, args ...interface{}) {
    time_string := utils.GetFormattedTimeString()
    msg := fmt.Sprintf(format, args...)
    log.Println(time_string, "===", msg)
    fmt.Println(time_string, "===", msg)
}

func Printf( format_string string , args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	sent_format := fmt.Sprintf( format_string , args... )
	log.Printf( "%s === %s" , time_string , sent_format )
}

func PrintfConsole( format_string string , args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	sent_format := fmt.Sprintf( format_string , args... )
	log.Printf( "%s === %s" , time_string , sent_format )
	fmt.Printf( "%s === %s" , time_string , sent_format )
}


func Debug( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ time_string , "===" , "DEBUG" , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Debug( args... )
	// log.Logrus.Debug( args... )
	log.Println( args... )
}

func Error( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ "ERROR !!!" , time_string , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Debug( args... )
	// log.Logrus.Debug( args... )
	log.Println( args... )
	fmt.Println( args... )
}
