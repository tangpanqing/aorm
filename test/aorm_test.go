package test

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tangpanqing/aorm"
	"github.com/tangpanqing/aorm/executor"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/null"
	"testing"
	"time"
)

type Article struct {
	Id          null.Int    `aorm:"primary;auto_increment;type:bigint" json:"id"`
	Type        null.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	ArticleBody null.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

type ArticleVO struct {
	Id          null.Int    `aorm:"primary;auto_increment;type:bigint" json:"id"`
	Type        null.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	PersonName  null.String `aorm:"comment:人员名称" json:"personName"`
	ArticleBody null.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

type Person struct {
	Id         null.Int    `aorm:"primary;auto_increment" json:"id"`
	Name       null.String `aorm:"size:100;not null;comment:名字" json:"name"`
	Sex        null.Bool   `aorm:"index;comment:性别" json:"sex"`
	Age        null.Int    `aorm:"index;comment:年龄" json:"age"`
	Type       null.Int    `aorm:"index;comment:类型" json:"type"`
	CreateTime null.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money      null.Float  `aorm:"comment:金额" json:"money"`
	Test       null.Float  `aorm:"type:double;comment:测试" json:"test"`
}

type PersonAge struct {
	Age      null.Int
	AgeCount null.Int
}

type PersonWithArticleCount struct {
	Id           null.Int    `aorm:"primary;auto_increment" json:"id"`
	Name         null.String `aorm:"size:100;not null;comment:名字" json:"name"`
	Sex          null.Bool   `aorm:"index;comment:性别" json:"sex"`
	Age          null.Int    `aorm:"index;comment:年龄" json:"age"`
	Type         null.Int    `aorm:"index;comment:类型" json:"type"`
	CreateTime   null.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money        null.Float  `aorm:"comment:金额" json:"money"`
	Test         null.Float  `aorm:"type:double;comment:测试" json:"test"`
	ArticleCount null.Int    `aorm:"comment:文章数量" json:"articleCount"`
}

func TestAll(t *testing.T) {
	//sqlite3Content, sqlite3Err := aorm.Open("sqlite3", "test.db")
	//if sqlite3Err != nil {
	//	panic(sqlite3Err)
	//}

	username := "root"
	password := "root"
	hostname := "localhost"
	port := "3306"
	dbname := "database_name"

	mysqlContent, mysqlErr := aorm.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
	if mysqlErr != nil {
		panic(mysqlErr)
	}

	dbList := make([]aorm.DbContent, 0)
	//dbList = append(dbList, sqlite3Content)
	dbList = append(dbList, mysqlContent)

	for i := 0; i < len(dbList); i++ {
		dbItem := dbList[i]

		testMigrate(dbItem.DriverName, dbItem.DbLink)

		testShowCreateTable(dbItem.DriverName, dbItem.DbLink)

		id := testInsert(dbItem.DriverName, dbItem.DbLink)
		testInsertBatch(dbItem.DriverName, dbItem.DbLink)

		testGetOne(dbItem.DriverName, dbItem.DbLink, id)
		testGetMany(dbItem.DriverName, dbItem.DbLink)
		testUpdate(dbItem.DriverName, dbItem.DbLink, id)
		testDelete(dbItem.DriverName, dbItem.DbLink, id)

		id2 := testInsert(dbItem.DriverName, dbItem.DbLink)
		testTable(dbItem.DriverName, dbItem.DbLink)
		testSelect(dbItem.DriverName, dbItem.DbLink)
		testSelectWithSub(dbItem.DriverName, dbItem.DbLink)
		testWhereWithSub(dbItem.DriverName, dbItem.DbLink)
		testWhere(dbItem.DriverName, dbItem.DbLink)
		testJoin(dbItem.DriverName, dbItem.DbLink)
		testGroupBy(dbItem.DriverName, dbItem.DbLink)
		testHaving(dbItem.DriverName, dbItem.DbLink)
		testOrderBy(dbItem.DriverName, dbItem.DbLink)
		testLimit(dbItem.DriverName, dbItem.DbLink)
		testLock(dbItem.DriverName, dbItem.DbLink, id2)

		testIncrement(dbItem.DriverName, dbItem.DbLink, id2)
		testDecrement(dbItem.DriverName, dbItem.DbLink, id2)

		testValue(dbItem.DriverName, dbItem.DbLink, id2)

		testPluck(dbItem.DriverName, dbItem.DbLink)

		testCount(dbItem.DriverName, dbItem.DbLink)
		testSum(dbItem.DriverName, dbItem.DbLink)
		testAvg(dbItem.DriverName, dbItem.DbLink)
		testMin(dbItem.DriverName, dbItem.DbLink)
		testMax(dbItem.DriverName, dbItem.DbLink)

		testExec(dbItem.DriverName, dbItem.DbLink)

		testTransaction(dbItem.DriverName, dbItem.DbLink)
		testTruncate(dbItem.DriverName, dbItem.DbLink)
		testHelper(dbItem.DriverName, dbItem.DbLink)
	}

	//
	//for _, db := range dbMap {
	//	db.Close()
	//}
}

func testMysqlConnect() *sql.DB {
	//replace this database param
	username := "root"
	password := "root"
	hostname := "localhost"
	port := "3306"
	dbname := "database_name"

	//connect
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	//ping test
	err1 := db.Ping()
	if err1 != nil {
		panic(err1)
	}

	return db
}

func testMigrate(name string, db *sql.DB) {
	//AutoMigrate
	aorm.Migrator(db).Driver(name).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").AutoMigrate(&Person{})
	aorm.Migrator(db).Driver(name).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "文章").AutoMigrate(&Article{})

	//Migrate
	aorm.Migrator(db).Driver(name).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").Migrate("person_1", &Person{})
}

func testShowCreateTable(name string, db *sql.DB) {
	aorm.Migrator(db).Driver(name).ShowCreateTable("person")
}

func testInsert(name string, db *sql.DB) int64 {

	id, errInsert := aorm.Use(db).Debug(false).Insert(&Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(false),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15987654321),
		Test:       null.FloatFrom(200.15987654321987654321),
	})
	if errInsert != nil {
		panic(name + "testInsert" + "found err")
	}

	aorm.Use(db).Debug(false).Insert(&Article{
		Type:        null.IntFrom(0),
		PersonId:    null.IntFrom(id),
		ArticleBody: null.StringFrom("文章内容"),
	})

	return id
}

