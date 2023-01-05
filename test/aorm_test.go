package test

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tangpanqing/aorm"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"github.com/tangpanqing/aorm/null"
	"testing"
	"time"
)

type Article struct {
	Id          null.Int    `aorm:"primary;auto_increment" json:"id"`
	Type        null.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	ArticleBody null.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

type ArticleVO struct {
	Id          null.Int    `aorm:"primary;auto_increment" json:"id"`
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
	dbList := make([]aorm.DbContent, 0)
	dbList = append(dbList, testSqlite3Connect())
	dbList = append(dbList, testMysqlConnect())
	dbList = append(dbList, testPostgresConnect())
	dbList = append(dbList, testMssqlConnect())

	for i := 0; i < len(dbList); i++ {
		dbItem := dbList[i]

		testMigrate(dbItem.DriverName, dbItem.DbLink)

		testShowCreateTable(dbItem.DriverName, dbItem.DbLink)

		id := testInsert(dbItem.DriverName, dbItem.DbLink)
		testInsertBatch(dbItem.DriverName, dbItem.DbLink)

		testGetOne(dbItem.DriverName, dbItem.DbLink, id)
		testGetMany(dbItem.DriverName, dbItem.DbLink)
		testUpdate(dbItem.DriverName, dbItem.DbLink, id)
		isExists := testExists(dbItem.DriverName, dbItem.DbLink, id)
		if isExists != true {
			panic("应该存在，但是数据库不存在")
		}

		testDelete(dbItem.DriverName, dbItem.DbLink, id)
		isExists2 := testExists(dbItem.DriverName, dbItem.DbLink, id)
		if isExists2 == true {
			panic("应该不存在，但是数据库存在")
		}

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
}

func testSqlite3Connect() aorm.DbContent {
	sqlite3Content, sqlite3Err := aorm.Open("sqlite3", "test.db")
	if sqlite3Err != nil {
		panic(sqlite3Err)
	}
	return sqlite3Content
}

func testMysqlConnect() aorm.DbContent {
	username := "root"
	password := "root"
	hostname := "localhost"
	port := "3306"
	dbname := "database_name"

	mysqlContent, mysqlErr := aorm.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
	if mysqlErr != nil {
		panic(mysqlErr)
	}

	err := mysqlContent.DbLink.Ping()
	if err != nil {
		panic(err)
	}

	return mysqlContent
}

func testPostgresConnect() aorm.DbContent {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "root", "postgres")

	postgresContent, postgresErr := aorm.Open("postgres", psqlInfo)
	if postgresErr != nil {
		panic(postgresErr)
	}

	err := postgresContent.DbLink.Ping()
	if err != nil {
		panic(err)
	}

	return postgresContent
}

func testMssqlConnect() aorm.DbContent {
	mssqlInfo := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", "localhost", "database_name", "sa", "root", 1433)
	mssqlContent, mssqlErr := aorm.Open("mssql", mssqlInfo)
	if mssqlErr != nil {
		panic(mssqlErr)
	}

	err := mssqlContent.DbLink.Ping()
	if err != nil {
		panic(err)
	}

	return mssqlContent
}

func testMigrate(driver string, db *sql.DB) {
	//AutoMigrate
	aorm.Migrator(db).Driver(driver).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").AutoMigrate(&Person{})
	aorm.Migrator(db).Driver(driver).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "文章").AutoMigrate(&Article{})

	//Migrate
	aorm.Migrator(db).Driver(driver).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").Migrate("person_1", &Person{})
}

func testShowCreateTable(driver string, db *sql.DB) {
	aorm.Migrator(db).Driver(driver).ShowCreateTable("person")
}

