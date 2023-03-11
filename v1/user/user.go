package user

import (
	"fmt"
	// "reflect"
	"strings"
	"time"
	json "encoding/json"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt "github.com/boltdb/bolt"
	// uuid "github.com/satori/go.uuid"
	encrypt "github.com/0187773933/MastersClosetTracker/v1/encryption"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

type CheckIn struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Type string `json:"type"`
}

type FailedCheckIn struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Type string `json:"type"`
	DaysRemaining int `json:"remaining_days"`
}

type BalanceItem struct {
	Available int `json:"available"`
	Limit int `json:"limit"`
	Used int `json:"used"`
}

type GeneralClothes struct {
	Total int `json:"total"`
	Available int `json:"available"`
	Tops BalanceItem `json:"tops"`
	Bottoms BalanceItem `json:"bottoms"`
	Dresses BalanceItem `json:"dresses"`
}

type Balance struct {
	General GeneralClothes `json:"general"`
	Shoes BalanceItem `json:"shoes"`
	Jackets BalanceItem `json:"jacketes"`
	Accessories BalanceItem `json:"accessories"`
}

type DateOfBirth struct {
	Month string `json:"month"`
	Day int `json:"day"`
	Year int `json:"year"`
}

type Address struct {
	StreetNumber string `json:"street_number"`
	StreetName string `json:"street_name"`
	AddressTwo string `json:"address_two"`
	City string `json:"city"`
	State string `json:"state"`
	ZipCode string `json:"zipcode"`
}

type Person struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Address Address`json:"address"`
	DateOfBirth DateOfBirth `json:"date_of_birth"`
	Age int `json:"age"`
	Sex string `json:"sex"`
	Height string `json:"height"`
	EyeColor string `json:"eye_color"`
}

type User struct {
	Username string `json:"username"`
	UUID string `json:"uuid"`
	Barcodes []string `json:"barcodes"`
	EmailAddress string `json:"email_address"`
	PhoneNumber string `json:"phone_number"`
	Identity Person `json:"identity"`
	AuthorizedAliases []Person `json:"authorized_aliases"`
	FamilySize int `json:"family_size"`
	FamilyMembers []Person `json:"family_members"`
	CreatedDate string `json:"created_date"`
	CreatedTime string `json:"created_time"`
	CheckIns []CheckIn `json:"check_ins"`
	FailedCheckIns []FailedCheckIn `json:"failed_check_ins"`
	Balance Balance `json:"balance"`
}

type GetUserResult struct {
	Username string `json:"username"`
	UUID string `json:"uuid"`
	LastCheckIn CheckIn `json:"last_check_in"`
}

// func New( username string , db *bolt.DB , encryption_key string ) ( new_user User ) {
// 	now := time.Now()
// 	new_user_uuid := uuid.NewV4().String()
// 	new_user.Username = username
// 	new_user.UUID = new_user_uuid
// 	new_user.CreatedDate = now.Format( "02JAN2006" )
// 	new_user.CreatedTime = now.Format( "15:04:05.000" )
// 	new_user_byte_object , _ := json.Marshal( new_user )
// 	new_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , new_user_byte_object )
// 	db_result := db.Update( func( tx *bolt.Tx ) error {
// 		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
// 		users_bucket.Put( []byte( new_user_uuid ) , new_user_byte_object_encrypted )
// 		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
// 		// something something holographic encryption would be nice here
// 		usernames_bucket.Put( []byte( username ) , []byte( "1" ) )
// 		return nil
// 	})
// 	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
// 	return
// }

func UserNameExists( username string , db *bolt.DB ) ( result bool , uuid string ) {
	result = false
	db.Update( func( tx *bolt.Tx ) error {
		bucket , tx_error := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		if tx_error != nil { fmt.Println( tx_error ); return nil }
		bucket_value := bucket.Get( []byte( username ) )
		if bucket_value == nil { return nil }
		result = true
		uuid = string( bucket_value )
		return nil
	})
	return
}

func GetByUUID( user_uuid string , db *bolt.DB , encryption_key string ) ( viewed_user User ) {
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		// fmt.Println( string( decrypted_bucket_value ) )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	return
}

func RefillBalance( user_uuid string , db *bolt.DB , encryption_key string , balance_config types.BalanceConfig ) ( new_balance Balance ) {
	var viewed_user User
	db.Update( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )

		viewed_user.Balance.General.Total = balance_config.General.Total
		viewed_user.Balance.General.Available = balance_config.General.Total
		viewed_user.Balance.General.Tops.Limit = balance_config.General.Tops
		viewed_user.Balance.General.Tops.Available = balance_config.General.Tops
		viewed_user.Balance.General.Bottoms.Limit = balance_config.General.Bottoms
		viewed_user.Balance.General.Bottoms.Available = balance_config.General.Bottoms
		viewed_user.Balance.General.Dresses.Limit = balance_config.General.Dresses
		viewed_user.Balance.General.Dresses.Available = balance_config.General.Dresses
		viewed_user.Balance.Shoes.Limit = balance_config.Shoes
		viewed_user.Balance.Shoes.Available = balance_config.Shoes
		viewed_user.Balance.Jackets.Limit = balance_config.Jackets
		viewed_user.Balance.Jackets.Available = balance_config.Jackets
		viewed_user.Balance.Accessories.Limit = balance_config.Accessories
		viewed_user.Balance.Accessories.Available = balance_config.Accessories

		viewed_user_byte_object , _ := json.Marshal( viewed_user )
		viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
		bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	new_balance = viewed_user.Balance
	return
}


