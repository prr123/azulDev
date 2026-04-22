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
// V4 refactor
//
// V5 parse Tables
//

package main

import (
    "fmt"
    "os"
    "log"
    "strings"
	"reflect"
    "bytes"
	"context"

    "github.com/goccy/go-yaml"

	"github.com/jackc/pgx/v5"
    util "github.com/prr123/utility/utilLib"
)

type table struct {
	name string
	fields []field
}

type field struct {
	name string
	prop []string
}

type dbObj struct {
	dbg bool
	dbctx context.Context
	sit map[string]any
	dbConn *pgx.Conn
	sqlFil *os.File
	base string
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

	yamlFilDir := "yaml/"
    yamlFilnam := yamlFilDir + yFil + ".yaml"
    yamlB, err := os.ReadFile(yamlFilnam)
    if err != nil {log.Fatalf("error -- cannot read yaml file: %v\n", err)}

	outFilDir := "out/"
	sqlFilnam := outFilDir + yFil + ".sql"
	jsFilnam :=  outFilDir + yFil + ".js"
//    if dbg {fmt.Printf("yaml len: %d\n", len(yamlB))}

    if dbg {
        fmt.Println("****** files ******")
        fmt.Printf("  domain:       %s\n", domain)
        fmt.Printf("  yaml file:    %s\n", yFil + ".yaml")
        fmt.Printf("  sql file:     %s\n", sqlFilnam)
        fmt.Printf("  js base file: %s\n", jsFilnam)
//        fmt.Printf("  go file:      %s\n", goFilnam)
        fmt.Println("**** end files ****")
    }

    sqlFil, err := os.Create(sqlFilnam)
    if err != nil {log.Fatalf("error -- creating sql file: %v!\n", err)}
    defer sqlFil.Close()

    jsFil, err := os.Create(jsFilnam)
    if err != nil {log.Fatalf("error -- creating js file: %v!\n", err)}
    defer jsFil.Close()


	sit := map[string]any{}
    err = yaml.Unmarshal(yamlB, &sit)
    if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	if dbg {PrintMap(sit)}
	db:= &dbObj{dbg: dbg, sit: sit, base: yFil}

	err = db.buildDb()
	if err != nil {log.Fatalf("error -- buildDb: %v\n", err)}
//	if dbConn == nil {log.Fatalf("error -- no db conn\n")}

	err = db.initdbconn()
	if err != nil {log.Fatalf("error -- initdb for %s: %v\n", db.sit["db"], err)}
	defer db.dbConn.Close(db.dbctx)

	tables, err := db.ParseTables()
	if err != nil {log.Fatalf("error -- ParseTables: %v\n", err)}

	if dbg {PrintTables(tables)}

	err = db.buildDbTables(tables)
	if err != nil {log.Fatalf("error -- buildDbTables: %v\n", err)}

	err = db.buildGoCode(tables)
	if err != nil {log.Fatalf("error -- buildGoCode: %v\n", err)}

	err = db.buildGoTestCode(tables)
	if err != nil {log.Fatalf("error -- buildGoTestCode: %v\n", err)}

	fmt.Println("*** success azulDbBuild ***")
}

func (db *dbObj) buildDbTables(tables []table)(err error) {

	var q strings.Builder
	q.Grow(1024)

	dbg:= db.dbg
	dbConn := db.dbConn
	if dbConn == nil {return fmt.Errorf("not db Conn!")}
	if dbg {fmt.Printf("dbg info: number of tables: %d\n", len(tables))}

	bctx := context.Background()

	for i:=0; i< len(tables); i++ {
		q.Reset()
		tbl := tables[i]
		// check whether table exists
		var exists bool
		query := `SELECT EXISTS (
    SELECT FROM information_schema.tables 
    WHERE  table_schema = 'public' 
    AND    table_name   = $1
);`

		err := dbConn.QueryRow(bctx, query, strings.ToLower(tbl.name)).Scan(&exists)
		if err != nil {return fmt.Errorf("table query: %v", err)}

		if exists {
			fmt.Printf("info -- table %s already exists! skipping!\n", tbl.name)
			continue
		}

		creTbl := fmt.Sprintf("create table if not exists %s (", tbl.name)
		q.WriteString(creTbl)
		for icol:= 0; icol< len(tbl.fields); icol++ {
//			colStr := fmt.Sprintf("%s",fields[i].name)
			fld := tbl.fields[icol]
			q.WriteString(fld.name)
			for iprop:=0; iprop<len(fld.prop); iprop++ {
				q.WriteByte(' ')
				q.WriteString(fld.prop[iprop])
			}
			if icol < len(tbl.fields) -1 {q.WriteByte(',')}
		}

		q.WriteString(");")
		if dbg {fmt.Printf("dbg info -- query: %s\n",q.String())}
	    tag, err := dbConn.Exec(bctx, q.String())
		if err != nil {return fmt.Errorf(" create table %s failed, query: >%s<: %v\n", tbl.name, q.String(), err)}
	    if dbg {fmt.Printf("info -- tag: %s\n", tag.String())}

	}
	return nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}



