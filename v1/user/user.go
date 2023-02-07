package user

import (
	// "fmt"
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
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	return
}