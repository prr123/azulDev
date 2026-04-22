// buildAzul: program that builds a website
//
// v2: build basic idx file
//  -- mv files to domain
//	-- add fav file

package main

import (
    "fmt"
    "os"
    "log"
    "strings"
	"bytes"

	"github.com/goccy/go-yaml"
    util "github.com/prr123/utility/utilLib"
)

type azulStruct struct {
	Title string `yaml:"title,omitempty"`
	Meta map[string]string `yaml:"meta"`
	Fav string `yaml:"favicon,omitempty"`
	Font string `yaml:"font,omitempty"`
	Nav map[string]string `yaml:"nav"`
	Blogs []string `yaml:"blogs,omitempty"`
	Foot azulFootStruct `yaml:"Footer,omitempty"`
	land string
}

type azulFootStruct struct {
	FootCol []footCol
}

type footCol struct {
	Name []string `yaml:"name,omitempty"`
}

func main() {

	numarg := len(os.Args)
    flags:=[]string{"dbg", "yaml", "domain"}

    useStr := "/yaml=<yamlfile> /domain=<domain> [/dbg]"
    helpStr := "azul build site program"

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

    yFil := ""
   	yval, ok := flagMap["yaml"]
    if !ok {
        log.Fatalf("error -- no yaml flag provided!\n")
    } else {
        if yval.(string) == "none" {log.Fatalf("error -- no yaml file name provided!\n")}
        yFil = yval.(string)
        idx := strings.IndexByte(yFil, '.')
        if idx > -1 {log.Fatalf("error -- yaml file <%s> has an extension!\n", yFil)}
    }

    domain := "test.imgerp.eu"
    dval, ok := flagMap["domain"]
    if !ok {
        log.Fatalf("error -- no domain flag provided!\n")
    } else {
        if dval.(string) == "none" {log.Fatalf("error -- no domain provided!\n")}
        domain = dval.(string)
    }

	yamlFilDir := "/home/peter/cloud/domains/" + domain + "/yaml/"
	_, err = os.Stat(yamlFilDir)
	if err != nil {log.Fatalf("error -- no yaml dir: %v\n", err)}

	domDir:=  "/home/peter/cloud/domains/" + domain

	favFilnam := domDir + "/img/fav.b64"

	yamlFilnam := yamlFilDir + yFil + ".yaml"
	yamlB, err := os.ReadFile(yamlFilnam)
	if err != nil {log.Fatalf("error -- cannot read yaml file: %v\n", err)}

	idxFilnam := domDir + "/html/" + yFil + ".html"
	jsFilnam :=  domDir + "/js/" + yFil + "_ld.js"

	if dbg {fmt.Printf("yaml len: %d\n", len(yamlB))}
	if dbg {
		fmt.Println("****** files ******")
		fmt.Printf("  domain:       %s\n", domain)
		fmt.Printf("  yaml file:    %s\n", yFil + ".yaml")
		fmt.Printf("  index file:   %s\n", idxFilnam)
		fmt.Printf("  js base file: %s\n", jsFilnam)
		fmt.Println("**** end files ****")
	}


	navmap := make(map[string]string)
	bloglist := make([]string,0, 128)

	sit := azulStruct{Nav: navmap, Blogs:bloglist}
	sit.land = yFil

	err = yaml.Unmarshal(yamlB, &sit)
	if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	fav, err := os.ReadFile(favFilnam)
	if err != nil {log.Fatalf("error -- cannot read favicon file: %v\n", err)}
	sit.Fav = string(fav)

	if dbg { printInp(sit) }

	hd, err := buildHd(sit)
	if err != nil {log.Fatalf("error -- buildHd: %v!\n", err)}

	ldJs, err := buildJs(sit)
	if err != nil {log.Fatalf("error -- buildJs: %v!\n", err)}

	idxFil, err := os.Create(idxFilnam)
	if err != nil {log.Fatalf("error -- idx file: %v!\n", err)}
	defer idxFil.Close()
	hd.WriteTo(idxFil)

	jsFil, err := os.Create(jsFilnam)
	if err != nil {log.Fatalf("error -- js file: %v!\n", err)}
	defer jsFil.Close()
	ldJs.WriteTo(jsFil)

	fmt.Println("*** buildAzul success *****")
}

func buildHd(sit azulStruct)(hd bytes.Buffer, err error) {

	hd.Grow(1024)
	hd.WriteString("<!DOCTYPE html>\n")
	hd.WriteString("<html lang=\"en\">\n")
	hd.WriteString("<head>\n")
	if len(sit.Title) > 0 {
		hd.WriteString("<title>")
		hd.WriteString(sit.Title)
		hd.WriteString("</title>\n")
	}
	hd.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	for k,v := range sit.Meta {
		out := "<meta name=\"" + k + "\" content=\"" + v + "\">\n"
		hd.WriteString(out)
	}
	if len(sit.Fav) > 0 {
		hd.WriteString("  <link id=\"favicon\" rel=\"icon\" type=\"image/png\" href=\"data:image/png;base64,")
		hd.WriteString(sit.Fav)
		hd.WriteString("\">\n")
	} else {
		hd.WriteString("  <!-- no favicon -->\n")
	}
	if len(sit.Font) > 0 {
		out := "  <link rel=\"stylesheet\" href=\"<url>\">\n"
		hd.WriteString(out)
	} else {
		hd.WriteString("  <!-- no font -->\n")
	}


	hd.WriteString("</head>\n")
	hd.WriteString("<body>\n")
	hd.WriteString("  <script src=\"/js/azulNLib.js\"></script>\n")
	land := "  <script src=\"/js/" + sit.land +"_ld.js\"></script>\n"
	hd.WriteString(land)
	hd.WriteString("</body>\n")
	hd.WriteString("</html>\n")

	return hd, nil
}

func buildJs(sit azulStruct)(js bytes.Buffer, err error) {

	js.Grow(1024)
	lpg := `const lpg = {
	divMainObj: {
	typ: 'div',
	style: {
		border: '1px solid blue',
		},
	},
	render() {azul.divMain = azul.addElement(this.divMainObj);},
	};
`
	js.WriteString(lpg)

	fin:= `lpg.render();
document.body.appendChild(azul.divMain);
`

	js.WriteString(fin)

	return js, nil
}


func printInp(sit azulStruct) {
	fmt.Println("****** yaml input ******")
	fmt.Printf("  Title: %s\n", sit.Title)
	fmt.Printf("  Fav Size: %d\n",len(sit.Fav))
	fmt.Printf("  Meta:\n")
	for k,v := range sit.Meta {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}

	fmt.Printf("  Nav Links:\n")
	for k,v := range sit.Nav {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}

	fmt.Printf("  Blogs:\n")
	for i:=0; i< len(sit.Blogs); i++ {
		fmt.Printf(" --%d: %s\n",i+1, sit.Blogs[i])
	}

	fmt.Println("**** end yaml input ****")
}

