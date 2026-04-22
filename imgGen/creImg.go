package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

    "github.com/gogpu/gg"
    "github.com/gogpu/gg/text"

    util "github.com/prr123/utility/utilLib"
)

type imgDat struct {
	W int
	H int
	Txt string
}

func main() {

    numarg := len(os.Args)
    flags:=[]string{"dbg", "w", "h", "txt"}

    useStr := "/w=width /h=height /txt=text [/dbg]"
    helpStr := "azul create image program"

    if numarg > len(flags) +1 {
        fmt.Println("too many arguments in cl!")
        fmt.Println("usage: %s %s\n", os.Args[0], useStr)
        os.Exit(-1)
    }

    if numarg == 1 || (numarg > 1 && os.Args[1] == "help") {
        fmt.Printf("help: %s\n", helpStr)
        fmt.Printf("usage is: %s %s\n", os.Args[0], useStr)
        os.Exit(1)
    }

    flagMap, err := util.ParseFlags(os.Args, flags)
    if err != nil {log.Fatalf("util.ParseFlags: %v\n", err)}

    dbg:= false
    _, ok := flagMap["dbg"]
    if ok {dbg = true}

	ImgDat := imgDat{W:-1,H:-1,Txt:"none"}

	wval,ok := flagMap["w"]
	if !ok {
		log.Fatalf("error -- no width flag!\n")
	} else {
		if wval.(string) == "none" {log.Fatalf("error -- no yaml file name provided!\n")}
		ImgDat.W, err = strconv.Atoi(wval.(string))
		if err != nil {log.Fatalf("error -- width is no int: %v\n", err)}
	}

	hval,ok := flagMap["h"]
	if !ok {
		log.Fatalf("error -- no height flag!\n")
	} else {
		if hval.(string) == "none" {log.Fatalf("error -- no yaml file name provided!\n")}
		ImgDat.H, err = strconv.Atoi(hval.(string))
		if err != nil {log.Fatalf("error -- height is no int: %v\n", err)}
	}

	txtval,ok := flagMap["txt"]
	if !ok {
		log.Fatalf("error -- no txt flag!\n")
	} else {
		if txtval.(string) == "none" {log.Fatalf("error -- no yaml file name provided!\n")}
		ImgDat.Txt = txtval.(string)
	}

	if dbg {
		fmt.Printf("txt: %s\n", ImgDat.Txt)
		fmt.Printf("width: %d height: %d\n", ImgDat.W, ImgDat.H)
	}

    // Create drawing context
  	
//	dc := gg.NewContext(512, 512)
	dc := gg.NewContext(ImgDat.W, ImgDat.H)
    defer dc.Close()

    dc.ClearWithColor(gg.White)

    // Draw shapes
    dc.SetHexColor("#3498db")
//    dc.DrawCircle(256, 256, 100)
//    dc.Fill()

    // Render text
	ttfFil :=  "/usr/share/fonts/truetype/lato/Lato-Regular.ttf"
//	source, err := text.NewFontSourceFromFile("arial.ttf")
	source, err := text.NewFontSourceFromFile(ttfFil)
	if err !=nil {log.Fatalf("error -- cannot find logfile: %v\n", err)}
    defer source.Close()

    dc.SetFont(source.Face(32))
    dc.SetColor(gg.Black)
    dc.DrawStringAnchored(ImgDat.Txt, float64(ImgDat.W/2), float64(ImgDat.H/2), 0.5, 0.5)

    dc.SavePNG("output.png")

	fmt.Println("*** success ***")
}
