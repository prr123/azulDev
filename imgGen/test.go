package main

import (
	"fmt"
	"log"

    "github.com/gogpu/gg"
    "github.com/gogpu/gg/text"
)

func main() {
    // Create drawing context
    dc := gg.NewContext(512, 512)
    defer dc.Close()

    dc.ClearWithColor(gg.White)

    // Draw shapes
    dc.SetHexColor("#3498db")
    dc.DrawCircle(256, 256, 100)
    dc.Fill()

    // Render text
	ttfFil :=  "/usr/share/fonts/truetype/lato/Lato-Regular.ttf"
//	source, err := text.NewFontSourceFromFile("arial.ttf")
	source, err := text.NewFontSourceFromFile(ttfFil)
	if err !=nil {log.Fatalf("error -- cannot find logfile: %v\n", err)}
    defer source.Close()

    dc.SetFont(source.Face(32))
    dc.SetColor(gg.Black)
    dc.DrawString("Hello, GoGPU!", 180, 260)

    dc.SavePNG("output.png")

	fmt.Println("*** success ***")
}
