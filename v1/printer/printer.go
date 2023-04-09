package printer

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	// "bufio"
	"runtime"
	"io/ioutil"
	"os"
	"os/exec"
	// "reflect"
	"path/filepath"
	"image/png"
	// "github.com/boombuler/barcode"
	// "github.com/boombuler/barcode/code128"
	"github.com/ppsleep/barcode"
	"github.com/ppsleep/barcode/code128"
	types "github.com/0187773933/MastersClosetTracker/v1/types"
)

func write_barcode_image( image_path string , barcode_number string ) {
	code , _ := code128.A( barcode_number )
	r := barcode.Encode( code , 2 , 50 )
	file , _ := os.Create( image_path )
	defer file.Close()
	png.Encode( file , r )
}

func clear_printer_que_mac_osx( printer_name string ) {
	cmd := exec.Command( "cancel" , "-a" , printer_name )
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println( "Error Clearing Printer Que:", err )
	}
}

// lpstat -v
// lp -d "_4BARCODE_4B_2054N" -o PrintSpeed=2 -o fit-to-page "/Users/morpheous/WORKSPACE/GO/TMP2/BarcodePrinterTest/output.pdf"
func print_pdf_mac_osx( printer_name string , pdf_file_path string ) {
	args := []string{ "lp" , "-d" , printer_name , "-o" , "PrintSpeed=2" , "-o" , "fit-to-page" , pdf_file_path }
	cmd := exec.Command( args[ 0 ] , args[ 1 : ]... )
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println( "Error Printing PDF: " , err )
	}
}

func clear_printer_que_windows( printer_name string ) {
	cmd := exec.Command( "cancel" , "-a" , printer_name )
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println( "Error Clearing Printer Que:", err )
	}
}

// Get Printer Names
	// wmic printer get name
// Get Printer Options
	// wmic printer where "name='Brother MFC-J4535DW'" get /value
// PDFtoPrinter.exe /printer "Brother MFC-J4535DW" /param "PrintSpeed=2" "myfile.pdf"

// print /D:"'Brother MFC-J4535DW'" "test.pdf"
// print /D:"printer_name" /o"option1=value1" /o"option2=value2" "file_name"
// print.exe only prints fucking plaintext ????
func print_pdf_windows( printer_name string , pdf_file_path string ) {
	sumatra_file_path  , _ := filepath.Abs( "SumatraPDF.exe" )
	args := []string{ sumatra_file_path , "-print-to" , printer_name , pdf_file_path }
	cmd := exec.Command( args[ 0 ] , args[ 1 : ]... )
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println( "Error Printing PDF: " , err )
	}
}

type PrintJob struct {
	FamilySize int `json:"family_size"`
	TotalClothingItems int `json:"total_clothing_items"`
	Shoes int `json:"shoes"`
	ShoesLimit int `json:"shoes_limit"`
	Accessories int `json:"accessories"`
	AccessoriesLimit int `json:"accessories_limit"`
	Seasonal int `json:"seasonal"`
	SeasonalLimit int `json:"seasonal_limit"`
	FamilyName string `json:"family_name"`
	BarcodeNumber string `json:"barcode_number"`
	Spanish bool `json:"spanish"`
}

func add_centered_text( pdf *gofpdf.Fpdf , text string , font_name string , font_size float64 , at_page_height float64 ) {
	page_width , _ := pdf.GetPageSize()
	// margin_left , margin_top , margin_right , margin_bottom := pdf.GetMargins()
	// fmt.Printf( "Page Width === %v || Page Height === %v\n" , page_width , page_height )
	// fmt.Printf( "Margin Left === %v || Margin Right === %v\n" , margin_left , margin_right )
	// fmt.Printf( "Margin Top === %v || Margin Bottom === %v\n" , margin_top , margin_bottom )
	pdf.SetFont( font_name , "" , font_size )
	string_width := pdf.GetStringWidth( text )
	// fmt.Printf( "String Width === %v" , string_width )
	page_center_x := ( page_width / 2 )
	starting_x := ( page_center_x - ( string_width / 2 ) )
	pdf.Text( starting_x , at_page_height , text )
}


