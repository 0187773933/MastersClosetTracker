package userroutes

import (
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	fmt.Println( "we out here in user routes" )
	fmt.Println( fiber_app )
}