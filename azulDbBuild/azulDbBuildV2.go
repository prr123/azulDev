// azulDbBuild
// program to generate files for db
//
// changes
// -- check string for date
//
// V2 build go
//
// V3 build js
//

package main

import (
    "fmt"
    "os"
    "log"
    "strings"
	"reflect"
    "bytes"

    "github.com/goccy/go-yaml"
    util "github.com/prr123/utility/utilLib"
)



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

	yamlFilDir := "yaml/"
    yamlFilnam := yamlFilDir + yFil + ".yaml"
    yamlB, err := os.ReadFile(yamlFilnam)
    if err != nil {log.Fatalf("error -- cannot read yaml file: %v\n", err)}

	outFilDir := "out/"
	sqlFilnam := outFilDir + yFil + ".sql"
	jsFilnam :=  outFilDir + yFil + ".js"
	goFilnam :=  outFilDir + yFil + ".go"
//    if dbg {fmt.Printf("yaml len: %d\n", len(yamlB))}

    if dbg {
        fmt.Println("****** files ******")
        fmt.Printf("  domain:       %s\n", domain)
        fmt.Printf("  yaml file:    %s\n", yFil + ".yaml")
        fmt.Printf("  sql file:     %s\n", sqlFilnam)
        fmt.Printf("  js base file: %s\n", jsFilnam)
        fmt.Printf("  go file:      %s\n", goFilnam)
        fmt.Println("**** end files ****")
    }

    sqlFil, err := os.Create(sqlFilnam)
    if err != nil {log.Fatalf("error -- creating sql file: %v!\n", err)}
    defer sqlFil.Close()

    jsFil, err := os.Create(jsFilnam)
    if err != nil {log.Fatalf("error -- creating js file: %v!\n", err)}
    defer jsFil.Close()

    goFil, err := os.Create(goFilnam)
    if err != nil {log.Fatalf("error -- creating go file: %v!\n", err)}
    defer goFil.Close()

	sit := map[string]any{}
    err = yaml.Unmarshal(yamlB, &sit)
    if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	PrintMap(sit)

//	fmt.Printf("map:\n%v\n*******\n",sit)
	tbl, ctyp, err := buildSql(sit)
    if err != nil {log.Fatalf("error -- buildSql: %v!\n", err)}
	tbl.WriteTo(sqlFil)
	ctyp.WriteTo(sqlFil)

	goBuf, err := buildGoStruct(sit)
    if err != nil {log.Fatalf("error -- buildGoStruct: %v!\n", err)}
	goBuf.WriteTo(goFil)

	fmt.Println("*** success azulDbBuild ***")
}


