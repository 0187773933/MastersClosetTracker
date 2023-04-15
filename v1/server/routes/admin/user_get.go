package adminroutes

import (
	"time"
	"fmt"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
)

// http://localhost:5950/user/get/04b5fba6-6d76-42e0-a543-863c3f0c252c
func GetUser( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/:uuid" ,
		"result": viewed_user ,
	})
}

func GetAllUsers( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	var result []user.GetUserResult
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			var get_user_result user.GetUserResult
			get_user_result.Username = viewed_user.Username
			get_user_result.UUID = viewed_user.UUID
			if len( viewed_user.CheckIns ) > 0 {
				get_user_result.LastCheckIn = viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
			}
			result = append( result , get_user_result )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all" ,
		"result": result ,
	})
}

func GetAllCheckIns( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	var result [][]user.CheckIn
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) > 0 {
				result = append( result , viewed_user.CheckIns )
			}
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/checkins" ,
		"result": result ,
	})
}

func GetAllEmails( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	var result [][]string
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.EmailAddress != "" {
				x_user := []string{ viewed_user.UUID , viewed_user.NameString , viewed_user.EmailAddress }
				result = append( result , x_user )
			}
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/checkins" ,
		"result": result ,
	})
}


func GetUserViaBarcode( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	barcode := context.Params( "barcode" )
	var viewed_user user.User
	db.View( func( tx *bolt_api.Tx ) error {
		barcode_bucket := tx.Bucket( []byte( "barcodes" ) )
		x_uuid := barcode_bucket.Get( []byte( barcode ) )
		if x_uuid == nil { return nil }
		fmt.Printf( "Barcode : %s || UUID : %s\n" , barcode , x_uuid )
		user_bucket := tx.Bucket( []byte( "users" ) )
		x_user := user_bucket.Get( []byte( x_uuid ) )
		decrypted_user := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , x_user )
		json.Unmarshal( decrypted_user , &viewed_user )
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/barcode" ,
		"result": viewed_user ,
	})
}
