package dbyaml

import (
	"log"
	"testing"
)

func Test_rdyaml (t *testing.T) {

	yamlFilnam := "testDbAltV2"
	dbinfo, err := RdYaml(yamlFilnam)
	if err !=nil {t.Errorf("rdyaml failed: %v", err)}

	if len(dbinfo.dbTbls) != 2  {t.Errorf("incorrect db tables: %d not 2!", len(dbinfo.dbTbls))}
	if len(dbinfo.layTbls) != 2  {t.Errorf("incorrect layout tables: %d not 2!", len(dbinfo.layTbls))}

	for _,dbTbl := range dbinfo.dbTbls {
		switch dbTbl.name {
		case "Person", "Notes":
		default:
			t.Errorf("incorrect name in dbTabl: %s!", dbTbl.name)
		}
	}

	for _, layTbl := range dbinfo.layTbls {
		switch layTbl.name {
		case "Person", "Notes":
		default:
			t.Errorf("incorrect name in dbTabl: %s!", layTbl.name)
		}
	}

	log.Println("*** rdyaml success ***")
}

func Test_DbTest(t *testing.T) {

    yamlFilnam := "testDbAltV2"
    db, err := RdYaml(yamlFilnam)
    if err !=nil {t.Errorf("rdyaml failed: %v", err)}

	err = DbTest(&db)
    if err !=nil {
		t.Errorf("DbTest failed: %v", err)
	} else {
		log.Printf("*** DbTest success ***\n")
	}
}

func Test_DbTest2(t *testing.T) {

    yamlFilnam := "testDbAltV2"
    db, err := RdYaml(yamlFilnam)
    if err !=nil {t.Errorf("rdyaml failed: %v", err)}

	err = DbTest(&db)
    if err !=nil {
		t.Errorf("DbTest failed: %v", err)
	} else {
		log.Printf("*** DbTest success ***\n")
	}

	err = db.TstTbls()
    if err !=nil {t.Errorf("TstTbls failed: %v", err)}

	err = db.RmTbls()
    if err !=nil {t.Errorf("RmTbls failed: %v", err)}
}

func Test_BldTbls(t *testing.T) {

    yamlFilnam := "testDbAltV2"
    db, err := RdYaml(yamlFilnam)
    if err !=nil {t.Errorf("rdyaml failed: %v", err)}

	err = DbTest(&db)
    if err !=nil {t.Errorf("DbInit failed: %v", err)}

	err = db.BldTbls()
    if err !=nil {t.Errorf("BldTbls failed: %v", err)}

}
