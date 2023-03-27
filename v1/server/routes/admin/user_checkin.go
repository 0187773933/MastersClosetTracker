package adminroutes

import (
	"fmt"
	"time"
	"strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	printer "github.com/0187773933/MastersClosetTracker/v1/printer"
)

type CheckInBalanceForm struct {
	TopsAvailable int `json:"balance_tops_available"`
	TopsLimit int `json:"balance_tops_limit"`
	TopsUsed int `json:"balance_tops_used"`
	BottomsAvailable int `json:"balance_bottoms_available"`
	BottomsLimit int `json:"balance_bottoms_limit"`
	BottomsUsed int `json:"balance_bottoms_used"`
	DressesAvailable int `json:"balance_dresses_available"`
	DressesLimit int `json:"balance_dresses_limit"`
	DressesUsed int `json:"balance_dresses_used"`
	ShoesAvailable int `json:"balance_shoes_available"`
	ShoesLimit int `json:"balance_shoes_limit"`
	ShoesUsed int `json:"balance_shoes_used"`
	SeasonalsAvailable int `json:"balance_seasonals_available"`
	SeasonalsLimit int `json:"balance_seasonals_limit"`
	SeasonalsUsed int `json:"balance_seasonals_used"`
	AccessoriesAvailable int `json:"balance_accessories_available"`
	AccessoriesLimit int `json:"balance_accessories_limit"`
	AccessoriesUsed int `json:"balance_accessories_used"`
}


// We changed this to a POST Form , so now we have to parse it
func UserCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// 1.) Prep
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// 2.) Grab the User
	uploaded_uuid := context.Params( "uuid" )
	x_uuid := utils.SanitizeInputString( uploaded_uuid )
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
	var balance_form CheckInBalanceForm
	json.Unmarshal( []byte( context.Body() ), &balance_form )
	fmt.Println( balance_form )
	viewed_user.Balance.General.Tops.Available = balance_form.TopsAvailable
	viewed_user.Balance.General.Tops.Limit = balance_form.TopsLimit
	viewed_user.Balance.General.Tops.Used = balance_form.TopsUsed

	viewed_user.Balance.General.Bottoms.Available = balance_form.BottomsAvailable
	viewed_user.Balance.General.Bottoms.Limit = balance_form.BottomsLimit
	viewed_user.Balance.General.Bottoms.Used = balance_form.BottomsUsed

	viewed_user.Balance.General.Dresses.Available = balance_form.DressesAvailable
	viewed_user.Balance.General.Dresses.Limit = balance_form.DressesLimit
	viewed_user.Balance.General.Dresses.Used = balance_form.DressesUsed

	viewed_user.Balance.Shoes.Available = balance_form.ShoesAvailable
	viewed_user.Balance.Shoes.Limit = balance_form.ShoesLimit
	viewed_user.Balance.Shoes.Used = balance_form.ShoesUsed

	viewed_user.Balance.Seasonals.Available = balance_form.SeasonalsAvailable
	viewed_user.Balance.Seasonals.Limit = balance_form.SeasonalsLimit
	viewed_user.Balance.Seasonals.Used = balance_form.SeasonalsUsed

	viewed_user.Balance.Accessories.Available = balance_form.AccessoriesAvailable
	viewed_user.Balance.Accessories.Limit = balance_form.AccessoriesLimit
	viewed_user.Balance.Accessories.Used = balance_form.AccessoriesUsed

	fmt.Println( "Checking In With Balance :" )
	fmt.Println( viewed_user.Balance )

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
	// TODO : clarify calculation ????
	// total_clothing_items := ( balance_form.TopsAvailable + balance_form.BottomsAvailable + balance_form.DressesAvailable )
	// total_clothing_items := ( balance_form.TopsAvailable + balance_form.ShoesAvailable + balance_form.SeasonalsAvailable + balance_form.AccessoriesAvailable )
	total_clothing_items := ( balance_form.TopsAvailable )
	family_size := viewed_user.FamilySize
	if family_size < 1 { family_size = 1 } // this is what happens when you don't just use sql
	barcode_number := ""
	if len( viewed_user.Barcodes ) > 0 { barcode_number = viewed_user.Barcodes[ 0 ] }
	family_name := viewed_user.NameString
	// if len( family_name ) > 20 ? // TODO : Find max length of family string
	print_job := printer.PrintJob{
		FamilySize: family_size ,
		TotalClothingItems: total_clothing_items ,
		Shoes: balance_form.ShoesAvailable ,
		ShoesLimit: GlobalConfig.Balance.Shoes ,
		Accessories: balance_form.AccessoriesAvailable ,
		AccessoriesLimit: GlobalConfig.Balance.Accessories ,
		Seasonal: balance_form.SeasonalsAvailable ,
		SeasonalLimit: GlobalConfig.Balance.Seasonals ,
		FamilyName: family_name ,
		BarcodeNumber: barcode_number ,
	}
	fmt.Println( "Printing :" )
	fmt.Println( print_job )
	printer.PrintTicket( GlobalConfig.Printer , print_job )

	// 7.) Return Result
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/:uuid" ,
		"result": true ,
	})
}

func UserCheckInTestV2( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	x_user_uuid := context.Params( "uuid" )
	x_user := user.GetViaUUID( x_user_uuid , GlobalConfig )
	check_in_test := x_user.CheckInTest()
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/testv2/:uuid" ,
		"result": check_in_test ,
		"user": x_user ,
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