func testInsert(driver string, db *sql.DB) int64 {
	obj := Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(false),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(1),
		Test:       null.FloatFrom(2),
	}

	id, errInsert := aorm.Use(db).Debug(false).Driver(driver).Insert(&obj)
	if errInsert != nil {
		panic(driver + " testInsert " + "found err: " + errInsert.Error())
	}
	aorm.Use(db).Debug(false).Driver(driver).Insert(&Article{
		Type:        null.IntFrom(0),
		PersonId:    null.IntFrom(id),
		ArticleBody: null.StringFrom("文章内容"),
	})

	var person Person
	err := aorm.Use(db).Table("person").Debug(false).Driver(driver).WhereEq("id", id).OrderBy("id", "DESC").GetOne(&person)
	if err != nil {
		fmt.Println(err.Error())
	}

	if obj.Name.String != person.Name.String {
		fmt.Println(driver + ",Name not match, expected: " + obj.Name.String + " ,but real is : " + person.Name.String)
	}

	if obj.Sex.Bool != person.Sex.Bool {
		fmt.Println(driver + ",Sex not match, expected: " + fmt.Sprintf("%v", obj.Sex.Bool) + " ,but real is : " + fmt.Sprintf("%v", person.Sex.Bool))
	}

	if obj.Age.Int64 != person.Age.Int64 {
		fmt.Println(driver + ",Age not match, expected: " + fmt.Sprintf("%v", obj.Age.Int64) + " ,but real is : " + fmt.Sprintf("%v", person.Age.Int64))
	}

	if obj.Type.Int64 != person.Type.Int64 {
		fmt.Println(driver + ",Type not match, expected: " + fmt.Sprintf("%v", obj.Type.Int64) + " ,but real is : " + fmt.Sprintf("%v", person.Type.Int64))
	}

	if obj.Money.Float64 != person.Money.Float64 {
		fmt.Println(driver + ",Money not match, expected: " + fmt.Sprintf("%v", obj.Money.Float64) + " ,but real is : " + fmt.Sprintf("%v", person.Money.Float64))
	}

	if obj.Test.Float64 != person.Test.Float64 {
		fmt.Println(driver + ",Test not match, expected: " + fmt.Sprintf("%v", obj.Test.Float64) + " ,but real is : " + fmt.Sprintf("%v", person.Test.Float64))
	}

	return id
}

func testInsertBatch(driver string, db *sql.DB) int64 {
	var batch []Person
	batch = append(batch, Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(false),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	batch = append(batch, Person{
		Name:       null.StringFrom("Bob"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	count, err := aorm.Use(db).Debug(false).Driver(driver).InsertBatch(&batch)
	if err != nil {
		panic(driver + " testInsertBatch " + "found err:" + err.Error())
	}

	return count
}

func testGetOne(driver string, db *sql.DB, id int64) {
	var person Person
	errFind := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Id: null.IntFrom(id)}).GetOne(&person)
	if errFind != nil {
		panic(driver + "testGetOne" + "found err")
	}
}

func testGetMany(driver string, db *sql.DB) {
	var list []Person
	errSelect := aorm.Use(db).Driver(driver).Debug(false).Where(&Person{Type: null.IntFrom(0)}).GetMany(&list)
	if errSelect != nil {
		panic(driver + " testGetMany " + "found err:" + errSelect.Error())
	}
}

func testUpdate(driver string, db *sql.DB, id int64) {
	_, errUpdate := aorm.Use(db).Debug(false).Driver(driver).Where(&Person{Id: null.IntFrom(id)}).Update(&Person{Name: null.StringFrom("Bob")})
	if errUpdate != nil {
		panic(driver + "testGetMany" + "found err")
	}
}

func testDelete(driver string, db *sql.DB, id int64) {
	_, errDelete := aorm.Use(db).Driver(driver).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Delete()
	if errDelete != nil {
		panic(driver + "testDelete" + "found err")
	}
}

func testExists(driver string, db *sql.DB, id int64) bool {
	exists, err := aorm.Use(db).Driver(driver).Debug(false).Where(&Person{Id: null.IntFrom(id)}).OrderBy("id", "DESC").Exists()
	if err != nil {
		panic(driver + " testExists " + "found err:" + err.Error())
	}
	return exists
}

func testTable(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Driver(driver).Table("person_1").Insert(&Person{Name: null.StringFrom("Cherry")})
	if err != nil {
		panic(driver + " testTable " + "found err:" + err.Error())
	}
}

func testSelect(driver string, db *sql.DB) {
	var listByFiled []Person
	err := aorm.Use(db).Debug(false).Driver(driver).Select("name,age").Where(&Person{Age: null.IntFrom(18)}).GetMany(&listByFiled)
	if err != nil {
		panic(driver + " testSelect " + "found err:" + err.Error())
	}
}

func testSelectWithSub(driver string, db *sql.DB) {
	var listByFiled []PersonWithArticleCount

	sub := aorm.Sub().Table("article").SelectCount("id", "article_count_tem").WhereRaw("person_id", "=person.id")
	err := aorm.Use(db).Debug(false).
		Driver(driver).
		SelectExp(&sub, "article_count").
		Select("*").
		Where(&Person{Age: null.IntFrom(18)}).
		GetMany(&listByFiled)

	if err != nil {
		panic(driver + " testSelectWithSub " + "found err:" + err.Error())
	}
}

func testWhereWithSub(driver string, db *sql.DB) {
	var listByFiled []Person

	sub := aorm.Sub().Table("article").Select("person_id").GroupBy("person_id").HavingGt("count(person_id)", 0)

	err := aorm.Use(db).Debug(false).
		Table("person").
		Driver(driver).
		WhereIn("id", &sub).
		GetMany(&listByFiled)

	if err != nil {
		panic(driver + " testWhereWithSub " + "found err:" + err.Error())
	}
}

func testWhere(driver string, db *sql.DB) {
	var listByWhere []Person

	var where1 []builder.WhereItem
	where1 = append(where1, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})
	where1 = append(where1, builder.WhereItem{Field: "age", Opt: builder.In, Val: []int{18, 20}})
	where1 = append(where1, builder.WhereItem{Field: "money", Opt: builder.Between, Val: []float64{100.1, 200.9}})
	where1 = append(where1, builder.WhereItem{Field: "money", Opt: builder.Eq, Val: 100.15})
	where1 = append(where1, builder.WhereItem{Field: "name", Opt: builder.Like, Val: []string{"%", "li", "%"}})

	err := aorm.Use(db).Debug(false).Driver(driver).Table("person").WhereArr(where1).GetMany(&listByWhere)
	if err != nil {
		panic(driver + "testWhere" + "found err")
	}
}

func testJoin(driver string, db *sql.DB) {
	var list2 []ArticleVO
	var where2 []builder.WhereItem
	where2 = append(where2, builder.WhereItem{Field: "o.type", Opt: builder.Eq, Val: 0})
	where2 = append(where2, builder.WhereItem{Field: "p.age", Opt: builder.In, Val: []int{18, 20}})
	err := aorm.Use(db).Debug(false).
		Table("article o").
		LeftJoin("person p", "p.id=o.person_id").
		Select("o.*").
		Select("p.name as person_name").
		WhereArr(where2).
		Driver(driver).
		GetMany(&list2)
	if err != nil {
		panic(driver + " testWhere " + "found err " + err.Error())
	}
}

func testGroupBy(driver string, db *sql.DB) {
	var personAge PersonAge
	var where []builder.WhereItem
	where = append(where, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where).
		Driver(driver).
		OrderBy("age", "DESC").
		GetOne(&personAge)
	if err != nil {
		panic(driver + "testGroupBy" + "found err")
	}
}

func testHaving(driver string, db *sql.DB) {
	var listByHaving []PersonAge

	var where3 []builder.WhereItem
	where3 = append(where3, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})

	var having []builder.WhereItem
	having = append(having, builder.WhereItem{Field: "count(age)", Opt: builder.Gt, Val: 4})

	err := aorm.Use(db).Debug(false).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where3).
		Driver(driver).
		OrderBy("age", "DESC").
		HavingArr(having).
		GetMany(&listByHaving)
	if err != nil {
		panic(driver + " testHaving " + "found err")
	}
}

