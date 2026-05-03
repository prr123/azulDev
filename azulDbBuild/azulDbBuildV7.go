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
// V6 create arrays for tables
//
// V7 change table cmd string
// -- create test json string
// --         js build

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
	ptyp string
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

	err = db.buildAzulCode(tables)
	if err != nil {log.Fatalf("error -- buildGoCode: %v\n", err)}

	err = db.buildAzulTestCode(tables)
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
	"strings"
//	"strconv"

    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

`
	goBuf.WriteString(imp)

	tblStr := `
type table struct {
    name string
    fields []string
	prop []string
}

type dbObj struct {
	tables []table
	dbg bool
	dbPool *pgxpool.Pool
    dbctx context.Context

}
`
	goBuf.WriteString(tblStr)

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
			tables[i].fields[ifld].ptyp = prop

			fldStr := fmt.Sprintf("  %s %s `db:\"%s\"`\n", capitalize(fld.name), fldprop, fld.name)
			goBuf.WriteString(fldStr)

		}
		goBuf.WriteString("}\n\n")
	}
//	fmt.Println("**** End Tables ******")


	goBuf.WriteString("func DbInit(dbg bool) (db *dbObj, err error) {\n")

	dbconStr := fmt.Sprintf("host=/var/run/postgresql user=azuldbadmin dbname=%s", dbNam)

	conStr := fmt.Sprintf("\n    dbConnStr := \"%s\"\n", dbconStr)

	goBuf.WriteString(conStr)

    init := `
	dbctx := context.Background()
    dbPool, err := pgxpool.New(context.Background(), dbConnStr)
    if err != nil {return nil, fmt.Errorf("DbInit: Unable to create pool connection: %v\n", err)}
	dbx := dbObj {dbg: dbg, dbPool: dbPool, dbctx: dbctx}
    return &dbx , nil
}
`
    goBuf.WriteString(init)

	// init tables
	goBuf.WriteString("\nfunc initTbl()(tables []table, err error) {\n")

	tStr:=fmt.Sprintf("  tables = make([]table, %d)\n", len(tables))
	goBuf.WriteString(tStr)

	for itbl:=0; itbl<len(tables); itbl++ {
		tbl:= tables[itbl]
		fldStr:=fmt.Sprintf("  tables[%d].fields = make([]string, %d)\n", itbl,len(tbl.fields))
		goBuf.WriteString(fldStr)
		propStr:=fmt.Sprintf("  tables[%d].prop = make([]string, %d)\n", itbl, len(tbl.fields))
		goBuf.WriteString(propStr)

//		goBuf.WriteString("  for il:=0; il<len(tables); il++ {\n")
		tblStr := fmt.Sprintf("    tables[%d].name=\"%s\"\n",itbl, tbl.name)
		goBuf.WriteString(tblStr)
		fldCnt := -1
		for ifl:=0; ifl<len(tbl.fields); ifl++ {
			fld := tbl.fields[ifl]
			if fld.prop[0] == "serial" {continue}
			fldCnt++
			fldStr := fmt.Sprintf("    tables[%d].fields[%d]=\"%s\"\n", itbl, fldCnt, fld.name)
			goBuf.WriteString(fldStr)
			propStr := fmt.Sprintf("    tables[%d].prop[%d]=\"%s\"\n", itbl, fldCnt, fld.ptyp)
			goBuf.WriteString(propStr)
		}
	}
	goBuf.WriteString("\n")

	goBuf.WriteString("  return tables, nil\n")
	goBuf.WriteString("}\n")



	// add data to table
	goBuf.WriteString("\nfunc (db *dbObj) TblCmd(jsonStr string)(res string, err error) {\n")
	goBuf.WriteString("\n")

	goBuf.WriteString("  var kval, val, upd strings.Builder\n")
	goBuf.WriteString("  kval.Grow(256)\n")
	goBuf.WriteString("  val.Grow(256)\n")
	goBuf.WriteString("  upd.Grow(256)\n")

	goBuf.WriteString("  if db == nil {return res, fmt.Errorf(\"no dbObj found!\")}\n")
	goBuf.WriteString("  dbg:=db.dbg\n")
	goBuf.WriteString("  dbPool:=db.dbPool\n")
	goBuf.WriteString("  dbctx:=db.dbctx\n")
	goBuf.WriteString("  jsMap := make(map[string]any)\n")
	goBuf.WriteString("  err = json.Unmarshal([]byte(jsonStr), &jsMap)\n")
	goBuf.WriteString("  if err != nil {return res, fmt.Errorf(\"Unmarshal json: %v\", err)}\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- jsonMap: %d\\n\", len(jsMap))}\n")
	goBuf.WriteString("  tblNam, ok := jsMap[\"table\"]\n")
	goBuf.WriteString("  if !ok {return res, fmt.Errorf(\"no table name found!\")}\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- table name: %s\\n\", tblNam.(string))}\n")
	goBuf.WriteString("  cmdStr, ok := jsMap[\"cmd\"]\n")
	goBuf.WriteString("  if !ok {return res, fmt.Errorf(\"no cmd found!\")}\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- cmd name: %s\\n\", cmdStr.(string))}\n")

	goBuf.WriteString("  flds, ok := jsMap[\"fields\"].(map[string]any)\n")
	goBuf.WriteString("  if !ok {return res, fmt.Errorf(\"fields not found!\")}\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- fld map name: %v\\n\", flds)}\n")

	goBuf.WriteString("  cond, ok := jsMap[\"cond\"].(map[string]any)\n")
