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

// Onboarding Experience
// 1. New QR code is generated at signup desk for new user
// 2. New user scans QR code with their phone
// 3. Takes them to a silent login page that stores a permanent login cookie.

// To Re-Enter
// 1. They scan a QR code on a poster at the front door or just go to the website.
// 2. If they have a cookie stored it returns a webpage with their unique QR code.
// 3. Displayed QR code gets scanned and validated

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	fiber_app.Get( "/checkin" , CheckIn )
	user_route_group := fiber_app.Group( "/user" )
	user_route_group.Get( "/login/:uuid" , Login )
	user_route_group.Get( "/checkin/display/:uuid" , Login )
	user_route_group.Get( "/checkin" , CheckIn )
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

func serve_failed_check_in_attempt( context *fiber.Ctx ) ( error ) {
	// return context.Redirect( "/join" )
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>check-in failed</h1>" )
}

func Login( context *fiber.Ctx ) ( error ) {
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	x_user_uuid := context.Params( "uuid" )
	x_user := user.GetByUUID( x_user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" {
		context.Set( "Content-Type" , "text/html" )
		return context.SendString( "<h1>Login Failed</h1>" )
	}
	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-user" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , x_user.UUID ) ,
			Secure: true , // dev
			Path: "/" , // fucking webkit
			// Domain: "9686-208-38-225-121.ngrok.io" , // probably should set this for webkit
			HTTPOnly: true ,
			SameSite: "Lax" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	return context.Redirect( fmt.Sprintf( "/user/checkin/display/%s" , x_user.UUID ) )
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new/:username
func CheckIn( context *fiber.Ctx ) ( error ) {

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// validate they have a stored user cookie
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie == "" { return serve_failed_check_in_attempt( context ) }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" { return serve_failed_check_in_attempt( context ) }

	return context.Redirect( fmt.Sprintf( "/user/checkin/display/%s" , x_user.UUID ) )
}

func CheckInDisplay( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/user_check_in.html"  )
}