func testOrderBy(driver string, db *sql.DB) {
	var listByOrder []Person
	var where []builder.WhereItem
	where = append(where, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where).
		OrderBy("age", builder.Desc).
		Driver(driver).
		GetMany(&listByOrder)
	if err != nil {
		panic(driver + "testOrderBy" + "found err")
	}
}

func testLimit(driver string, db *sql.DB) {
	var list3 []Person
	var where1 []builder.WhereItem
	where1 = append(where1, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})
	err1 := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where1).
		Limit(50, 10).
		Driver(driver).
		OrderBy("id", "DESC").
		GetMany(&list3)
	if err1 != nil {
		panic(driver + "testLimit" + "found err")
	}

	var list4 []Person
	var where2 []builder.WhereItem
	where2 = append(where2, builder.WhereItem{Field: "type", Opt: builder.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where2).
		Page(3, 10).
		Driver(driver).
		OrderBy("id", "DESC").
		GetMany(&list4)
	if err != nil {
		panic(driver + "testPage" + "found err")
	}
}

func testLock(driver string, db *sql.DB, id int64) {
	if driver == model.Sqlite3 || driver == model.Mssql {
		return
	}

	var itemByLock Person
	err := aorm.Use(db).
		Debug(false).
		LockForUpdate(true).
		Where(&Person{Id: null.IntFrom(id)}).
		Driver(driver).
		OrderBy("id", "DESC").
		GetOne(&itemByLock)
	if err != nil {
		panic(driver + "testLock" + "found err")
	}
}

func testIncrement(driver string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Driver(driver).Increment("age", 1)
	if err != nil {
		panic(driver + " testIncrement " + "found err:" + err.Error())
	}
}

func testDecrement(driver string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: null.IntFrom(id)}).Driver(driver).Decrement("age", 2)
	if err != nil {
		panic(driver + "testDecrement" + "found err")
	}
}