func (db *dbObj) buildDb () (err error) {

	mp := db.sit
	dbg := db.dbg
	dbnam, ok := mp["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)
	if len(dbNam)< 3 {return fmt.Errorf("db name too short: %s", len(dbNam))}
	if dbg {fmt.Printf("info -- db name: %s\n", dbNam)}

	dbuser, ok := mp["dbuser"]
	if !ok {return fmt.Errorf("no dbuser found!")}
	dbUser := dbuser.(string)
	if len(dbUser)< 3 {return fmt.Errorf("db name too short: %s", len(dbUser))}
	if dbg {fmt.Printf("info -- db user: %s\n", dbUser)}

    dbConnStr := "host=/var/run/postgresql user=azuldbadmin dbname=postgres"

	bctx := context.Background()

    dbConn, err := pgx.Connect(bctx, dbConnStr)
    if err != nil {return fmt.Errorf("error -- Unable to connect to database: %v\n", err)}
    defer dbConn.Close(bctx)
//	db.dbConn = dbConn

	if dbg {fmt.Printf("info -- conn established!\n")}

    query := fmt.Sprintf("SELECT count(*) FROM pg_database WHERE datname = '%s';",dbNam)

	dbCnt:= -1
	err = dbConn.QueryRow(bctx, query).Scan(&dbCnt)
    if err != nil {return fmt.Errorf(" db query failed: %v\n", err)}
    if dbg {fmt.Printf("info -- db count: %d\n", dbCnt)}
//	if dbCnt > 0 {return fmt.Errorf("db already exists!")}

	// check whther role exists
	roleCnt := -1
	roleQuery:= fmt.Sprintf("select count(*) FROM pg_roles WHERE rolname = '%s';", dbUser)
	err = dbConn.QueryRow(bctx, roleQuery).Scan(&roleCnt)
    if err != nil {return fmt.Errorf(" db roleQuery %s failed: %v\n", roleQuery, err)}
    if dbg {fmt.Printf("info -- role count: %d\n", roleCnt)}

	//create role
	if roleCnt == 0 {
		fmt.Printf("info -- creating role: %s\n", dbUser)
		rol := fmt.Sprintf("create role %s;", dbUser)
	    tag, err := dbConn.Exec(bctx, rol)
		if err != nil {return fmt.Errorf(" create role %s failed: %v\n", dbUser, err)}
	    if dbg {fmt.Printf("info -- tag: %s\n", tag.String())}

	} else {
		fmt.Printf("info -- role: %s already exists\n", dbUser)
	}

	// create database
	if dbCnt == 0 {
		fmt.Printf("info -- creating db: %s\n", dbNam)
		dbCre := fmt.Sprintf("create database %s;", dbNam)
	    tag, err := dbConn.Exec(bctx, dbCre)
		if err != nil {return fmt.Errorf(" create db %s failed: %v\n", dbNam, err)}
	    if dbg {fmt.Printf("info -- tag: %s\n", tag.String())}

	} else {
		fmt.Printf("info -- db: %s already exists\n", dbNam)
	}
	return nil
}


func (db *dbObj) initdbconn() (err error) {

	dbg := db.dbg
	dbnam, ok := db.sit["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)
	if len(dbNam)< 3 {return fmt.Errorf("db name too short: %s", len(dbNam))}
	if dbg {fmt.Printf("info -- db name: %s\n", dbNam)}

	bctx := context.Background()
	db.dbctx = bctx

    dbConnStr := fmt.Sprintf("host=/var/run/postgresql user=azuldbadmin dbname=%s", dbnam)
    dbConn, err := pgx.Connect(bctx, dbConnStr)
    if err != nil {return fmt.Errorf("error -- Unable to connect to database %s: %v\n", err)}
//    defer dbConn.Close(bctx)
	db.dbConn = dbConn

	if dbg {fmt.Printf("info -- conn established!\n")}

	return nil
}

func (db *dbObj) buildGoCode(tables []table) (err error) {


	dbg := db.dbg
	goBuf := new(bytes.Buffer)
	goBuf.Grow(4096)

	dbnam, ok := db.sit["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)

	goFilnam :=  "dbgo/azul_" + dbNam + ".go"
    goFil, err := os.Create(goFilnam)
    if err != nil {return fmt.Errorf("creating go file: %v!\n", err)}
    defer goFil.Close()


	top := fmt.Sprintf("// %s\npackage azul_%s\n", db.base, dbNam)
	goBuf.WriteString(top)

	imp := `
import (
//    "os"
    "fmt"
	"time"
	"context"

//    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

`
	goBuf.WriteString(imp)

	tbl := `
type table struct {
    name string
    fields []string
	fldtyp []string
}

`
	goBuf.WriteString(tbl)

	goBuf.WriteString("\n// ****** Tables ********\n")

	for i:=0; i< len(tables); i++ {

		tbl := tables[i]
		tblStr := fmt.Sprintf("type %s_tbl struct {\n", tbl.name)
		goBuf.WriteString(tblStr)
		for ifld:=0; ifld< len(tbl.fields); ifld++ {

			fld:=tbl.fields[ifld]

			fldprop := ""
			prop, _,_ := strings.Cut(fld.prop[0],"(")
			if dbg {fmt.Printf("info dbg -- table: %s field: %s prop: %s %s\n",tbl.name, fld.name, fld.prop[0], prop)}
			switch prop {
			case "serial", "int":
				fldprop="int"
			case "text","varchar":
				fldprop="string"
			case "date":
				fldprop="time.Time"
			default:
				fldprop="unknown"
			}
			fldStr := fmt.Sprintf("  %s %s `db:\"%s\"`\n", capitalize(fld.name), fldprop, fld.name)
			goBuf.WriteString(fldStr)

//			for prop :=0; prop<len(tables[i].fields[fld].prop); prop++ {
//				fmt.Printf("    Prop %d: %s\n", prop+1, tables[i].fields[fld].prop[prop])
//			}
		}
		goBuf.WriteString("}\n\n")
	}
//	fmt.Println("**** End Tables ******")


	goBuf.WriteString("func DbInit() (dbPool *pgxpool.Pool, err error) {\n")

	dbconStr := fmt.Sprintf("host=/var/run/postgresql user=azuldbadmin dbname=%s", dbNam)

	conStr := fmt.Sprintf("\n    dbConnStr := \"%s\"\n", dbconStr)

	goBuf.WriteString(conStr)

    init := `
    dbPool, err = pgxpool.New(context.Background(), dbConnStr)
    if err != nil {return nil, fmt.Errorf("DbInit: Unable to create pool connection: %v\n", err)}
    return dbPool, nil
}
`
    goBuf.WriteString(init)

	// add data to table
	addfun := "\nfunc add(tbl table, json string)(err error) {\n"
	goBuf.WriteString(addfun)

	goBuf.WriteString("\n")

	goBuf.WriteString("  return nil\n")
	goBuf.WriteString("}\n")

	
	goBuf.WriteTo(goFil)
	return nil
}


func (db *dbObj) buildGoTestCode(tables []table) (err error) {


//	dbg := db.dbg
	goBuf := new(bytes.Buffer)
	goBuf.Grow(4096)

	dbnam, ok := db.sit["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)

	goFilnam :=  "dbgo/azul_" + dbNam + "_test.go"
    goFil, err := os.Create(goFilnam)
    if err != nil {return fmt.Errorf("creating go file: %v!\n", err)}
    defer goFil.Close()


	top := fmt.Sprintf("// %s\npackage azul_%s\n", db.base, dbNam)
	goBuf.WriteString(top)

	imp := `
import (
//    "os"
//    "fmt"
	"testing"

//    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
//    "github.com/jackc/pgx/v5/pgxpool"
)

`
	goBuf.WriteString(imp)

	tstAdd := "\nfunc TestAdd(t *testing.T) {\n"
	goBuf.WriteString(tstAdd)

	goBuf.WriteString("\n")
	goBuf.WriteString("}\n")

	goBuf.WriteTo(goFil)
	return nil
}

func (db *dbObj) ParseTables() (tables []table, err error){

//	dbg := db.dbg
	mp := db.sit

	for k,v := range mp {
		typ := reflect.TypeOf(v)
		nam := typ.Name()
		knd:= typ.Kind()
		fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)
		//test for map
		switch knd {
		case 21: 
			// table
			tbl := table{name: k}
			mps := v.(map[string]any)
			for ks, vs := range mps {
				typs := reflect.TypeOf(vs)
				nams := typs.Name()
				knds:= typs.Kind()
				fmt.Printf("  ks: %s %s nam: %s kind: %d vs: %v\n", ks, knds.String(), nams, knds, vs)
				switch knds {

				case 23:
					fld := field{name:ks}
					ar := vs.([]any)
					fmt.Printf("  %v: len: %d\n", ar, len(ar))
					elChar := make([]string,len(ar))
					for ak, av := range ar {
            			switch vv := av.(type) {
            			case string:
                			fmt.Printf("    %v: is string - %q\n", ak, vv)
							elChar[ak] = vv
            			case int:
                			fmt.Printf("    %v: is int - %q\n", ak, vv)
            			default:
                			fmt.Printf("    %v unknown type", ak)
//                WTHisThisJSON(v)
            			}
        			}
					fld.prop = elChar
					tbl.fields = append(tbl.fields, fld)

				default:
					return tables, fmt.Errorf("table: %s key: %s invalid kind %d in tables!",k,  ks, knds)
				}
			}
			tables = append(tables, tbl)

		default:

		}
	}
	return tables, nil
}

func PrintTables (tables []table) {

	fmt.Println("****** Tables ********")
	for i:=0; i< len(tables); i++ {
		fmt.Printf("Table %d: %s\n", i+1, tables[i].name)
		for fld:=0; fld< len(tables[i].fields); fld++ {
			fmt.Printf("  field %d: %s\n",fld+1, tables[i].fields[fld].name)
			for prop :=0; prop<len(tables[i].fields[fld].prop); prop++ {
				fmt.Printf("    Prop %d: %s\n", prop+1, tables[i].fields[fld].prop[prop])
			}
		}
	}
	fmt.Println("**** End Tables ******")
}

func PrintMap(mp map[string]any) {

	fmt.Printf("******** yaml map *****\n%v\n*** detail ****\n",mp)

	for k,v := range mp {
		typ := reflect.TypeOf(v)
		nam := typ.Name()
		knd:= typ.Kind()
		fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)
		//test for map
		switch knd {
		case 21: 
			mps := v.(map[string]any)
			for ks, vs := range mps {
				typs := reflect.TypeOf(vs)
				nams := typs.Name()
				knds:= typs.Kind()
				fmt.Printf("  ks: %s %s nam: %s kind: %d vs: %v\n", ks, knds.String(), nams, knds, vs)
				switch knds {
				case 21:
					mp3 := vs.(map[string]any)
					for k3, v3 := range mp3 {
						typ3 := reflect.TypeOf(v3)
						nam3 := typ3.Name()
						knd3:= typ3.Kind()
						fmt.Printf("     k3: %s nam: %s kind: %d v3: %v\n", k3, nam3, knd3, v3)
					}
				case 23:
					ar := vs.([]any)
					fmt.Printf("  %v: len: %d\n", ar, len(ar))
					elChar := make([]string,len(ar))
					for ak, av := range ar {
            			switch vv := av.(type) {
            			case string:
                			fmt.Printf("    %v: is string - %q\n", ak, vv)
							elChar[ak] = vv
            			case int:
                			fmt.Printf("    %v: is int - %q\n", ak, vv)
            			default:
                			fmt.Printf("    %v unknown type", ak)
//                WTHisThisJSON(v)
            			}
        			}
					for i := 0; i < len(ar); i++ {
						fmt.Printf("    --%d: %s\n",i, elChar[i])
					}
				default:
				}
			}
		// array
		case 23:
			ar := v.([]any)
			fmt.Printf("  array: %d/n", len(ar))

		default:

		}
	}
	fmt.Println("******** end yaml map *****\n")
}
