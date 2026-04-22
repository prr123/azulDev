// testdbAlt
package azul_testdb

import (
//    "os"
    "fmt"
	"time"
	"context"
	"strings"
	"strconv"

    "github.com/goccy/go-json"
//    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)


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

// ****** Tables ********
type Person_tbl struct {
  NIE string `db:"NIE"`
  Email string `db:"email"`
  Id int `db:"id"`
  First string `db:"first"`
  Middle string `db:"middle"`
  Last string `db:"last"`
  Bdate time.Time `db:"bdate"`
}

type Notes_tbl struct {
  Id int `db:"id"`
  Pid int `db:"pid"`
  Date time.Time `db:"date"`
  Content string `db:"content"`
  Edited time.Time `db:"edited"`
}

func DbInit() (dbPool *pgxpool.Pool, err error) {

    dbConnStr := "host=/var/run/postgresql user=azuldbadmin dbname=testdb"

    dbPool, err = pgxpool.New(context.Background(), dbConnStr)
    if err != nil {return nil, fmt.Errorf("DbInit: Unable to create pool connection: %v\n", err)}
    return dbPool, nil
}

func initTbl()(tables []table, err error) {
  tables = make([]table, 2)
  tables[0].fields = make([]string, 7)
  tables[0].prop = make([]string, 7)
    tables[0].name="Person"
    tables[0].fields[0]="NIE"
    tables[0].prop[0]="varchar"
    tables[0].fields[1]="email"
    tables[0].prop[1]="varchar"
    tables[0].fields[2]="first"
    tables[0].prop[2]="varchar"
    tables[0].fields[3]="middle"
    tables[0].prop[3]="varchar"
    tables[0].fields[4]="last"
    tables[0].prop[4]="varchar"
    tables[0].fields[5]="bdate"
    tables[0].prop[5]="date"
  tables[1].fields = make([]string, 5)
  tables[1].prop = make([]string, 5)
    tables[1].name="Notes"
    tables[1].fields[0]="pid"
    tables[1].prop[0]="int"
    tables[1].fields[1]="date"
    tables[1].prop[1]="date"
    tables[1].fields[2]="content"
    tables[1].prop[2]="text"
    tables[1].fields[3]="edited"
    tables[1].prop[3]="date"

  return tables, nil
}

func (db *dbObj) addDb(jsonStr string)(err error) {

  var kval, val, upd strings.Builder
  kval.Grow(256)
  val.Grow(256)
  upd.Grow(256)
  if db == nil {return fmt.Errorf("no dbObj found!")}
  dbg:=db.dbg
  jsMap := make(map[string]string)
  err = json.Unmarshal([]byte(jsonStr), &jsMap)
  if err != nil {return fmt.Errorf("Unmarshal json: %v", err)}
  if dbg {fmt.Printf("info dbg -- jsonMap: %d\n", len(jsMap))}
  tblNam, ok := jsMap["table"]
  if !ok {return fmt.Errorf("no table name found!")}
  if dbg {fmt.Printf("info dbg -- table name: %s\n", tblNam)}
  cmdStr, ok := jsMap["cmd"]
  if !ok {return fmt.Errorf("no cmd found!")}
  if dbg {fmt.Printf("info dbg -- cmd name: %s\n", cmdStr)}
  idx1 := strings.IndexByte(jsonStr, ',') +1
  idx2 := strings.IndexByte(string(jsonStr[idx1:]), ',') +1
  idx := idx1 + idx2
  njStr := "{" + jsonStr[idx:]
  if dbg {fmt.Printf("info dbg -- json: %s\n",njStr)}
  switch cmdStr {
  case "add":
  for k, v := range jsMap {
     if k == "table" {continue}
     if k == "cmd" {continue}
     kval.WriteString(","+k)
     vstr := fmt.Sprintf(",'%s'",v)
     val.WriteString(vstr)
  }
  q:= fmt.Sprintf("insert into %s (%s) values (%s);",tblNam, kval.String()[1:], val.String()[1:])
  if dbg {fmt.Printf("info dbg -- q: %s\n",q)}

  case "upd":
    idStr, ok := jsMap["id"]
    if !ok {return fmt.Errorf("upd no id!")}
    id, err2 := strconv.Atoi(idStr)
    if err2 != nil {return fmt.Errorf("id is not int!")}
    for k, v := range jsMap {
      if k == "table" {continue}
      if k == "cmd" {continue}
      if k == "id" {continue}
      vstr := fmt.Sprintf(",%s = '%s'",k, v)
      upd.WriteString(vstr)
    }
    q:= fmt.Sprintf("update %s set %s where id = %d;",tblNam, upd.String()[1:], id)
    if dbg {fmt.Printf("info dbg -- q: %s\n",q)}

  default:
  return fmt.Errorf("invalid cmd: %s", cmdStr)
  }

  return nil
}
