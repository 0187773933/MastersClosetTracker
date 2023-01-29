package userroutes

import (
	// "fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	user := fiber_app.Group( "/user" )
	user.Get( "/new" , New )
}

// https://docs.gofiber.io/api/ctx#cookie
func New( context *fiber.Ctx ) ( error ) {
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