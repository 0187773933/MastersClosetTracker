package server

import (
	"fmt"
	"time"
	// index_sort "github.com/mkmik/argsort"
	fiber "github.com/gofiber/fiber/v2"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	// try "github.com/manucorporat/try"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	// redis "github.com/0187773933/MastersClosetTracker/v1/redis"
	user_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/user"
	// admin_routes "github.com/0187773933/MastersClosetTracker/v1/server/routes/admin"
)

type Server struct {
	FiberApp *fiber.App `json:"fiber_app"`
	Config types.ConfigFile `json:"config"`
}

func New( config types.ConfigFile ) ( server Server ) {

	server.FiberApp = fiber.New()
	server.Config = config

	ip_addresses := utils.GetLocalIPAddresses()
	fmt.Println( "Server's IP Addresses === " , ip_addresses )
	// https://docs.gofiber.io/api/middleware/limiter
	server.FiberApp.Use( rate_limiter.New( rate_limiter.Config{
		Max: 2 ,
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
	server.SetupRoutes()
	return
}

func ( s *Server ) SetupRoutes() {
	user_routes.RegisterRoutes( s.FiberApp , &s.Config )
}

func ( s *Server ) Start() {
	fmt.Printf( "Listening on %s\n" , s.Config.ServerPort )
	s.FiberApp.Listen( fmt.Sprintf( ":%s" , s.Config.ServerPort ) )
}

