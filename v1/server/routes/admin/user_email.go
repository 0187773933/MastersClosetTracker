package adminroutes

import (
	fmt "fmt"
	time "time"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	// log "github.com/0187773933/MastersClosetTracker/v1/log"
)

func EmailAllUsers( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	email_message := context.FormValue( "email_message" )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.EmailAddress == "" { return nil; }
			fmt.Printf( "%s === %s\n" , "from@example.com" , viewed_user.EmailAddress )
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/user/email/all" ,
		"email_message": email_message ,
		"result": "not implemented" ,
	})
}