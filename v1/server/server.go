package server

import (
	"fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	fiber_cookie "github.com/gofiber/fiber/v2/middleware/encryptcookie"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	// try "github.com/manucorporat/try"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	user_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/user"
	admin_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/admin"
)

type Server struct {
	FiberApp *fiber.App `json:"fiber_app"`
	Config types.ConfigFile `json:"config"`
}

func request_logging_middleware( context *fiber.Ctx ) ( error ) {
	fmt.Printf( "%s === %s === %s\n" , context.IP() , context.Method() , context.Path() )
	return context.Next()
}

// https://docs.gofiber.io/api/middleware/encryptcookie/
func New( config types.ConfigFile ) ( server Server ) {

	server.FiberApp = fiber.New()
	server.Config = config

	ip_addresses := utils.GetLocalIPAddresses()
	fmt.Println( "Server's IP Addresses === " , ip_addresses )
	// https://docs.gofiber.io/api/middleware/limiter
	server.FiberApp.Use( request_logging_middleware )
	server.FiberApp.Get( "/favicon.ico" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/favicon.ico" ) } )
	server.FiberApp.Use( rate_limiter.New( rate_limiter.Config{
		Max: 4 ,
		Expiration: ( 4 * time.Second ) ,
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
	// has to be fucking base64 encoded ??? who does this
	// whatever , lets let it generate us one and see what it looks like
	// temp_key := fiber_cookie.GenerateKey()
	// fmt.Println( temp_key )
	server.FiberApp.Use( fiber_cookie.New( fiber_cookie.Config{
		Key: server.Config.ServerCookieSecret ,
		// Key: temp_key ,
	}))
	// server.FiberApp.Static( "/cdn" , "./v1/server/cdn" )
	// just white-list static stuff
	server.FiberApp.Get( "/logo.png" , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/logo.png" ) } )
	server.SetupRoutes()
	return
}

func ( s *Server ) SetupRoutes() {
	user_routes.RegisterRoutes( s.FiberApp , &s.Config )
	admin_routes.RegisterRoutes( s.FiberApp , &s.Config )
}

func ( s *Server ) Start() {
	fmt.Printf( "Listening on %s\n" , s.Config.ServerPort )
	s.FiberApp.Listen( fmt.Sprintf( ":%s" , s.Config.ServerPort ) )
}

