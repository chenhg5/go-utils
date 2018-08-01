package database

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type SqlTxStruct struct {
	Tx *sql.Tx
}

var (
	sqlDBmap        = make(map[string]*sql.DB, 0)
	SqlDB           *sql.DB
	SqlTxStructPool = sync.Pool{
		New: func() interface{} {
			return new(SqlTxStruct)
		},
	}
)

type Config struct {
	UserName     string
	Password     string
	Port         string
	Ip           string
	DatabaseName string
	Charset      string
	MaxIdleConns int
	MaxOpenConns int
}

func InitDefaultDB(config Config) {

	// 初始化默认连接
	var err error
	SqlDB, err = sql.Open("mysql", config.UserName+":"+config.Password+"@tcp("+config.Ip+":"+config.Port+")/"+config.DatabaseName+"?charset="+config.Charset)

	if err != nil {
		SqlDB.Close()
		panic(err.Error())
	} else {

		sqlDBmap = map[string]*sql.DB{
			"default": SqlDB,
		}

		// 设置数据库最大连接 减少timewait 正式环境调大
		SqlDB.SetMaxIdleConns(config.MaxIdleConns) // 连接池连接数 = mysql最大连接数/2
		SqlDB.SetMaxOpenConns(config.MaxOpenConns) // 最大打开连接 = mysql最大连接数
	}
}

func InitCons(cons map[string]Config) {
	for k, v := range cons {
		tempSql, openErr := sql.Open("mysql", v.UserName+":"+v.Password+"@tcp("+v.Ip+":"+v.Port+")/"+v.DatabaseName+"?charset="+v.Charset)
		if openErr != nil {
			tempSql.Close()
			panic(openErr.Error())
		}
		tempSql.SetMaxIdleConns(v.MaxIdleConns) // 连接池连接数 = mysql最大连接数/2
		tempSql.SetMaxOpenConns(v.MaxOpenConns) // 最大打开连接 = mysql最大连接数
		sqlDBmap[k] = tempSql
	}
}

func QueryWithConnection(con string, query string, args ...interface{}) ([]map[string]interface{}, *sql.Rows) {

	rs, err := sqlDBmap[con].Query(query, args...)

	if err != nil {
		if rs != nil {
			rs.Close()
		}
		panic(err)
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		if rs != nil {
			rs.Close()
		}
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		if rs != nil {
			rs.Close()
		}
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
			rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		if rs != nil {
			rs.Close()
		}
		panic(err)
	}
	rs.Close()
	return results, rs
}

func Query(query string, args ...interface{}) ([]map[string]interface{}, *sql.Rows) {

	rs, err := sqlDBmap["default"].Query(query, args...)

	if err != nil {
		if rs != nil {
			rs.Close()
		}
		panic(err)
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		if rs != nil {
			rs.Close()
		}
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		if rs != nil {
			rs.Close()
		}
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
			rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		if rs != nil {
			rs.Close()
		}
		panic(err)
	}
	rs.Close()
	return results, rs
}

func Exec(query string, args ...interface{}) sql.Result {

	rs, err := sqlDBmap["default"].Exec(query, args...)
	if err != nil {
		panic(err.Error())
	}
	return rs
}

func BeginTransactionsByLevel() *SqlTxStruct {

	//LevelDefault IsolationLevel = iota
	//LevelReadUncommitted
	//LevelReadCommitted
	//LevelWriteCommitted
	//LevelRepeatableRead
	//LevelSnapshot
	//LevelSerializable
	//LevelLinearizable

	var SqlTx *SqlTxStruct
	SqlTx = SqlTxStructPool.Get().(*SqlTxStruct)
	if SqlTx == nil {
		SqlTx = new(SqlTxStruct)
	}

	tx, err := SqlDB.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	if err != nil {
		panic(err)
	}
	(*SqlTx).Tx = tx
	return SqlTx
}

func BeginTransactions() *SqlTxStruct {
	tx, err := SqlDB.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		panic(err)
	}

	var SqlTx *SqlTxStruct
	SqlTx = SqlTxStructPool.Get().(*SqlTxStruct)
	if SqlTx == nil {
		SqlTx = new(SqlTxStruct)
	}

	(*SqlTx).Tx = tx
	return SqlTx
}

func (SqlTx *SqlTxStruct) Exec(query string, args ...interface{}) (sql.Result, error) {
	rs, err := SqlTx.Tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (SqlTx *SqlTxStruct) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rs, err := SqlTx.Tx.Query(query, args...)

	if err != nil {
		return nil, err
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		rs.Close()
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		rs.Close()
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
			rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		rs.Close()
		panic(err)
	}
	return results, nil
}

type TxFn func(*SqlTxStruct) (error, map[string]interface{})

func WithTransaction(fn TxFn) (err error, res map[string]interface{}) {

	SqlTx := BeginTransactions()

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			SqlTx.Tx.Rollback()
			PutAnEndToTransaction(SqlTx)
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			SqlTx.Tx.Rollback()
			PutAnEndToTransaction(SqlTx)
		} else {
			// all good, commit
			err = SqlTx.Tx.Commit()
			PutAnEndToTransaction(SqlTx)
		}
	}()

	err, res = fn(SqlTx)
	return
}

func PutAnEndToTransaction(SqlTx *SqlTxStruct) {
	SqlTxStructPool.Put(SqlTx)
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
