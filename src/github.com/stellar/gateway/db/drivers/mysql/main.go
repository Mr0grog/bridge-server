package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate"
	"github.com/stellar/gateway/db/entities"
)

type MysqlDriver struct {
	database *sqlx.DB
}

func (d *MysqlDriver) Init(url string) (err error) {
	d.database, err = sqlx.Connect("mysql", url)
	return
}

// go-bindata -ignore .+\.go$ -pkg mysql -o bindata.go ./migrations*
func (d *MysqlDriver) MigrateUp(component string) (migrationsApplied int, err error) {
	source := d.getAssetMigrationSource(component)
	migrationsApplied, err = migrate.Exec(d.database.DB, "mysql", source, migrate.Up)
	return
}

func (d *MysqlDriver) GetLastReceivedPayment() (*entities.ReceivedPayment, error) {
	var receivedPayment entities.ReceivedPayment
	err := d.database.Get(&receivedPayment, "SELECT * FROM ReceivedPayment ORDER BY id DESC LIMIT 1")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &receivedPayment, nil
}

func (d *MysqlDriver) Insert(object entities.Entity) (id int64, err error) {
	value, tableName, err := getTypeData(object)

	if err != nil {
		return 0, err
	}

	fieldsCount := value.NumField()
	var fieldNames []string
	var fieldValues []string

	for i := 0; i < fieldsCount; i++ {
		field := value.Field(i)
		tag := field.Tag.Get("db")
		if tag == "" {
			continue
		}
		fieldNames = append(fieldNames, tag)
		fieldValues = append(fieldValues, ":"+tag)
	}

	query := "INSERT INTO " + tableName + " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ");"

	var result sql.Result
	switch object := object.(type) {
	case *entities.AuthorizedTransaction:
		result, err = d.database.NamedExec(query, object)
	case *entities.AllowedFi:
		result, err = d.database.NamedExec(query, object)
	case *entities.AllowedUser:
		result, err = d.database.NamedExec(query, object)
	case *entities.SentTransaction:
		result, err = d.database.NamedExec(query, object)
	case *entities.ReceivedPayment:
		result, err = d.database.NamedExec(query, object)
	}

	if err != nil {
		return
	}

	id, err = result.LastInsertId()

	if id == 0 {
		// Not autoincrement
		if object.GetId() == nil {
			return 0, fmt.Errorf("Not autoincrement but ID nil")
		}
		id = *object.GetId()
	}

	if err == nil {
		object.SetId(id)
		object.SetExists()
	}

	return
}

func (d *MysqlDriver) Update(object entities.Entity) (err error) {
	value, tableName, err := getTypeData(object)

	if err != nil {
		return err
	}

	fieldsCount := value.NumField()

	query := "UPDATE " + tableName + " SET "
	var fields []string

	for i := 0; i < fieldsCount; i++ {
		field := value.Field(i)
		if field.Tag.Get("db") == "id" || field.Tag.Get("db") == "" {
			continue
		}
		fields = append(fields, field.Tag.Get("db")+" = :"+field.Tag.Get("db"))
	}

	query += strings.Join(fields, ", ") + " WHERE id = :id;"

	switch object := object.(type) {
	case *entities.AuthorizedTransaction:
		_, err = d.database.NamedExec(query, object)
	case *entities.AllowedFi:
		_, err = d.database.NamedExec(query, object)
	case *entities.AllowedUser:
		_, err = d.database.NamedExec(query, object)
	case *entities.SentTransaction:
		_, err = d.database.NamedExec(query, object)
	case *entities.ReceivedPayment:
		_, err = d.database.NamedExec(query, object)
	}

	return
}

func (d *MysqlDriver) Delete(object entities.Entity) (err error) {
	_, tableName, err := getTypeData(object)

	if err != nil {
		return
	}

	query := "DELETE FROM " + tableName + " WHERE id = :id;"
	_, err = d.database.NamedExec(query, object)

	return
}

func (d *MysqlDriver) GetOne(object entities.Entity, where string, params ...interface{}) (entities.Entity, error) {
	_, tableName, err := getTypeData(object)
	if err != nil {
		return nil, err
	}

	err = d.database.Get(object, "SELECT * FROM "+tableName+" WHERE "+where+" LIMIT 1;", params...)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		} else {
			return nil, err
		}
	}
	object.SetExists() // Mark this entity as existing
	return object, err
}

func getTypeData(object interface{}) (typeValue reflect.Type, tableName string, err error) {
	switch object := object.(type) {
	case *entities.AuthorizedTransaction:
		typeValue = reflect.TypeOf(*object)
		tableName = "AuthorizedTransaction"
	case *entities.AllowedFi:
		typeValue = reflect.TypeOf(*object)
		tableName = "AllowedFi"
	case *entities.AllowedUser:
		typeValue = reflect.TypeOf(*object)
		tableName = "AllowedUser"
	case *entities.SentTransaction:
		typeValue = reflect.TypeOf(*object)
		tableName = "SentTransaction"
	case *entities.ReceivedPayment:
		typeValue = reflect.TypeOf(*object)
		tableName = "ReceivedPayment"
	default:
		return typeValue, tableName, fmt.Errorf("Unknown entity type: %T", object)
	}
	return
}

func (d *MysqlDriver) getAssetMigrationSource(component string) (source *migrate.AssetMigrationSource) {
	source = &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "migrations_" + component,
	}
	return
}
