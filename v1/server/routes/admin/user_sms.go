package adminroutes

import (
	"fmt"
	time "time"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	twilio "github.com/sfreiberg/gotwilio"
	log "github.com/0187773933/MastersClosetTracker/v1/log"
)

func SMSAllUsers( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	sms_message := context.FormValue( "sms_message" )

	twilio_client := twilio.NewTwilioClient( GlobalConfig.TwilioClientID , GlobalConfig.TwilioAuthToken )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.PhoneNumber == "" { return nil; }

			to_number := fmt.Sprintf( "+1%s" , viewed_user.PhoneNumber )
			// https://github.com/sfreiberg/gotwilio/blob/master/sms.go#L12
			result , _ , _ := twilio_client.SendSMS( GlobalConfig.TwilioSMSFromNumber , to_number , sms_message , "" , "" )
			log.Printf( "%s === %s\n" , to_number , result.Status )
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/user/sms/all" ,
		"sms_message": sms_message ,
		"result": "success" ,
	})
}