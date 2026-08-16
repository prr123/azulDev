package dbyaml

// change db

import (
    "os"
    "fmt"
//	"log"
//    "time"
    "context"
    "strings"
//  "strconv"

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
    fldList []string
    propList []string
}

type layTable struct {
	name string
	cmds []string
	layCmdList []layCmd
}

type layCmd struct {
	nrows int
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
	for nam, v := range dblMap {
//		fmt.Printf("name: %s, v: %v\n", nam, v)
		fldMap := v.(map[string]any)
		flds := make([]string,0, 24)
		props := make([]string, 0, 24)
		cnt:=0
		for fk, fv := range fldMap {
			cnt++
//			fmt.Printf("  --%d: fk: %s fv: %v\n", cnt, fk, fv)
			xval := fv.([]any)
//			fmt.Printf("  --%d: fk: %s xval: %v\n", cnt, fk, xval)
			flds = append(flds, fk)
			props = append(props, xval[0].(string))
/*
			for i:=0; i< len(xval); i++ {
				str:= xval[i].(string)
				fmt.Printf(" %d: %s,", i, str)
			}
			fmt.Println()
*/
		}
		dbTbl := dbTable{name: nam, fldList: flds, propList: props}
		dbtab = append(dbtab, dbTbl)
	}
	db.dbTbls = dbtab

	layTab := make([]layTable,0,24)
	for nam, v := range layMap {
//		fmt.Printf("name: %s, v: %v\n", nam, v)
		cmdMap := v.(map[string]any)
		cmdList := make([]string,0, 24)
		layCmdValList := make([]layCmd,0, 24)
		cnt:=0
		for cmd, cv := range cmdMap {
			cnt++
//			fmt.Printf("  --%d: cmd: %s cv: %v\n", cnt, cmd, cv)
			cmdList = append(cmdList,cmd)

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
			layCmdVal := layCmd{nrows: len(layCmdValMap), rows: rowDescList}
			layCmdValList = append(layCmdValList, layCmdVal)
		}
//		dbTbl := dbTable{name: nam, fields: flds}
		layTbl :=  layTable{name:nam, cmds: cmdList, layCmdList: layCmdValList}
		layTab = append(layTab, layTbl)
	}
	db.dbTbls = dbtab
	db.layTbls = layTab

//	fmt.Printf("dblMap: %d loutMap: %d\n", len(dblMap), len(layMap))
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
	for cnt, dbTbl := range db.dbTbls {
		fmt.Printf("  Table %d: %s\n", cnt, dbTbl.name)
		for ifld, fld := range dbTbl.fldList {
			fmt.Printf("    field %2d: %-10s %-20s\n", ifld, fld, dbTbl.propList[ifld])
		}
	}

	fmt.Println("***** layout Tables *****")
	for cnt, layTbl := range db.layTbls {
		fmt.Printf("  Table %d: %s\n", cnt, layTbl.name)
		for icmd, cmd := range layTbl.cmds {
			fmt.Printf("    cmd %d: %-10s\n", icmd, cmd)
			laycmd := layTbl.layCmdList[icmd]
			for ir, row := range laycmd.rows {
				fmt.Printf("      %d: %s\n", ir, row.rnam)
				for ifd, fld:= range row.fields {
					fmt.Printf("        fld %d: %s\n", ifd, fld)
				}
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
				if strings.ToLower(fld) == columnName {
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

//    var q strings.Builder
//    q.Grow(1024)

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
		for i:=0; i<len(tbl.fldList); i++ {
			q.WriteString(tbl.fldList[i])
			q.WriteString(" ")
			q.WriteString(tbl.propList[i])
			if i<len(tbl.fldList)-1 {q.WriteString(",")}
		}
		q.WriteString(" );")
//		if db.Dbg {fmt.Printf("q: %s\n", q.String())}
        _, err := dbConn.Exec(db.dbInfo.dbctx, q.String())
        if err != nil {return fmt.Errorf(" create table %s failed, query: >%s<: %v\n", tbl.name, q.String(), err)}
	}

	return nil
}
