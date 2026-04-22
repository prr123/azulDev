// testdbAlt
package azul_testdb

import (
//    "os"
    "fmt"
	"strings"
	"context"
	"testing"


    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
//    "github.com/jackc/pgx/v5/pgxpool"
)


func TestInitDb(t *testing.T) {

  dbPool,err := DbInit()
  if err != nil {t.Errorf("error -- could not connect to pool!")}
  defer dbPool.Close()
  //check tables
  tblq := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE';
	`

	dbctx := context.Background()
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
}

func TestAddDb(t *testing.T) {

  dbPool,err := DbInit()
  if err != nil {t.Errorf("error -- could not connect to pool!")}
  defer dbPool.Close()
  db := dbObj {dbg:true, dbPool: dbPool, dbctx: context.Background()}
  dbg := db.dbg
  jsonStr := `{"table": "person", "cmd":"add", "first": "peter", "middle": "rich", "last":"Smith", "nie":"1234561A", "email": "peter@test.com"}`
  if dbg {fmt.Printf("info dbg -- json: %s\n",jsonStr)}
    err = db.addDb(jsonStr)
    if err!=nil {t.Errorf("addDb: %v", err) }
  jsonStr = `{"table": "person", "cmd":"upd", "first": "peter2", "last":"Smith2", "id":"2"}`
  if dbg {fmt.Printf("info dbg -- json: %s\n",jsonStr)}
  err = db.addDb(jsonStr)
  if err!=nil {t.Errorf("addDb: %v", err) }
  jsMap := make(map[string]any, 24)
  err1 := json.Unmarshal([]byte(jsonStr), &jsMap)
  if err1 != nil {t.Errorf("Unmarshal %v ", err1)}
}
