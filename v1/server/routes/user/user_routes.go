package userroutes

import (
	"fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt_db "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	user_route_group := fiber_app.Group( "/user" )
	user_route_group.Get( "/new" , New )
	user_route_group.Get( "/get/:uuid" , GetUser )
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new
func New( context *fiber.Ctx ) ( error ) {
	db , _ := bolt_db.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_db.Options{ Timeout: ( 3 * time.Second ) } )
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
	db , _ := bolt_db.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_db.Options{ Timeout: ( 3 * time.Second ) } )
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	fmt.Println( viewed_user )
	return context.JSON( fiber.Map{
		"route": "/user/get/:uuid" ,
		"result": "temp" ,
	})
}