//	goBuf.WriteString("  if !ok {return res, fmt.Errorf(\"condition not found!\")}\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- cond: %v\\n\", cond)}\n")


	goBuf.WriteString("  switch cmdStr {\n")

	goBuf.WriteString("  case \"add\":\n")
	goBuf.WriteString("    kval.Reset()\n")
	goBuf.WriteString("    val.Reset()\n")
	goBuf.WriteString("    for k, v := range flds {\n")
	goBuf.WriteString("      kval.WriteString(\",\"+k)\n")
	goBuf.WriteString("      vstr := fmt.Sprintf(\",'%s'\",v)\n")
	goBuf.WriteString("      val.WriteString(vstr)\n")
	goBuf.WriteString("    }\n")
	goBuf.WriteString("    q:= fmt.Sprintf(\"insert into %s (%s) values (%s) returning id\",tblNam, kval.String()[1:], val.String()[1:])\n")
	goBuf.WriteString("    if dbg {fmt.Printf(\"info dbg -- q: %s\\n\",q)}\n")

	goBuf.WriteString("    var newId int\n")
	goBuf.WriteString("    err = dbPool.QueryRow(dbctx, q).Scan(&newId)\n")
	goBuf.WriteString("    if err != nil {return res, fmt.Errorf(\"add db: %v\", err)}\n")

	goBuf.WriteString("    res = fmt.Sprintf(\"{\\\"add\\\":\\\"%d\\\"}\", newId)\n")
	goBuf.WriteString("    return res, nil\n")
	goBuf.WriteString("\n")

	goBuf.WriteString("  case \"upd\":\n")
	goBuf.WriteString("    if cond == nil {return res, fmt.Errorf(\"no cond!\")}\n")
//	goBuf.WriteString("    kval.Reset()")
	goBuf.WriteString("    upd.Reset()\n")
	goBuf.WriteString("    cndCnt := 0\n")
	goBuf.WriteString("    condStr := \"\"\n")
	goBuf.WriteString("    for ck, cv := range cond {\n")
	goBuf.WriteString("      andStr :=\"\"\n")
	goBuf.WriteString("      if cndCnt > 0 {andStr = \" and \"}\n")
	goBuf.WriteString("      cStr := fmt.Sprintf(\"%s = '%s'\",ck,cv)\n")
	goBuf.WriteString("      if ck == \"id\" { cStr = fmt.Sprintf(\"%s = %s\",ck,cv)}\n")
	goBuf.WriteString("      condStr = condStr + andStr + cStr\n")
	goBuf.WriteString("      cndCnt++\n")
	goBuf.WriteString("    }\n")

	goBuf.WriteString("    for k, v := range flds {\n")
	goBuf.WriteString("      vstr := fmt.Sprintf(\",%s = '%s'\",k, v)\n")
	goBuf.WriteString("      upd.WriteString(vstr)\n")
	goBuf.WriteString("    }\n")
	goBuf.WriteString("    q:= fmt.Sprintf(\"update %s set %s where %s;\",tblNam, upd.String()[1:], condStr)\n")
	goBuf.WriteString("    if dbg {fmt.Printf(\"info dbg -- q: %s\\n\",q)}\n")
	goBuf.WriteString("    cmdTag, err := dbPool.Exec(dbctx, q)\n")
	goBuf.WriteString("    if err != nil {return res, fmt.Errorf(\"upd db: %v\", err)}\n")
	goBuf.WriteString("    if dbg {fmt.Printf(\"info dbg -- rows: %d\\n\", cmdTag.RowsAffected())}\n")

	goBuf.WriteString("    res = fmt.Sprintf(\"{\\\"update\\\":\\\"%d\\\"}\",cmdTag.RowsAffected())\n")
	goBuf.WriteString("    return res, nil\n")
	goBuf.WriteString("\n")

	goBuf.WriteString("  case \"list\":\n")

	goBuf.WriteString("    kval.Reset()\n")
	goBuf.WriteString("    val.Reset()\n")
	goBuf.WriteString("    for k, _ := range flds {\n")
	goBuf.WriteString("      kval.WriteString(\",\"+k)\n")