// non-volitle? / passive checkin
// just sees if its possible , or if the user is currently timed-out
// 8e1bb28c-8868-448f-a07e-f0d270b4bbee === should be able to check-in
// d1e22369-6777-4eff-bf6a-0bf46a343a72
func CheckInTest( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result bool , time_remaining int , balance Balance ) {
	result = false
	time_remaining = -1
	// 1.) grab the user from the db
	var viewed_user User
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	if viewed_user.UUID == "" { fmt.Println( "user UUID doesn't exist" ); result = false; return }

	// 2.) Test if Check-In is possible
	var new_check_in CheckIn
	now := time.Now()
	now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )

	if len( viewed_user.CheckIns ) < 1 {
		result = true
		time_remaining = 0
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		last_check_in_date , _ := time.ParseInLocation( "02Jan2006" , last_check_in.Date , now_time_zone )
		fmt.Println( "Now ===" , now )
		fmt.Println( "Last ===" , last_check_in_date )

		cool_off_hours := ( 24 * cool_off_days )
		fmt.Println( "Cooloff Hours ===" , cool_off_hours )
		cool_off_duration , _ := time.ParseDuration( fmt.Sprintf( "%dh" , cool_off_hours ) )
		fmt.Println( "Cooloff Duration ===" , cool_off_duration )

		check_in_date_difference := now.Sub( last_check_in_date )
		fmt.Println( "Difference ===" , check_in_date_difference )

		// Negative Values Mean The User Has Waited Long Enough
		// Positive Values Mean the User Still has to wait
		time_remaining_duration := ( cool_off_duration - check_in_date_difference )
		fmt.Println( "Time Remaining ===" , time_remaining_duration )

		if time_remaining_duration < 0 {
			// "the user waited long enough before checking in again"
			result = true
			time_remaining = 0
		} else {

			days_remaining := int( time_remaining_duration.Hours() / 24 )
			time_remaining_string := time_remaining_duration.String()
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait : %d days , or %s" , days_remaining , time_remaining_string )

			result = false
			time_remaining = int( time_remaining_duration.Milliseconds() )
		}
	}
	balance = viewed_user.Balance
	return
}

func CheckInUser( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result bool , time_remaining int ) {
	result = false
	time_remaining = -1
	var viewed_user User
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	if viewed_user.UUID == "" { fmt.Println( "user UUID doesn't exist" ); result = false; return }
	var new_check_in CheckIn
	now := time.Now()
	now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )
	if len( viewed_user.CheckIns ) < 1 {
		new_check_in.Type = "first"
		new_check_in.Date = strings.ToUpper( new_check_in.Date )
		viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
		result = true
		time_remaining = 0
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		last_check_in_date , _ := time.ParseInLocation( "02Jan2006" , last_check_in.Date , now_time_zone )
		fmt.Println( "Now/New ===" , now )
		fmt.Println( "Last ===" , last_check_in_date )

		cool_off_hours := ( 24 * cool_off_days )
		fmt.Println( "Cooloff Hours ===" , cool_off_hours )
		cool_off_duration , _ := time.ParseDuration( fmt.Sprintf( "%dh" , cool_off_hours ) )
		fmt.Println( "Cooloff Duration ===" , cool_off_duration )

		check_in_date_difference := now.Sub( last_check_in_date )
		fmt.Println( "Difference ===" , check_in_date_difference )

		// Negative Values Mean The User Has Waited Long Enough
		// Positive Values Mean the User Still has to wait
		time_remaining_duration := ( cool_off_duration - check_in_date_difference )
		fmt.Println( "Time Remaining ===" , time_remaining_duration )

		if time_remaining_duration < 0 {
			// "the user waited long enough before checking in again"
			new_check_in.Type = "new"
			viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
			result = true
			time_remaining = 0
		} else {

			days_remaining := int( time_remaining_duration.Hours() / 24 )
			time_remaining_string := time_remaining_duration.String()
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait : %d days , or %s" , days_remaining , time_remaining_string )

			var new_failed_check_in FailedCheckIn
			new_failed_check_in.Date = strings.ToUpper( new_check_in.Date )
			new_failed_check_in.Time = new_check_in.Time
			new_failed_check_in.Type = "normal"
			new_failed_check_in.DaysRemaining = days_remaining
			viewed_user.FailedCheckIns = append( viewed_user.FailedCheckIns , new_failed_check_in )

			result = false
			time_remaining = int( time_remaining_duration.Milliseconds() )
		}
	}
	viewed_user_byte_object , _ := json.Marshal( viewed_user )
	viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
	db_result := db.Update( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
	return
}