func testInsertBatch(name string, db *sql.DB) int64 {
	var batch []Person
	batch = append(batch, Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(false),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15987654321),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	batch = append(batch, Person{
		Name:       null.StringFrom("Bob"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15987654321),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	count, err := aorm.Use(db).Debug(false).InsertBatch(&batch)
	if err != nil {
		panic(name + "testInsertBatch" + "found err")
	}

	return count
}

func testGetOne(name string, db *sql.DB, id int64) {
	var person Person
	errFind := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).GetOne(&person)
	if errFind != nil {
		panic(name + "testGetOne" + "found err")
	}
}

func testGetMany(name string, db *sql.DB) {
	var list []Person
	errSelect := aorm.Use(db).Debug(false).Where(&Person{Type: null.IntFrom(0)}).GetMany(&list)
	if errSelect != nil {
		panic(name + " testGetMany " + "found err:" + errSelect.Error())
	}
}

func testUpdate(name string, db *sql.DB, id int64) {
	_, errUpdate := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Update(&Person{Name: null.StringFrom("Bob")})
	if errUpdate != nil {
		panic(name + "testGetMany" + "found err")
	}
}

func testDelete(name string, db *sql.DB, id int64) {
	_, errDelete := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Delete()
	if errDelete != nil {
		panic(name + "testDelete" + "found err")
	}
}

func testTable(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Table("person_1").Insert(&Person{Name: null.StringFrom("Cherry")})
	if err != nil {
		panic(name + "testTable" + "found err")
	}
}

func testSelect(name string, db *sql.DB) {
	var listByFiled []Person
	err := aorm.Use(db).Debug(false).Select("name,age").Where(&Person{Age: null.IntFrom(18)}).GetMany(&listByFiled)
	if err != nil {
		panic(name + " testSelect " + "found err:" + err.Error())
	}
}

func testSelectWithSub(name string, db *sql.DB) {
	var listByFiled []PersonWithArticleCount

	sub := aorm.Sub().Table("article").SelectCount("id", "article_count_tem").WhereRaw("person_id", "=person.id")
	err := aorm.Use(db).Debug(false).
		SelectExp(&sub, "article_count").
		Select("*").
		Where(&Person{Age: null.IntFrom(18)}).
		GetMany(&listByFiled)

	if err != nil {
		panic(name + " testSelectWithSub " + "found err:" + err.Error())
	}
}

func testWhereWithSub(name string, db *sql.DB) {
	var listByFiled []Person

	sub := aorm.Sub().Table("article").Select("person_id").GroupBy("person_id").HavingGt("count(person_id)", 0)

	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereIn("id", &sub).
		GetMany(&listByFiled)

	if err != nil {
		panic(name + " testWhereWithSub " + "found err:" + err.Error())
	}
}

func testWhere(name string, db *sql.DB) {
	var listByWhere []Person

	var where1 []executor.WhereItem
	where1 = append(where1, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})
	where1 = append(where1, executor.WhereItem{Field: "age", Opt: executor.In, Val: []int{18, 20}})
	where1 = append(where1, executor.WhereItem{Field: "money", Opt: executor.Between, Val: []float64{100.1, 200.9}})
	where1 = append(where1, executor.WhereItem{Field: "money", Opt: executor.Eq, Val: 100.15})
	where1 = append(where1, executor.WhereItem{Field: "name", Opt: executor.Like, Val: []string{"%", "li", "%"}})

	err := aorm.Use(db).Debug(false).Table("person").WhereArr(where1).GetMany(&listByWhere)
	if err != nil {
		panic(name + "testWhere" + "found err")
	}
}

