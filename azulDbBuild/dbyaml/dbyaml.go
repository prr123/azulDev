package dbyaml

// change db
// change to V3

import (
    "os"
    "fmt"
//	"log"
//    "time"
    "context"
    "strings"
	"strconv"

	"github.com/goccy/go-yaml"
//    "github.com/goccy/go-json"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"
)

type dbData struct {
	dbInfo db
	dbTbls []dbTable
	layTbls []layTable
	Dbg bool
}

type db struct {
	DB string
	User string
	dbConn *pgx.Conn
	dbPool *pgxpool.Pool
	dbctx context.Context
}

type dbTable struct {
    name string
    fldList []fld
}

type fld struct {
	fldnam string
	fldtyp string
	fldattList []string
}

type layTable struct {
	name string
	layCmdList []layCmd
}

type layCmd struct {
	nrows int
	cmd string
	rows []rowDesc
}

type rowDesc struct {
	numFlds int
	rnam string
	fields []string
}


func RdYaml (yFil string) (db dbData, err error) {


	yamlFilDir := "/home/peter/go/src/azulDev/azulDbBuild/yaml/"
    yamlFilnam := yamlFilDir + yFil + ".yaml"
    yamlB, err := os.ReadFile(yamlFilnam)
    if err != nil {return db, fmt.Errorf("cannot read yaml file: %v\n", err)}

    rawMap := map[string]any{}
    err = yaml.Unmarshal(yamlB, &rawMap)
    if err != nil {return db, fmt.Errorf("cannot decode yaml file: %v\n", err)}

	// db
	dbnam, ok := rawMap["db"]
	if !ok {return db, fmt.Errorf("no db field!")}
	db.dbInfo.DB = dbnam.(string)

	dbuser, ok := rawMap["dbuser"]
	if !ok {return db, fmt.Errorf("no dbuser field!")}
	db.dbInfo.User = dbuser.(string)

	// get table names
	dblMapRaw, ok := rawMap["dbl"]
	if !ok {return db, fmt.Errorf("no dbl field!")}
	dblMap := dblMapRaw.(map[string]any)

	layoutMapRaw, ok := rawMap["layout"]
	if !ok {return db, fmt.Errorf("no layout field!")}
	layMap := layoutMapRaw.(map[string]any)

	dbtab := make([]dbTable, 0, 24)
	for tblNam, v := range dblMap {
//		fmt.Printf("name: %s, v: %v\n", nam, v)
		fldMap := v.(map[string]any)
		flds := make([]fld, 0, 24)
		for fnam, fldv := range fldMap {
			fldEl := fld {fldnam: fnam}
			fldvMap := fldv.(map[string]any)
//			fmt.Printf("  -- fnam: %s fldv: %v\n", fnam, fldv)
			ftyp, ok := fldvMap["typ"]
			if !ok {return db, fmt.Errorf("dbTable %s no typ defined for el %s", tblNam, fnam)}
			fldEl.fldtyp = ftyp.(string)
			fldatts := make([]string, 0, 24)
			fattsv, ok := fldvMap["att"]
			if ok {
//				fmt.Printf("    - fldatts: %v\n", fattsv)
				fattsRaw := fattsv.([]any)
				for _, fldatt := range fattsRaw {
					fldattStr := fldatt.(string)
					fldatts = append(fldatts, fldattStr)
				}
			}
			fldEl.fldattList = fldatts
			flds = append(flds, fldEl)
		}
		dbTbl := dbTable{name: tblNam, fldList: flds}
		dbtab = append(dbtab, dbTbl)
	}
	db.dbTbls = dbtab

	layTab := make([]layTable,0,24)
	for nam, v := range layMap {
//		fmt.Printf("name: %s, v: %v\n", nam, v)
		cmdMap := v.(map[string]any)
		layCmdValList := make([]layCmd,0, 24)
		cnt:=0
		for cmdStr, cv := range cmdMap {
			cnt++
//			fmt.Printf("  --%d: cmd: %s cv: %v\n", cnt, cmd, cv)
			layCmdValMap := cv.(map[string]any)
//			fmt.Printf("  --%d: cmd: %s cv val: %v\n", cnt, cmd, layCmdValMap)

			rowDescList := make([]rowDesc, 0, 24)
			for row, vrow := range layCmdValMap {
//				fmt.Printf("    -- %s: %v\n", row, vrow)
				vlist := vrow.([]any)
				fmt.Printf("      ")
				rowList := make([]string, 0, 24)
				for i:=0; i< len(vlist); i++ {
					str:= vlist[i].(string)
//					fmt.Printf("<row %d: %s>", i, str)
					rowList = append(rowList, str)
				}
//				fmt.Println()
				rowDescVal := rowDesc{ numFlds: len(vlist), rnam: row, fields:rowList}
				rowDescList = append(rowDescList, rowDescVal)
			}
			layCmdVal := layCmd{nrows: len(layCmdValMap), cmd: cmdStr , rows: rowDescList}
			layCmdValList = append(layCmdValList, layCmdVal)
		}
//		dbTbl := dbTable{name: nam, fields: flds}
		layTbl :=  layTable{name:nam, layCmdList: layCmdValList}
		layTab = append(layTab, layTbl)
	}
	db.dbTbls = dbtab
	db.layTbls = layTab

//	fmt.Printf("dblMap: %d loutMap: %d\n", len(dblMap), len(layMap))
	PrintDb(db)
	return db, nil
}

