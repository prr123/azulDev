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
// V6: simplify Nav
//  -- add foot content
//  -- add outline code for virtual pages
//
// V7: add virtual page
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
	Nav []string `yaml:"nav"`
	Blogs []string `yaml:"blogs,omitempty"`
	Foot azulFootStruct `yaml:"footer,omitempty"`
	FCols []azulFColStruct `yaml:"fcols,omitempty"`
	domDir string
	land string
}

type azulFootStruct struct {
	Nrows int `yaml:"nrow"`
	Ncols int `yaml:"ncol"`
}

type azulFColStruct struct {
	Col footColItem `yaml:"col"`
}

type footColItem struct {
	CNum int `yaml:"cnum"`
	CNam []string `yaml:"cnam,omitempty"`
}

type vpgObj struct {
	dbg bool
	vpg bool
}

func main() {

	bldVpg := vpgObj{dbg: false, vpg: false}
	numarg := len(os.Args)
    flags:=[]string{"dbg", "yaml", "domain", "vpg"}

    useStr := "/yaml=<yamlfile> /domain=<domain> [/vpg] [/dbg]"
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

    _, ok := flagMap["dbg"]
    if ok {bldVpg.dbg = true}

    _, ok = flagMap["vpg"]
    if ok {bldVpg.vpg = true}

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

	metaMap := make(map[string]string)
	bloglist := make([]string,0, 128)
	navList := make([]string, 0, 24)

	sit := azulStruct{Nav: navList, Meta: metaMap, Blogs:bloglist}
	sit.land = yFil

	yamlFilDir := "/home/peter/cloud/domains/" + domain + "/yaml/"
	_, err = os.Stat(yamlFilDir)
	if err != nil {log.Fatalf("error -- no yaml dir: %v\n", err)}

	sit.domDir =  "/home/peter/cloud/domains/" + domain
	
	favFilnam := sit.domDir + "/img/fav.b64"

	yamlFilnam := yamlFilDir + yFil + ".yaml"
	yamlB, err := os.ReadFile(yamlFilnam)
	if err != nil {log.Fatalf("error -- cannot read yaml file: %v\n", err)}

	idxFilnam := sit.domDir + "/html/" + yFil + ".html"
	jsFilnam :=  sit.domDir + "/js/" + yFil + "_ld.js"
	jsNavFilnam := sit.domDir + "/js/" + yFil + "_butserv.js"

	if len(yamlB) == 0 {log.Fatalf("error -- yaml file empty!\n")}
	if bldVpg.dbg {
		fmt.Println("****** files ******")
		fmt.Printf("  domain:       %s\n", domain)
		fmt.Printf("  yaml file:    %s\n", yFil + ".yaml")
		fmt.Printf("  index file:   %s\n", idxFilnam)
		fmt.Printf("  js base file: %s\n", jsFilnam)
		fmt.Printf("  js Nav file:  %s\n", jsNavFilnam)
		fmt.Printf("  vpg:			%t\n", bldVpg.vpg)
		fmt.Println("**** end files ****")
	}


	err = yaml.Unmarshal(yamlB, &sit)
	if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	if bldVpg.dbg { printInp(sit) }

	fav, err := os.ReadFile(favFilnam)
	if err != nil {log.Fatalf("error -- cannot read favicon file: %v\n", err)}
	sit.Fav = string(fav)

	idxFil, err := os.Create(idxFilnam)
	if err != nil {log.Fatalf("error -- idx file: %v!\n", err)}
	defer idxFil.Close()

	jsFil, err := os.Create(jsFilnam)
	if err != nil {log.Fatalf("error -- js file: %v!\n", err)}
	defer jsFil.Close()

	jsNavFil, err := os.Create(jsNavFilnam)
	if err != nil {log.Fatalf("error -- js Nav file: %v!\n", err)}
	defer jsNavFil.Close()

	err = bldVpg.buildVPages(sit)
	if err != nil {log.Fatalf("error -- buildVPages: %v!\n", err)}

	hd, err := bldVpg.buildHd(sit)
	if err != nil {log.Fatalf("error -- buildHd: %v!\n", err)}

	mainJs, err := bldVpg.buildPage(sit)
	if err != nil {log.Fatalf("error -- buildMain: %v!\n", err)}

	err = bldVpg.buildNavCont(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildNav: %v!\n", err)}
	err = bldVpg.buildNav(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildNav: %v!\n", err)}

	butJs , err  := bldVpg.buildButServePage(sit)
	if err != nil {log.Fatalf("error -- buildNavPages: %v!\n", err)}

	err = bldVpg.buildMain(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildMain: %v!\n", err)}

	err = bldVpg.buildFootCont(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildFootCont: %v!\n", err)}
	err = bldVpg.buildFoot(sit, &mainJs)
	if err != nil {log.Fatalf("error -- buildFoot: %v!\n", err)}

	bldVpg.renderSPA(&mainJs)

	hd.WriteTo(idxFil)

	mainJs.WriteTo(jsFil)

	butJs.WriteTo(jsNavFil)

	fmt.Println("*** buildAzul success *****")
}

func (vpg vpgObj)buildHd(sit azulStruct)(hd bytes.Buffer, err error) {

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
/*
	// aux pages
	for i:=0; i<len(sit.Nav); i++ {
		v:=sit.Nav[i]
//		k:= fmt.Sprintf("navbut%d",i+1)
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		_, err  := os.Stat(filnam)
//		if err != nil {return hd, fmt.Errorf("%s not found: %v", filnam, err)}
		if err != nil {
			fmt.Printf("info -- file: %s not found!\n", filnam)
			continue
		}
		script := fmt.Sprintf("<script src=\"/js/%s_%s.js\" async></script>", sit.land, v)
		hd.WriteString(script)
	}

   var fbutNam [24]string
    cnt:=0
    for ic:=0; ic< len(sit.FCols); ic++ {
        for j:=0; j< len(sit.FCols[ic].Col.CNam); j++ {
            fbutNam[cnt] = sit.FCols[ic].Col.CNam[j]
            cnt++
        }
    }

    for i:=0; i<cnt; i++ {
        v:=fbutNam[i]
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		_, err  := os.Stat(filnam)
//		if err != nil {return hd, fmt.Errorf("%s not found: %v", filnam, err)}
		if err != nil {
			fmt.Printf("info -- file: %s not found!\n", filnam)
			continue
		}
		script := fmt.Sprintf("<script src=\"/js/%s_%s.js\" async></script>", sit.land, v)
		hd.WriteString(script)
	}
*/
	lnav := "  <script src=\"/js/" + sit.land +"_butserv.js\" type='module' async></script>\n"
	hd.WriteString(lnav)
	hd.WriteString("</html>\n")

	return hd, nil
}

func (bld vpgObj) buildVPages(sit azulStruct)(err error) {

	var js bytes.Buffer

	if !bld.vpg {return nil}

	js.Grow(1024)

	for i:=0; i<len(sit.Nav); i++ {
		js.Reset()
		v:=sit.Nav[i]
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		fil, err  := os.Create(filnam)
		if err != nil {return fmt.Errorf("cannot create %s: %v", filnam, err)}

		out := fmt.Sprintf("export const %s_%sVpg = {\n",sit.land,v)
		js.WriteString(out)

		out2 := `  vpgObj: {
    divStyl: {border: '1px dashed green'},
  },
  render(vpgObj) {
    const elDiv = document.createElement('div');
    Object.assign(elDiv.style, vpgObj.divStyl);
`
		js.WriteString(out2)

		// page els
		codStr := "    const p1 = document.createElement('p');\n"
		js.WriteString(codStr)
		cod2 := fmt.Sprintf("	p1.textContent = 'found %s vpg obj!';\n", v)
		js.WriteString(cod2)
		cod3 := "    elDiv.appendChild(p1);\n"
		js.WriteString(cod3)

		js.WriteString("   return elDiv;\n  },\n};\n")
		js.WriteTo(fil)
		fil.Close()
	}

   var fbutNam [24]string
    cnt:=0
    for ic:=0; ic< len(sit.FCols); ic++ {
        for j:=0; j< len(sit.FCols[ic].Col.CNam); j++ {
            fbutNam[cnt] = sit.FCols[ic].Col.CNam[j]
            cnt++
        }
    }

    for i:=0; i<cnt; i++ {
		js.Reset()
        v:=fbutNam[i]
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		fil, err  := os.Create(filnam)
		if err != nil {return fmt.Errorf("cannot create %s: %v", filnam, err)}

		out := fmt.Sprintf("export const %s_%sVpg = {\n",sit.land,v)
		js.WriteString(out)

		out2 := `  vpgObj: {
    divStyl: {border: '1px dashed green'},
  },
  render(vpgObj) {
    const elDiv = document.createElement('div');
    Object.assign(elDiv.style, vpgObj.divStyl);
`
		js.WriteString(out2)

		// page els
		codStr := "    const p1 = document.createElement('p');\n"
		js.WriteString(codStr)
		cod2 := fmt.Sprintf("	p1.textContent = 'found %s vpg obj!';\n", v)
		js.WriteString(cod2)
		cod3 := "    elDiv.appendChild(p1);\n"
		js.WriteString(cod3)

		js.WriteString("   return elDiv;\n  },\n};\n")
		js.WriteTo(fil)
		fil.Close()
	}

	return nil
}

func (bld vpgObj) buildPage(sit azulStruct)(js bytes.Buffer, err error) {

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

func (vpg vpgObj) buildButServePage(sit azulStruct)(js bytes.Buffer, err error) {

	js.Grow(4096)

	for i:=0; i<len(sit.Nav); i++ {
		v:=sit.Nav[i]
//		k:= fmt.Sprintf("navbut%d",i+1)
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		_, err  := os.Stat(filnam)
//		if err != nil {return hd, fmt.Errorf("%s not found: %v", filnam, err)}
		if err != nil {
			fmt.Printf("info -- file: %s not found!\n", filnam)
			continue
		}
		nam:= fmt.Sprintf("%s_%s",sit.land, v)
		script := fmt.Sprintf("import {%sVpg} from '/js/%s.js';\n", nam, nam)
		js.WriteString(script)
	}

   var fbutNam [24]string
    cnt:=0
    for ic:=0; ic< len(sit.FCols); ic++ {
        for j:=0; j< len(sit.FCols[ic].Col.CNam); j++ {
            fbutNam[cnt] = sit.FCols[ic].Col.CNam[j]
            cnt++
        }
    }

    for i:=0; i<cnt; i++ {
        v:=fbutNam[i]
		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
		_, err  := os.Stat(filnam)
//		if err != nil {return hd, fmt.Errorf("%s not found: %v", filnam, err)}
		if err != nil {
			fmt.Printf("info -- file: %s not found!\n", filnam)
			continue
		}
		nam:= fmt.Sprintf("%s_%s",sit.land, v)
		script := fmt.Sprintf("import {%sVpg} from '/js/%s.js';\n", nam, nam)
		js.WriteString(script)
	}


	js.WriteString("const ButServ = {\n")
	for i:=0; i<len(sit.Nav); i++ {
		v:=sit.Nav[i]
		k:= fmt.Sprintf("navbut%d",i+1)

		out := fmt.Sprintf("%s: {\n",v )
		js.WriteString(out)
		js.WriteString("  render() {\n")

//		filnam := fmt.Sprintf("%s/js/%s_%s.js", sit.domDir, sit.land, v)
//		_, errfil := os.Stat(filnam)
//		if errfil != nil {
			codStr := `    vpg = document.createElement('div');
    const p1 = document.createElement('p');
`
			js.WriteString(codStr)
			cod2 := fmt.Sprintf("	p1.textContent = 'could not find %s obj!';\n", v)
			js.WriteString(cod2)
			cod3 := "    vpg.appendChild(p1);\n"
			js.WriteString(cod3)
			js.WriteString("    return vpg;\n")
//		} else {
//			codAlt := fmt.Sprintf("const %sDiv = %s_%sVpg.render();\n",v, sit.land, v)
//			js.WriteString(codAlt)
//		}
		js.WriteString("  },\n")
		js.WriteString("  navfun() {\n")
		js.WriteString("    let vpg = {};\n")
		msg := fmt.Sprintf("    //testing %s!\n", k)
		js.WriteString(msg)
		obj := fmt.Sprintf("%s_%sVpg", sit.land, v)
		tst:= fmt.Sprintf("    if (typeof %s != \"undefined\") {\n", obj)
		js.WriteString(tst)
        rStr := fmt.Sprintf("      vpg = %s.render(%s.vpgObj);}\n",obj, obj)
        js.WriteString(rStr)
		js.WriteString("    else\n")
		rStr2 := fmt.Sprintf("    {vpg = ButServ.%s.render();}\n",v)
		js.WriteString(rStr2)
		js.WriteString("    azul.rplSect(vpg);\n")
		js.WriteString("  },\n")
		js.WriteString("},\n")
	}

//	var fbutNam [24]string
	cnt=0
	for ic:=0; ic< len(sit.FCols); ic++ {
		for j:=0; j< len(sit.FCols[ic].Col.CNam); j++ {
			fbutNam[cnt] = sit.FCols[ic].Col.CNam[j]
			cnt++
		}
	}

	for i:=0; i<cnt; i++ {
		v:=fbutNam[i]
		k:= fmt.Sprintf("fbut%d",i+1)
		out := fmt.Sprintf("%s: {\n",v )
		js.WriteString(out)

//		js.WriteString("navfun() {\n")

		js.WriteString("  render() {\n")

		codStr := `    vpg = document.createElement('div');
    const p1 = document.createElement('p');
`
		js.WriteString(codStr)
		cod2 := fmt.Sprintf("	p1.textContent = 'could not find %s obj!';\n", v)
		js.WriteString(cod2)
		cod3 := "    vpg.appendChild(p1);\n"
		js.WriteString(cod3)
		js.WriteString("    return vpg;\n")
//		} else {
//			codAlt := fmt.Sprintf("const %sDiv = %s_%sVpg.render();\n",v, sit.land, v)
//			js.WriteString(codAlt)
//		}
		js.WriteString("  },\n")
		js.WriteString("  navfun() {\n")
		js.WriteString("    let vpg = {};\n")
		msg := fmt.Sprintf("    //testing %s!\n", k)
		js.WriteString(msg)
		obj := fmt.Sprintf("%s_%sVpg", sit.land, v)
		tst:= fmt.Sprintf("    if (typeof %s != \"undefined\") {\n", obj)
		js.WriteString(tst)
        rStr := fmt.Sprintf("      vpg = %s.render(%s.vpgObj);}\n",obj, obj)
        js.WriteString(rStr)
		js.WriteString("    else\n")
		rStr2 := fmt.Sprintf("    {vpg = ButServ.%s.render();}\n",v)
		js.WriteString(rStr2)
		js.WriteString("    azul.rplSect(vpg);\n")
		js.WriteString("  },\n")
		js.WriteString("},\n")

	}

	js.WriteString("};\n")

	for i:=0; i<len(sit.Nav); i++ {
		v:=sit.Nav[i]
		k:= fmt.Sprintf("navbut%d",i+1)
		fun := fmt.Sprintf("ButServ.%s.navfun",v)
		ev := fmt.Sprintf("azul.Nav.%s.addEventListener('click', %s);\n", k, fun)
		js.WriteString(ev)
	}

	for i:=0; i<cnt; i++ {
		v:=fbutNam[i]
		k:= fmt.Sprintf("fbut%d",i+1)
		fun := fmt.Sprintf("ButServ.%s.navfun",v)
		ev := fmt.Sprintf("azul.Foot.%s.addEventListener('click', %s);\n", k, fun)
		js.WriteString(ev)
	}

	return js, nil
}


func (vpg vpgObj) buildNavCont(sit azulStruct, js *bytes.Buffer)(err error) {

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

	for ic:=0; ic<len(sit.Nav); ic++ {
		k:= fmt.Sprintf("navbut%d", ic+1)
		v:= sit.Nav[ic]
		nBut := fmt.Sprintf("const %s = azul.addElement(SPA.navcont.butObj);\n", k)
		js.WriteString(nBut)
		nB1 := fmt.Sprintf("%s.textContent = '%s';\n", k, v)
		js.WriteString(nB1)

		nB2 := fmt.Sprintf("azul.Nav.%s = %s;\n",k, k)
		js.WriteString(nB2)

		it := fmt.Sprintf("it%d",ic+1)
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

func (vpg vpgObj) buildNav(sit azulStruct, js *bytes.Buffer)(err error) {

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

func (vpg vpgObj) buildMain(sit azulStruct, js *bytes.Buffer)(err error) {

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

func (vpg vpgObj) buildFootCont(sit azulStruct, js *bytes.Buffer)(err error) {

	footCont := `SPA.footcont = {
	butObj: {
		style: {background: 'none',border: 'none',padding: '5px 0 0 10px',cursor: 'pointer',},
        typ: 'button'
	},
    gridObj: {
`
	footCont2 := `        style: {display: 'grid',border: '1px solid blue',margin: '5px',},
        elStyle: {border: '1px dashed green',minHeight: '30px',},
    },

	creGrid () {
		const fgrid = azul.addGrid(this.gridObj);
`

	footCont4 := `	return fgrid;
	}
};
`

	rep:=""
	for icol:=0; icol< sit.Foot.Ncols; icol++ {
		rep +="1fr "
	}

	footCont1 := fmt.Sprintf("		nrow: %d, ncol: %d, cols: '%s',\n", sit.Foot.Nrows, sit.Foot.Ncols, rep)

	js.WriteString(footCont)
	js.WriteString(footCont1)
	js.WriteString(footCont2)

	// adding buttons
	ic:=1
	for i:=0; i< len(sit.FCols); i++ {
		colNum := sit.FCols[i].Col.CNum
        for j:=0; j<len(sit.FCols[i].Col.CNam); j++ {
				nam:=sit.FCols[i].Col.CNam[j]
				fbStr := fmt.Sprintf("const fbut%d = azul.addElement(this.butObj);\n", ic)
				js.WriteString(fbStr)
				fbStr1:= fmt.Sprintf("fbut%d.textContent = '%s';\n", ic, nam)
				js.WriteString(fbStr1)
				fbStr2 := fmt.Sprintf("fgrid.els[%d][%d].appendChild(fbut%d);\n",j, colNum-1, ic)
				js.WriteString(fbStr2)
				fbStr3 := fmt.Sprintf("azul.Foot.fbut%d = fbut%d;\n", ic, ic)
				js.WriteString(fbStr3)
				ic++
		}
	}
	fbStr4 := fmt.Sprintf("azul.Foot.nbut = %d;\n", ic-1)
	js.WriteString(fbStr4)


	js.WriteString(footCont4)

	return nil
}

func (vpg vpgObj) buildFoot(sit azulStruct, js *bytes.Buffer)(err error) {

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

func (vpg vpgObj) renderSPA (js *bytes.Buffer) {

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
	for i:=0; i< len(sit.Nav); i++ {
		fmt.Printf("  -- %d v: %s\n",i+1, sit.Nav[i])
	}

	fmt.Printf("  Blogs:\n")
	for i:=0; i< len(sit.Blogs); i++ {
		fmt.Printf(" --%d: %s\n",i+1, sit.Blogs[i])
	}

	fmt.Printf("  Footer:\n")
	fmt.Printf("    nrows: %d ncols: %d\n", sit.Foot.Nrows, sit.Foot.Ncols)

	fmt.Printf("  Footer Cols:\n")
	for i:=0; i< len(sit.FCols); i++ {
		fmt.Printf(" --%d: col %d\n",i+1, sit.FCols[i].Col.CNum)
		for j:=0; j<len(sit.FCols[i].Col.CNam); j++ {
			fmt.Printf("   --%d: %s\n", j+1, sit.FCols[i].Col.CNam[j])
		}
	}


	fmt.Println("**** end yaml input ****")
}

