package duckdb

import (
	"gorm.io/gorm"
)

// CustomRowQuery is a debugging version of GORM's RowQuery callback
func CustomRowQuery(db *gorm.DB) {
	debugLog(" CustomRowQuery called")
	debugLog(" db.Error: %v", db.Error)
	debugLog(" db.DryRun: %t", db.DryRun)

	if db.Error == nil {
		debugLog(" No error, calling BuildQuerySQL")
		// This is what GORM's BuildQuerySQL does for Raw queries
		if db.Statement.SQL.Len() == 0 {
			debugLog(" SQL is empty, this shouldn't happen for Raw() queries")
		}

		// Check for DryRun or Error before proceeding
		if db.DryRun || db.Error != nil {
			debugLog(" DryRun=%t or Error=%v, returning early", db.DryRun, db.Error)
			return
		}

		debugLog(" Checking for 'rows' setting")
		if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
			debugLog(" isRows=true, calling QueryContext")
			db.Statement.Settings.Delete("rows")
			db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
		} else {
			debugLog(" isRows=false or not found, calling QueryRowContext")
			debugLog(" SQL: %s", db.Statement.SQL.String())
			debugLog(" Vars: %v", db.Statement.Vars...)
			debugLog(" ConnPool type: %T", db.Statement.ConnPool)

			result := db.Statement.ConnPool.QueryRowContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			debugLog(" QueryRowContext returned: %v (nil: %t)", result, result == nil)

			db.Statement.Dest = result
			debugLog(" After assignment - Statement.Dest: %v (nil: %t)", db.Statement.Dest, db.Statement.Dest == nil)
		}

		debugLog(" Setting RowsAffected to -1")
		db.RowsAffected = -1
	} else {
		debugLog(" db.Error is not nil: %v", db.Error)
	}
}
