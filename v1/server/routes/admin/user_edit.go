package adminroutes

import (
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	// pp "github.com/k0kubun/pp/v3"
	// pp.Println( viewed_user )
)

// could move to user , but edit should be the only thing like this
func HandleUserEdit( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	viewed_user.Config = GlobalConfig
	viewed_user.Save();
	return context.JSON( fiber.Map{
		"route": "/admin/user/edit" ,
		"result": true ,
		"user": viewed_user ,
	})
}