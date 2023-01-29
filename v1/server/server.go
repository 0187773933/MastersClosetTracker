package server

import (
	"os"
	"fmt"
	"time"
	"sort"
	// index_sort "github.com/mkmik/argsort"
	"strconv"
	"net"
	net_url "net/url"
	"strings"
	"context"
	"reflect"
	"sync"
	fiber "github.com/gofiber/fiber/v2"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	redis_manager "github.com/0187773933/RedisManagerUtils/manager"
	redis_lib "github.com/go-redis/redis/v8"
	// try "github.com/manucorporat/try"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

func GetRedisConnection( host string , port string , db int , password string ) ( redis_client redis_manager.Manager ) {
	redis_client.Connect( fmt.Sprintf( "%s:%s" , host , port ) , db , password )
	// TODO ===
	// https://stackoverflow.com/questions/49910285/redis-dial-tcp-redis-address-connect-connection-refused#49910392
	// redis := &redis.Pool{
	// 	MaxActive: idleConnections,
	// 	MaxIdle:   idleConnections,
	// 	Wait:      true,
	// 	Dial: func() (redis.Conn, error) {
	// 		c, err := redis.Dial("tcp", address, options...)
	// 		for retries := 0; err != nil && retries < 5; retries++ {
	// 			time.Sleep((50 << retries) * time.Millisecond)
	// 			c, err = redis.Dial("tcp", address, options...)
	// 		}
	// 		return c, err
	// 	},
	// }
	return
}

func RedisSetAdd( redis redis_manager.Manager , set_key string , value string ) {
	var ctx = context.Background()
	_ , set_add_error := redis.Redis.SAdd( ctx , set_key , value ).Result()
	if set_add_error != nil { fmt.Println( set_add_error ); }
	return
}

func RedisGetList( redis redis_manager.Manager , list_key string ) ( result []string ) {
	var ctx = context.Background()
	list_result , list_get_error := redis.Redis.LRange( ctx , list_key , 0 , -1 ).Result()
	if list_get_error != nil { fmt.Println( list_get_error ); }
	result = list_result
	return
}

func RedisGet( redis redis_manager.Manager , key string , wg *sync.WaitGroup , result_channel chan string ) {
	defer wg.Done()
	var ctx = context.Background()
	fmt.Println( key )
	string_result , string_result_error := redis.Redis.Get( ctx , key ).Result()
	if string_result_error != nil { fmt.Println( string_result_error ); }
	// result = string_result
	if len( string_result ) > 0 {
		fmt.Println( "1 - here" , reflect.TypeOf( string_result ) , len( string_result ) , string_result )
		result_channel <- string_result
	} else {
		result_channel <- "empty"
	}
	return
}

func RedisGetMulti( redis redis_manager.Manager , uuids []string ) ( result []string ) {
	var ctx = context.Background()
	pipe := redis.Redis.Pipeline()
	var values []*redis_lib.StringCmd
	for _ , uuid := range uuids {
		// this assumes all commands in multi are "GETS"
		values = append( values , pipe.Get( ctx , uuid ) )
		// but we don't care about values , we need :
			// SMEMBERS for ANALYTICS.{UUID}.IPS
			// LRANGE for ANALYTICS.{UUID}.RECORDS
			// GET for ANALYTICS.{UUID}.TOTAL
	}
	pipe.Exec( ctx )
	pipe.Close()
	for _ , i := range values {
		result = append( result , i.Val() )
	}
	return
}

type CustomMulti struct {
	Totals []int
	IPS [][]string
	Records [][]string
	Names []string
}
func RedisMulti( redis redis_manager.Manager , uuids []string ) ( results CustomMulti ) {
	var ctx = context.Background()
	pipe := redis.Redis.Pipeline()

	// so each multi requires a currated set of typed receivers that can later be "unwrapped"
	// aka .Result() can be called once the pipe.Exec() is finished
	// you have to know the type each pipe.Function() will return before hand , so you know what type of bins to create
	// TODO = Add Name if Set

	// could make just an Interface{}
	// then it doesn't matter the type
	// but still here , we have to have a custom Redis->Multi() order
	var ones []*redis_lib.StringCmd
	var twos []*redis_lib.StringSliceCmd
	var threes []*redis_lib.StringSliceCmd
	var names []*redis_lib.StringCmd
	for _ , uuid := range uuids {
		one := pipe.Get( ctx , fmt.Sprintf( "ANALYTICS.%s.TOTAL" , uuid ) )
		two := pipe.SMembers( ctx , fmt.Sprintf( "ANALYTICS.%s.IPS" , uuid ) )
		three := pipe.LRange( ctx , fmt.Sprintf( "ANALYTICS.%s.RECORDS" , uuid ) , 0 , -1 )
		name := pipe.Get( ctx , fmt.Sprintf( "ANALYTICS.%s.NAME" , uuid ) )
		ones = append( ones , one )
		twos = append( twos , two )
		threes = append( threes , three )
		names = append( names , name )
	}
	pipe.Exec( ctx )
	pipe.Close()

	// unwrap the results
	for _ , e := range ones {
		value , _ := e.Result()
		int_value , _ := strconv.Atoi( value )
		results.Totals = append( results.Totals , int_value )
	}
	// fmt.Println( "original , original totals" , results.Totals )
	for _ , e := range twos {
		value , _ := e.Result()
		results.IPS = append( results.IPS , value )
	}
	for _ , e := range threes {
		value , _ := e.Result()
		results.Records = append( results.Records , value )
	}
	for _ , e := range names {
		value , _ := e.Result()
		results.Names = append( results.Names , value )
	}
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

func (s Slice) Swap(i, j int) {
	s.IntSlice.Swap(i, j)
	s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]
}

func NewSlice(n []int) *Slice {
	s := &Slice{IntSlice: sort.IntSlice(n), indexes: make([]int, len(n))}
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

func New( config types.ConfigFile ) ( app *fiber.App ) {
	app = fiber.New()
	ip_addresses := GetLocalIPAddresses()
	fmt.Println( ip_addresses )
	// https://docs.gofiber.io/api/middleware/limiter
	app.Use( rate_limiter.New( rate_limiter.Config{
		Max: 2 ,
		Expiration: ( 4 * time.Second ) ,
		// Next: func( c *fiber.Ctx ) bool {
		// 	ip := c.IP()
		// 	fmt.Println( ip )
		// 	return ip == "127.0.0.1"
		// } ,
		LimiterMiddleware: rate_limiter.SlidingWindow{} ,
		KeyGenerator: func( c *fiber.Ctx ) string {
			return c.Get( "x-forwarded-for" )
		} ,
		LimitReached: func( c *fiber.Ctx ) error {
			ip := c.IP()
			fmt.Printf( "%s === limit reached\n" , ip )
			c.Set( "Content-Type" , "text/html" )
			return c.SendString( "<html><h1>why</h1></html>" )
		} ,
		// Storage: myCustomStorage{}
		// monkaS
		// https://github.com/gofiber/fiber/blob/master/middleware/limiter/config.go#L53
	}))
	// tracking := app.Group( "/t" )
	app.Get( "/t/:id" , func( fiber_context *fiber.Ctx ) ( error ) {
		id := fiber_context.Params( "id" )
		ip := fiber_context.IP()
		ips := fiber_context.IPs()
		// fmt.Println( ip )
		if ip == "127.0.0.1" {
			if len( ips ) > 0 {
				fmt.Println( ips )
				ip = ips[ 0 ]
			}
		} else if ip == "172.17.0.1" {
			if len( ips ) > 0 {
				fmt.Println( ips )
				ip = ips[ 0 ]
			}
		}
		global_total_key := fmt.Sprintf( "ANALYTICS.%s.TOTAL" , id )
		global_ips_key := fmt.Sprintf( "ANALYTICS.%s.IPS" , id )
		global_records_key := fmt.Sprintf( "ANALYTICS.%s.RECORDS" , id )
		// ip_total_key := fmt.Sprintf( "ANALYTICS.%s.%s.TOTAL" , id , ip )
		// ip_times_key := fmt.Sprintf( "ANALYTICS.%s.%s.TIMES" , id , ip )

		time_string := GetFormattedTimeString()
		record := fmt.Sprintf( "%s === %s" , time_string , ip )

		redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
		redis.Increment( global_total_key )
		// redis.Increment( ip_total_key )
		// redis.ListPushRight( ip_times_key , time_string )
		redis.ListPushRight( global_records_key , record )
		RedisSetAdd( redis , global_ips_key , ip )

		fmt.Printf( "%s === %s === new tracking\n" , time_string , ip )
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( "<html><h1>new tracking</h1></html>" )
	})

	app.Get( "/a/:id" , func( fiber_context *fiber.Ctx ) ( error ) {
		id := fiber_context.Params( "id" )
		ip := fiber_context.IP()
		global_total_key := fmt.Sprintf( "ANALYTICS.%s.TOTAL" , id )
		// global_ips_key := fmt.Sprintf( "ANALYTICS.%s.IPS" , id )
		global_records_key := fmt.Sprintf( "ANALYTICS.%s.RECORDS" , id )

		redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
		total_views := redis.Get( global_total_key )

		records := RedisGetList( redis , global_records_key )
		records_html_string := "<ol>\n"
		for _ , v := range records {
			records_html_string = records_html_string + fmt.Sprintf( "<li>%v</li>\n" , v )
		}
		records_html_string = records_html_string + "</ol>"
		fmt.Println( records )

		fmt.Printf( "Total Views === %v\n" , total_views )
		fmt.Printf( "%s === analytics\n" , ip )
		html_result_string := fmt.Sprintf( "<html>\n\t<h1>Total Views = %s</h1>\n%s\n</html>" , total_views , records_html_string )
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( html_result_string )
	})

	app.Post( "/a-list" , func( fiber_context *fiber.Ctx ) ( error ) {
		sent_api_key := fiber_context.Get( "key" )

		// Timing attacks possible ? , yes ?
		// just add random delay if wrong
		if sent_api_key != config.ServerAPIKey {
			fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
			return fiber_context.SendString( "?" )
		}

		// Parse Uploaded UUIDs
		var uploaded types.AListResponse
		fiber_context.BodyParser( &uploaded )

		// Get Values From Redis
		redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
		values := RedisGetMulti( redis , uploaded.UUIDS )
		result := map[string]string{}
		for i , e := range values {
			result[ uploaded.UUIDS[ i ] ] = e
		}

		// Send Results
		return fiber_context.JSON( result )
		// fmt.Println( fiber )
		// fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		// return fiber_context.SendString( "testing" )
	})

	app.Get( "/a-list/*" , func( fiber_context *fiber.Ctx ) ( error ) {

		// Parse UUIDs Separated by "/" in URL
		uuids := strings.Split( fiber_context.Params( "*" ) , "/" )

		// Get Values From Redis
		redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
		redis_results := RedisMulti( redis , uuids )

		// Build HTML String
		html_string := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Results</title>
	<style>
		table {
			border-collapse: collapse;
			font-family: Tahoma, Geneva, sans-serif;
		}
		table td {
			padding: 15px;
			text-align: center;
		}
		table th td {
			background-color: #54585d;
			color: #ffffff;
			font-weight: bold;
			font-size: 13px;
			border: 1px solid #54585d;
		}
		table tbody td {
			color: #636363;
			border: 1px solid #dddfe1;
		}
		table tbody tr {
			background-color: #f9fafb;
		}
		table tbody tr:nth-child(odd) {
			background-color: #ffffff;
		}
		th {
			padding-left: 10px;
			padding-right: 10px;
		}
	</style>
</head>
<body>
	<table id="result_table">
		<tr>
			<th>UUID</th>
			<th>Total</th>
			<th>Unique</th>
			<th>Name</th>
			<th>Latest Record</th>
		</tr>
`
		html_string_suffix := `	</table>
</body>
</html>`

		// Sort By Total Views
		// have to make a hard copy , so it doesn't get changed during sort
		var totals_copy []int
		for _ , e := range redis_results.Totals {
			totals_copy = append( totals_copy , e )
		}
		sorted_indexes_by_totals := NewSlice( totals_copy )
		sort.Sort( sorted_indexes_by_totals )
		reverse_sorted := ReverseInts( sorted_indexes_by_totals.indexes )
		// fmt.Println( "Lengths Match ? ===" , len( uuids ) , len( redis_results.Totals ) , len( redis_results.Names ) , len( redis_results.Records ) )

		// Populate HTML Table with Values
		for _ , sorted_index := range reverse_sorted {
			// html_string += fmt.Sprintf( "\t\t<li>%s === %s === %s</li>\n" , uuid , redis_results.Totals[ i ] , redis_results.Records[ i ][ ( len( redis_results.Records[ i ] ) - 1 ) ] )
			// fmt.Println( i , sorted_index , redis_results.Totals[ sorted_index ] , redis_results.Names[ sorted_index ] )
			uuid_link := fmt.Sprintf( "<a target=\"_blank\" href=\"%s/a/%s\">%s</a>" , config.ServerBaseUrl , uuids[ sorted_index ] , uuids[ sorted_index ] )
			// fmt.Println( uuid_link )

			unique_views := CountUniqueViewsInRecords( redis_results.Records[ sorted_index ] )
			// fmt.Println( "Unique Views ===" , unique_views )

			html_string += fmt.Sprintf( "\t\t<tr>\n" )
			html_string += fmt.Sprintf( "\t\t\t<td>%s</td>\n" , uuid_link )
			html_string += fmt.Sprintf( "\t\t\t<td>%d</td>\n" , redis_results.Totals[ sorted_index ] )
			html_string += fmt.Sprintf( "\t\t\t<td>%d</td>\n" , unique_views )
			html_string += fmt.Sprintf( "\t\t\t<td>%s</td>\n" , redis_results.Names[ sorted_index ] )
			if len( redis_results.Records[ sorted_index ] ) > 0 {
				html_string += fmt.Sprintf( "\t\t\t<td>%s</td>\n" , redis_results.Records[ sorted_index ][ ( len( redis_results.Records[ sorted_index ] ) - 1 ) ] )
			} else {
				html_string += fmt.Sprintf( "\t\t\t<td></td>\n" )
			}
			html_string += fmt.Sprintf( "\t\t</tr>\n" )
		}

		html_string += html_string_suffix

		// Send Results
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( html_string )
	})

	// app.Get( "/a-list" ) // returns HTML list
	// app.Get( "/set/name/:id" )
	app.Get( "/set/name/:id/:name" , func( fiber_context *fiber.Ctx ) ( error ) {
		id := fiber_context.Params( "id" )
		name , _ := net_url.PathUnescape( fiber_context.Params( "name" ) )
		redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
		name_key := fmt.Sprintf( "ANALYTICS.%s.NAME" , id )
		result := redis.Set( name_key , name )
		html_result_string := fmt.Sprintf( "<html>\n\t<h1>%s === %s === %s</h1>\n</html>" , id , name , result )
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( html_result_string )
	})

	return
}