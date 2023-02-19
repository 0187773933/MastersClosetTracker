package user

import (
	"fmt"
	// "reflect"
	"time"
	json "encoding/json"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt "github.com/boltdb/bolt"
	// uuid "github.com/satori/go.uuid"
	encrypt "github.com/0187773933/MastersClosetTracker/v1/encryption"
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
	MiddleName string `json:"last_name"`
	Address Address`json:"address"`
	DateOfBirth DateOfBirth `json:"date_of_birth"`
	Sex string `json:"sex"`
	Height string `json:"height"`
	EyeColor string `json:"eye_color"`
}

type User struct {
	Username string `json:"username"`
	UUID string `json:"uuid"`
	EmailAddress string `json:"email_address"`
	Identity Person `json:"identity"`
	AuthorizedAliases []Person `json:"authorized_aliases"`
	CreatedDate string `json:"created_date"`
	CreatedTime string `json:"created_time"`
	CheckIns []CheckIn `json:"check_ins"`
	FailedCheckIns []FailedCheckIn `json:"failed_check_ins"`
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

func UserNameExists( username string , db *bolt.DB ) ( result bool ) {
	result = false
	db.Update( func( tx *bolt.Tx ) error {
		bucket , tx_error := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		if tx_error != nil { fmt.Println( tx_error ); return nil }
		bucket_value := bucket.Get( []byte( username ) )
		if bucket_value == nil { return nil }
		result = true
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
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	return
}

func CheckInUser( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result bool ) {
	result = false
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
	new_check_in.Date = now.Format( "02JAN2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )
	if len( viewed_user.CheckIns ) < 1 {
		new_check_in.Type = "first"
		viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
		result = true
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		cool_off_duration := ( time.Duration( cool_off_days ) * 24 * time.Hour )
		last_check_in_date , _ := time.Parse( "02JAN2006" , last_check_in.Date )
		check_in_date_difference := last_check_in_date.Sub( last_check_in_date )
		if check_in_date_difference >= cool_off_duration {
			// "the user waited long enough before checking in again"
			new_check_in.Type = "new"
			viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
			result = true
		} else {
			days_remaining := ( cool_off_days - int( check_in_date_difference / ( 24 * time.Hour ) ) )
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait %d days\n" , days_remaining )
			var new_failed_check_in FailedCheckIn
			new_failed_check_in.Date = now.Format( "02JAN2006" )
			new_failed_check_in.Time = now.Format( "15:04:05.000" )
			new_failed_check_in.Type = "normal"
			new_failed_check_in.DaysRemaining = days_remaining
			viewed_user.FailedCheckIns = append( viewed_user.FailedCheckIns , new_failed_check_in )
			result = false
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