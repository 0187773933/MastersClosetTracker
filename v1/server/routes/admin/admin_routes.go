package adminroutes

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"bytes"
	json "encoding/json"
	net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
	bcrypt "golang.org/x/crypto/bcrypt"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	utils "github.com/0187773933/MastersClosetTracker/v1/utils"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersClosetTracker/v1/user"
	encryption "github.com/0187773933/MastersClosetTracker/v1/encryption"
	bleve "github.com/blevesearch/bleve/v2"
	printer "github.com/0187773933/MastersClosetTracker/v1/printer"
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	admin_route_group := fiber_app.Group( "/admin" )
	admin_route_group.Get( "/logout" , Logout )
	admin_route_group.Get( "/login" , ServeLoginPage )
	admin_route_group.Post( "/login" , HandleLogin )
	admin_route_group.Get( "/" , AdminPage )

	admin_route_group.Get( "/users" , ViewUsersPage )
	admin_route_group.Get( "/checkins" , ViewCheckInsPage )

	// admin_route_group.Get( "/user/check/username" , CheckIfFirstNameLastNameAlreadyExists )
	admin_route_group.Get( "/user/new" , NewUserSignUpPage )
	admin_route_group.Post( "/user/new" , HandleNewUserJoin )
	admin_route_group.Post( "/user/edit" , HandleUserEdit )
	admin_route_group.Get( "/user/new/handoff/:uuid" , NewUserSignUpHandOffPage )

	admin_route_group.Get( "/user/checkin" , CheckInUserPage )
	admin_route_group.Post( "/user/checkin/:uuid" , UserCheckIn )
	admin_route_group.Get( "/user/checkin/test/:uuid" , UserCheckInTest )
	admin_route_group.Get( "/user/get/all" , GetAllUsers )
	admin_route_group.Get( "/user/get/all/checkins" , GetAllCheckIns )
	admin_route_group.Get( "/user/get/:uuid" , GetUser )
	admin_route_group.Get( "/user/get/barcode/:barcode" , GetUserViaBarcode )
	admin_route_group.Get( "/user/search/username/:username" , UserSearch )
	admin_route_group.Get( "/user/edit/:uuid" , EditUserPage )
	admin_route_group.Get( "/user/delete/:uuid" , DeleteUser )

	admin_route_group.Get( "/user/search/username/fuzzy/:username" , UserSearchFuzzy )

	admin_route_group.Get( "/print-test" , PrintTest )
}

// GET http://localhost:5950/admin/login
func ServeLoginPage( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func validate_login_credentials( context *fiber.Ctx ) ( result bool ) {
	result = false
	uploaded_username := context.FormValue( "username" )
	if uploaded_username == "" { fmt.Println( "username empty" ); return }
	if uploaded_username != GlobalConfig.AdminUsername { fmt.Println( "username not correct" ); return }
	uploaded_password := context.FormValue( "password" )
	if uploaded_password == "" { fmt.Println( "password empty" ); return }
	password_matches := bcrypt.CompareHashAndPassword( []byte( uploaded_password ) , []byte( GlobalConfig.AdminPassword ) )
	if password_matches != nil { fmt.Println( "bcrypted password doesn't match" ); return }
	result = true
	return
}

func serve_failed_attempt( context *fiber.Ctx ) ( error ) {
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>no</h1>" )
}

func Logout( context *fiber.Ctx ) ( error ) {
	context.Cookie( &fiber.Cookie{
		Name: "the-masters-closet-admin" ,
		Value: "" ,
		Expires: time.Now().Add( -time.Hour ) , // set the expiration to the past
		HTTPOnly: true ,
		Secure: true ,
	})
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>Logged Out</h1>" )
}

// POST http://localhost:5950/admin/login
func HandleLogin( context *fiber.Ctx ) ( error ) {
	valid_login := validate_login_credentials( context )
	if valid_login == false { return serve_failed_attempt( context ) }
	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-admin" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , GlobalConfig.ServerCookieAdminSecretMessage ) ,
			Secure: true ,
			Path: "/" ,
			// Domain: "blah.ngrok.io" , // probably should set this for webkit
			HTTPOnly: true ,
			SameSite: "Lax" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	return context.Redirect( "/admin" )
}

func validate_admin_cookie( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie == "" { fmt.Println( "admin cookie was blank" ); return }
	admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
	if admin_cookie_value != GlobalConfig.ServerCookieAdminSecretMessage { fmt.Println( "admin cookie secret message was not equal" ); return }
	result = true
	return
}

func AdminPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin.html" )
}

func NewUserSignUpPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_user_new.html" )
}

func CheckInUserPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_user_checkin.html" )
}

func ViewUsersPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_view_users.html" )
}

func ViewCheckInsPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_view_total_checkins.html" )
}

func EditUserPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_user_edit.html" )
}


// weak attempt at sanitizing form input to build a "username"
func SanitizeUsername( first_name string , last_name string ) ( username string ) {
	if first_name == "" { first_name = "Not Provided" }
	if last_name == "" { last_name = "Not Provided" }
	sanitized_first_name := utils.SanitizeInputString( first_name )
	sanitized_last_name := utils.SanitizeInputString( last_name )
	username = fmt.Sprintf( "%s-%s" , sanitized_first_name , sanitized_last_name )
	return
}
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

	if new_user.Identity.MiddleName != "" {
		new_user.Username = fmt.Sprintf( "%s-%s-%s" , new_user.Identity.FirstName , new_user.Identity.MiddleName , new_user.Identity.LastName )
	} else {
		new_user.Username = fmt.Sprintf( "%s-%s" , new_user.Identity.FirstName , new_user.Identity.LastName )
	}
	new_user.UUID = uuid.NewV4().String()

	now := time.Now()
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )

	return
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
	search_index.Index( new_user.UUID , new_search_item )

	//return context.Redirect( fmt.Sprintf( "/admin/user/new/handoff/%s" , new_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/new" ,
		"result": fiber.Map{
			"uuid": new_user.UUID ,
		} ,
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



func NewUserSignUpHandOffPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_user_new_handoff.html" )
}

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
	return context.JSON( fiber.Map{
		"route": "/admin/user/delete/:uuid" ,
		"result": "deleted" ,
	})
}

func parse_form_value_as_int( context *fiber.Ctx , form_key string ) ( result int ) {
	result = -1
	uploaded := context.FormValue( form_key )
	sanitized := utils.SanitizeInputString( uploaded )
	parsed_int , _ := strconv.Atoi( sanitized )
	result = parsed_int
	return
}

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


func UserSearch( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	username := context.Params( "username" )
	escaped_username , _ := net_url.QueryUnescape( username )
	formated_username := strings.Replace( escaped_username , " " , "-" , -1 )
	formated_username_bytes := []byte( formated_username )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	found_uuid := "not found"
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "usernames" ) )
		bucket.ForEach( func( k , v []byte ) error {
			if bytes.Equal( k , formated_username_bytes ) == false { return nil }
			found_uuid = string( v )
			return nil
		})
		return nil
	})
	fmt.Printf( "Searched : %s || Result === %s\n" , formated_username , found_uuid )
	return context.JSON( fiber.Map{
		"route": "/admin/user/search/username/:username" ,
		"result": found_uuid ,
	})
}

func UserSearchFuzzy( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	username := context.Params( "username" )
	escaped_username , _ := net_url.QueryUnescape( username )

	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
	defer search_index.Close()

	// Only works for single words
	// query := bleve.NewFuzzyQuery( escaped_username )
	// query.Fuzziness = 2
	// // query.Fuzziness = 1
	// search_request := bleve.NewSearchRequest( query )
	// search_results , _ := search_index.Search( search_request )

	words := strings.Fields( escaped_username )
	boolean_query := bleve.NewBooleanQuery()
	for _, word := range words {
		q := bleve.NewFuzzyQuery( word )
		q.Fuzziness = 2
		boolean_query.AddMust( q )
	}
	search_request := bleve.NewSearchRequest( boolean_query )
	search_results , _ := search_index.Search( search_request )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	var search_results_users []user.User
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		for _ , hit := range search_results.Hits {
			x_user := bucket.Get( []byte( hit.ID ) )
			if x_user == nil { continue } // so this is needed because we didn't delete search indexes when deleting a user
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , x_user )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			search_results_users = append( search_results_users , viewed_user )
		}
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/search/username/fuzzy/:username" ,
		"result": search_results_users ,
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
		"route": "/admin/user/get/all" ,
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

func PrintTest( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	printer.PrintTicket( GlobalConfig.Printer , printer.PrintJob{
		FamilySize: 5 ,
		TotalClothingItems: 23 ,
		Shoes: 1 ,
		Accessories: 2 ,
		Seasonal: 1 ,
		FamilyName: "Cerbus" ,
	})
	return context.JSON( fiber.Map{
		"route": "/admin/print-test" ,
		"result": "success" ,
	})
}