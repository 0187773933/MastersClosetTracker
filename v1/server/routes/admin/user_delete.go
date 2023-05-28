package adminroutes

import (
	"time"
	fiber "github.com/gofiber/fiber/v2"
	bleve "github.com/blevesearch/bleve/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	log "github.com/0187773933/MastersClosetTracker/v1/log"
)

func DeleteUser( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )

	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
	defer search_index.Close()
	search_index.Delete( user_uuid )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		users_bucket.Delete( []byte( user_uuid ) )
		usernames_bucket := tx.Bucket( []byte( "usernames" ) )
		usernames_bucket.Delete( []byte( viewed_user.Username ) )
		return nil
	})
	log.PrintlnConsole( viewed_user.UUID , "===" , "Deleted" )
	return context.JSON( fiber.Map{
		"route": "/admin/user/delete/:uuid" ,
		"result": "deleted" ,
	})
}