//	goBuf.WriteString("      vstr := fmt.Sprintf(\",'%s'\",v)\n")
//	goBuf.WriteString("      val.WriteString(vstr)\n")
	goBuf.WriteString("    }\n")
	goBuf.WriteString("    if len(flds) == 0 {kval.WriteString(\"-*\")}\n")

	goBuf.WriteString("    cndCnt := 0\n")
	goBuf.WriteString("    condStr := \"\"\n")
	goBuf.WriteString("    for ck, cv := range cond {\n")
	goBuf.WriteString("      andStr :=\"\"\n")
	goBuf.WriteString("      if cndCnt > 0 {andStr = \" and \"}\n")
	goBuf.WriteString("      cStr := fmt.Sprintf(\"%s = '%s'\",ck,cv)\n")
	goBuf.WriteString("      if ck == \"id\" { cStr = fmt.Sprintf(\"%s = %s\",ck,cv)}\n")
	goBuf.WriteString("      condStr = condStr + andStr + cStr\n")
	goBuf.WriteString("      cndCnt++\n")
	goBuf.WriteString("    }\n")
	goBuf.WriteString("    if cndCnt > 0 {condStr = \" where \" + condStr}\n")
	goBuf.WriteString("\n")
	goBuf.WriteString("    selStr:= fmt.Sprintf(\"SELECT json_agg(row_to_json(t)) FROM (select %s from %s%s) t\",kval.String()[1:], tblNam, condStr)\n")
	goBuf.WriteString("    if dbg {fmt.Printf(\"info dbg -- q: %s\\n\",selStr)}\n")

	goBuf.WriteString("    var jsonOut []byte\n")

	goBuf.WriteString("    err = dbPool.QueryRow(dbctx, selStr).Scan(&jsonOut)\n")

	goBuf.WriteString("    if dbg {fmt.Printf(\"info dbg -- jsonOut: %s\\n\",jsonOut)}\n")

	goBuf.WriteString("    res = string(jsonOut)\n")
	goBuf.WriteString("    return res, nil\n")

	goBuf.WriteString("\n")
	goBuf.WriteString("  default:\n")
	goBuf.WriteString("  return res, fmt.Errorf(\"invalid cmd: %s\", cmdStr)\n")
	goBuf.WriteString("  }\n")


//	goBuf.WriteString("  q:= fmt.Sprintf(\"insert into %s select * from json_to_record('%s') ax x(%s);\", tblNam, njStr, colStr)\n")
//	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- q: %s\\n\",q)}\n")

	goBuf.WriteString("\n")

	

	goBuf.WriteString("  return res, nil\n")
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
    "fmt"
	"strings"
//	"context"
	"testing"


    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
//    "github.com/jackc/pgx/v5/pgxpool"
)

