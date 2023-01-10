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
	"github.com/tangpanqing/aorm/model"
	"github.com/tangpanqing/aorm/null"
	"testing"
	"time"
)

type Student struct {
	StudentId null.Int    `aorm:"primary;auto_increment" json:"studentId"`
	Name      null.String `aorm:"size:100;not null;comment:名字" json:"name"`
}

func (s *Student) TableOpinion() map[string]string {
	return map[string]string{
		"ENGINE":  "InnoDB",
		"COMMENT": "学生表",
	}
}

type Article struct {
	Id          null.Int    `aorm:"primary;auto_increment" json:"id"`
	Type        null.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	ArticleBody null.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

func (a *Article) TableOpinion() map[string]string {
	return map[string]string{
		"ENGINE":  "InnoDB",
		"COMMENT": "文章表",
	}
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

func (p *Person) TableOpinion() map[string]string {
	return map[string]string{
		"ENGINE":  "InnoDB",
		"COMMENT": "人员表",
	}
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

var student = Student{}
var person = Person{}
var article = Article{}
var articleVO = ArticleVO{}
var personAge = PersonAge{}
var personWithArticleCount = PersonWithArticleCount{}

func TestAll(t *testing.T) {
	aorm.Store(&person, &article, &student)
	aorm.Store(&articleVO)
	aorm.Store(&personAge, &personWithArticleCount)

	var dbList = []aorm.DbContent{
		testMysqlConnect(),
		testSqlite3Connect(),
		testPostgresConnect(),
		testMssqlConnect(),
	}

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
		testJoinWithAlias(dbItem.DriverName, dbItem.DbLink)

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

		testDistinct(dbItem.DriverName, dbItem.DbLink)

		testExec(dbItem.DriverName, dbItem.DbLink)

		testTransaction(dbItem.DriverName, dbItem.DbLink)
		testTruncate(dbItem.DriverName, dbItem.DbLink)
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
	aorm.Migrator(db).Driver(driver).AutoMigrate(&person, &article, &student)

	aorm.Migrator(db).Driver(driver).Migrate("person_1", &person)
}

func testShowCreateTable(driver string, db *sql.DB) {
	aorm.Migrator(db).Driver(driver).ShowCreateTable("person")
}

func testInsert(driver string, db *sql.DB) int64 {
	obj := Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(1),
		Test:       null.FloatFrom(2),
	}

	id, errInsert := aorm.Db(db).Debug(false).Driver(driver).Insert(&obj)
	if errInsert != nil {
		panic(driver + " testInsert " + "found err: " + errInsert.Error())
	}
	aorm.Db(db).Debug(false).Driver(driver).Insert(&Article{
		Type:        null.IntFrom(0),
		PersonId:    null.IntFrom(id),
		ArticleBody: null.StringFrom("文章内容"),
	})

	var personItem Person
	err := aorm.Db(db).Table(&person).Debug(false).Driver(driver).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
	if err != nil {
		fmt.Println(err.Error())
	}

	if obj.Name.String != personItem.Name.String {
		fmt.Println(driver + ",Name not match, expected: " + obj.Name.String + " ,but real is : " + personItem.Name.String)
	}

	if obj.Sex.Bool != personItem.Sex.Bool {
		fmt.Println(driver + ",Sex not match, expected: " + fmt.Sprintf("%v", obj.Sex.Bool) + " ,but real is : " + fmt.Sprintf("%v", personItem.Sex.Bool))
	}

	if obj.Age.Int64 != personItem.Age.Int64 {
		fmt.Println(driver + ",Age not match, expected: " + fmt.Sprintf("%v", obj.Age.Int64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Age.Int64))
	}

	if obj.Type.Int64 != personItem.Type.Int64 {
		fmt.Println(driver + ",Type not match, expected: " + fmt.Sprintf("%v", obj.Type.Int64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Type.Int64))
	}

	if obj.Money.Float64 != personItem.Money.Float64 {
		fmt.Println(driver + ",Money not match, expected: " + fmt.Sprintf("%v", obj.Money.Float64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Money.Float64))
	}

	if obj.Test.Float64 != personItem.Test.Float64 {
		fmt.Println(driver + ",Test not match, expected: " + fmt.Sprintf("%v", obj.Test.Float64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Test.Float64))
	}

	//测试非id主键
	aorm.Db(db).Debug(false).Driver(driver).Insert(&Student{
		Name: null.StringFrom("new student"),
	})

	return id
}

func testInsertBatch(driver string, db *sql.DB) int64 {
	var batch []*Person
	batch = append(batch, &Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(false),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	batch = append(batch, &Person{
		Name:       null.StringFrom("Bob"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(100.15),
		Test:       null.FloatFrom(200.15987654321987654321),
	})

	count, err := aorm.Db(db).Debug(false).Driver(driver).InsertBatch(&batch)
	if err != nil {
		panic(driver + " testInsertBatch " + "found err:" + err.Error())
	}

	return count
}

func testGetOne(driver string, db *sql.DB, id int64) {
	var personItem Person
	errFind := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).GetOne(&personItem)
	if errFind != nil {
		panic(driver + "testGetOne" + "found err")
	}
}

func testGetMany(driver string, db *sql.DB) {
	var list []Person
	errSelect := aorm.Db(db).Driver(driver).Debug(false).Table(&person).WhereEq(&person.Type, 0).GetMany(&list)
	if errSelect != nil {
		panic(driver + " testGetMany " + "found err:" + errSelect.Error())
	}
}

func testUpdate(driver string, db *sql.DB, id int64) {
	_, errUpdate := aorm.Db(db).Debug(false).Driver(driver).WhereEq(&person.Id, id).Update(&Person{Name: null.StringFrom("Bob")})
	if errUpdate != nil {
		panic(driver + "testUpdate" + "found err")
	}
}

func testDelete(driver string, db *sql.DB, id int64) {
	_, errDelete := aorm.Db(db).Driver(driver).Debug(false).Table(&person).WhereEq(&person.Id, id).Delete()
	if errDelete != nil {
		panic(driver + "testDelete" + "found err")
	}

	_, errDelete2 := aorm.Db(db).Driver(driver).Debug(false).Delete(&Person{
		Id: null.IntFrom(id),
	})
	if errDelete2 != nil {
		panic(driver + "testDelete" + "found err")
	}
}

func testExists(driver string, db *sql.DB, id int64) bool {
	exists, err := aorm.Db(db).Driver(driver).Debug(false).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).Exists()
	if err != nil {
		panic(driver + " testExists " + "found err:" + err.Error())
	}
	return exists
}

func testTable(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Driver(driver).Table("person_1").Insert(&Person{Name: null.StringFrom("Cherry")})
	if err != nil {
		panic(driver + " testTable " + "found err:" + err.Error())
	}

	_, err2 := aorm.Db(db).Debug(false).Driver(driver).Table(&person).Insert(&Person{Name: null.StringFrom("Cherry")})
	if err2 != nil {
		panic(driver + " testTable " + "found err:" + err2.Error())
	}
}

func testSelect(driver string, db *sql.DB) {
	var listByFiled []Person
	err := aorm.Db(db).Debug(false).Driver(driver).Table(&person).Select(&person.Name).Select(&person.Age).WhereEq(&person.Age, 18).GetMany(&listByFiled)
	if err != nil {
		panic(driver + " testSelect " + "found err:" + err.Error())
	}
}

func testSelectWithSub(driver string, db *sql.DB) {
	var listByFiled []PersonWithArticleCount

	sub := aorm.Db().Table(&article).SelectCount(&article.Id, "article_count_tem").WhereRawEq(&article.PersonId, &person.Id)
	err := aorm.Db(db).Debug(false).
		Driver(driver).
		SelectExp(&sub, &personWithArticleCount.ArticleCount).
		SelectAll(&person).
		Table(&person).
		WhereEq(&person.Age, 18).
		GetMany(&listByFiled)

	if err != nil {
		panic(driver + " testSelectWithSub " + "found err:" + err.Error())
	}
}

func testWhereWithSub(driver string, db *sql.DB) {
	var listByFiled []Person
	sub := aorm.Db().Table(&article).Driver(driver).SelectCount(&article.PersonId, "count_person_id").GroupBy(&article.PersonId).HavingGt("count_person_id", 0)
	err := aorm.Db(db).Debug(false).
		Table(&person).
		Driver(driver).
		WhereIn(&person.Id, &sub).
		GetMany(&listByFiled)
	if err != nil {
		panic(driver + " testWhereWithSub " + "found err:" + err.Error())
	}
}

func testWhere(driver string, db *sql.DB) {
	var listByWhere []Person
	err := aorm.Db(db).Debug(false).Driver(driver).Table(&person).WhereArr([]builder.WhereItem{
		builder.GenWhereItem(&person.Type, builder.Eq, 0),
		builder.GenWhereItem(&person.Age, builder.In, []int{18, 20}),
		builder.GenWhereItem(&person.Money, builder.Between, []float64{100.1, 200.9}),
		builder.GenWhereItem(&person.Money, builder.Eq, 100.15),
		builder.GenWhereItem(&person.Name, builder.Like, []string{"%", "li", "%"}),
	}).GetMany(&listByWhere)
	if err != nil {
		panic(driver + "testWhere" + "found err")
	}
}

func testJoin(driver string, db *sql.DB) {
	var list2 []ArticleVO
	err := aorm.Db(db).Debug(false).Driver(driver).
		Table(&article).
		LeftJoin(
			&person,
			[]builder.JoinCondition{
				builder.GenJoinCondition(&person.Id, builder.RawEq, &article.PersonId),
			},
		).
		SelectAll(&article).
		SelectAs(&person.Name, &articleVO.PersonName).
		WhereEq(&article.Type, 0).
		WhereIn(&person.Age, []int{18, 20}).
		GetMany(&list2)
	if err != nil {
		panic(driver + " testWhere " + "found err " + err.Error())
	}
}

func testJoinWithAlias(driver string, db *sql.DB) {
	var list2 []ArticleVO
	err := aorm.Db(db).Debug(false).Driver(driver).
		Table(&article, "o").
		LeftJoin(
			&person,
			[]builder.JoinCondition{
				builder.GenJoinCondition(&person.Id, builder.RawEq, &article.PersonId, "o"),
			},
			"p",
		).
		Select("*", "o").
		SelectAs(&person.Name, &articleVO.PersonName, "p").
		WhereEq(&article.Type, 0, "o").
		WhereIn(&person.Age, []int{18, 20}, "p").
		GetMany(&list2)
	if err != nil {
		panic(driver + " testWhere " + "found err " + err.Error())
	}
}

func testGroupBy(driver string, db *sql.DB) {
	var personAgeItem PersonAge
	err := aorm.Db(db).Debug(false).
		Table(&person).
		Select(&person.Age).
		SelectCount(&person.Age, &personAge.AgeCount).
		GroupBy(&person.Age).
		WhereEq(&person.Type, 0).
		Driver(driver).
		OrderBy(&person.Age, builder.Desc).
		GetOne(&personAgeItem)
	if err != nil {
		panic(driver + "testGroupBy" + "found err")
	}
}

func testHaving(driver string, db *sql.DB) {
	var listByHaving []PersonAge

	err := aorm.Db(db).Debug(false).Driver(driver).
		Table(&person).
		Select(&person.Age).
		SelectCount(&person.Age, &personAge.AgeCount).
		GroupBy(&person.Age).
		WhereEq(&person.Type, 0).
		OrderBy(&person.Age, builder.Desc).
		HavingGt(&personAge.AgeCount, 4).
		GetMany(&listByHaving)
	if err != nil {
		panic(driver + " testHaving " + "found err")
	}
}

func testOrderBy(driver string, db *sql.DB) {
	var listByOrder []Person
	err := aorm.Db(db).Debug(false).Driver(driver).
		Table(&person).
		WhereEq(&person.Type, 0).
		OrderBy(&person.Age, builder.Desc).
		GetMany(&listByOrder)
	if err != nil {
		panic(driver + "testOrderBy" + "found err")
	}

	var listByOrder2 []Person
	err2 := aorm.Db(db).Debug(false).Driver(driver).
		Table(&person, "o").
		WhereEq(&person.Type, 0, "o").
		OrderBy(&person.Age, builder.Desc, "o").
		GetMany(&listByOrder2)
	if err2 != nil {
		panic(driver + "testOrderBy" + "found err")
	}
}

func testLimit(driver string, db *sql.DB) {
	var list3 []Person
	err1 := aorm.Db(db).Debug(false).
		Table(&person).
		WhereEq(&person.Type, 0).
		Limit(50, 10).
		Driver(driver).
		OrderBy(&person.Id, builder.Desc).
		GetMany(&list3)
	if err1 != nil {
		panic(driver + "testLimit" + "found err")
	}

	var list4 []Person
	err := aorm.Db(db).Debug(false).
		Driver(driver).
		Table(&person).
		WhereEq(&person.Type, 0).
		Page(3, 10).
		OrderBy(&person.Id, builder.Desc).
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
	err := aorm.Db(db).
		Debug(false).
		LockForUpdate(true).
		Table(&person).
		WhereEq(&person.Id, id).
		Driver(driver).
		OrderBy(&person.Id, builder.Desc).
		GetOne(&itemByLock)
	if err != nil {
		panic(driver + "testLock" + "found err")
	}
}

func testIncrement(driver string, db *sql.DB, id int64) {
	_, err := aorm.Db(db).Debug(false).Driver(driver).Table(&person).WhereEq(&person.Id, id).Increment(&person.Age, 1)
	if err != nil {
		panic(driver + " testIncrement " + "found err:" + err.Error())
	}
}

func testDecrement(driver string, db *sql.DB, id int64) {
	_, err := aorm.Db(db).Debug(false).Driver(driver).Table(&person).WhereEq(&person.Id, id).Decrement(&person.Age, 2)
	if err != nil {
		panic(driver + "testDecrement" + "found err")
	}
}

func testValue(driver string, db *sql.DB, id int64) {

	var name string
	errName := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Name, &name)
	if errName != nil {
		panic(driver + "testValue" + "found err")
	}

	var age int64
	errAge := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Age, &age)
	if errAge != nil {
		panic(driver + "testValue" + "found err")
	}

	var money float32
	errMoney := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Money, &money)
	if errMoney != nil {
		panic(driver + "testValue" + "found err")
	}

	var test float64
	errTest := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Test, &test)
	if errTest != nil {
		panic(driver + "testValue" + "found err")
	}
}

