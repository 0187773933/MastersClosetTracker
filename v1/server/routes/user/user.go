package userroutes

import (
	// "fmt"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	user := fiber_app.Group( "/user" )
	user.Get( "/new" , New )
}

func New( context *fiber.Ctx ) ( error ) {
	return context.JSON( fiber.Map{
		"route": "/user/new" ,
		"result": "temp" ,
	})
}