`
	goBuf.WriteString(imp)

	tstInit := "\nfunc TestInitDb(t *testing.T) {\n"
	goBuf.WriteString(tstInit)
	goBuf.WriteString("\n")
	goBuf.WriteString("  db, err := DbInit(true)\n")
	goBuf.WriteString("  if err != nil {t.Errorf(\"error -- could not connect to pool!\")}\n")
	goBuf.WriteString("  defer db.dbPool.Close()\n")
	goBuf.WriteString("  dbPool := db.dbPool\n")

	goBuf.WriteString("  //check tables\n")
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE';
	`
	tblqc := fmt.Sprintf("  tblq := `%s`\n",query)
	goBuf.WriteString(tblqc)

	query2 := `
	dbctx := db.dbctx
	tblnam := make([]string, 0, 24)
	rows, err := dbPool.Query(dbctx, tblq)
	if err != nil {t.Errorf("table query failed: %v\n", err)}
	defer rows.Close()
	tblcnt:=-1
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Errorf("Row scan failed: %v\n", err)
			continue
		}
		tblcnt++
		tblnam = append(tblnam, tableName)
//		fmt.Printf("- %s\n", tableName)
	}

	tables, err := initTbl()
	if err !=nil {t.Errorf("Table init failed: %v\n", err)}

	for i:=0; i< len(tblnam); i++ {
		fmt.Printf("  -%d: %s\n", i+1, tblnam[i])
		match := false
		for j:=0; j<len(tables); j++ {
//			fmt.Printf("table name: %s\n",tables[j].name)
			nam := strings.ToLower(tables[j].name)
			if nam == tblnam[i] {match = true; break;}
		}
		if !match {t.Errorf("no table name match for %s!",tblnam[i])}

	}
`
	goBuf.WriteString(query2)

	colStr := `
	for i:=0; i< len(tblnam); i++ {
		colqStr := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s' and table_schema='public';", tblnam[i])
		crows, err := dbPool.Query(dbctx, colqStr)
		if err != nil {t.Errorf("col name query failed: %v\n", err)}
		defer crows.Close()
		colCnt:=-1
		for crows.Next() {
			colnam := ""
			coltyp := ""
			if err := crows.Scan(&colnam,&coltyp); err != nil {
				t.Errorf("Col row scan failed: %v\n", err)
				continue
			}
			colCnt++
//			tblnam = append(tblnam, tableName)
			fmt.Printf("-%d: %s %s\n", colCnt, colnam, coltyp)
		}
	}
`
	goBuf.WriteString(colStr)


	goBuf.WriteString("}\n")

	tstAdd := "\nfunc TestTblCmd(t *testing.T) {\n"
	goBuf.WriteString(tstAdd)
	goBuf.WriteString("\n")
	goBuf.WriteString("  db, err := DbInit(true)\n")
	goBuf.WriteString("  if err != nil {t.Errorf(\"error -- could not connect to pool!\")}\n")
	goBuf.WriteString("  defer db.dbPool.Close()\n")

//	goBuf.WriteString("  db := dbObj {dbg:true, dbPool: dbPool, dbctx: context.Background()}\n")
	goBuf.WriteString("  dbg := db.dbg\n\n")

	goBuf.WriteString("  // db add\n")
	jsonAdd := `{"table": "person", "cmd":"add", "fields": {"first": "peter", "middle": "rich", "last":"Smith", "nie":"1234561A", "email": "peter@test.com"}}`
	jsonStr := fmt.Sprintf("  jsonStr := `%s`\n", jsonAdd)
	goBuf.WriteString(jsonStr)
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- json: %s\\n\",jsonStr)}\n")

	goBuf.WriteString("  res, err := db.TblCmd(jsonStr)\n")
	goBuf.WriteString("  if err != nil {t.Errorf(\"TblCmd: %v\", err) }\n\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- res add: %s\\n\", res)}\n")

	goBuf.WriteString("  // db update\n")
	jsonUpd := `{"table": "person", "cmd":"upd", "fields":{"first": "peter2", "last":"Smith2"}, "cond": {"id":"2"}}`
	jsonStr = fmt.Sprintf("  jsonStr = `%s`\n", jsonUpd)
	goBuf.WriteString(jsonStr)
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- json: %s\\n\",jsonStr)}\n")

	goBuf.WriteString("  res, err = db.TblCmd(jsonStr)\n")
	goBuf.WriteString("  if err != nil {t.Errorf(\"TblCmd: %v\", err) }\n\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- res upd: %s\\n\", res)}\n")

	goBuf.WriteString("  // db list\n")
	jsonList := `{"table": "person", "cmd":"list", "fields":{"first": "", "last":""}, "cond": {}}`
	jsonStr = fmt.Sprintf("  jsonStr = `%s`\n", jsonList)
	goBuf.WriteString(jsonStr)
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- json: %s\\n\",jsonStr)}\n")

	goBuf.WriteString("  res, err = db.TblCmd(jsonStr)\n")
	goBuf.WriteString("  if err != nil {t.Errorf(\"TblCmd: %v\", err) }\n\n")
	goBuf.WriteString("  if dbg {fmt.Printf(\"info dbg -- res list: %s\\n\", res)}\n")

	goBuf.WriteString("  // test unmarshal\n")
	goBuf.WriteString("  jsMap := make(map[string]any, 24)\n")
	goBuf.WriteString("  err1 := json.Unmarshal([]byte(jsonStr), &jsMap)\n")
	goBuf.WriteString("  if err1 != nil {t.Errorf(\"Unmarshal %v \", err1)}\n\n")

	goBuf.WriteString("}\n")

	goBuf.WriteTo(goFil)
	return nil
}

