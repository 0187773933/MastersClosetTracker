package userroutes

import (
	"fmt"
	"time"
	// net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	// utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	fiber_app.Get( "/" , RenderHomePage )
	fiber_app.Get( "/checkin" , CheckIn )
	user_route_group := fiber_app.Group( "/user" )
	user_route_group.Get( "/login/fresh/:uuid" , LoginFresh )
	// user_route_group.Get( "/login/success/:uuid" , LoginSuccess )
	user_route_group.Get( "/checkin/display/:uuid" , CheckInDisplay )
	user_route_group.Get( "/checkin" , CheckIn )
	user_route_group.Get( "/checkin/silent/:uuid" , CheckInSilentTest )
}

func check_if_user_cookie_exists( context *fiber.Ctx ) ( result bool ) {
	result = false
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie == "" { return }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	// result = uuid.IsUUID( user_cookie_value )
	if user_cookie_value != "" { result = true }
	return
}

func check_if_admin_cookie_exists( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie == "" { return }
	admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
	if admin_cookie_value != GlobalConfig.ServerCookieAdminSecretMessage { fmt.Println( "admin cookie secret message was not equal" ); return }
	result = true
	return
}

func serve_failed_check_in_attempt( context *fiber.Ctx ) ( error ) {
	// return context.Redirect( "/join" )
	context.Set( "Content-Type" , "text/html" )
	// return context.SendString( "<h1>check-in failed</h1>" )
	return context.SendFile( "./v1/server/html/user_check_in_failed.html" )
}

func RenderHomePage( context *fiber.Ctx ) ( error ) {
	context.Set( "Content-Type" , "text/html" )
	admin_logged_in := check_if_admin_cookie_exists( context )
	if admin_logged_in == true {
		return context.SendFile( "./v1/server/html/admin.html" )
	}
	user_logged_in := check_if_user_cookie_exists( context )
	if user_logged_in == true {
		return context.SendFile( "./v1/server/html/user_home.html" )
	}
	return context.SendFile( "./v1/server/html/home.html" )
}


func LoginFresh( context *fiber.Ctx ) ( error ) {
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	x_user_uuid := context.Params( "uuid" )
	x_user := user.GetByUUID( x_user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" {
		context.Set( "Content-Type" , "text/html" )
		return context.SendString( "<h1>Login Failed</h1>" )
	}

	// Manual Check In For First Time Login
	user.CheckInUser( x_user.UUID , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )

	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-user" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , x_user.UUID ) ,
			Secure: true , // dev
			Path: "/" , // fucking webkit
			// Domain: "blah.ngrok.io" , // probably should set this for webkit
			HTTPOnly: true ,
			SameSite: "Lax" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)

	return context.SendFile( "./v1/server/html/user_login_success.html" )
}

func CheckInSilentTest( context *fiber.Ctx ) ( error ) {
	x_user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	check_in_result , milliseconds_remaining := user.CheckInTest( x_user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	return context.JSON( fiber.Map{
		"route": "/user/checkin/silent/:uuid" ,
		"result": fiber.Map{
			"check_in_possible": check_in_result ,
			"milliseconds_remaining": milliseconds_remaining ,
		} ,
	})
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new/:username
func CheckIn( context *fiber.Ctx ) ( error ) {

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// validate they have a stored user cookie
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie == "" { fmt.Println( "user cookie was blank" ); return serve_failed_check_in_attempt( context ) }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" { fmt.Println( "UUID stored in user cookie was blank" ); return serve_failed_check_in_attempt( context ) }

	// TODO : render a different page if the check-in would fail ?
	check_in_test , time_remaining := user.CheckInUser( x_user.UUID , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	fmt.Println( "Pre-Check-In Test Result ===" , check_in_test , "Time Remaining ===" , time_remaining )
	return context.Redirect( fmt.Sprintf( "/user/checkin/display/%s" , x_user.UUID ) )
}

func CheckInDisplay( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/user_check_in.html" )
}