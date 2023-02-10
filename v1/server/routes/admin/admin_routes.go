package adminroutes

import (
	// "fmt"
	"time"
	// net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	bcrypt "golang.org/x/crypto/bcrypt"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	// bolt_api "github.com/boltdb/bolt"
	// user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	admin_route_group := fiber_app.Group( "/admin" )
	admin_route_group.Get( "/login" , ServeLoginPage )
	admin_route_group.Post( "/login" , HandleLogin )
	admin_route_group.Get( "/" , AdminPage )
}

// GET http://localhost:5950/admin/login
func ServeLoginPage( context *fiber.Ctx ) ( error ) {
	// context.Set( "Content-Type" , "text/html" )
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func validate_login_credentials( context *fiber.Ctx ) ( result bool ) {
	result = false
	uploaded_username := context.FormValue( "username" )
	if uploaded_username == "" { return }
	if uploaded_username != GlobalConfig.AdminUsername { return }
	uploaded_password := context.FormValue( "password" )
	if uploaded_password == "" { return }
	password_matches := bcrypt.CompareHashAndPassword( []byte( uploaded_password ) , []byte( GlobalConfig.AdminPassword ) )
	if password_matches != nil { return }
	result = true
	return
}

func serve_failed_attempt( context *fiber.Ctx ) ( error ) {
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>no</h1>" )
}

// POST http://localhost:5950/admin/login
func HandleLogin( context *fiber.Ctx ) ( error ) {
	valid_login := validate_login_credentials( context )
	if valid_login == false { return serve_failed_attempt( context ) }
	// if valid_login == false {
	// 	context.Set( "Content-Type" , "text/html" )
	// 	return context.SendString( "<h1>no</h1>" )
	// }
	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-admin" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , GlobalConfig.ServerCookieAdminSecretMessage ) ,
			Secure: true ,
			SameSite: "strict" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	return context.Redirect( "/admin" )
}

func validate_admin_cookie( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie == "" { return }
	admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
	if admin_cookie_value != GlobalConfig.ServerCookieAdminSecretMessage { return }
	result = true
	return
}


func AdminPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin.html" )
}