func testPluck(driver string, db *sql.DB) {
	var nameList []string
	errNameList := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Name, &nameList)
	if errNameList != nil {
		panic(driver + "testPluck" + "found err")
	}

	var ageList []int64
	errAgeList := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Age, &ageList)
	if errAgeList != nil {
		panic(driver + "testPluck" + "found err:" + errAgeList.Error())
	}

	var moneyList []float32
	errMoneyList := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Money, &moneyList)
	if errMoneyList != nil {
		panic(driver + "testPluck" + "found err")
	}

	var testList []float64
	errTestList := aorm.Db(db).Debug(false).Driver(driver).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Test, &testList)
	if errTestList != nil {
		panic(driver + "testPluck" + "found err")
	}
}

func testCount(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Table(&person).WhereEq(&person.Age, 18).Driver(driver).Count("*")
	if err != nil {
		panic(driver + "testCount" + "found err")
	}
}

func testSum(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Table(&person).WhereEq(&person.Age, 18).Driver(driver).Sum(&person.Age)
	if err != nil {
		panic(driver + "testSum" + "found err")
	}
}

func testAvg(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Table(&person).WhereEq(&person.Age, 18).Driver(driver).Avg(&person.Age)
	if err != nil {
		panic(driver + "testAvg" + "found err")
	}
}