func testJoin(name string, db *sql.DB) {
	var list2 []ArticleVO
	var where2 []executor.WhereItem
	where2 = append(where2, executor.WhereItem{Field: "o.type", Opt: executor.Eq, Val: 0})
	where2 = append(where2, executor.WhereItem{Field: "p.age", Opt: executor.In, Val: []int{18, 20}})
	err := aorm.Use(db).Debug(false).
		Table("article o").
		LeftJoin("person p", "p.id=o.person_id").
		Select("o.*").
		Select("p.name as person_name").
		WhereArr(where2).
		GetMany(&list2)
	if err != nil {
		panic(name + " testWhere " + "found err " + err.Error())
	}
}

func testGroupBy(name string, db *sql.DB) {
	var personAge PersonAge
	var where []executor.WhereItem
	where = append(where, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where).
		GetOne(&personAge)
	if err != nil {
		panic(name + "testGroupBy" + "found err")
	}
}

func testHaving(name string, db *sql.DB) {
	var listByHaving []PersonAge

	var where3 []executor.WhereItem
	where3 = append(where3, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})

	var having []executor.WhereItem
	having = append(having, executor.WhereItem{Field: "age_count", Opt: executor.Gt, Val: 4})

	err := aorm.Use(db).Debug(false).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where3).
		HavingArr(having).
		GetMany(&listByHaving)
	if err != nil {
		panic(name + "testHaving" + "found err")
	}
}

func testOrderBy(name string, db *sql.DB) {
	var listByOrder []Person
	var where []executor.WhereItem
	where = append(where, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where).
		OrderBy("age", executor.Desc).
		GetMany(&listByOrder)
	if err != nil {
		panic(name + "testOrderBy" + "found err")
	}
}

func testLimit(name string, db *sql.DB) {
	var list3 []Person
	var where1 []executor.WhereItem
	where1 = append(where1, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})
	err1 := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where1).
		Limit(50, 10).
		GetMany(&list3)
	if err1 != nil {
		panic(name + "testLimit" + "found err")
	}

	var list4 []Person
	var where2 []executor.WhereItem
	where2 = append(where2, executor.WhereItem{Field: "type", Opt: executor.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where2).
		Page(3, 10).
		GetMany(&list4)
	if err != nil {
		panic(name + "testPage" + "found err")
	}
}

func testLock(name string, db *sql.DB, id int64) {

	var itemByLock Person
	err := aorm.Use(db).Debug(false).LockForUpdate(true).Where(&Person{Id: null.IntFrom(id)}).GetOne(&itemByLock)
	if err != nil {
		panic(name + "testLock" + "found err")
	}
}

