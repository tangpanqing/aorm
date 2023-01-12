package test

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tangpanqing/aorm"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/driver"
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
	Type        null.Int    `aorm:"index;comment:类型" json:"driver"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	ArticleBody null.String `aorm:"driver:text;comment:文章内容" json:"articleBody"`
}

func (a *Article) TableOpinion() map[string]string {
	return map[string]string{
		"ENGINE":  "InnoDB",
		"COMMENT": "文章表",
	}
}

type ArticleVO struct {
	Id          null.Int    `aorm:"primary;auto_increment" json:"id"`
	Type        null.Int    `aorm:"index;comment:类型" json:"driver"`
	PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
	PersonName  null.String `aorm:"comment:人员名称" json:"personName"`
	ArticleBody null.String `aorm:"driver:text;comment:文章内容" json:"articleBody"`
}

type Person struct {
	Id         null.Int    `aorm:"primary;auto_increment" json:"id"`
	Name       null.String `aorm:"size:100;not null;comment:名字" json:"name"`
	Sex        null.Bool   `aorm:"index;comment:性别" json:"sex"`
	Age        null.Int    `aorm:"index;comment:年龄" json:"age"`
	Type       null.Int    `aorm:"index;comment:类型" json:"driver"`
	CreateTime null.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money      null.Float  `aorm:"comment:金额" json:"money"`
	Test       null.Float  `aorm:"driver:double;comment:测试" json:"test"`
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
	Type         null.Int    `aorm:"index;comment:类型" json:"driver"`
	CreateTime   null.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money        null.Float  `aorm:"comment:金额" json:"money"`
	Test         null.Float  `aorm:"driver:double;comment:测试" json:"test"`
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

	var dbList = []*model.AormDB{
		testMysqlConnect(),
		testSqlite3Connect(),
		testPostgresConnect(),
		testMssqlConnect(),
	}
	defer closeAll(dbList)

	for i := 0; i < len(dbList); i++ {
		dbItem := dbList[i]

		testMigrate(dbItem)
		testShowCreateTable(dbItem)

		id := testInsert(dbItem)
		testInsertBatch(dbItem)
		testGetOne(dbItem, id)
		testGetMany(dbItem)
		testUpdate(dbItem, id)

		isExists := testExists(dbItem, id)
		if isExists != true {
			panic("应该存在，但是数据库不存在")
		}

		testDelete(dbItem, id)
		isExists2 := testExists(dbItem, id)
		if isExists2 == true {
			panic("应该不存在，但是数据库存在")
		}

		id2 := testInsert(dbItem)
		testTable(dbItem)
		testSelect(dbItem)
		testSelectWithSub(dbItem)
		testWhereWithSub(dbItem)
		testWhere(dbItem)
		testJoin(dbItem)
		testJoinWithAlias(dbItem)

		testGroupBy(dbItem)
		testHaving(dbItem)
		testOrderBy(dbItem)
		testLimit(dbItem)
		testLock(dbItem, id2)

		testIncrement(dbItem, id2)
		testDecrement(dbItem, id2)

		testValue(dbItem, id2)
		testPluck(dbItem)

		testCount(dbItem)
		testSum(dbItem)
		testAvg(dbItem)
		testMin(dbItem)
		testMax(dbItem)

		testDistinct(dbItem)
		testRawSql(dbItem, id2)

		testTransaction(dbItem)
		testTruncate(dbItem)

	}

	testPreview()
	testDbContent()
}

func testSqlite3Connect() *model.AormDB {
	sqlite3Content, sqlite3Err := aorm.Open(driver.Sqlite3, "test.db")
	if sqlite3Err != nil {
		panic(sqlite3Err)
	}

	sqlite3Content.SetDebugMode(false)
	return sqlite3Content
}

func testMysqlConnect() *model.AormDB {
	username := "root"
	password := "root"
	hostname := "localhost"
	port := "3306"
	dbname := "database_name"

	mysqlContent, mysqlErr := aorm.Open(driver.Mysql, username+":"+password+"@tcp("+hostname+":"+port+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
	if mysqlErr != nil {
		panic(mysqlErr)
	}

	mysqlContent.SetDebugMode(false)
	return mysqlContent
}

func testPostgresConnect() *model.AormDB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "root", "postgres")

	postgresContent, postgresErr := aorm.Open(driver.Postgres, psqlInfo)
	if postgresErr != nil {
		panic(postgresErr)
	}

	postgresContent.SetDebugMode(false)

	return postgresContent
}

func testMssqlConnect() *model.AormDB {
	mssqlInfo := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", "localhost", "database_name", "sa", "root", 1433)

	mssqlContent, mssqlErr := aorm.Open(driver.Mssql, mssqlInfo)
	if mssqlErr != nil {
		panic(mssqlErr)
	}

	mssqlContent.SetDebugMode(false)
	return mssqlContent
}

func testMigrate(db *model.AormDB) {
	aorm.Migrator(db).AutoMigrate(&person, &article, &student)

	aorm.Migrator(db).Migrate("person_1", &person)
}

func testShowCreateTable(db *model.AormDB) {
	aorm.Migrator(db).ShowCreateTable("person")
}

func testInsert(db *model.AormDB) int64 {
	obj := Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(1),
		Test:       null.FloatFrom(2),
	}

	id, errInsert := aorm.Db(db).Insert(&obj)
	if errInsert != nil {
		panic(db.DriverName() + " testInsert " + "found err: " + errInsert.Error())
	}
	aorm.Db(db).Insert(&Article{
		Type:        null.IntFrom(0),
		PersonId:    null.IntFrom(id),
		ArticleBody: null.StringFrom("文章内容"),
	})

	var personItem Person
	err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
	if err != nil {
		fmt.Println(err.Error())
	}

	if obj.Name.String != personItem.Name.String {
		fmt.Println(db.DriverName() + ",Name not match, expected: " + obj.Name.String + " ,but real is : " + personItem.Name.String)
	}

	if obj.Sex.Bool != personItem.Sex.Bool {
		fmt.Println(db.DriverName() + ",Sex not match, expected: " + fmt.Sprintf("%v", obj.Sex.Bool) + " ,but real is : " + fmt.Sprintf("%v", personItem.Sex.Bool))
	}

	if obj.Age.Int64 != personItem.Age.Int64 {
		fmt.Println(db.DriverName() + ",Age not match, expected: " + fmt.Sprintf("%v", obj.Age.Int64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Age.Int64))
	}

	if obj.Type.Int64 != personItem.Type.Int64 {
		fmt.Println(db.DriverName() + ",Type not match, expected: " + fmt.Sprintf("%v", obj.Type.Int64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Type.Int64))
	}

	if obj.Money.Float64 != personItem.Money.Float64 {
		fmt.Println(db.DriverName() + ",Money not match, expected: " + fmt.Sprintf("%v", obj.Money.Float64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Money.Float64))
	}

	if obj.Test.Float64 != personItem.Test.Float64 {
		fmt.Println(db.DriverName() + ",Test not match, expected: " + fmt.Sprintf("%v", obj.Test.Float64) + " ,but real is : " + fmt.Sprintf("%v", personItem.Test.Float64))
	}

	//测试非id主键
	aorm.Db(db).Insert(&Student{
		Name: null.StringFrom("new student"),
	})

	return id
}

func testInsertBatch(db *model.AormDB) int64 {
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

	count, err := aorm.Db(db).InsertBatch(&batch)
	if err != nil {
		panic(db.DriverName() + " testInsertBatch " + "found err:" + err.Error())
	}

	return count
}

func testGetOne(db *model.AormDB, id int64) {
	var personItem Person
	errFind := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).GetOne(&personItem)
	if errFind != nil {
		panic(db.DriverName() + "testGetOne" + "found err")
	}
}

func testGetMany(db *model.AormDB) {
	var list []Person
	errSelect := aorm.Db(db).Table(&person).WhereEq(&person.Type, 0).GetMany(&list)
	if errSelect != nil {
		panic(db.DriverName() + " testGetMany " + "found err:" + errSelect.Error())
	}
}

func testUpdate(db *model.AormDB, id int64) {
	_, errUpdate := aorm.Db(db).WhereEq(&person.Id, id).Update(&Person{Name: null.StringFrom("Bob")})
	if errUpdate != nil {
		panic(db.DriverName() + "testUpdate" + "found err")
	}
}

func testDelete(db *model.AormDB, id int64) {
	_, errDelete := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Delete()
	if errDelete != nil {
		panic(db.DriverName() + "testDelete" + "found err")
	}

	_, errDelete2 := aorm.Db(db).Delete(&Person{
		Id: null.IntFrom(id),
	})
	if errDelete2 != nil {
		panic(db.DriverName() + "testDelete" + "found err")
	}
}

func testExists(db *model.AormDB, id int64) bool {
	exists, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).Exists()
	if err != nil {
		panic(db.DriverName() + " testExists " + "found err:" + err.Error())
	}
	return exists
}

func testTable(db *model.AormDB) {
	_, err := aorm.Db(db).Table("person_1").Insert(&Person{Name: null.StringFrom("Cherry")})
	if err != nil {
		panic(db.DriverName() + " testTable " + "found err:" + err.Error())
	}

	_, err2 := aorm.Db(db).Table(&person).Insert(&Person{Name: null.StringFrom("Cherry")})
	if err2 != nil {
		panic(db.DriverName() + " testTable " + "found err:" + err2.Error())
	}
}

func testSelect(db *model.AormDB) {
	var listByFiled []Person
	err := aorm.Db(db).Table(&person).Select(&person.Name).Select(&person.Age).WhereEq(&person.Age, 18).GetMany(&listByFiled)
	if err != nil {
		panic(db.DriverName() + " testSelect " + "found err:" + err.Error())
	}
}

func testSelectWithSub(db *model.AormDB) {
	var listByFiled []PersonWithArticleCount

	sub := aorm.Db(db).Table(&article).SelectCount(&article.Id, "article_count_tem").WhereRawEq(&article.PersonId, &person.Id)
	err := aorm.Db(db).
		SelectExp(&sub, &personWithArticleCount.ArticleCount).
		SelectAll(&person).
		Table(&person).
		WhereEq(&person.Age, 18).
		GetMany(&listByFiled)

	if err != nil {
		panic(db.DriverName() + " testSelectWithSub " + "found err:" + err.Error())
	}
}

func testWhereWithSub(db *model.AormDB) {
	var listByFiled []Person
	sub := aorm.Db(db).Table(&article).SelectCount(&article.PersonId, "count_person_id").GroupBy(&article.PersonId).HavingGt("count_person_id", 0)
	err := aorm.Db(db).
		Table(&person).
		WhereIn(&person.Id, &sub).
		GetMany(&listByFiled)
	if err != nil {
		panic(db.DriverName() + " testWhereWithSub " + "found err:" + err.Error())
	}
}

func testWhere(db *model.AormDB) {
	var listByWhere []Person
	err := aorm.Db(db).Table(&person).WhereArr([]builder.WhereItem{
		builder.GenWhereItem(&person.Type, builder.Eq, 0),
		builder.GenWhereItem(&person.Age, builder.In, []int{18, 20}),
		builder.GenWhereItem(&person.Money, builder.Between, []float64{100.1, 200.9}),
		builder.GenWhereItem(&person.Money, builder.Eq, 100.15),
		builder.GenWhereItem(&person.Name, builder.Like, []string{"%", "li", "%"}),
	}).GetMany(&listByWhere)
	if err != nil {
		panic(db.DriverName() + "testWhere" + "found err")
	}
}

func testJoin(db *model.AormDB) {
	var list2 []ArticleVO
	err := aorm.Db(db).
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
		panic(db.DriverName() + " testWhere " + "found err " + err.Error())
	}
}

func testJoinWithAlias(db *model.AormDB) {
	var list2 []ArticleVO
	err := aorm.Db(db).
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
		panic(db.DriverName() + " testWhere " + "found err " + err.Error())
	}
}

func testGroupBy(db *model.AormDB) {
	var personAgeItem PersonAge
	err := aorm.Db(db).
		Table(&person).
		Select(&person.Age).
		SelectCount(&person.Age, &personAge.AgeCount).
		GroupBy(&person.Age).
		WhereEq(&person.Type, 0).
		OrderBy(&person.Age, builder.Desc).
		GetOne(&personAgeItem)
	if err != nil {
		panic(db.DriverName() + "testGroupBy" + "found err")
	}
}

func testHaving(db *model.AormDB) {
	var listByHaving []PersonAge

	err := aorm.Db(db).
		Table(&person).
		Select(&person.Age).
		SelectCount(&person.Age, &personAge.AgeCount).
		GroupBy(&person.Age).
		WhereEq(&person.Type, 0).
		OrderBy(&person.Age, builder.Desc).
		HavingGt(&personAge.AgeCount, 4).
		GetMany(&listByHaving)
	if err != nil {
		panic(db.DriverName() + " testHaving " + "found err")
	}
}

func testOrderBy(db *model.AormDB) {
	var listByOrder []Person
	err := aorm.Db(db).
		Table(&person).
		WhereEq(&person.Type, 0).
		OrderBy(&person.Age, builder.Desc).
		GetMany(&listByOrder)
	if err != nil {
		panic(db.DriverName() + "testOrderBy" + "found err")
	}

	var listByOrder2 []Person
	err2 := aorm.Db(db).
		Table(&person, "o").
		WhereEq(&person.Type, 0, "o").
		OrderBy(&person.Age, builder.Desc, "o").
		GetMany(&listByOrder2)
	if err2 != nil {
		panic(db.DriverName() + "testOrderBy" + "found err")
	}
}

func testLimit(db *model.AormDB) {
	var list3 []Person
	err1 := aorm.Db(db).
		Table(&person).
		WhereEq(&person.Type, 0).
		Limit(50, 10).
		OrderBy(&person.Id, builder.Desc).
		GetMany(&list3)
	if err1 != nil {
		panic(db.DriverName() + "testLimit" + "found err")
	}

	var list4 []Person
	err := aorm.Db(db).
		Table(&person).
		WhereEq(&person.Type, 0).
		Page(3, 10).
		OrderBy(&person.Id, builder.Desc).
		GetMany(&list4)
	if err != nil {
		panic(db.DriverName() + "testPage" + "found err")
	}
}

func testLock(db *model.AormDB, id int64) {
	if db.DriverName() == driver.Sqlite3 || db.DriverName() == driver.Mssql {
		return
	}

	var itemByLock Person
	err := aorm.Db(db).
		LockForUpdate(true).
		Table(&person).
		WhereEq(&person.Id, id).
		OrderBy(&person.Id, builder.Desc).
		GetOne(&itemByLock)
	if err != nil {
		panic(db.DriverName() + "testLock" + "found err")
	}
}

func testIncrement(db *model.AormDB, id int64) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Increment(&person.Age, 1)
	if err != nil {
		panic(db.DriverName() + " testIncrement " + "found err:" + err.Error())
	}
}

func testDecrement(db *model.AormDB, id int64) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Decrement(&person.Age, 2)
	if err != nil {
		panic(db.DriverName() + "testDecrement" + "found err")
	}
}

func testValue(db *model.AormDB, id int64) {

	var name string
	errName := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Name, &name)
	if errName != nil {
		panic(db.DriverName() + "testValue" + "found err")
	}

	var age int64
	errAge := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Age, &age)
	if errAge != nil {
		panic(db.DriverName() + "testValue" + "found err")
	}

	var money float32
	errMoney := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Money, &money)
	if errMoney != nil {
		panic(db.DriverName() + "testValue" + "found err")
	}

	var test float64
	errTest := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Test, &test)
	if errTest != nil {
		panic(db.DriverName() + "testValue" + "found err")
	}
}

func testPluck(db *model.AormDB) {
	var nameList []string
	errNameList := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Name, &nameList)
	if errNameList != nil {
		panic(db.DriverName() + "testPluck" + "found err")
	}

	var ageList []int64
	errAgeList := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Age, &ageList)
	if errAgeList != nil {
		panic(db.DriverName() + "testPluck" + "found err:" + errAgeList.Error())
	}

	var moneyList []float32
	errMoneyList := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Money, &moneyList)
	if errMoneyList != nil {
		panic(db.DriverName() + "testPluck" + "found err")
	}

	var testList []float64
	errTestList := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Test, &testList)
	if errTestList != nil {
		panic(db.DriverName() + "testPluck" + "found err")
	}
}

func testCount(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Count("*")
	if err != nil {
		panic(db.DriverName() + "testCount" + "found err")
	}
}

func testSum(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Sum(&person.Age)
	if err != nil {
		panic(db.DriverName() + "testSum" + "found err")
	}
}

func testAvg(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Avg(&person.Age)
	if err != nil {
		panic(db.DriverName() + "testAvg" + "found err")
	}
}

func testMin(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Min(&person.Age)
	if err != nil {
		panic(db.DriverName() + "testMin" + "found err")
	}
}

func testMax(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Max(&person.Age)
	if err != nil {
		panic(db.DriverName() + "testMax" + "found err")
	}
}

func testDistinct(db *model.AormDB) {
	var listByFiled []Person
	err := aorm.Db(db).Distinct(true).Table(&person).Select(&person.Name).Select(&person.Age).WhereEq(&person.Age, 18).GetMany(&listByFiled)
	if err != nil {
		panic(db.DriverName() + " testSelect " + "found err:" + err.Error())
	}
}

func testRawSql(db *model.AormDB, id2 int64) {
	var list []Person
	err1 := aorm.Db(db).RawSql("SELECT * FROM person WHERE id=?", id2).GetMany(&list)
	if err1 != nil {
		panic(err1)
	}

	_, err := aorm.Db(db).RawSql("UPDATE person SET name = ? WHERE id=?", "Bob2", id2).Exec()
	if err != nil {
		panic(db.DriverName() + "testRawSql" + "found err")
	}
}

func testTransaction(db *model.AormDB) {
	tx := db.Begin()

	id, errInsert := aorm.Db(tx).Insert(&Person{
		Name: null.StringFrom("Alice"),
	})

	if errInsert != nil {
		tx.Rollback()
		panic(db.DriverName() + " testTransaction " + "found err:" + errInsert.Error())
		return
	}

	_, errCount := aorm.Db(tx).Table(&person).WhereEq(&person.Id, id).Count("*")
	if errCount != nil {
		tx.Rollback()
		panic(db.DriverName() + "testTransaction" + "found err")
		return
	}

	var personItem Person
	errPerson := aorm.Db(tx).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
	if errPerson != nil {
		tx.Rollback()
		panic(db.DriverName() + "testTransaction" + "found err")
		return
	}

	_, errUpdate := aorm.Db(tx).Where(&Person{
		Id: null.IntFrom(id),
	}).Update(&Person{
		Name: null.StringFrom("Bob"),
	})

	if errUpdate != nil {
		tx.Rollback()
		panic(db.DriverName() + "testTransaction" + "found err")
		return
	}

	tx.Commit()
}

func testTruncate(db *model.AormDB) {
	_, err := aorm.Db(db).Table(&person).Truncate()
	if err != nil {
		panic(db.DriverName() + " testTruncate " + "found err")
	}
}

func testPreview() {

	//Content Mysql
	db, _ := aorm.Open(driver.Mysql, "root:root@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local")
	defer db.Close()

	//Insert a Person
	personId, _ := aorm.Db(db).Insert(&Person{
		Name:       null.StringFrom("Alice"),
		Sex:        null.BoolFrom(true),
		Age:        null.IntFrom(18),
		Type:       null.IntFrom(0),
		CreateTime: null.TimeFrom(time.Now()),
		Money:      null.FloatFrom(1),
		Test:       null.FloatFrom(2),
	})

	//Insert a Article
	articleId, _ := aorm.Db(db).Insert(&Article{
		Type:        null.IntFrom(0),
		PersonId:    null.IntFrom(personId),
		ArticleBody: null.StringFrom("文章内容"),
	})

	//GetOne
	var personItem Person
	err := aorm.Db(db).Table(&person).WhereEq(&person.Id, personId).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
	if err != nil {
		fmt.Println(err.Error())
	}

	//Join
	var list2 []ArticleVO
	aorm.
		Db(db).
		Table(&article).
		LeftJoin(&person, []builder.JoinCondition{
			builder.GenJoinCondition(&person.Id, builder.RawEq, &article.PersonId),
		}).
		SelectAll(&article).SelectAs(&person.Name, &articleVO.PersonName).
		WhereEq(&article.Id, articleId).
		GetMany(&list2)

	//Join With Alias
	var list3 []ArticleVO
	aorm.
		Db(db).
		Table(&article, "o").
		LeftJoin(&person, []builder.JoinCondition{
			builder.GenJoinCondition(&person.Id, builder.RawEq, &article.PersonId, "o"),
		}, "p").
		Select("*", "o").SelectAs(&person.Name, &articleVO.PersonName, "p").
		WhereEq(&article.Id, articleId, "o").
		GetMany(&list3)
}

func testDbContent() {
	db, err := aorm.Open(driver.Mysql, "root:root@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(5)
	db.SetDebugMode(false)
	defer db.Close()

	aorm.Db(db).Insert(&Person{
		Name: null.StringFrom("test name"),
	})

	tx := db.Begin()
	aorm.Db(tx).Insert(&Person{
		Name: null.StringFrom("test name"),
	})

	tx.Commit()
}

func closeAll(dbList []*model.AormDB) {
	for i := 0; i < len(dbList); i++ {
		dbList[i].Close()
	}
}
