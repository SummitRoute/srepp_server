////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////



package models

import (
	"database/sql"

	"github.com/coopernurse/gorp"
)

// InitDB initialize the database by creating the needed tables
func InitDB(ConnectionString string) (*gorp.DbMap, error) {
	// Connect to db
	db, err := sql.Open("postgres", ConnectionString)
	if err != nil {
		return nil, err
	}

	// Construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	// Add a table, setting the table
	dbmap.AddTableWithName(SystemSet{}, "systemsets").SetKeys(true, "ID")
	dbmap.AddTableWithName(System{}, "systems").SetKeys(true, "ID")
	dbmap.AddTableWithName(RuleSet{}, "rulesets").SetKeys(true, "ID")
	dbmap.AddTableWithName(Rule{}, "rules").SetKeys(true, "ID")

	tbl := dbmap.AddTableWithName(ExecutableFile{}, "executablefiles").SetKeys(true, "ID")
	tbl.ColMap("Md5").SetMaxSize(32)
	tbl.ColMap("Sha1").SetMaxSize(40)
	tbl.ColMap("Sha256").SetMaxSize(64)
	tbl.ColMap("CodeSectionSha256").SetMaxSize(64)
	tbl.ColMap("AuthenticodeMd5").SetMaxSize(32)
	tbl.ColMap("AuthenticodeSha1").SetMaxSize(40)
	tbl.ColMap("AuthenticodeSha256").SetMaxSize(64)

	dbmap.AddTableWithName(FileToSignerMap{}, "filetosignermap").SetKeys(false, "FileID", "SignerID")
	dbmap.AddTableWithName(FileToCounterSignerMap{}, "filetocountersignermap").SetKeys(false, "FileID", "SignerID")
	dbmap.AddTableWithName(Signer{}, "signers").SetKeys(true, "ID")
	dbmap.AddTableWithName(CatalogFile{}, "catalogfiles").SetKeys(true, "ID")
	dbmap.AddTableWithName(CertificateTrustList{}, "certificatetrustlist").SetKeys(false, "CatalogID", "Hash")

	dbmap.AddTableWithName(ProcessEvent{}, "processevents").SetKeys(true, "ID")
	dbmap.AddTableWithName(FileToSystemMap{}, "filetosystemmap").SetKeys(false, "FileID", "SystemID")

	tbl = dbmap.AddTableWithName(Customer{}, "customers").SetKeys(true, "ID")
	tbl.ColMap("UUID").SetMaxSize(16)
	tbl.ColMap("UUID").SetUnique(true)

	tbl = dbmap.AddTableWithName(User{}, "users").SetKeys(true, "ID")
	tbl.ColMap("PasswordHash").SetMaxSize(60)

	tbl = dbmap.AddTableWithName(BrowserSession{}, "browserSessions").SetKeys(true, "ID")
	tbl.ColMap("NonceHash").SetMaxSize(32)
	tbl = dbmap.AddTableWithName(PasswordReset{}, "passwordResets").SetKeys(true, "ID")
	tbl.ColMap("Nonce").SetMaxSize(64)

	dbmap.AddTableWithName(Task{}, "tasks").SetKeys(true, "ID")

	dbmap.AddTableWithName(Updates{}, "updates").SetKeys(true, "ID")

	// Create the tables
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}

	return dbmap, nil
}