func PrintDb(db dbData) {

	fmt.Println()
	fmt.Println("***** database *****")
	fmt.Printf("  db:   %s\n", db.dbInfo.DB)
	fmt.Printf("  User: %s\n", db.dbInfo.User)
	fmt.Println("********************")
	fmt.Println("***** db Tables *****")
	fmt.Println("*********************")
	for _, dbTbl := range db.dbTbls {
		fmt.Printf("  Table: %s\n", dbTbl.name)
		for _, fld := range dbTbl.fldList {
			fmt.Printf("     fld: %-10s typ: %-15s att: ", fld.fldnam, fld.fldtyp)
			for atNum, fatt :=range fld.fldattList {
				fmt.Printf("%s", fatt)
				if atNum < len(fld.fldattList) -1 {fmt.Printf(",")}
			}
			fmt.Printf("\n")
		}
	}

	fmt.Println("***** layout Tables *****")
	for _, layTbl := range db.layTbls {
		fmt.Printf("  Layout Table: %s\n", layTbl.name)
		for _, lcmd := range layTbl.layCmdList {
			fmt.Printf("    %-10s\n", lcmd.cmd)
			for _, row := range lcmd.rows {
				fmt.Printf("     %-5s: ", row.rnam)
				for _, fld:= range row.fields {
					fmt.Printf(" %s,", fld)
				}
				fmt.Printf("\n")
			}
			fmt.Println()
		}
	}

	fmt.Println("*************************")
}


func DbTest(db *dbData) (err error) {


	db.Dbg = true
	bctx :=context.Background()
	db.dbInfo.dbctx = bctx

    dbConnStr := fmt.Sprintf("host=/var/run/postgresql user=azuldbadmin dbname=%s", db.dbInfo.DB)
    dbConn, err := pgx.Connect(bctx, dbConnStr)
    if err != nil {return fmt.Errorf("error -- Unable to connect to database %s: %v\n", db.dbInfo.DB, err)}
    db.dbInfo.dbConn = dbConn
//	defer db.dbInfo.dbConn.Close(bctx)

	return nil
}