func testMin(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Table(&person).WhereEq(&person.Age, 18).Driver(driver).Min(&person.Age)
	if err != nil {
		panic(driver + "testMin" + "found err")
	}
}

func testMax(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Table(&person).WhereEq(&person.Age, 18).Driver(driver).Max(&person.Age)
	if err != nil {
		panic(driver + "testMax" + "found err")
	}
}

func testDistinct(driver string, db *sql.DB) {
	var listByFiled []Person
	err := aorm.Db(db).Debug(false).Driver(driver).Distinct(true).Table(&person).Select(&person.Name).Select(&person.Age).WhereEq(&person.Age, 18).GetMany(&listByFiled)
	if err != nil {
		panic(driver + " testSelect " + "found err:" + err.Error())
	}
}

func testExec(driver string, db *sql.DB) {
	_, err := aorm.Db(db).Debug(false).Driver(driver).Exec("UPDATE person SET name = ? WHERE person.id=?", "Bob", 3)
	if err != nil {
		panic(driver + "testExec" + "found err")
	}
}

func testTransaction(driver string, db *sql.DB) {
	tx, _ := db.Begin()

	id, errInsert := aorm.Db(tx).Debug(false).Driver(driver).Insert(&Person{
		Name: null.StringFrom("Alice"),
	})

	if errInsert != nil {
		tx.Rollback()
		panic(driver + " testTransaction " + "found err:" + errInsert.Error())
		return
	}

	_, errCount := aorm.Db(tx).Debug(false).Driver(driver).Table(&person).WhereEq(&person.Id, id).Count("*")
	if errCount != nil {
		tx.Rollback()
		panic(driver + "testTransaction" + "found err")
		return
	}

	var personItem Person
	errPerson := aorm.Db(tx).Debug(false).Driver(driver).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
	if errPerson != nil {
		tx.Rollback()
		panic(driver + "testTransaction" + "found err")
		return
	}

	_, errUpdate := aorm.Db(tx).Debug(false).Driver(driver).Where(&Person{
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
	_, err := aorm.Db(db).Debug(false).Driver(driver).Table(&person).Truncate()
	if err != nil {
		panic(driver + " testTruncate " + "found err")
	}
}
