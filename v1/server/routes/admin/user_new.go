package adminroutes

import (
	"fmt"
	"time"
	"strings"
	"reflect"
	"strconv"
	json "encoding/json"
	uuid "github.com/satori/go.uuid"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	bleve "github.com/blevesearch/bleve/v2"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
)

func ProcessNewUserForm( context *fiber.Ctx ) ( new_user user.User ) {

	uploaded_first_name := context.FormValue( "user_first_name" )
	uploaded_last_name := context.FormValue( "user_last_name" )
	uploaded_user_middle_name := context.FormValue( "user_middle_name" )
	uploaded_user_email := context.FormValue( "user_email" )
	uploaded_phone_number := context.FormValue( "user_phone_number" )
	uploaded_user_street_number := context.FormValue( "user_street_number" )
	uploaded_user_street_name := context.FormValue( "user_street_name" )
	uploaded_user_address_two := context.FormValue( "user_address_two" )
	uploaded_user_city := context.FormValue( "user_city" )
	uploaded_user_state := context.FormValue( "user_state" )
	uploaded_user_zip_code := context.FormValue( "user_zip_code" )
	uploaded_user_birth_day := context.FormValue( "user_birth_day" )
	uploaded_user_birth_month := context.FormValue( "user_birth_month" )
	uploaded_user_birth_year := context.FormValue( "user_birth_year" )
	uploaded_user_family_size := context.FormValue( "user_family_size" )
	uploaded_total_barcodes := context.FormValue( "total_barcodes" )

	new_user.EmailAddress = utils.SanitizeInputString( uploaded_user_email )
	new_user.PhoneNumber = utils.SanitizeInputString( uploaded_phone_number )
	new_user.Identity.FirstName = utils.SanitizeInputString( uploaded_first_name )
	new_user.Identity.MiddleName = utils.SanitizeInputString( uploaded_user_middle_name )
	new_user.Identity.LastName = utils.SanitizeInputString( uploaded_last_name )
	new_user.Identity.Address.StreetNumber = utils.SanitizeInputString( uploaded_user_street_number )
	new_user.Identity.Address.StreetName = utils.SanitizeInputString( uploaded_user_street_name )
	new_user.Identity.Address.AddressTwo = utils.SanitizeInputString( uploaded_user_address_two )
	new_user.Identity.Address.City = utils.SanitizeInputString( uploaded_user_city )
	new_user.Identity.Address.State = utils.SanitizeInputString( uploaded_user_state )
	new_user.Identity.Address.ZipCode = utils.SanitizeInputString( uploaded_user_zip_code )

	sanitized_birth_day := utils.SanitizeInputString( uploaded_user_birth_day )
	sanitized_birth_day_int , _ := strconv.Atoi( sanitized_birth_day )
	new_user.Identity.DateOfBirth.Day = sanitized_birth_day_int
	new_user.Identity.DateOfBirth.Month = utils.SanitizeInputString( uploaded_user_birth_month )
	sanitized_birth_year := utils.SanitizeInputString( uploaded_user_birth_year )
	sanitized_birth_year_int , _ := strconv.Atoi( sanitized_birth_year )
	new_user.Identity.DateOfBirth.Year = sanitized_birth_year_int

	sanitized_total_barcodes := utils.SanitizeInputString( uploaded_total_barcodes )
	sanitized_total_barcodes_int , _ := strconv.Atoi( sanitized_total_barcodes )
	if sanitized_total_barcodes_int > 0 {
		for i := 0; i < sanitized_total_barcodes_int; i++ {
			uploaded_barcode := context.FormValue( fmt.Sprintf( "user_barcode_%d" , ( i + 1 ) ) )
			sanitized_barcode := utils.SanitizeInputString( uploaded_barcode )
			new_user.Barcodes = append( new_user.Barcodes , sanitized_barcode )
		}
	}

	sanitized_family_size := utils.SanitizeInputString( uploaded_user_family_size )
	sanitized_family_size_int , _ := strconv.Atoi( sanitized_family_size )
	new_user.FamilySize = sanitized_family_size_int
	if sanitized_family_size_int > 0 {
		for i := 0; i < sanitized_family_size_int; i++ {
			uploaded_family_member_age := context.FormValue( fmt.Sprintf( "user_family_member_%d_age" , ( i + 1 ) ) )
			sanitized_family_member_age := utils.SanitizeInputString( uploaded_family_member_age )
			family_member_age_int , _ := strconv.Atoi( sanitized_family_member_age )
			var family_member user.Person
			family_member.Age = family_member_age_int
			new_user.FamilyMembers = append( new_user.FamilyMembers , family_member )
			// fmt.Printf( "Adding Family Member - %d - Age = %d\n" , ( i + 1 ) , family_member_age_int )
		}
	}

	user.FormatUsername( &new_user )

	new_user.UUID = uuid.NewV4().String()

	now := time.Now()
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )

	return
}

func HandleNewUserJoin2( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	// pp.Println( viewed_user )
	fmt.Println( viewed_user )
	viewed_user.Config = GlobalConfig

	viewed_user.FormatUsername()
	new_user := user.New( viewed_user.Username , GlobalConfig )
	fmt.Println( new_user )
	viewed_user.UUID = new_user.UUID
	viewed_user.CreatedDate = new_user.CreatedDate
	viewed_user.CreatedTime = new_user.CreatedTime
	viewed_user.Save()

	// viewed_user.Save();
	return context.JSON( fiber.Map{
		"route": "/admin/user/new2" ,
		"result": viewed_user ,
	})
}

func HandleNewUserJoin( context *fiber.Ctx ) ( error ) {

	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// 1.) Create New User From Uploaded Form Fields
	new_user := ProcessNewUserForm( context )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()


	// 2.) Early Return if User Already Exists
	// TODO : Add more sophisticated exists? check
	username_exists , exists_uuid := user.UserNameExists( new_user.Username , db )
	if username_exists == true {
		fmt.Printf( "User Already Exists with Username === %s === %s\n" , new_user.Username , exists_uuid )
		return context.JSON( fiber.Map{
			"route": "/admin/user/new" ,
			"result": fiber.Map{
				"error": "already exists" ,
				"uuid": exists_uuid ,
			} ,
		})
	}

	fmt.Println( "New User Created :" )
	fmt.Println( new_user )

	// we just need a way to map multiple barcodes --> uuid

	// 3.) Store User in DB
	new_user_byte_object , _ := json.Marshal( new_user )
	new_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , new_user_byte_object )
	db_result := db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		users_bucket.Put( []byte( new_user.UUID ) , new_user_byte_object_encrypted )
		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		// something something holographic encryption would be nice here
		usernames_bucket.Put( []byte( new_user.Username ) , []byte( new_user.UUID ) )

		barcodes_bucket , _ := tx.CreateBucketIfNotExists( []byte( "barcodes" ) )
		for i := 0; i < len( new_user.Barcodes ); i++ {
			barcodes_bucket.Put( []byte( new_user.Barcodes[ i ] ) , []byte( new_user.UUID ) )
		}
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }

	// 4.) Update User Bleve Search Index
	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
	defer search_index.Close()
	new_search_item := types.SearchItem{
		UUID: new_user.UUID ,
		Name: strings.ReplaceAll( new_user.Username , "-" , " " ) ,
	}
	fmt.Println( reflect.TypeOf( search_index ) )
	search_index.Index( new_user.UUID , new_search_item )

	//return context.Redirect( fmt.Sprintf( "/admin/user/new/handoff/%s" , new_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/new" ,
		"result": fiber.Map{
			"uuid": new_user.UUID ,
		} ,
	})
}