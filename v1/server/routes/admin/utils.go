package adminroutes

import (
	"fmt"
	"strings"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	printer "github.com/0187773933/MastersClosetTracker/v1/printer"
)

// weak attempt at sanitizing form input to build a "username"
func SanitizeUsername( first_name string , last_name string ) ( username string ) {
	if first_name == "" { first_name = "Not Provided" }
	if last_name == "" { last_name = "Not Provided" }
	sanitized_first_name := utils.SanitizeInputString( first_name )
	sanitized_last_name := utils.SanitizeInputString( last_name )
	username = fmt.Sprintf( "%s-%s" , sanitized_first_name , sanitized_last_name )
	return
}

func serve_failed_attempt( context *fiber.Ctx ) ( error ) {
	// context.Set( "Content-Type" , "text/html" )
	// return context.SendString( "<h1>no</h1>" )
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func ServeLoginPage( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func ServeAuthenticatedPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	x_path := context.Route().Path
	url_key := strings.Split( x_path , "/admin" )
	if len( url_key ) < 2 { return context.SendFile( "./v1/server/html/admin_login.html" ) }
	// fmt.Println( "Sending -->" , url_key[ 1 ] , x_path )
	return context.SendFile( ui_html_pages[ url_key[ 1 ] ] )
}

func PrintTest( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	printer.PrintTicket( GlobalConfig.Printer , printer.PrintJob{
		FamilySize: 5 ,
		TotalClothingItems: 23 ,
		Shoes: 1 ,
		Accessories: 2 ,
		Seasonal: 1 ,
		FamilyName: "Cerbus" ,
	})
	return context.JSON( fiber.Map{
		"route": "/admin/print-test" ,
		"result": "success" ,
	})
}