package database

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"
)

type SqlTx struct {
	Tx *sql.Tx
}

type SqlDB struct {
	db *sql.DB
}

var sqlDB SqlDB

type Config struct {
	mysql.Config
	MaxIdleConn int
	MaxOpenConn int
	MaxLifetime int
}

func (c *Config) FormatDSN() string {
	return c.Config.FormatDSN()
}

func InitDefaultDB(config Config) *SqlDB {

	if config.DBName == "" {
		panic("empty database name")
	}

	// 初始化默认连接
	var err error
	sqlDB.db, err = sql.Open("mysql", config.FormatDSN())

	if err != nil {
		_ = sqlDB.db.Close()
		panic(err.Error())
	} else {
		// 设置数据库最大连接 减少 time wait 正式环境调大
		sqlDB.db.SetMaxIdleConns(config.MaxIdleConn) // 连接池连接数 = mysql最大连接数/2
		sqlDB.db.SetMaxOpenConns(config.MaxOpenConn) // 最大打开连接 = mysql最大连接数

		// 设置链接重置时间
		sqlDB.db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
	}

	return &sqlDB
}

func (db *SqlDB) Query(query string, args ...interface{}) []map[string]interface{} {

	rs, err := db.db.Query(query, args...)

	if err != nil || rs == nil {
		panic(err)
	}

	defer func() {
		_ = rs.Close()
	}()

	col, colErr := rs.Columns()

	if colErr != nil {
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		panic(err)
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			SetColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		panic(err)
	}
	return results
}

func (db *SqlDB) Exec(query string, args ...interface{}) (sql.Result, int64) {

	rs, err := db.db.Exec(query, args...)
	if err != nil {
		panic(err)
	}

	rows, execError := rs.RowsAffected()

	if execError != nil {
		panic(execError)
	}

	return rs, rows
}

func (db *SqlDB) BeginTransactionsByLevel() *SqlTx {

	//LevelDefault IsolationLevel = iota
	//LevelReadUncommitted
	//LevelReadCommitted
	//LevelWriteCommitted
	//LevelRepeatableRead
	//LevelSnapshot
	//LevelSerializable
	//LevelLinearizable

	tx, err := db.db.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	if err != nil {
		panic(err)
	}
	return new(SqlTx).WithTx(tx)
}

func (db *SqlDB) BeginTransactions() *SqlTx {
	tx, err := db.db.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		panic(err)
	}
	return new(SqlTx).WithTx(tx)
}

func (SqlTx *SqlTx) Exec(query string, args ...interface{}) (sql.Result, int64) {
	rs, err := SqlTx.Tx.Exec(query, args...)
	if err != nil {
		panic(err)
	}

	rows, execError := rs.RowsAffected()

	if execError != nil {
		panic(execError)
	}

	return rs, rows
}

func (SqlTx *SqlTx) WithTx(tx *sql.Tx) *SqlTx {
	SqlTx.Tx = tx
	return SqlTx
}

func (SqlTx *SqlTx) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rs, err := SqlTx.Tx.Query(query, args...)

	if err != nil || rs == nil {
		return nil, err
	}

	defer func() {
		_ = rs.Close()
	}()

	col, colErr := rs.Columns()

	if colErr != nil {
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		panic(err)
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			SetColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}

	if err := rs.Err(); err != nil {
		panic(err)
	}

	return results, nil
}

type TxFn func(*SqlTx) (error, map[string]interface{})

func (db *SqlDB) WithTransaction(fn TxFn) (err error, res map[string]interface{}) {

	SqlTx := db.BeginTransactions()

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			_ = SqlTx.Tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			_ = SqlTx.Tx.Rollback()
		} else {
			// all good, commit
			err = SqlTx.Tx.Commit()
		}
	}()

	err, res = fn(SqlTx)
	return
}

func SetColVarType(colVar *[]interface{}, i int, typeName string) {
	switch typeName {
	case "INT":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "TINYINT":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "MEDIUMINT":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "SMALLINT":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "BIGINT":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "FLOAT":
		var s sql.NullFloat64
		(*colVar)[i] = &s
	case "DOUBLE":
		var s sql.NullFloat64
		(*colVar)[i] = &s
	case "DECIMAL":
		var s []uint8
		(*colVar)[i] = &s
	case "DATE":
		var s sql.NullString
		(*colVar)[i] = &s
	case "TIME":
		var s sql.NullString
		(*colVar)[i] = &s
	case "YEAR":
		var s sql.NullString
		(*colVar)[i] = &s
	case "DATETIME":
		var s sql.NullString
		(*colVar)[i] = &s
	case "TIMESTAMP":
		var s sql.NullString
		(*colVar)[i] = &s
	case "VARCHAR":
		var s sql.NullString
		(*colVar)[i] = &s
	case "MEDIUMTEXT":
		var s sql.NullString
		(*colVar)[i] = &s
	case "LONGTEXT":
		var s sql.NullString
		(*colVar)[i] = &s
	case "TINYTEXT":
		var s sql.NullString
		(*colVar)[i] = &s
	case "TEXT":
		var s sql.NullString
		(*colVar)[i] = &s
	default:
		var s interface{}
		(*colVar)[i] = &s
	}
}

func SetResultValue(result *map[string]interface{}, index string, colVar interface{}, typeName string) {
	switch typeName {
	case "INT":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "TINYINT":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "MEDIUMINT":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "SMALLINT":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "BIGINT":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "FLOAT":
		temp := *(colVar.(*sql.NullFloat64))
		if temp.Valid {
			(*result)[index] = temp.Float64
		} else {
			(*result)[index] = nil
		}
	case "DOUBLE":
		temp := *(colVar.(*sql.NullFloat64))
		if temp.Valid {
			(*result)[index] = temp.Float64
		} else {
			(*result)[index] = nil
		}
	case "DECIMAL":
		//temp := *(colVar.(*sql.NullInt64))
		//if temp.Valid {
		//	(*result)[index] = temp.Int64
		//} else {
		//	(*result)[index] = nil
		//}
		(*result)[index] = *(colVar.(*[]uint8))
	case "DATE":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "TIME":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "YEAR":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "DATETIME":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "TIMESTAMP":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "VARCHAR":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "MEDIUMTEXT":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "LONGTEXT":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "TINYTEXT":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "TEXT":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	default:
		(*result)[index] = colVar
	}
}