func testValue(driver string, db *sql.DB, id int64) {

	var name string
	errName := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Id: null.IntFrom(id)}).Value("name", &name)
	if errName != nil {
		panic(driver + "testValue" + "found err")
	}

	var age int64
	errAge := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Id: null.IntFrom(id)}).Value("age", &age)
	if errAge != nil {
		panic(driver + "testValue" + "found err")
	}

	var money float32
	errMoney := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Id: null.IntFrom(id)}).Value("money", &money)
	if errMoney != nil {
		panic(driver + "testValue" + "found err")
	}

	var test float64
	errTest := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Id: null.IntFrom(id)}).Value("test", &test)
	if errTest != nil {
		panic(driver + "testValue" + "found err")
	}
}

func testPluck(driver string, db *sql.DB) {

	var nameList []string
	errNameList := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("name", &nameList)
	if errNameList != nil {
		panic(driver + "testPluck" + "found err")
	}

	var ageList []int64
	errAgeList := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("age", &ageList)
	if errAgeList != nil {
		panic(driver + "testPluck" + "found err:" + errAgeList.Error())
	}

	var moneyList []float32
	errMoneyList := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("money", &moneyList)
	if errMoneyList != nil {
		panic(driver + "testPluck" + "found err")
	}

	var testList []float64
	errTestList := aorm.Use(db).Debug(false).Driver(driver).OrderBy("id", "DESC").Where(&Person{Type: null.IntFrom(0)}).Limit(0, 3).Pluck("test", &testList)
	if errTestList != nil {
		panic(driver + "testPluck" + "found err")
	}
}

func testCount(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Driver(driver).Count("*")
	if err != nil {
		panic(driver + "testCount" + "found err")
	}
}

func testSum(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Driver(driver).Sum("age")
	if err != nil {
		panic(driver + "testSum" + "found err")
	}
}

func testAvg(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Driver(driver).Avg("age")
	if err != nil {
		panic(driver + "testAvg" + "found err")
	}
}

func testMin(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Driver(driver).Min("age")
	if err != nil {
		panic(driver + "testMin" + "found err")
	}
}

func testMax(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: null.IntFrom(18)}).Driver(driver).Max("age")
	if err != nil {
		panic(driver + "testMax" + "found err")
	}
}

func testExec(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Driver(driver).Exec("UPDATE person SET name = ? WHERE id=?", "Bob", 3)
	if err != nil {
		panic(driver + "testExec" + "found err")
	}
}

func testTransaction(driver string, db *sql.DB) {
	tx, _ := db.Begin()

	id, errInsert := aorm.Use(tx).Debug(false).Driver(driver).Insert(&Person{
		Name: null.StringFrom("Alice"),
	})

	if errInsert != nil {
		tx.Rollback()
		panic(driver + " testTransaction " + "found err:" + errInsert.Error())
		return
	}

	_, errCount := aorm.Use(tx).Debug(false).Driver(driver).Where(&Person{
		Id: null.IntFrom(id),
	}).Count("*")
	if errCount != nil {
		tx.Rollback()
		panic(driver + "testTransaction" + "found err")
		return
	}

	var person Person
	errPerson := aorm.Use(tx).Debug(false).Where(&Person{
		Id: null.IntFrom(id),
	}).Driver(driver).OrderBy("id", "DESC").GetOne(&person)
	if errPerson != nil {
		tx.Rollback()
		panic(driver + "testTransaction" + "found err")
		return
	}

	_, errUpdate := aorm.Use(tx).Debug(false).Driver(driver).Where(&Person{
		Id: null.IntFrom(id),
	}).Update(&Person{
		Name: null.StringFrom("Bob"),
	})

	if errUpdate != nil {
		tx.Rollback()
		panic(driver + "testTransaction" + "found err")
		return
	}

	tx.Commit()
}

func testTruncate(driver string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Driver(driver).Table("person").Truncate()
	if err != nil {
		panic(driver + " testTruncate " + "found err")
	}
}

func testHelper(driver string, db *sql.DB) {
	var list2 []ArticleVO
	var where2 []builder.WhereItem
	where2 = append(where2, builder.WhereItem{Field: "o.type", Opt: builder.Eq, Val: 0})
	where2 = append(where2, builder.WhereItem{Field: "p.age", Opt: builder.In, Val: []int{18, 20}})
	err := aorm.Use(db).Debug(false).
		Table("article o").
		LeftJoin("person p", helper.Ul("p.id=o.personId")).
		Select("o.*").
		Select(helper.Ul("p.name as personName")).
		WhereArr(where2).
		Driver(driver).
		GetMany(&list2)
	if err != nil {
		panic(driver + "testHelper" + "found err")
	}
}