func (db *dbData)TstTbls() (err error) {
    var tblExists bool
    var q strings.Builder
    q.Grow(1024)

	dbConn := db.dbInfo.dbConn
	if dbConn == nil {return fmt.Errorf("no dbConn!")}

	bctx := db.dbInfo.dbctx

	// check wehter db tables exist
	for _, tbl := range db.dbTbls {
		q.Reset()
        // check whether table exists
        query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE  table_schema = 'public' AND table_name   = $1);`

        err := dbConn.QueryRow(bctx, query, strings.ToLower(tbl.name)).Scan(&tblExists)
        if err != nil {return fmt.Errorf("table %s query: %v", tbl.name, err)}
        if !tblExists {return fmt.Errorf("table %s not in db", tbl.name)}


		// check columns of table
		q.WriteString("SELECT column_name, data_type, character_maximum_length AS max_length ")
		q.WriteString("FROM information_schema.columns WHERE table_schema = 'public' and table_name=$1;")
		query = q.String()
        rows, err := dbConn.Query(bctx, query, strings.ToLower(tbl.name))
        if err != nil {return fmt.Errorf("table %s col query %v", tbl.name, err)}
		defer rows.Close()

		fmt.Printf("\n*** table: %s\n", tbl.name)
		for rows.Next() {
			var columnName string
			var dataType string
			var mlen pgtype.Int4

			err := rows.Scan(&columnName, &dataType, &mlen)
			if err != nil {return fmt.Errorf("Row scan failed: %v\n", err)}

			if mlen.Valid {
				fmt.Printf("Column: %-20s | Type: %s | %d \n", columnName, dataType, mlen.Int32)
			} else {
				fmt.Printf("Column: %-20s | Type: %s \n", columnName, dataType)
			}

			colFound := false
			for _, fld := range tbl.fldList {
				if strings.ToLower(fld.fldnam) == columnName {
					colFound = true
					break
				}
			}
			if !colFound {return fmt.Errorf("table %s col %s not found!", tbl.name, columnName)}

		}

		if rows.Err() != nil {fmt.Errorf("Rows error: %v\n", rows.Err())}

	}

	// role
	roleCnt := -1
    roleQuery:= fmt.Sprintf("select count(*) FROM pg_roles WHERE rolname = '%s';", db.dbInfo.User)
    err = dbConn.QueryRow(bctx, roleQuery).Scan(&roleCnt)
    if err != nil {return fmt.Errorf(" db roleQuery %s failed: %v\n", roleQuery, err)}
//    if dbg {fmt.Printf("info -- role count: %d\n", roleCnt)}

	return nil
}

func DbInit(db *dbData) (err error) {

	db.Dbg = true
	bctx :=context.Background()
	db.dbInfo.dbctx = bctx


    dbConnStr := fmt.Sprintf("host=/var/run/postgresql user=%s dbname=%s", db.dbInfo.User, db.dbInfo.DB)
	dbPool, err := pgxpool.New(bctx, dbConnStr)
    if err != nil {return fmt.Errorf("error -- Unable to connect to database %s: %v\n", db.dbInfo.DB, err)}
    db.dbInfo.dbPool = dbPool
	defer db.dbInfo.dbPool.Close()

	err = dbPool.Ping(bctx)
	if err != nil {return fmt.Errorf("Unable to ping database: %v\n", err)}

	return nil
}

func (db *dbData) RmTbls() (err error) {

    var q strings.Builder
    q.Grow(1024)

//	query := "drop table if exists $1"

	dbConn := db.dbInfo.dbConn
	if dbConn == nil {return fmt.Errorf("no dbConn")}

	for _, tbl := range db.dbTbls {
		q.Reset()
		q.WriteString("drop table if exists ")
		q.WriteString(strings.ToLower(tbl.name))
		q.WriteString(";")
		_, err = dbConn.Exec(db.dbInfo.dbctx, q.String())
		if err != nil {return fmt.Errorf("drop tables %s: %v", q.String(), err)}
	}
	return nil
}

func (db *dbData) BldTbls() (err error) {

    var q strings.Builder
    q.Grow(1024)

	dbConn := db.dbInfo.dbConn
	if dbConn == nil {return fmt.Errorf("no dbConn")}


	for _, tbl := range db.dbTbls {
		q.Reset()
		q.WriteString("create table if not exists ")
		q.WriteString(strings.ToLower(tbl.name))
		q.WriteString(" (")
		// fields
		for ifld, fld := range tbl.fldList {
			q.WriteString(fld.fldnam)
			q.WriteString(" ")
			q.WriteString(fld.fldtyp)
			for _, attStr := range fld.fldattList {
				q.WriteString(" ")
				q.WriteString(attStr)
			}
			if ifld<len(tbl.fldList)-1 {q.WriteString(",")}
		}


		q.WriteString(" );")
//		if db.Dbg {fmt.Printf("q: %s\n", q.String())}
        _, err := dbConn.Exec(db.dbInfo.dbctx, q.String())
        if err != nil {return fmt.Errorf(" create table %s failed, query: >%s<: %v\n", tbl.name, q.String(), err)}
	}

	return nil
}

func (db *dbData) BldGoCode() (err error) {

    var q strings.Builder
    q.Grow(1024)

	dbConn := db.dbInfo.dbPool
	if dbConn == nil {return fmt.Errorf("no dbPool")}

	// open file
	q.WriteString("../dbgo/")
	q.WriteString(db.dbInfo.DB)
	q.WriteString(".go")
	goFilnam := q.String()
//	fmt.Printf("go Fil: %s\n", goFilnam)
	q.Reset()

	goFil, err := os.Create(goFilnam)
	if err != nil {return fmt.Errorf("cannot open go file %s: %v", goFilnam, err)}
	defer goFil.Close()

	goFil.WriteString("// golang  libary for db handler\n")
	goFil.WriteString("\n")
	goFil.WriteString("package dbgo\n")
	goFil.WriteString("\n")

    imp := `
import (
//    "os"
    "fmt"
//    "time"
    "context"
//    "strings"
//  "strconv"

    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

`
    goFil.WriteString(imp)
	goFil.WriteString("\n")

	//type
typDecl := `
type dbgoObj struct {
	dbPool *pgxpool.Pool
	Dbg	bool
	dbctx context.Context
	tbls map[string][]string
}
`

    goFil.WriteString(typDecl)
	goFil.WriteString("\n")

	dbStr := "const dbnam = \"" + db.dbInfo.DB + "\"\n"
	goFil.WriteString(dbStr)
	dbStr = "const dbuser = \"" + db.dbInfo.User + "\"\n"
	goFil.WriteString(dbStr)
	goFil.WriteString("\n")

	// write dbinit
	dbInitStr := `
func DbInit()(dbp *dbgoObj, err error) {

	var db dbgoObj
	db.Dbg = true
	bctx :=context.Background()
	db.dbctx = bctx


    dbConnStr := fmt.Sprintf("host=/var/run/postgresql user=%s dbname=%s", dbuser, dbnam)
	dbPool, err := pgxpool.New(bctx, dbConnStr)
    if err != nil {return nil, fmt.Errorf("error -- Unable to connect to database %s: %v\n", dbnam, err)}
    db.dbPool = dbPool
//	defer db.dbPool.Close()
`
    goFil.WriteString(dbInitStr)
	goFil.WriteString("\n")

	q.WriteString("    db.tbls = make(map[string][]string, ")
	numtbls := strconv.Itoa(len(db.dbTbls))
	q.WriteString(numtbls)
	q.WriteString(")\n")
	goFil.WriteString(q.String())
	q.Reset()
	goFil.WriteString("\n")

	// init
	for _, tbl := range db.dbTbls {
		q.WriteString("    db.tbls[\"")
		q.WriteString(tbl.name)
		q.WriteString("\"] = []string{")
		for fcnt, fld := range tbl.fldList {
			q.WriteString("\"")
			q.WriteString(fld.fldnam)
			q.WriteString("\"")
			if fcnt < len(tbl.fldList) -1 {q.WriteString(", ")}
		}
		q.WriteString("}\n")
		goFil.WriteString(q.String())
		q.Reset()
	}
	goFil.WriteString("\n")

	goFil.WriteString("    return &db, nil\n}\n\n")


dbPoolTstStr := `
func (db *dbgoObj) DbPoolCheck()(err error) {

	dbPool := db.dbPool
	err = dbPool.Ping(db.dbctx)
	if err != nil {return fmt.Errorf("Unable to ping database: %v\n", err)}

	return nil
}
`
    goFil.WriteString(dbPoolTstStr)
	goFil.WriteString("\n")

cmdParStr := `
func (db *dbgoObj) DbCmdParse(clientCmdStr string)(err error) {

	cmdList := []string{"add","upd","sel","del"}

	cmdMap := make(map[string]string, 24)

	err = json.Unmarshal([]byte(clientCmdStr), &cmdMap)
	if err != nil {return fmt.Errorf("Unmarshal: %v", err)}

	tblnam, ok := cmdMap["tbl"]
	if !ok {return fmt.Errorf("no tbl from Client")}

	_, tok := db.tbls[tblnam]
	if !tok {return fmt.Errorf("invalid tablename: %s", tblnam)}

	cmd, ok := cmdMap["cmd"]
	if !ok {return fmt.Errorf("no cmd from Client")}

	found := false
	for _, cmdl := range cmdList {
		if cmd == cmdl {
			found = true
			break
		}
	}

	if !found {return fmt.Errorf("illegal cmd: %s", cmd)}

	return nil
}
`
    goFil.WriteString(cmdParStr)
	goFil.WriteString("\n")

/*
	for _, tbl := range db.layTbls {
		q.Reset()
		tblnam:= strings.ToLower(tbl.name)
		//get cmd list

		//build add


		// build update

		//build delete


		//build select

		q.WriteString("select ")
		// col nams
		q.WriteString("from ")
		q.WriteString(tblnam)


	}
*/
	return nil
}

func (db *dbData) BldGoTestCode() (err error) {

    var q strings.Builder
    q.Grow(1024)

	dbConn := db.dbInfo.dbPool
	if dbConn == nil {return fmt.Errorf("no dbPool")}

	// open file
	q.WriteString("../dbgo/")
	q.WriteString(db.dbInfo.DB)
	q.WriteString("_test.go")
	goFilnam := q.String()
	fmt.Printf("go Fil: %s\n", goFilnam)
	goFil, err := os.Create(goFilnam)
	if err != nil {return fmt.Errorf("cannot open go file %s: %v", goFilnam, err)}
	defer goFil.Close()

	goFil.WriteString("// test for golang libary for db handler\n")
	goFil.WriteString("\n")
	goFil.WriteString("package dbgo\n")
	goFil.WriteString("\n")

   imp := `
import (
	"testing"
)

`
    goFil.WriteString(imp)
    goFil.WriteString("\n")

	goFil.WriteString("func TestDBInit(t *testing.T) {\n")
    goFil.WriteString("\n")
	goFil.WriteString("  db, err := DbInit()\n")
	goFil.WriteString("  if err !=nil {t.Errorf(\"error DBINIT: %v\", err)}\n")
	goFil.WriteString("  if db.dbPool ==nil {t.Errorf(\"error DBPool is nil!\")}\n")
    goFil.WriteString("\n")
	goFil.WriteString("  err = db.DbPoolCheck()\n")
	goFil.WriteString("  if err !=nil {t.Errorf(\"error DbPoolCheck: %v\", err)}\n")
	goFil.WriteString("}\n\n")

	// check BldGoCode
	goFil.WriteString("func TestDBCmdParse(t *testing.T) {\n")

	goFil.WriteString("  db, err := DbInit()\n")
	goFil.WriteString("  if err !=nil {t.Errorf(\"error DbInit: %v\", err)}\n")
    goFil.WriteString("\n")
	jsStr := `{"tbl":"Person", "cmd":"sel"}`

	goFil.WriteString("  jsonCmdStr := `" + jsStr + "`\n")
//    goFil.WriteString("\n")
	goFil.WriteString("  err = db.DbCmdParse(jsonCmdStr)\n")
	goFil.WriteString("  if err !=nil {t.Errorf(\"error DBCmdParse: %v\", err)}\n")
    goFil.WriteString("\n")

	goFil.WriteString("}\n\n")

	return nil
}


func (db *dbData) BldJsCode() (err error) {

    var q strings.Builder
    q.Grow(1024)

	dbConn := db.dbInfo.dbPool
	if dbConn == nil {return fmt.Errorf("no dbPool")}

	// open file
	q.WriteString("../dbjs/")
	q.WriteString(db.dbInfo.DB)
	q.WriteString(".js")
	jsFilnam := q.String()
	fmt.Printf("js Fil: %s\n", jsFilnam)
	jsFil, err := os.Create(jsFilnam)
	if err != nil {return fmt.Errorf("cannot open go file %s: %v", jsFilnam, err)}
	defer jsFil.Close()
/*
	for _, tbl := range db.dbTbls {
		q.Reset()
		tblnam:= strings.ToLower(tbl.name)

	}
*/
	return nil
}

