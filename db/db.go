package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DriverSqlite3 = "sqlite3"
	DriverMysql   = "mysql"
)

type Config struct {
	Driver   string
	Name     string
	Host     string
	User     string
	Password string
}

type Repository[T any] struct {
	db    *sql.DB
	table string
}

func NewRepository[T any](db *sql.DB, table string) Repository[T] {
	return Repository[T]{db: db, table: table}
}

func (r Repository[T]) Create(item T) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("db: repository requires a non-nil database")
	}

	columns, values, args, err := r.structInsertParts(item)
	if err != nil {
		return 0, err
	}
	if len(columns) == 0 || len(values) == 0 {
		return 0, fmt.Errorf("db: repository insert produced no columns")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", r.table, strings.Join(columns, ", "), strings.TrimSuffix(strings.Repeat("?, ", len(values)), ", "))
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r Repository[T]) FindAll() ([]T, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db: repository requires a non-nil database")
	}

	query := fmt.Sprintf("SELECT * FROM %s", r.table)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := []T{}
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		obj := new(T)
		fieldMap, err := structFieldMap(obj, columns)
		if err != nil {
			return nil, err
		}
		ref := reflect.ValueOf(obj).Elem()
		for i, col := range columns {
			field := fieldMap[col]
			if !field.IsValid() {
				continue
			}
			value, err := coerceValue(values[i], field.Type())
			if err != nil {
				return nil, err
			}
			field.Set(reflect.ValueOf(value))
		}
		results = append(results, ref.Interface().(T))
	}
	return results, rows.Err()
}

func (r Repository[T]) FindByID(id any) (T, error) {
	var zero T
	if r.db == nil {
		return zero, fmt.Errorf("db: repository requires a non-nil database")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1", r.table)
	rows, err := r.db.Query(query, id)
	if err != nil {
		return zero, err
	}
	defer rows.Close()

	if !rows.Next() {
		return zero, nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return zero, err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return zero, err
	}

	obj := new(T)
	fieldMap, err := structFieldMap(obj, columns)
	if err != nil {
		return zero, err
	}
	ref := reflect.ValueOf(obj).Elem()
	for i, col := range columns {
		field := fieldMap[col]
		if !field.IsValid() {
			continue
		}
		value, err := coerceValue(values[i], field.Type())
		if err != nil {
			return zero, err
		}
		field.Set(reflect.ValueOf(value))
	}
	return ref.Interface().(T), nil
}

func (r Repository[T]) UpdateByID(id any, updates map[string]any) error {
	if r.db == nil {
		return fmt.Errorf("db: repository requires a non-nil database")
	}
	if len(updates) == 0 {
		return nil
	}

	parts := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)
	for key, value := range updates {
		parts = append(parts, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", r.table, strings.Join(parts, ", "))
	_, err := r.db.Exec(query, args...)
	return err
}

func (r Repository[T]) DeleteByID(id any) error {
	if r.db == nil {
		return fmt.Errorf("db: repository requires a non-nil database")
	}
	_, err := r.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.table), id)
	return err
}

func (r Repository[T]) structInsertParts(item T) ([]string, []string, []any, error) {
	ref := reflect.ValueOf(item)
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			return nil, nil, nil, fmt.Errorf("db: repository insert cannot use a nil pointer")
		}
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil, nil, nil, fmt.Errorf("db: repository insert requires a struct type")
	}

	fields := make([]string, 0)
	args := make([]any, 0)
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Type().Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := field.Name
		if tag := field.Tag.Get("db"); tag != "" {
			name = strings.TrimSpace(strings.Split(tag, ",")[0])
		}
		if name == "" {
			continue
		}
		if strings.EqualFold(name, "id") {
			v := ref.Field(i)
			if v.Kind() >= reflect.Int && v.Kind() <= reflect.Int64 && v.Int() == 0 {
				continue
			}
		}
		fields = append(fields, snakeCaseName(name))
		args = append(args, ref.Field(i).Interface())
	}
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return fields, placeholders, args, nil
}

func structFieldMap(target any, columns []string) (map[string]reflect.Value, error) {
	ref := reflect.ValueOf(target)
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			return nil, fmt.Errorf("db: repository field map requires a non-nil pointer")
		}
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil, fmt.Errorf("db: repository requires a struct type")
	}

	fieldMap := map[string]reflect.Value{}
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Type().Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := field.Name
		if tag := field.Tag.Get("db"); tag != "" {
			name = strings.TrimSpace(strings.Split(tag, ",")[0])
		}
		fieldMap[snakeCaseName(name)] = ref.Field(i)
	}
	return fieldMap, nil
}

func snakeCaseName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			prev := rune(name[i-1])
			if unicode.IsLower(prev) || (i+1 < len(name) && unicode.IsLower(rune(name[i+1]))) {
				b.WriteRune('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func coerceValue(raw sql.RawBytes, fieldType reflect.Type) (any, error) {
	if raw == nil {
		return reflect.Zero(fieldType).Interface(), nil
	}

	value := string(raw)
	kind := fieldType.Kind()
	switch kind {
	case reflect.String:
		return value, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(v).Convert(fieldType).Interface(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(v).Convert(fieldType).Interface(), nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(v).Convert(fieldType).Interface(), nil
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		return value, nil
	}
}

func ValidateConfig(cfg Config) error {
	driver := cfg.Driver
	if driver == "" {
		driver = DriverSqlite3
	}

	switch driver {
	case DriverSqlite3:
		return nil
	default:
		return fmt.Errorf("invalid database driver (%s): currently only sqlite3 is supported", driver)
	}
}

func NewSQL(cfg Config) (*sql.DB, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	driver := cfg.Driver
	if driver == "" {
		driver = DriverSqlite3
	}

	name := cfg.Name
	if len(name) == 0 {
		name = "app_db"
	}
	return sql.Open(driver, name)
}

func WithTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("db: transaction requires a non-nil database")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx rollback failed: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx commit failed: %w (rollback error: %v)", err, rbErr)
		}
		return err
	}

	return nil
}
