// buildAzul: program that builds a website
//
// v2: build basic idx file
//  -- mv files to domain
//	-- add fav file
//
// V3 new yaml file
//
// V4  aggreate objs unde azul.SPA
//
// V5: add nav content
//

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
	Pages map[string]string `yaml:"pages,omitempty"`
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
	jsNavFilnam :=  domDir + "/js/" + yFil + "_nav.js"

	if dbg {fmt.Printf("yaml len: %d\n", len(yamlB))}
	if dbg {
		fmt.Println("****** files ******")
		fmt.Printf("  domain:       %s\n", domain)
		fmt.Printf("  yaml file:    %s\n", yFil + ".yaml")
		fmt.Printf("  index file:   %s\n", idxFilnam)
		fmt.Printf("  js base file: %s\n", jsFilnam)
		fmt.Printf("  js Nav file:  %s\n", jsNavFilnam)
		fmt.Println("**** end files ****")
	}


	navMap := make(map[string]string)
	metaMap := make(map[string]string)
	pageMap := make(map[string]string)
	bloglist := make([]string,0, 128)


	sit := azulStruct{Nav: navMap, Meta: metaMap, Pages: pageMap, Blogs:bloglist}
	sit.land = yFil

	err = yaml.Unmarshal(yamlB, &sit)
	if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	fav, err := os.ReadFile(favFilnam)
	if err != nil {log.Fatalf("error -- cannot read favicon file: %v\n", err)}
	sit.Fav = string(fav)

	if dbg { printInp(sit) }

	idxFil, err := os.Create(idxFilnam)
	if err != nil {log.Fatalf("error -- idx file: %v!\n", err)}
	defer idxFil.Close()

	jsFil, err := os.Create(jsFilnam)
	if err != nil {log.Fatalf("error -- js file: %v!\n", err)}
	defer jsFil.Close()

	jsNavFil, err := os.Create(jsNavFilnam)
	if err != nil {log.Fatalf("error -- js Nav file: %v!\n", err)}
	defer jsNavFil.Close()

	hd, err := buildHd(sit)
	if err != nil {log.Fatalf("error -- buildHd: %v!\n", err)}

	mainJs, err := buildPage(sit)
	if err != nil {log.Fatalf("error -- buildMain: %v!\n", err)}

	err = buildNavCont(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildNav: %v!\n", err)}
	err = buildNav(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildNav: %v!\n", err)}

	navJs , err  := buildNavPages(sit)
	if err != nil {log.Fatalf("error -- buildNavPages: %v!\n", err)}

	err = buildMain(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildMain: %v!\n", err)}

	err = buildFootCont(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildFootCont: %v!\n", err)}
	err = buildFoot(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildFoot: %v!\n", err)}

	renderSPA(&mainJs)

	hd.WriteTo(idxFil)

	mainJs.WriteTo(jsFil)

	navJs.WriteTo(jsNavFil)

	if dbg {
		fmt.Println("****** js file ******")
		mainJs.Reset()
		mainJs.WriteTo(os.Stdout)
		fmt.Println("**** end js file ****")
	}
	fmt.Println("*** buildAzul success *****")
}

func buildHd(sit azulStruct)(hd bytes.Buffer, err error) {

	hd.Grow(4096)
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
	lnav := "  <script src=\"/js/" + sit.land +"_nav.js\" async></script>\n"
	hd.WriteString(lnav)
	hd.WriteString("</html>\n")

	return hd, nil
}

func buildPage(sit azulStruct)(js bytes.Buffer, err error) {

	js.Grow(1024)
	js.WriteString("const SPA = {};\n")
	lpg := `SPA.lpg = {
	divSPAObj: {
	typ: 'div',
	style: {border: '1px solid blue',},
	},
	render() {azul.SPA = azul.addElement(SPA.lpg.divSPAObj);},
};
`
	js.WriteString(lpg)

	js.WriteString("SPA.lpg.render();\n")

	return js, nil
}

func buildNavPages(sit azulStruct)(js bytes.Buffer, err error) {

	js.Grow(1024)

	js.WriteString("const NavServ = {\n")
	for k, v := range sit.Pages {
		out := fmt.Sprintf("%s: {\n", v)
		js.WriteString(out)
		js.WriteString("navfun() {\n")
		msg := fmt.Sprintf("'testing %s!'", k)
		fun := fmt.Sprintf("  console.log(%s);\n", msg)
		js.WriteString(fun)
		js.WriteString("},\n")
		js.WriteString("},\n")
	}

	js.WriteString("};\n")
	for k, v := range sit.Pages {
		fun := fmt.Sprintf("NavServ.%s.navfun",v)
		ev := fmt.Sprintf("azul.Nav.%s.addEventListener('click', %s);\n", k, fun)
		js.WriteString(ev)
	}

	return js, nil
}


func buildNavCont(sit azulStruct, js *bytes.Buffer)(err error) {

	navCont := `SPA.navcont = {
	butObj: {
		style: {background: 'none',border: 'none',padding: '0',cursor: 'pointer',},
        typ: 'button'
	},
    menuInlineObj: {
        style: {display: 'flex',justifyContent: 'flex-end',border: '1px dashed green',minHeight: '30px', margin: '5px'},
        typ: 'div',
    },
    itemObj: {
        style: {border: '1px dashed orange',margin: '5px',minWidth: '100px', textAlign: 'center',},
        typ: 'div',
    },
	creCont () {
		const cel = azul.addElement(SPA.navcont.menuInlineObj);
`
	js.WriteString(navCont)

//	itLen := len(sit.Nav)
//	for i:=0; i< itLen; i++ {
	ic :=0
	for k, v := range sit.Nav {
		ic++
		nButObj := fmt.Sprintf("const %sObj = {textContent: '%s'};\n", k, v)
		js.WriteString(nButObj)
		nB1 := fmt.Sprintf("Object.assign(%sObj,SPA.navcont.butObj);\n", k)
		js.WriteString(nB1)

		nBut := fmt.Sprintf("const %s = azul.addElement(%sObj);\n", k, k)
		js.WriteString(nBut)
		nB2 := fmt.Sprintf("azul.Nav.%s = %s;\n",k, k)
		js.WriteString(nB2)

		it := fmt.Sprintf("it%d",ic)
		itEl := fmt.Sprintf("const %s = azul.addElement(SPA.navcont.itemObj);\n", it)
		js.WriteString(itEl)
		nB3 := fmt.Sprintf("%s.appendChild(%s);\n",it, k)
		js.WriteString(nB3)
		out := fmt.Sprintf("cel.appendChild(%s);\n",it)
		js.WriteString(out)
	}

navContEnd := `		return cel;
	},
};
`
	js.WriteString(navContEnd)

	return nil
}

func buildNav(sit azulStruct, js *bytes.Buffer)(err error) {

	nav := `SPA.nav = {
    divNavObj: {
    typ: 'div',
    style: {minHeight: '100px',margin: '5px',border: '1px solid red',},
    },
    render() {
		azul.Nav = azul.addElement(SPA.nav.divNavObj);
		if (SPA.navcont !== undefined) {
			const cel= SPA.navcont.creCont();
			azul.Nav.appendChild(cel);
		};
		azul.SPA.appendChild(azul.Nav);
	},
};
`
	js.WriteString(nav)

	js.WriteString("SPA.nav.render();\n")

	return nil
}

func buildMain(sit azulStruct, js *bytes.Buffer)(err error) {

	main := `SPA.main = {
    divMainObj: {
    typ: 'div',
    style: {minHeight: '100px',margin: '5px',border: '1px solid green',},
    },
    render() {
		azul.Main = azul.addElement(SPA.main.divMainObj);
		azul.SPA.appendChild(azul.Main);
	},
};
`
	js.WriteString(main)

	js.WriteString("SPA.main.render();\n")

	return nil
}

func buildFootCont(sit azulStruct, js *bytes.Buffer)(err error) {

	footCont := `SPA.footcont = {
	butObj: {
		style: {background: 'none',border: 'none',padding: '0',cursor: 'pointer',},
        typ: 'button'
	},
    gridObj: {
        nrow: 3, ncol: 4,
        cols: '1fr 1fr 1fr 1fr',
        style: {display: 'grid',border: '1px solid blue',margin: '10px',},
        elStyle: {border: '1px dashed green',minHeight: '30px',},
    },
	creGrid () {
	const fgrid = azul.addGrid(this.gridObj);
	return fgrid;
	}
};
`
	js.WriteString(footCont)

	return nil
}

func buildFoot(sit azulStruct, js *bytes.Buffer)(err error) {

	foot := `SPA.foot = {
    divFootObj: {
    typ: 'div',
    style: {minHeight: '100px',margin: '5px', border: '1px solid purple',},
    },
    render() {
		azul.Foot = azul.addElement(SPA.foot.divFootObj);
		if (SPA.footcont !== undefined) {
			const cel= SPA.footcont.creGrid();
			azul.Foot.appendChild(cel);
		};
		azul.SPA.appendChild(azul.Foot);
	},
};
`
	js.WriteString(foot)

	js.WriteString("SPA.foot.render();\n")

	return nil
}

func renderSPA (js *bytes.Buffer) {

	fin := "document.body.appendChild(azul.SPA);\n"

	js.WriteString(fin)

}

func printInp(sit azulStruct) {
	fmt.Println("****** yaml input ******")
	fmt.Printf("  Title: %s\n", sit.Title)
	fmt.Printf("  Fav Size: %d\n",len(sit.Fav))
	fmt.Printf("  Meta:\n")
	for k,v := range sit.Meta {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}

	fmt.Printf("  Nav Buttons:\n")
	for k,v := range sit.Nav {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}

	fmt.Printf("  Pages:\n")
	for k,v := range sit.Pages {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}


	fmt.Printf("  Blogs:\n")
	for i:=0; i< len(sit.Blogs); i++ {
		fmt.Printf(" --%d: %s\n",i+1, sit.Blogs[i])
	}

	fmt.Println("**** end yaml input ****")
}

