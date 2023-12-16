package logger

import (
	"os"
	"fmt"
	"strings"
	"io"
	// "encoding/json"
	types "github.com/0187773933/FireC2Server/v1/types"
	utils "github.com/0187773933/FireC2Server/v1/utils"
	logrus "github.com/sirupsen/logrus"
	// ulid "github.com/oklog/ulid/v2"
)

var Log *logrus.Logger
var config *types.ConfigFile

type CustomTextFormatter struct {
	logrus.TextFormatter
}

// https://github.com/sirupsen/logrus/blob/v1.9.3/entry.go#L44
// https://pkg.go.dev/github.com/sirupsen/logrus#Entry
func ( f *CustomTextFormatter ) Format( entry *logrus.Entry ) ( result_bytes []byte , result_error error ) {
	time_string := utils.FormatTime( &entry.Time )
	// result_bytes , result_error = f.TextFormatter.Format( entry )

	var result_string string
	if entry.Caller != nil {
		var caller_function string
		test_parts := strings.Split( entry.Caller.Function, "github.com/0187773933/MastersClosetTracker/v1/" )
		if len( test_parts ) > 1 {
			caller_function = test_parts[ 1 ]
		} else {
			caller_function = entry.Caller.Function
		}
		result_string = fmt.Sprintf( "%s === %s():%d === %s\n" , time_string , caller_function , entry.Caller.Line , entry.Message )
	} else {
		result_string = fmt.Sprintf( "%s === %s\n" , time_string , entry.Message )
	}
	result_bytes = []byte( result_string )
	result_error = nil

	// DB.Update( func( tx *bolt_api.Tx ) error {
	// 	b_logs := tx.Bucket( []byte( "logs" ) )
	// 	b_today , _ := b_logs.CreateBucketIfNotExists( []byte( db_log_prefix ) )
	// 	b_today.Put( []byte( ulid_prefix ) , message_bytes )
	// 	return nil
	// })

	// message := &CustomLogMessage{
	// 	Message: result_string ,
	// 	Fields: entry.Data ,
	// 	Time: time_string ,
	// 	Level: entry.Level.String() ,
	// }
	// if entry.Caller != nil {
	// 	message.Frame = CustomLogMessageFrame{
	// 		// Function: entry.Caller.Function ,
	// 		File: entry.Caller.File ,
	// 		Line: entry.Caller.Line ,
	// 	}
	// }
	// db_log_prefix := utils.FormatDBLogPrefix( &entry.Time )
	// ulid_prefix := ulid.Make().String()
	// message_bytes , _ := json.Marshal( message )
	// DB.Update( func( tx *bolt_api.Tx ) error {
	// 	b_logs := tx.Bucket( []byte( "logs" ) )
	// 	b_today , _ := b_logs.CreateBucketIfNotExists( []byte( db_log_prefix ) )
	// 	b_today.Put( []byte( ulid_prefix ) , message_bytes )
	// 	return nil
	// })

	return result_bytes , result_error
}

type CustomLogrusWriter struct {
	io.Writer
}

func ( w *CustomLogrusWriter ) Write( p []byte ) ( n int , err error ) {
	message := string( p )
	n , err = fmt.Fprint( os.Stdout , message )
	return n , err
}

type CustomJSONFormatter struct {
	logrus.JSONFormatter
}

func ( f *CustomJSONFormatter ) Format( entry *logrus.Entry ) ( []byte , error ) {
	time_string := utils.FormatTime( &entry.Time )
	fmt.Println( time_string )
	fmt.Println( entry )
	return f.JSONFormatter.Format( entry )
}




// so apparently The limitation arises due to the Go language's initialization order:
// Package-level variables are initialized before main() is called.
// Functions in main() execute after package-level initializations.
// something something , singleton
func GetLogger() *logrus.Logger {
	if Log == nil { Init() }
	return Log
}

func Init() {
	Log = logrus.New()
	log_level := os.Getenv( "LOG_LEVEL" )
	fmt.Printf( "LOG_LEVEL=%s\n" , log_level )
	switch log_level {
		case "debug":
			Log.SetReportCaller( true )
			Log.SetLevel( logrus.DebugLevel )
		default:
			Log.SetReportCaller( false )
			Log.SetLevel( logrus.InfoLevel )
	}
	Log.SetFormatter( &CustomTextFormatter{
		TextFormatter: logrus.TextFormatter{
			DisableColors: false ,
		} ,
	})
	// log.SetFormatter( &CustomJSONFormatter{
	// 	JSONFormatter: logrus.JSONFormatter{} ,
	// })

	// log.SetOutput( os.Stdout )
	Log.SetOutput( &CustomLogrusWriter{} )
}