func testIncrement(name string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Increment("age", 1)
	if err != nil {
		panic(name + "testIncrement" + "found err")
	}
}

func testDecrement(name string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Decrement("age", 2)
	if err != nil {
		panic(name + "testDecrement" + "found err")
	}
}

func testValue(dbName string, db *sql.DB, id int64) {

	var name string
	errName := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Value("name", &name)
	if errName != nil {
		panic(dbName + "testValue" + "found err")
	}

	var age int64
	errAge := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Value("age", &age)
	if errAge != nil {
		panic(dbName + "testValue" + "found err")
	}

	var money float32
	errMoney := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Value("money", &money)
	if errMoney != nil {
		panic(dbName + "testValue" + "found err")
	}

	var test float64
	errTest := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Value("test", &test)
	if errTest != nil {
		panic(dbName + "testValue" + "found err")
	}
}

func testPluck(name string, db *sql.DB) {

	var nameList []string
	errNameList := aorm.Use(db).Debug(false).Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("name", &nameList)
	if errNameList != nil {
		panic(name + "testPluck" + "found err")
	}

	var ageList []int64
	errAgeList := aorm.Use(db).Debug(false).Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("age", &ageList)
	if errAgeList != nil {
		panic(name + "testPluck" + "found err:" + errAgeList.Error())
	}

	var moneyList []float32
	errMoneyList := aorm.Use(db).Debug(false).Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("money", &moneyList)
	if errMoneyList != nil {
		panic(name + "testPluck" + "found err")
	}

	var testList []float64
	errTestList := aorm.Use(db).Debug(false).Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("test", &testList)
	if errTestList != nil {
		panic(name + "testPluck" + "found err")
	}
}

func testCount(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Count("*")
	if err != nil {
		panic(name + "testCount" + "found err")
	}
}

func testSum(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Sum("age")
	if err != nil {
		panic(name + "testSum" + "found err")
	}
}

func testAvg(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Avg("age")
	if err != nil {
		panic(name + "testAvg" + "found err")
	}
}

func testMin(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Min("age")
	if err != nil {
		panic(name + "testMin" + "found err")
	}
}

func testMax(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Max("age")
	if err != nil {
		panic(name + "testMax" + "found err")
	}
}

func testExec(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Exec("UPDATE person SET name = ? WHERE id=?", "Bob", 3)
	if err != nil {
		panic(name + "testExec" + "found err")
	}
}

func testTransaction(name string, db *sql.DB) {

	tx, _ := db.Begin()

	id, errInsert := aorm.Use(tx).Debug(false).Insert(&Person{
		Name: null.StringFrom("Alice"),
	})

	if errInsert != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	_, errCount := aorm.Use(tx).Debug(false).Where(&Person{
		Id: null.IntFrom(id),
	}).Count("*")
	if errCount != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	var person Person
	errPerson := aorm.Use(tx).Debug(false).Where(&Person{
		Id: null.IntFrom(id),
	}).GetOne(&person)
	if errPerson != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	_, errUpdate := aorm.Use(tx).Debug(false).Where(&Person{
		Id: null.IntFrom(id),
	}).Update(&Person{
		Name: null.StringFrom("Bob"),
	})

	if errUpdate != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	tx.Commit()
}

func testTruncate(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Table("person").Truncate()
	if err != nil {
		panic(name + "testTruncate" + "found err")
	}
}

func testHelper(name string, db *sql.DB) {

	var list2 []ArticleVO
	var where2 []executor.WhereItem
	where2 = append(where2, executor.WhereItem{Field: "o.type", Opt: executor.Eq, Val: 0})
	where2 = append(where2, executor.WhereItem{Field: "p.age", Opt: executor.In, Val: []int{18, 20}})
	err := aorm.Use(db).Debug(false).
		Table("article o").
		LeftJoin("person p", helper.Ul("p.id=o.personId")).
		Select("o.*").
		Select(helper.Ul("p.name as personName")).
		WhereArr(where2).
		GetMany(&list2)
	if err != nil {
		panic(name + "testHelper" + "found err")
	}
}
