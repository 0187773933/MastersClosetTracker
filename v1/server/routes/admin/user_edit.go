package adminroutes

import (
	"fmt"
	"time"
	"strings"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	bleve "github.com/blevesearch/bleve/v2"
	// pp "github.com/k0kubun/pp/v3"
		// pp.Println( viewed_user )
)

func HandleUserEdit2( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	// pp.Println( viewed_user )
	viewed_user.FormatUsername()
	viewed_user.Config = GlobalConfig
	viewed_user.Save();
	return context.JSON( fiber.Map{
		"route": "/admin/user/edit2" ,
		"result": viewed_user ,
	})
}

func HandleUserEdit( context *fiber.Ctx ) ( error ) {

	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// 1.) Create New User From Uploaded Form Fields
	new_user := ProcessNewUserForm( context )
	editing_uuid := context.FormValue( "editing_uuid" )
	new_user.UUID = editing_uuid

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// 2.) Grab Old Username so we can check to make sure it wasn't changed
	old_user := user.GetByUUID( editing_uuid , db , GlobalConfig.BoltDBEncryptionKey )

	fmt.Println( "Editing User :" )
	fmt.Println( new_user )
	fmt.Println( "barcodes ===" , new_user.Barcodes )

	// 3.) Store User in DB
	new_user_byte_object , _ := json.Marshal( new_user )
	new_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , new_user_byte_object )
	db_result := db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		users_bucket.Put( []byte( new_user.UUID ) , new_user_byte_object_encrypted )
		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		// something something holographic encryption would be nice here

		if old_user.Username != new_user.Username {
			usernames_bucket.Delete( []byte( old_user.Username ) )
			search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
			defer search_index.Close()
			edited_search_item := types.SearchItem{
				UUID: new_user.UUID ,
				Name: strings.ReplaceAll( new_user.Username , "-" , " " ) ,
			}
			search_index.Index( new_user.UUID , edited_search_item )
		}
		usernames_bucket.Put( []byte( new_user.Username ) , []byte( new_user.UUID ) )

		barcodes_bucket , _ := tx.CreateBucketIfNotExists( []byte( "barcodes" ) )
		for i := 0; i < len( new_user.Barcodes ); i++ {
			barcodes_bucket.Put( []byte( new_user.Barcodes[ i ] ) , []byte( new_user.UUID ) )
			// TODO , handle what happens if we remove a barcode from a user
			// Not really that big of a problem , since this just updates the barcode for the right uuid anyway
		}

		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }

	//return context.Redirect( fmt.Sprintf( "/admin/user/new/handoff/%s" , new_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/edit" ,
		"result": "saved" ,
	})
}