package mysql

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"utils/database/performer"
)

type Tx struct {
	Tx *sql.Tx
}

type DB struct {
	db *sql.DB
}

var (
	conns  = make(map[string]DB, 0)
	SqlDB  DB
	txPool = sync.Pool{
		New: func() interface{} {
			return new(Tx)
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

func InitDefaultDB(config Config) *DB {

	// 初始化默认连接
	var err error
	SqlDB.db, err = sql.Open("mysql", config.UserName+":"+config.Password+"@tcp("+config.Ip+":"+config.Port+")/"+config.DatabaseName+"?charset="+config.Charset)

	if err != nil {
		_ = SqlDB.db.Close()
		panic(err.Error())
	} else {

		conns = map[string]DB{
			"default": SqlDB,
		}

		// 设置数据库最大连接 减少timewait 正式环境调大
		SqlDB.db.SetMaxIdleConns(config.MaxIdleConns) // 连接池连接数 = mysql最大连接数/2
		SqlDB.db.SetMaxOpenConns(config.MaxOpenConns) // 最大打开连接 = mysql最大连接数
	}

	return &SqlDB
}

func InitCons(cons map[string]Config) *map[string]DB {
	for k, v := range cons {
		tempSql, openErr := sql.Open("mysql", v.UserName+":"+v.Password+"@tcp("+v.Ip+":"+v.Port+")/"+v.DatabaseName+"?charset="+v.Charset)
		if openErr != nil {
			tempSql.Close()
			panic(openErr.Error())
		}
		tempSql.SetMaxIdleConns(v.MaxIdleConns) // 连接池连接数 = mysql最大连接数/2
		tempSql.SetMaxOpenConns(v.MaxOpenConns) // 最大打开连接 = mysql最大连接数
		conns[k] = DB{
			tempSql,
		}
	}
	return &conns
}

func (db *DB) QueryWithConnection(con string, query string, args ...interface{}) ([]map[string]interface{}, *sql.Rows) {

	rs, err := conns[con].db.Query(query, args...)

	if err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			performer.SetColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			_ = rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			performer.SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}
	_ = rs.Close()
	return results, rs
}

func (db *DB) Query(query string, args ...interface{}) ([]map[string]interface{}, *sql.Rows) {

	rs, err := db.db.Query(query, args...)

	if err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			performer.SetColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			_ = rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			performer.SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		if rs != nil {
			_ = rs.Close()
		}
		panic(err)
	}
	_ = rs.Close()
	return results, rs
}

func (db *DB) Exec(query string, args ...interface{}) sql.Result {

	rs, err := db.db.Exec(query, args...)
	if err != nil {
		panic(err.Error())
	}
	return rs
}

func (db *DB) BeginTransactionsByLevel() *Tx {
	var (
		sqltx *Tx
		ok    bool
	)

	if sqltx, ok = txPool.Get().(*Tx); !ok {
		sqltx = new(Tx)
	}

	tx, err := db.db.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	if err != nil {
		panic(err)
	}
	(*sqltx).Tx = tx
	return sqltx
}

func (db *DB) BeginTransactions() *Tx {
	tx, err := db.db.BeginTx(context.Background(),
		&sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		panic(err)
	}

	var (
		sqltx *Tx
		ok    bool
	)

	if sqltx, ok = txPool.Get().(*Tx); !ok {
		sqltx = new(Tx)
	}

	(*sqltx).Tx = tx
	return sqltx
}

func (SqlTx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	rs, err := SqlTx.Tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	if rows, execError := rs.RowsAffected(); execError != nil || rows == 0 {
		return nil, errors.New("exec fail")
	}

	return rs, nil
}

func (SqlTx *Tx) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rs, err := SqlTx.Tx.Query(query, args...)

	if err != nil {
		return nil, err
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		_ = rs.Close()
		panic(colErr)
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		_ = rs.Close()
		panic(err)
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			performer.SetColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			_ = rs.Close()
			panic(scanErr)
		}
		for j := 0; j < len(col); j++ {
			performer.SetResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		_ = rs.Close()
		panic(err)
	}
	return results, nil
}

type TxFn func(*Tx) (error, map[string]interface{})

func (db *DB) WithTransaction(fn TxFn) (err error, res map[string]interface{}) {

	SqlTx := db.BeginTransactions()

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			_ = SqlTx.Tx.Rollback()
			db.PutAnEndToTransaction(SqlTx)
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			_ = SqlTx.Tx.Rollback()
			db.PutAnEndToTransaction(SqlTx)
		} else {
			// all good, commit
			err = SqlTx.Tx.Commit()
			db.PutAnEndToTransaction(SqlTx)
		}
	}()

	err, res = fn(SqlTx)
	return
}

func (db *DB) PutAnEndToTransaction(SqlTx *Tx) {
	txPool.Put(SqlTx)
}
