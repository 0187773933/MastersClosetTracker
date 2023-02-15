package userroutes

import (
	"fmt"
	"time"
	// net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
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
	fiber_app.Get( "/join" , NewUserJoinPage )
	fiber_app.Post( "/join" , HandleNewUserJoin )
	fiber_app.Get( "/checkin" , CheckIn )
	fiber_app.Get( "/checkin/:uuid" , CheckInDisplay )
	// user_route_group := fiber_app.Group( "/user" )
	// user_route_group.Get( "/new/:username" , New )
	// user_route_group.Get( "/checkin/:uuid" , CheckIn )
}

// probably should just change this route to be the root "/" url ,
// and if it already detects a cookie , redirect elsewhere
func NewUserJoinPage( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/user_new.html" )
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


// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new/:username
func HandleNewUserJoin( context *fiber.Ctx ) ( error ) {

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// if they already have a stored user cookie
	// if they do , just redirect to /checkin
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie != "" {
		user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
		x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
		if x_user.UUID != "" {
			return context.Redirect( "/checkin" )
		}
	}

	// weak attempt at sanitizing form input to build a "username"
	uploaded_first_name := context.FormValue( "first_name" )
	if uploaded_first_name == "" { uploaded_first_name = "Not Provided" }
	uploaded_last_name := context.FormValue( "last_name" )
	if uploaded_last_name == "" { uploaded_last_name = "Not Provided" }
	sanitized_first_name := utils.SanitizeInputName( uploaded_first_name )
	sanitized_last_name := utils.SanitizeInputName( uploaded_last_name )
	username := fmt.Sprintf( "%s-%s" , sanitized_first_name , sanitized_last_name )

	new_user := user.New( username , db , GlobalConfig.BoltDBEncryptionKey )
	fmt.Println( new_user )
	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-user" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , new_user.UUID ) ,
			Secure: true ,
			SameSite: "strict" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	// context.Cookie(
	// 	&fiber.Cookie{
	// 		Name: "the-masters-closet-user-uuid" ,
	// 		Value: new_user.UUID ,
	// 		Secure: true ,
	// 		SameSite: "strict" ,
	// 		Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
	// 	} ,
	// )
	return context.Redirect( "/checkin" )
}

func serve_failed_attempt( context *fiber.Ctx ) ( error ) {
	// return context.Redirect( "/join" )
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>check-in failed</h1>" )
}

// http://localhost:5950/user/checkin/04b5fba6-6d76-42e0-a543-863c3f0c252c
func CheckIn( context *fiber.Ctx ) ( error ) {

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// validate there is a valid uuid stored in the user cookie
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie == "" { return serve_failed_attempt( context ) }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )

	// validate user's uuid exists
	x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" { return context.Redirect( "/join" ) }

	fmt.Println( x_user )
	// check_in_result := user.CheckInUser( user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	// return context.JSON( fiber.Map{
	// 	"route": "/user/checkin/:uuid" ,
	// 	"result": x_user ,
	// })

	// return context.SendFile( "./v1/server/html/user_check_in.html" , context.Query( "uuid" , x_user.UUID ) )

	// there has to be a better way of doing this
	// return context.SendFile( "./v1/server/html/user_check_in.html" )
	return context.Redirect( fmt.Sprintf( "/checkin/%s" , x_user.UUID ) )
}

func CheckInDisplay( context *fiber.Ctx ) ( error ) {

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	// user_cookie := context.Cookies( "the-masters-closet-user" )
	// if user_cookie == "" { return serve_failed_attempt( context ) }
	// user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )

	// // validate user's uuid exists
	// x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
	// if x_user.UUID == "" { return context.Redirect( "/join" ) }

	return context.SendFile( "./v1/server/html/user_check_in.html"  )
}