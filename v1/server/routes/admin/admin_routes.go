package adminroutes

import (
	"fmt"
	"time"
	"strconv"
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
)

var GlobalConfig *types.ConfigFile

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config
	admin_route_group := fiber_app.Group( "/admin" )
	admin_route_group.Get( "/logout" , Logout )
	admin_route_group.Get( "/login" , ServeLoginPage )
	admin_route_group.Post( "/login" , HandleLogin )
	admin_route_group.Get( "/" , AdminPage )

	admin_route_group.Get( "/user/check/username" , CheckIfFirstNameLastNameAlreadyExists )
	admin_route_group.Get( "/user/new" , NewUserSignUpPage )
	admin_route_group.Post( "/user/new" , HandleNewUserJoin )
	admin_route_group.Get( "/user/new/handoff/:uuid" , NewUserSignUpHandOffPage )

	admin_route_group.Get( "/user/checkin" , CheckInUserPage )
	admin_route_group.Get( "/user/checkin/:uuid" , UserCheckIn )
	admin_route_group.Get( "/user/get/:uuid" , GetUser )
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
	uploaded_user_street_number := context.FormValue( "user_street_number" )
	uploaded_user_street_name := context.FormValue( "user_street_name" )
	uploaded_user_address_two := context.FormValue( "user_address_two" )
	uploaded_user_city := context.FormValue( "user_city" )
	uploaded_user_state := context.FormValue( "user_state" )
	uploaded_user_zip_code := context.FormValue( "user_zip_code" )
	uploaded_user_birth_day := context.FormValue( "user_birth_day" )
	uploaded_user_birth_month := context.FormValue( "user_birth_month" )
	uploaded_user_birth_year := context.FormValue( "user_birth_year" )

	new_user.EmailAddress = utils.SanitizeInputString( uploaded_user_email )
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

	new_user.Username = fmt.Sprintf( "%s-%s" , new_user.Identity.FirstName , new_user.Identity.LastName )
	new_user.UUID = uuid.NewV4().String()

	now := time.Now()
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )

	return
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/admin/new/:username
func HandleNewUserJoin( context *fiber.Ctx ) ( error ) {

	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// sanitize input
	new_user := ProcessNewUserForm( context )

	// Store User in DB
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	new_user_byte_object , _ := json.Marshal( new_user )
	new_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , new_user_byte_object )
	db_result := db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		users_bucket.Put( []byte( new_user.UUID ) , new_user_byte_object_encrypted )
		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		// something something holographic encryption would be nice here
		usernames_bucket.Put( []byte( new_user.Username ) , []byte( "1" ) )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }

	//return context.Redirect( fmt.Sprintf( "/admin/user/new/handoff/%s" , new_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/new" ,
		"result": fiber.Map{
			"uuid": new_user.UUID ,
		} ,
	})
}

func NewUserSignUpHandOffPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	return context.SendFile( "./v1/server/html/admin_user_new_handoff.html" )
}

// TODO : this can fail though , what if somebody has the exact same first and last name ????
// just make them have an email ???
func CheckIfFirstNameLastNameAlreadyExists( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }

	// build username
	uploaded_first_name := context.Query( "fn" )
	uploaded_last_name := context.Query( "ln" )
	first_name , _ := net_url.QueryUnescape( uploaded_first_name )
	last_name , _ := net_url.QueryUnescape( uploaded_last_name )
	username := SanitizeUsername( first_name , last_name )
	// fmt.Println( username )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	username_exists := user.UserNameExists( username , db )
	return context.JSON( fiber.Map{
		"route": "/admin/user/check/username" ,
		"result": username_exists ,
	})
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

func UserCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_cookie( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	check_in_result := user.CheckInUser( user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/:uuid" ,
		"result": check_in_result ,
	})
}