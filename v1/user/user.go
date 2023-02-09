package user

import (
	"fmt"
	// "reflect"
	"time"
	json "encoding/json"
	// bolt "github.com/0187773933/MastersClosetTracker/v1/bolt"
	bolt "github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"
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
	RemainingTime string `json:"remaining_time"`
}

type User struct {
	UUID string `json:"uuid"`
	CreatedDate string `json:"created_date"`
	CreatedTime string `json:"created_time"`
	CheckIns []CheckIn `json:"check_ins"`
	FailedCheckIns []FailedCheckIn `json:"failed_check_ins"`
}

func New( username string , db *bolt.DB , encryption_key string ) ( new_user User ) {
	now := time.Now()
	new_user_uuid := uuid.NewV4().String()
	new_user.UUID = new_user_uuid
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )
	new_user_byte_object , _ := json.Marshal( new_user )
	new_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , new_user_byte_object )
	db_result := db.Update( func( tx *bolt.Tx ) error {
		bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		bucket.Put( []byte( new_user_uuid ) , new_user_byte_object_encrypted )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
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

func CheckInUser( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result string ) {
	var viewed_user User
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	if viewed_user.UUID == "" { result = "user UUID doesn't exist"; return }
	if len( viewed_user.CheckIns ) < 1 {
		var new_check_in CheckIn
		now := time.Now()
		new_check_in.Date = now.Format( "02JAN2006" )
		new_check_in.Time = now.Format( "15:04:05.000" )
		new_check_in.Type = "first"
		viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
		viewed_user_byte_object , _ := json.Marshal( viewed_user )
		viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
		db_result := db.Update( func( tx *bolt.Tx ) error {
			bucket := tx.Bucket( []byte( "users" ) )
			bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
			return nil
		})
		if db_result != nil { panic( "couldn't write to bolt db ??" ) }
		result = "success"
		return
	} else {
		fmt.Println( "user has checked in before , need to compare last check-in date to now" )
		now := time.Now()
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		// only comparing the dates , not the times
		// fmt.Println( viewed_user.CheckIns )
		cool_off_duration := ( time.Duration( cool_off_days ) * 24 * time.Hour )
		last_check_in_date , _ := time.Parse( "02JAN2006" , last_check_in.Date )
		fmt.Println( "last checkin date ===" , last_check_in_date )
		check_in_date_difference := last_check_in_date.Sub( last_check_in_date )
		fmt.Println( check_in_date_difference )
		if check_in_date_difference >= cool_off_duration {
			fmt.Println( "the user waited long enough before checking in again" )
			var new_check_in CheckIn
			new_check_in.Date = now.Format( "02JAN2006" )
			new_check_in.Time = now.Format( "15:04:05.000" )
			new_check_in.Type = "new"
			fmt.Println( len( viewed_user.CheckIns ) )
			viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
			fmt.Println( len( viewed_user.CheckIns ) )
			fmt.Println( viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ] )
			viewed_user_byte_object , _ := json.Marshal( viewed_user )
			viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
			db_result := db.Update( func( tx *bolt.Tx ) error {
				bucket := tx.Bucket( []byte( "users" ) )
				bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
				return nil
			})
			if db_result != nil { panic( "couldn't write to bolt db ??" ) }
			result = "success"
			return
		} else {
			days_remaining := ( cool_off_days - int( check_in_date_difference / ( 24 * time.Hour ) ) )
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait %d days\n" , days_remaining )
		}
	}
	return
}