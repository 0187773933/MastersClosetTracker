package utils

import (
	"os"
	"bufio"
	"time"
	"net"
	"fmt"
	// index_sort "github.com/mkmik/argsort"
	"sort"
	"strings"
	"unicode"
	"io/ioutil"
	"encoding/json"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

func ParseConfig( file_path string ) ( result types.ConfigFile ) {
	file_data , _ := ioutil.ReadFile( file_path )
	err := json.Unmarshal( file_data , &result )
	if err != nil { fmt.Println( err ) }
	return
}

// https://stackoverflow.com/a/28862477
func GetLocalIPAddresses() ( ip_addresses []string ) {
	host , _ := os.Hostname()
	addrs , _ := net.LookupIP( host )
	for _ , addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			// fmt.Println( "IPv4: " , ipv4 )
			ip_addresses = append( ip_addresses , ipv4.String() )
		}
	}
	return
}

func GetFormattedTimeString() ( result string ) {
	location , _ := time.LoadLocation( "America/New_York" )
	time_object := time.Now().In( location )
	// https://stackoverflow.com/a/51915792
	// month_name := strings.ToUpper( time_object.Format( "Feb" ) ) // monkaHmm
	month_name := strings.ToUpper( time_object.Format( "Jan" ) )
	milliseconds := time_object.Format( ".000" )
	date_part := fmt.Sprintf( "%02d%s%d" , time_object.Day() , month_name , time_object.Year() )
	time_part := fmt.Sprintf( "%02d:%02d:%02d%s" , time_object.Hour() , time_object.Minute() , time_object.Second() , milliseconds )
	result = fmt.Sprintf( "%s === %s" , date_part , time_part )
	return
}

type Slice struct {
	sort.IntSlice
	indexes []int
}
func ( s Slice ) Swap( i , j int ) {
	s.IntSlice.Swap(i, j)
	s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]
}

func NewSlice( n []int ) *Slice {
	s := &Slice{
		IntSlice: sort.IntSlice(n) ,
		indexes: make( []int , len( n ) ) ,
	}
	for i := range s.indexes {
		s.indexes[i] = i
	}
	return s
}

func ReverseInts( input []int ) []int {
	if len(input) == 0 {
		return input
	}
	return append(ReverseInts(input[1:]), input[0])
}

func CountUniqueViewsInRecords( records []string ) ( result int ) {
	ip_map := map[string]int{}
	for _ , record := range records {
		ip_address := strings.Split( record , " === " )[ 2 ]
		if _ , exists := ip_map[ ip_address ]; exists == false {
			ip_map[ ip_address ] = 1
		}
	}
	result = len( ip_map )
	return
}

func RemoveNonASCII( input string ) ( result string ) {
	for _ , i := range input {
		if i > unicode.MaxASCII { continue }
		result += string( i )
	}
	return
}

const NameSizeLimit = 20
func SanitizeInputName( input string ) ( result string ) {
	trimmed := strings.TrimSpace( input )
    if len( trimmed ) > NameSizeLimit { trimmed = strings.TrimSpace( trimmed[ 0 : NameSizeLimit ] ) }
	result = RemoveNonASCII( trimmed )
	return
}

func WriteAdminUserHandOffHTML( server_base_url string ) {
	file , _ := os.OpenFile( "./v1/server/html/admin_user_new_handoff.html" , os.O_RDWR , 0 )
	defer file.Close()
	reader := bufio.NewReader( file )
	line_number := 1
	var lines []string
	for {
		line , err := reader.ReadString( '\n' )
		if err != nil { break }
		if line_number == 48 { line = "\t\t\tconst QR_CODE_BASE_URL = \"" + server_base_url + "\";\n" }
		lines = append( lines , line )
		line_number = line_number + 1
	}
	file.Seek( 0 , 0 )
	file.Truncate( 0 )
	for _ , line := range lines {
		file.WriteString( line )
	}
}