func buildGoStruct (mp map[string]any) (goBuf *bytes.Buffer, err error) {

	dbg := true
	goBuf = new(bytes.Buffer)
	subStruct := new(bytes.Buffer)

    for k,v := range mp {
        typ := reflect.TypeOf(v)
        nam := typ.Name()
        knd:= typ.Kind()
        if dbg {fmt.Printf("goStruct k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}

		switch knd {
        case 21:
			field := fmt.Sprintf("type %s struct {\n", k)
			goBuf.WriteString(field)

			sub, err1 := buildSubStruct(v.(map[string]any), k)
			sub.WriteTo(goBuf)
			if err1 != nil {return goBuf, fmt.Errorf(" struct type: %v\n", err1)}
            mps := v.(map[string]any)
            for ks, vs := range mps {
                typs := reflect.TypeOf(vs)
                nams := typs.Name()
                knds:= typs.Kind()
                if dbg {fmt.Printf("  goStruct sub struct k: %s nam: %s kind: %d v: %v\n", ks, nams, knds, vs)}
            }
        default:
			return goBuf, fmt.Errorf("unimplemented type: %s -- %d: %v", k, knd, v)
		}
//		goBuf.WriteString("}\n")
    }
	subStruct.WriteTo(goBuf)
	return goBuf, nil
}

func buildSubStruct (mp map[string]any, nam string)(subStr *bytes.Buffer, err error) {

	dbg := true
	subStr = new(bytes.Buffer)
	subL2 := new(bytes.Buffer)

    for k,v := range mp {
        typ := reflect.TypeOf(v)
        nam := typ.Name()
        knd:= typ.Kind()
        if dbg {fmt.Printf("  goSubStruct k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}

        switch knd {
        // bool
        case 1:
            entry := fmt.Sprintf("  %s bool\n", k)
            subStr.WriteString(entry)

        // int
        case 2:
            entry := fmt.Sprintf("  %s int\n", k)
            subStr.WriteString(entry)

		case 21:
			entry := fmt.Sprintf("  %s %s_typ\n",k, k)
            subStr.WriteString(entry)

			field := fmt.Sprintf("type %s_typ struct {\n", k)
			subL2.WriteString(field)
			sub2, err := buildSubStruct( v.(map[string]any), nam)
			sub2.WriteTo(subL2)
			if err != nil {return subStr, fmt.Errorf("  buildSub map error: %v")}
        // string
        case 24:
            entry := fmt.Sprintf("  %s string\n", k)
            subStr.WriteString(entry)

        default:
			return subStr, fmt.Errorf(" SubStruct unimplemented type: %s -- %d: %v", k, knd, v)
		}
    }

	subStr.WriteString("}\n")
	subL2.WriteTo(subStr)
	return subStr, nil
}



func buildSql(mp map[string]any) (tbl, ctyp bytes.Buffer, err error) {

	dbg := true
	tbl.Grow(1024)
	ctyp.Grow(1024)

    for k,v := range mp {
        typ := reflect.TypeOf(v)
        nam := typ.Name()
        knd:= typ.Kind()
        if dbg {fmt.Printf("parse tables k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}

		if knd != 21 {return tbl, ctyp, fmt.Errorf("parse table: %s %d", k, knd)}
		tblStr := fmt.Sprintf("create table %s (\n", k)
		tbl.WriteString(tblStr)
		err1 := buildTbl(v.(map[string]any), &tbl, &ctyp)
		if err1 != nil {return tbl, ctyp, fmt.Errorf("parse table entry: %v", err1)}

		tbl.WriteString(");\n")
	}

	return tbl, ctyp, nil
}

func buildTbl (mp map[string]any, tbl, ctyp *bytes.Buffer) (err error) {

	dbg := true
    for k,v := range mp {
        typ := reflect.TypeOf(v)
        nam := typ.Name()
        knd:= typ.Kind()
        if dbg {fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}

		switch knd {
		// bool
		case 1:
            entry := fmt.Sprintf("  %s bool,\n", k)
            tbl.WriteString(entry)

		// int
		case 2:
            entry := fmt.Sprintf("  %s int,\n", k)
            tbl.WriteString(entry)
            field := fmt.Sprintf("  %s int\n", k)
            ctyp.WriteString(field)

		// string
		case 24:
			// check whether string is a date
			entry := fmt.Sprintf("  %s varchar(%d),\n", k, len(v.(string)))
			tbl.WriteString(entry)

		// map
        case 21:
			entry := fmt.Sprintf("  %s %s_typ,\n", k, k)
			tbl.WriteString(entry)

			err1 := buildCTyp(v.(map[string]any), k, ctyp)
			if err1 != nil {return fmt.Errorf("complex type: %v\n", err1)}
            mps := v.(map[string]any)
            for ks, vs := range mps {
                typs := reflect.TypeOf(vs)
                nams := typs.Name()
                knds:= typs.Kind()
                if dbg {fmt.Printf("  complex k: %s nam: %s kind: %d v: %v\n", ks, nams, knds, vs)}
            }
        default:
			return fmt.Errorf("unimplemented type: %s -- %d: %v", k, knd, v)
		}
    }
	return nil
}

func buildCTyp (mp map[string]any, nam string, ctyp *bytes.Buffer) (err error) {

	dbg := true
    cStr := fmt.Sprintf("create type %s_typ as (\n", nam)
    ctyp.WriteString(cStr)

    for k,v := range mp {
        typ := reflect.TypeOf(v)
        nam := typ.Name()
        knd:= typ.Kind()
        if dbg {fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}

        switch knd {
        // bool
        case 1:
            entry := fmt.Sprintf("  %s bool,\n", k)
            ctyp.WriteString(entry)

        // int
        case 2:
            entry := fmt.Sprintf("  %s int,\n", k)
            ctyp.WriteString(entry)

        // string
        case 24:
            entry := fmt.Sprintf("  %s varchar(%d),\n", k, len(v.(string)))
            ctyp.WriteString(entry)

        default:
			return fmt.Errorf("unimplemented type: %s -- %d: %v", k, knd, v)
		}
    }
	ctyp.WriteString(");\n")
	return nil
}


func PrintMap(mp map[string]any) {

	fmt.Printf("********\nmap:\n%v\n*******\n",mp)

	for k,v := range mp {
		typ := reflect.TypeOf(v)
		nam := typ.Name()
		knd:= typ.Kind()
		fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)
		//test for map
		if knd == 21 {
			mps := v.(map[string]any)
			for ks, vs := range mps {
				typs := reflect.TypeOf(vs)
				nams := typs.Name()
				knds:= typs.Kind()
				fmt.Printf("  ks: %s nam: %s kind: %d vs: %v\n", ks, nams, knds, vs)
				if knds == 21 {
					mp3 := v.(map[string]any)
					for k3, v3 := range mp3 {
						typ3 := reflect.TypeOf(v3)
						nam3 := typ3.Name()
						knd3:= typ3.Kind()
						fmt.Printf("     k3: %s nam: %s kind: %d v3: %v\n", k3, nam3, knd3, v3)
					}
				}
			}
		}
	}

}
