package adminroutes

import (
	"time"
	"strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
)

// We changed this to a POST Form , so now we have to parse it
func UserCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// 1.) Prep
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	uploaded_uuid := context.FormValue( "balance_tops_available" )
	x_uuid := utils.SanitizeInputString( uploaded_uuid )

	// 2.) Grab the User
	var viewed_user user.User
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( x_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})

	// 3.) Create a New Forced Check In
	var new_check_in user.CheckIn
	now := time.Now()
	// now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )
	new_check_in.Type = "forced"
	new_check_in.Date = strings.ToUpper( new_check_in.Date )
	viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )

	// 4.) Update the Balance
	viewed_user.Balance.General.Tops.Available = parse_form_value_as_int( context , "balance_tops_available" )
	viewed_user.Balance.General.Tops.Limit = parse_form_value_as_int( context , "balance_tops_limit" )
	viewed_user.Balance.General.Tops.Used = parse_form_value_as_int( context , "balance_tops_used" )

	viewed_user.Balance.General.Bottoms.Available = parse_form_value_as_int( context , "balance_bottoms_available" )
	viewed_user.Balance.General.Bottoms.Limit = parse_form_value_as_int( context , "balance_bottoms_limit" )
	viewed_user.Balance.General.Bottoms.Used = parse_form_value_as_int( context , "balance_bottoms_used" )

	viewed_user.Balance.General.Dresses.Available = parse_form_value_as_int( context , "balance_dresses_available" )
	viewed_user.Balance.General.Dresses.Limit = parse_form_value_as_int( context , "balance_dresses_limit" )
	viewed_user.Balance.General.Dresses.Used = parse_form_value_as_int( context , "balance_dresses_used" )

	viewed_user.Balance.Shoes.Available = parse_form_value_as_int( context , "balance_shoes_available" )
	viewed_user.Balance.Shoes.Limit = parse_form_value_as_int( context , "balance_shoes_limit" )
	viewed_user.Balance.Shoes.Used = parse_form_value_as_int( context , "balance_shoes_used" )

	viewed_user.Balance.Seasonals.Available = parse_form_value_as_int( context , "balance_seasonals_available" )
	viewed_user.Balance.Seasonals.Limit = parse_form_value_as_int( context , "balance_seasonals_limit" )
	viewed_user.Balance.Seasonals.Used = parse_form_value_as_int( context , "balance_seasonals_used" )

	viewed_user.Balance.Accessories.Available = parse_form_value_as_int( context , "balance_accessories_available" )
	viewed_user.Balance.Accessories.Limit = parse_form_value_as_int( context , "balance_accessories_limit" )
	viewed_user.Balance.Accessories.Used = parse_form_value_as_int( context , "balance_accessories_used" )

	// 5.) Re-Save the User
	viewed_user_byte_object , _ := json.Marshal( viewed_user )
	viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , viewed_user_byte_object )
	db_result := db.Update( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.Put( []byte( x_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }

	// 6.) Print Ticket
	// TODO !!!!! Where Barcode Numbers ??????
	// printer.PrintTicket( GlobalConfig.Printer , printer.PrintJob{
	// 	FamilySize: 5 ,
	// 	TotalClothingItems: 23 ,
	// 	Shoes: 1 ,
	// 	Accessories: 2 ,
	// 	Seasonal: 1 ,
	// 	FamilyName: "Cerbus" ,
	// })

	// 7.) Return Result
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/:uuid" ,
		"result": true ,
	})
}

func UserCheckInTest( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	check_in_test_result , time_remaining , balance , name_string , family_size := user.CheckInTest( user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )

	// idk where else to put this
	// only other option is maybe on the new user create form
	if check_in_test_result == true {
		balance = user.RefillBalance( user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.Balance , family_size )
	}

	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/test/:uuid" ,
		"result": check_in_test_result ,
		"time_remaining": time_remaining ,
		"balance": balance ,
		"name_string": name_string ,
		"family_size": family_size ,
	})
}