package services

import (
	"database/sql"
	"fmt"
	"generateTestData/backend/models"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct{}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// 打开数据库连接
func (s *DatabaseService) openConnection(ds *models.DataSource) (*sql.DB, error) {
	var dsn string

	switch strings.ToLower(ds.Type) {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			ds.Username, ds.Password, ds.Host, ds.Port, ds.Database)
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			ds.Host, ds.Port, ds.Username, ds.Password, ds.Database)
	case "sqlite":
		dsn = ds.Database
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", ds.Type)
	}

	return sql.Open(ds.Type, dsn)
}

// 测试数据库连接
func (s *DatabaseService) TestConnection(ds *models.DataSource) error {
	db, err := s.openConnection(ds)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}

// 获取数据库中的所有表
func (s *DatabaseService) GetTables(ds *models.DataSource) ([]string, error) {
	db, err := s.openConnection(ds)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	switch strings.ToLower(ds.Type) {
	case "mysql":
		query = "SHOW TABLES"
	case "postgresql":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	case "sqlite":
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", ds.Type)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// 获取表结构
func (s *DatabaseService) GetTableStructure(ds *models.DataSource, tableName string) (*models.TableInfo, error) {
	db, err := s.openConnection(ds)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var columns []models.ColumnInfo

	switch strings.ToLower(ds.Type) {
	case "mysql":
		columns, err = s.getMySQLColumns(db, tableName)
	case "postgresql":
		columns, err = s.getPostgreSQLColumns(db, tableName)
	case "sqlite":
		columns, err = s.getSQLiteColumns(db, tableName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", ds.Type)
	}

	if err != nil {
		return nil, err
	}

	return &models.TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// 获取MySQL表列信息
func (s *DatabaseService) getMySQLColumns(db *sql.DB, tableName string) ([]models.ColumnInfo, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY,
			EXTRA,
			CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_NAME = ? 
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var nullable, columnKey, extra string
		var defaultValue, maxLength sql.NullString

		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultValue, &columnKey, &extra, &maxLength)
		if err != nil {
			return nil, err
		}

		col.Nullable = nullable == "YES"
		col.IsPrimaryKey = columnKey == "PRI"
		col.IsAutoIncrement = strings.Contains(extra, "auto_increment")
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}
		if maxLength.Valid {
			fmt.Sscanf(maxLength.String, "%d", &col.MaxLength)
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// 获取PostgreSQL表列信息
func (s *DatabaseService) getPostgreSQLColumns(db *sql.DB, tableName string) ([]models.ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length
		FROM information_schema.columns 
		WHERE table_name = $1 
		ORDER BY ordinal_position
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var nullable string
		var defaultValue, maxLength sql.NullString

		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultValue, &maxLength)
		if err != nil {
			return nil, err
		}

		col.Nullable = nullable == "YES"
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}
		if maxLength.Valid {
			fmt.Sscanf(maxLength.String, "%d", &col.MaxLength)
		}

		columns = append(columns, col)
	}

	// 获取主键信息
	pkQuery := `
		SELECT column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = $1 AND tc.constraint_type = 'PRIMARY KEY'
	`

	pkRows, err := db.Query(pkQuery, tableName)
	if err == nil {
		defer pkRows.Close()
		pkColumns := make(map[string]bool)
		for pkRows.Next() {
			var colName string
			if pkRows.Scan(&colName) == nil {
				pkColumns[colName] = true
			}
		}

		for i := range columns {
			if pkColumns[columns[i].Name] {
				columns[i].IsPrimaryKey = true
			}
		}
	}

	return columns, nil
}

// 获取SQLite表列信息
func (s *DatabaseService) getSQLiteColumns(db *sql.DB, tableName string) ([]models.ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var cid int
		var col models.ColumnInfo
		var notNull int
		var defaultValue sql.NullString
		var pk int

		err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		col.Nullable = notNull == 0
		col.IsPrimaryKey = pk == 1
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}

		columns = append(columns, col)
	}

	return columns, nil
}