// https://pkg.go.dev/github.com/jung-kurt/gofpdf#example-Fpdf.PageSize
// https://pkg.go.dev/github.com/jung-kurt/gofpdf#Fpdf.ImageOptions
// https://pkg.go.dev/github.com/jung-kurt/gofpdf#Fpdf.Text
// https://pkg.go.dev/github.com/jung-kurt/gofpdf#example-Fpdf.TransformBegin
func PrintTicket( config types.PrinterConfig , job PrintJob ) {
	pdf := gofpdf.NewCustom( &gofpdf.InitType{
		UnitStr: "in" ,
		Size: gofpdf.SizeType{ Wd: config.PageWidth , Ht: config.PageHeight } ,
	})
	pdf.SetMargins( 0.5 , 1 , 0.5 )
	pdf.AddPage()

	// 1.) Add Logo
	pdf.ImageOptions(
		config.LogoFilePath ,
		0.5 , 0.25 ,
		3 , 0 ,
		false ,
		gofpdf.ImageOptions{
			ImageType: "PNG" ,
			ReadDpi: true ,
			AllowNegativePosition: false ,
		} ,
		0 , "" ,
	)

	// 2.) Add Middle Text
	if job.Spanish == true {
		// add_centered_text( pdf , fmt.Sprintf( "Tamaño Familiar ( %d )" , job.FamilySize ) , config.FontName , 20 , 2.0 )
		add_centered_text( pdf , fmt.Sprintf( "Tamano Familiar ( %d )" , job.FamilySize ) , config.FontName , 20 , 2.0 )
		add_centered_text( pdf , fmt.Sprintf( "Total Vestir Para La Familia ( %d )" , job.TotalClothingItems ) , config.FontName , 16 , 2.5 )
		if job.Shoes > 1 {
			add_centered_text( pdf , fmt.Sprintf( "%d Pares De Zapatos  , %d Por Persona" , job.Shoes , job.ShoesLimit ) , config.FontName , 14 , 3.0 )
		} else {
			add_centered_text( pdf , fmt.Sprintf( "%d Par de Zapatos , %d Por Persona" , job.Shoes , job.ShoesLimit ) , config.FontName , 14 , 3.0 )
		}
		if job.Accessories > 1 {
			add_centered_text( pdf , fmt.Sprintf( "%d Accesorios , %d Por Persona" , job.Accessories , job.AccessoriesLimit ) , config.FontName , 14 , 3.3 )
		} else {
			add_centered_text( pdf , fmt.Sprintf( "%d Accesorio , %d Por Persona" , job.Accessories , job.AccessoriesLimit ) , config.FontName , 14 , 3.3 )
		}
		if job.Seasonal > 1 {
			// add_centered_text( pdf , fmt.Sprintf( "%d Artículos De Temporada , %d Por Persona" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
			add_centered_text( pdf , fmt.Sprintf( "%d Articulos De Temporada , %d Por Persona" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
		} else {
			// add_centered_text( pdf , fmt.Sprintf( "%d Artículo De Temporada , %d Por Persona" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
			add_centered_text( pdf , fmt.Sprintf( "%d Articulo De Temporada , %d Por Persona" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
		}
	} else {
		add_centered_text( pdf , fmt.Sprintf( "Family Size ( %d )" , job.FamilySize ) , config.FontName , 20 , 2.0 )
		add_centered_text( pdf , fmt.Sprintf( "Total Clothing Items for Family ( %d )" , job.TotalClothingItems ) , config.FontName , 16 , 2.5 )
		if job.Shoes > 1 {
			add_centered_text( pdf , fmt.Sprintf( "%d Pairs of Shoes , %d Per Person" , job.Shoes , job.ShoesLimit ) , config.FontName , 14 , 3.0 )
		} else {
			add_centered_text( pdf , fmt.Sprintf( "%d Pair of Shoes , %d Per Person" , job.Shoes , job.ShoesLimit ) , config.FontName , 14 , 3.0 )
		}
		if job.Accessories > 1 {
			add_centered_text( pdf , fmt.Sprintf( "%d Accessories , %d Per Person" , job.Accessories , job.AccessoriesLimit ) , config.FontName , 14 , 3.3 )
		} else {
			add_centered_text( pdf , fmt.Sprintf( "%d Accessory , %d Per Person" , job.Accessories , job.AccessoriesLimit ) , config.FontName , 14 , 3.3 )
		}
		if job.Seasonal > 1 {
			add_centered_text( pdf , fmt.Sprintf( "%d Seasonal items , %d Per Person" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
		} else {
			add_centered_text( pdf , fmt.Sprintf( "%d Seasonal item , %d Per Person" , job.Seasonal , job.SeasonalLimit ) , config.FontName , 14 , 3.6 )
		}
	}


	add_centered_text( pdf , job.FamilyName , config.FontName , 16 , 4.4 )
	// 3.) Gen and Add Barcode
	barcode_temp_file , _ := ioutil.TempFile( "" , "barcode-*.png" )
	defer barcode_temp_file.Close()
	barcode_temp_file_path := barcode_temp_file.Name()
	defer func() {
		os.Remove( barcode_temp_file_path )
	}()
	write_barcode_image( barcode_temp_file_path , "123456" )
	pdf.ImageOptions(
		barcode_temp_file_path ,
		1.23 , 4.5 ,
		1.5 , 0 ,
		false ,
		gofpdf.ImageOptions{
			ImageType: "PNG" ,
			ReadDpi: true ,
			AllowNegativePosition: false ,
		} ,
		0 , "" ,
	)

	pdf_temp_file , _ := ioutil.TempFile( "" , "ticket-*.pdf" )
	defer pdf_temp_file.Close()
	pdf_temp_file_path := pdf_temp_file.Name()
	defer func() {
		os.Remove( pdf_temp_file_path )
	}()
	pdf.OutputFileAndClose( pdf_temp_file_path )
	if runtime.GOOS == "windows" {
		// clear_printer_que_windows( config.PrinterName )
		print_pdf_windows( config.PrinterName , pdf_temp_file_path )
	} else if runtime.GOOS == "darwin" {
		clear_printer_que_mac_osx( config.PrinterName )
		print_pdf_mac_osx( config.PrinterName , pdf_temp_file_path )
	}
}