func (db *dbObj) buildAzulCode(tables []table) (err error) {

//	dbg := db.dbg
	jsBuf := new(bytes.Buffer)
	jsBuf.Grow(4096)

	dbnam, ok := db.sit["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)

	jsFilnam :=  "dbgo/azul_" + dbNam + ".js"
    jsFil, err := os.Create(jsFilnam)
    if err != nil {return fmt.Errorf("creating azul js file: %v!\n", err)}
    defer jsFil.Close()


	top := fmt.Sprintf("// %s %s\n", db.base, dbNam)
	jsBuf.WriteString(top)

	jsBuf.WriteString("\n")
	for i:=0; i< len(tables); i++ {
		tblNam := tables[i].name
		jsBuf.WriteString("// table: " + tblNam + "\n")
		addStr := fmt.Sprintf("let %sAdd = {\n",tblNam)
		jsBuf.WriteString(addStr)
//		jsBuf.WriteString("\n")
		sbut := `
    subButObj: {
        text: 'submit new',
        style: {
            display: 'block',
            textAlign: 'center',
            width: '200px',
            margin: '20px auto',
        },
    },
`
		jsBuf.WriteString(sbut)
		jsBuf.WriteString("\n")
		jsBuf.WriteString("  subFun() {\n")
		jsBuf.WriteString("// submit processing\n")
		jsBuf.WriteString("  },\n")

		jsBuf.WriteString("\n")
		jsBuf.WriteString("  rendSubmit() {\n")
		subCode:=`    const subDiv = document.createElement('div');
    const subBut = new azulButton(this.subButObj);
    this.subButEl = subBut.el;
`
		jsBuf.WriteString(subCode)
		jsBuf.WriteString("    subBut.el.addEventListener('click', function() {")
		sfun := fmt.Sprintf("%sAdd.subFunc(dbData.gridDiv.inpEls);},false);\n", tblNam)
		jsBuf.WriteString(sfun)
		subCode2 :=
`    subDiv.appendChild(subBut.el);
    return subDiv;
`
		jsBuf.WriteString(subCode2)
		jsBuf.WriteString("  },\n")
		jsBuf.WriteString("\n")

		jsBuf.WriteString("  render() {\n")
		jsBuf.WriteString("    const root = document.createElement('div');\n")
		jsBuf.WriteString("// add content\n")
        rcode:= fmt.Sprintf("    const subDiv = %sAdd.rendSubmit();\n", tblNam)
		jsBuf.WriteString(rcode)
		jsBuf.WriteString("    root.appendChild(subDiv);\n")
		jsBuf.WriteString("    return root;\n")
		jsBuf.WriteString("  },\n")

		jsBuf.WriteString("  rendfun() {\n")
		jsBuf.WriteString("    console.log('add click!');\n")
		add2:= fmt.Sprintf("    const ldiv = %sAdd.render();\n    azul.rplDiv(dbMain.dbDat, ldiv);\n", tblNam)
	    jsBuf.WriteString(add2)
		jsBuf.WriteString("  },\n")
		jsBuf.WriteString("};\n\n")

		updStr := fmt.Sprintf("let %sUpd = {\n",tblNam)
		jsBuf.WriteString(updStr)
		jsBuf.WriteString("\n")

		jsBuf.WriteString("  render() {\n")
		jsBuf.WriteString("    const root = document.createElement('div');\n")
		jsBuf.WriteString("// add content\n")
		jsBuf.WriteString("    return root;\n")
		jsBuf.WriteString("  },\n")

		jsBuf.WriteString("  rendfun() {\n")
		jsBuf.WriteString("    console.log('upd click!');\n")
		upd2:= fmt.Sprintf("    const ldiv = %sUpd.render();\n    azul.rplDiv(dbMain.dbDat, ldiv);\n", tblNam)
	    jsBuf.WriteString(upd2)
		jsBuf.WriteString("  },\n")
		jsBuf.WriteString("};\n\n")
		lsStr := fmt.Sprintf("let %sList = {\n",tblNam)
		jsBuf.WriteString(lsStr)
		jsBuf.WriteString("\n")
		jsBuf.WriteString("  render() {\n")
		jsBuf.WriteString("    const root = document.createElement('div');\n")
		jsBuf.WriteString("// add content\n")
		jsBuf.WriteString("    return root;\n")
		jsBuf.WriteString("  },\n")
		jsBuf.WriteString("  rendfun() {\n")
		jsBuf.WriteString("    console.log('list click!');\n")
		ls2:= fmt.Sprintf("    const ldiv = %sList.render();\n    azul.rplDiv(dbMain.dbDat, ldiv);\n", tblNam)
	    jsBuf.WriteString(ls2)
		jsBuf.WriteString("  },\n")
		jsBuf.WriteString("};\n\n")
	}

	jsBuf.WriteTo(jsFil)
	return nil
}

func (db *dbObj) buildAzulTestCode(tables []table) (err error) {

	jsBuf := new(bytes.Buffer)
	jsBuf.Grow(4096)

	dbnam, ok := db.sit["db"]
	if !ok {return fmt.Errorf("no db name found!")}
	dbNam := dbnam.(string)

	jsFilnam :=  "dbgo/azul_" + dbNam + "_test.js"
    jsFil, err := os.Create(jsFilnam)
    if err != nil {return fmt.Errorf("creating azul js test file: %v!\n", err)}
    defer jsFil.Close()

	top := fmt.Sprintf("// %s %s\n", db.base, dbNam)
	jsBuf.WriteString(top)

	jsBuf.WriteTo(jsFil)
	return nil
}


func (db *dbObj) ParseTables() (tables []table, err error){

	dbg := db.dbg
	mp := db.sit

	for k,v := range mp {
		typ := reflect.TypeOf(v)
		nam := typ.Name()
		knd:= typ.Kind()
		if dbg {fmt.Printf("k: %s nam: %s kind: %d v: %v\n", k, nam, knd, v)}
		//test  for array
		switch knd {
		case 21: 
			// table
			tbl := table{name: k}
			mps := v.(map[string]any)
			for ks, vs := range mps {
				typs := reflect.TypeOf(vs)
				nams := typs.Name()
				knds:= typs.Kind()
				if dbg {fmt.Printf("  ks: %s %s nam: %s kind: %d vs: %v\n", ks, knds.String(), nams, knds, vs)}

				switch knds {
				case 23:
					fld := field{name:ks}
					ar := vs.([]any)
					if dbg {fmt.Printf("  %v: len: %d\n", ar, len(ar))}
					elChar := make([]string,len(ar))

/*
					for ak, av := range ar {
            			switch vv := av.(type) {
            			case string:
 	               			if dbg {fmt.Printf("    %v: is string - %q\n", ak, vv)}
							elChar[ak] = vv
            			case int:
                			if dbg {fmt.Printf("    %v: is int - %q\n", ak, vv)}
//							elChar[ak] = vv
            			default:
                			if dbg {fmt.Printf("    %v unknown type", ak)}
							typeName := reflect.TypeOf(av).String()
							return tables, fmt.Errorf("unknown typ in 2nd map: %s", typeName)
            			}
        			}
*/
					for ak, av := range ar {
						elChar[ak] = av.(string)
					}
					// this reduces varchar(num) to varchar
					fld.ptyp, _,_ = strings.Cut(elChar[0],"(")

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
