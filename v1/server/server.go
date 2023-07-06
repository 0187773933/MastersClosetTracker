package server

import (
	"fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	fiber_cookie "github.com/gofiber/fiber/v2/middleware/encryptcookie"
	fiber_cors "github.com/gofiber/fiber/v2/middleware/cors"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	favicon "github.com/gofiber/fiber/v2/middleware/favicon"
	// try "github.com/manucorporat/try"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	user_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/user"
	admin_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/admin"
	// "os"
	log "github.com/0187773933/MastersClosetTracker/v1/log"
)

var GlobalConfig *types.ConfigFile

type Server struct {
	FiberApp *fiber.App `json:"fiber_app"`
	Config types.ConfigFile `json:"config"`
}

func request_logging_middleware( context *fiber.Ctx ) ( error ) {
	ip_address := context.Get( "x-forwarded-for" )
	if ip_address == "" { ip_address = context.IP() }
	// log_message := fmt.Sprintf( "%s === %s === %s === %s === %s" , time_string , GlobalConfig.FingerPrint , ip_address , context.Method() , context.Path() )
	// log_message := fmt.Sprintf( "%s === %s === %s === %s" , time_string , GlobalConfig.FingerPrint , context.Method() , context.Path() )
	// log_message := fmt.Sprintf( "%s === %s" , context.Method() , context.Path() )
	log_message := fmt.Sprintf( "%s === %s === %s" , ip_address , context.Method() , context.Path() );
	log.Println( log_message )
	return context.Next()
}

func New( config types.ConfigFile ) ( server Server ) {

	server.FiberApp = fiber.New()
	server.Config = config
	GlobalConfig = &config

	// ip_addresses := utils.GetLocalIPAddresses()
	// fmt.Println( "Server's IP Addresses === " , ip_addresses )
	// https://docs.gofiber.io/api/middleware/limiter
	server.FiberApp.Use( request_logging_middleware )
	server.FiberApp.Use( favicon.New() )
	// server.FiberApp.Get( "/favicon.ico" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/favicon.ico" ) } )
	server.FiberApp.Use( rate_limiter.New( rate_limiter.Config{
		Max: 30 ,
		Expiration: ( 1 * time.Second ) ,
		// Next: func( c *fiber.Ctx ) bool {
		// 	ip := c.IP()
		// 	fmt.Println( ip )
		// 	return ip == "127.0.0.1"
		// } ,
		LimiterMiddleware: rate_limiter.SlidingWindow{} ,
		KeyGenerator: func( c *fiber.Ctx ) string {
			return c.Get( "x-forwarded-for" )
		} ,
		LimitReached: func( c *fiber.Ctx ) error {
			ip := c.IP()
			fmt.Printf( "%s === limit reached\n" , ip )
			c.Set( "Content-Type" , "text/html" )
			return c.SendString( "<html><h1>why</h1></html>" )
		} ,
		// Storage: myCustomStorage{}
		// monkaS
		// https://github.com/gofiber/fiber/blob/master/middleware/limiter/config.go#L53
	}))
	// temp_key := fiber_cookie.GenerateKey()
	// fmt.Println( temp_key )
	server.FiberApp.Use( fiber_cookie.New( fiber_cookie.Config{
		Key: server.Config.ServerCookieSecret ,
		// Key: temp_key ,
	}))

    server.FiberApp.Use( fiber_cors.New( fiber_cors.Config{
		AllowOrigins: server.Config.ServerBaseUrl ,
    }))

	// server.FiberApp.Static( "/cdn" , "./v1/server/cdn" )
	// just white-list static stuff
	server.FiberApp.Get( "/logo.png" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/logo.png" ) } )
	// server.FiberApp.Get( "/cdn/api.js" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/api.js" ) } )
	server.FiberApp.Get( "/cdn/utils.js" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/utils.js" ) } )
	server.FiberApp.Get( "/cdn/ui.js" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/ui.js" ) } )
	server.SetupRoutes()
	server.FiberApp.Get( "/*" , func( context *fiber.Ctx ) ( error ) { return context.Redirect( "/" ) } )
	return
}

func ( s *Server ) SetupRoutes() {
	user_routes.RegisterRoutes( s.FiberApp , &s.Config )
	admin_routes.RegisterRoutes( s.FiberApp , &s.Config )
}

func ( s *Server ) Start() {
	fmt.Println( "\n" )
	log.PrintfConsole( "Listening on http://localhost:%s\n" , s.Config.ServerPort )
	fmt.Printf( "Admin Login @ http://localhost:%s/admin/login\n" , s.Config.ServerPort )
	fmt.Printf( "Admin Login @ %s/admin/login\n" , s.Config.ServerLiveUrl )
	fmt.Printf( "Admin Username === %s\n" , s.Config.AdminUsername )
	fmt.Printf( "Admin Password === %s\n" , s.Config.AdminPassword )
	s.FiberApp.Listen( fmt.Sprintf( ":%s" , s.Config.ServerPort ) )
}

