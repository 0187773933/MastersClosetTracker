package userroutes

import (
	"fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
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
	user_route_group := fiber_app.Group( "/user" )
	user_route_group.Get( "/new" , New )
	user_route_group.Get( "/get/:uuid" , GetUser )
	user_route_group.Get( "/checkin/:uuid" , CheckIn )
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new
func New( context *fiber.Ctx ) ( error ) {
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	new_user := user.New( "new-user-test" , db , GlobalConfig.BoltDBEncryptionKey )
	fmt.Println( new_user )
	context.Cookie(
		&fiber.Cookie{
			Name: "new-user-test" ,
			Value: "blah blah blah , probably some nacl salsa boxed value" ,
			Secure: true ,
			SameSite: "strict" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	return context.JSON( fiber.Map{
		"route": "/user/new" ,
		"result": "temp" ,
	})
}

// http://localhost:5950/user/get/04b5fba6-6d76-42e0-a543-863c3f0c252c
func GetUser( context *fiber.Ctx ) ( error ) {
	user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	return context.JSON( fiber.Map{
		"route": "/user/get/:uuid" ,
		"result": viewed_user ,
	})
}

// http://localhost:5950/user/checkin/04b5fba6-6d76-42e0-a543-863c3f0c252c
func CheckIn( context *fiber.Ctx ) ( error ) {
	user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	check_in_result := user.CheckInUser( user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	return context.JSON( fiber.Map{
		"route": "/user/checkin/:uuid" ,
		"result": check_in_result ,
	})
}
