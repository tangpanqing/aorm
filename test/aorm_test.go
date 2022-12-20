package test

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tangpanqing/aorm"
	"testing"
	"time"
)

type Article struct {
	Id          aorm.Int    `aorm:"primary;auto_increment;type:bigint" json:"id"`
	Type        aorm.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    aorm.Int    `aorm:"comment:人员Id" json:"personId"`
	ArticleBody aorm.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

type ArticleVO struct {
	Id          aorm.Int    `aorm:"primary;auto_increment;type:bigint" json:"id"`
	Type        aorm.Int    `aorm:"index;comment:类型" json:"type"`
	PersonId    aorm.Int    `aorm:"comment:人员Id" json:"personId"`
	PersonName  aorm.String `aorm:"comment:人员名称" json:"personName"`
	ArticleBody aorm.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
}

type Person struct {
	Id         aorm.Int    `aorm:"primary;auto_increment" json:"id"`
	Name       aorm.String `aorm:"size:100;not null;comment:名字" json:"name"`
	Sex        aorm.Bool   `aorm:"index;comment:性别" json:"sex"`
	Age        aorm.Int    `aorm:"index;comment:年龄" json:"age"`
	Type       aorm.Int    `aorm:"index;comment:类型" json:"type"`
	CreateTime aorm.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money      aorm.Float  `aorm:"comment:金额" json:"money"`
	Test       aorm.Float  `aorm:"type:double;comment:测试" json:"test"`
}

type PersonAge struct {
	Age      aorm.Int
	AgeCount aorm.Int
}

type PersonWithArticleCount struct {
	Id           aorm.Int    `aorm:"primary;auto_increment" json:"id"`
	Name         aorm.String `aorm:"size:100;not null;comment:名字" json:"name"`
	Sex          aorm.Bool   `aorm:"index;comment:性别" json:"sex"`
	Age          aorm.Int    `aorm:"index;comment:年龄" json:"age"`
	Type         aorm.Int    `aorm:"index;comment:类型" json:"type"`
	CreateTime   aorm.Time   `aorm:"comment:创建时间" json:"createTime"`
	Money        aorm.Float  `aorm:"comment:金额" json:"money"`
	Test         aorm.Float  `aorm:"type:double;comment:测试" json:"test"`
	ArticleCount aorm.Int    `aorm:"comment:文章数量" json:"articleCount"`
}

func TestAll(t *testing.T) {
	name := "mysql"
	db := testConnect()
	defer db.Close()

	testMigrate(name, db)

	testShowCreateTable(name, db)

	id := testInsert(name, db)
	testInsertBatch(name, db)

	testGetOne(name, db, id)
	testGetMany(name, db)
	testUpdate(name, db, id)
	testDelete(name, db, id)

	id2 := testInsert(name, db)
	testTable(name, db)
	testSelect(name, db)
	testSelectWithSub(name, db)
	testWhereWithSub(name, db)
	testWhere(name, db)
	testJoin(name, db)
	testGroupBy(name, db)
	testHaving(name, db)
	testOrderBy(name, db)
	testLimit(name, db)
	testLock(name, db, id2)

	testIncrement(name, db, id2)
	testDecrement(name, db, id2)

	testValue(name, db, id2)

	testPluck(name, db)

	testCount(name, db)
	testSum(name, db)
	testAvg(name, db)
	testMin(name, db)
	testMax(name, db)

	testExec(name, db)

	testTransaction(name, db)
	testTruncate(name, db)
	testHelper(name, db)
}

func testConnect() *sql.DB {
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
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").AutoMigrate(&Person{})
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "文章").AutoMigrate(&Article{})

	//Migrate
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").Migrate("person_1", &Person{})
}

func testShowCreateTable(name string, db *sql.DB) {
	aorm.Use(db).ShowCreateTable("person")
}

func testInsert(name string, db *sql.DB) int64 {

	id, errInsert := aorm.Use(db).Debug(false).Insert(&Person{
		Name:       aorm.StringFrom("Alice"),
		Sex:        aorm.BoolFrom(false),
		Age:        aorm.IntFrom(18),
		Type:       aorm.IntFrom(0),
		CreateTime: aorm.TimeFrom(time.Now()),
		Money:      aorm.FloatFrom(100.15987654321),
		Test:       aorm.FloatFrom(200.15987654321987654321),
	})
	if errInsert != nil {
		panic(name + "testInsert" + "found err")
	}

	aorm.Use(db).Debug(false).Insert(&Article{
		Type:        aorm.IntFrom(0),
		PersonId:    aorm.IntFrom(id),
		ArticleBody: aorm.StringFrom("文章内容"),
	})

	return id
}

func testInsertBatch(name string, db *sql.DB) int64 {
	var batch []Person
	batch = append(batch, Person{
		Name:       aorm.StringFrom("Alice"),
		Sex:        aorm.BoolFrom(false),
		Age:        aorm.IntFrom(18),
		Type:       aorm.IntFrom(0),
		CreateTime: aorm.TimeFrom(time.Now()),
		Money:      aorm.FloatFrom(100.15987654321),
		Test:       aorm.FloatFrom(200.15987654321987654321),
	})

	batch = append(batch, Person{
		Name:       aorm.StringFrom("Bob"),
		Sex:        aorm.BoolFrom(true),
		Age:        aorm.IntFrom(18),
		Type:       aorm.IntFrom(0),
		CreateTime: aorm.TimeFrom(time.Now()),
		Money:      aorm.FloatFrom(100.15987654321),
		Test:       aorm.FloatFrom(200.15987654321987654321),
	})

	count, err := aorm.Use(db).Debug(false).InsertBatch(&batch)
	if err != nil {
		panic(name + "testInsertBatch" + "found err")
	}

	return count
}

func testGetOne(name string, db *sql.DB, id int64) {
	var person Person
	errFind := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).GetOne(&person)
	if errFind != nil {
		panic(name + "testGetOne" + "found err")
	}
}

func testGetMany(name string, db *sql.DB) {
	var list []Person
	errSelect := aorm.Use(db).Debug(false).Where(&Person{Type: aorm.IntFrom(0)}).GetMany(&list)
	if errSelect != nil {
		panic(name + " testGetMany " + "found err:" + errSelect.Error())
	}
}

func testUpdate(name string, db *sql.DB, id int64) {
	_, errUpdate := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Update(&Person{Name: aorm.StringFrom("Bob")})
	if errUpdate != nil {
		panic(name + "testGetMany" + "found err")
	}
}

func testDelete(name string, db *sql.DB, id int64) {
	_, errDelete := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Delete()
	if errDelete != nil {
		panic(name + "testDelete" + "found err")
	}
}

func testTable(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Table("person_1").Insert(&Person{Name: aorm.StringFrom("Cherry")})
	if err != nil {
		panic(name + "testTable" + "found err")
	}
}

func testSelect(name string, db *sql.DB) {
	var listByFiled []Person
	err := aorm.Use(db).Debug(false).Select("name,age").Where(&Person{Age: aorm.IntFrom(18)}).GetMany(&listByFiled)
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
		Where(&Person{Age: aorm.IntFrom(18)}).
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

	var where1 []aorm.WhereItem
	where1 = append(where1, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	where1 = append(where1, aorm.WhereItem{Field: "age", Opt: aorm.In, Val: []int{18, 20}})
	where1 = append(where1, aorm.WhereItem{Field: "money", Opt: aorm.Between, Val: []float64{100.1, 200.9}})
	where1 = append(where1, aorm.WhereItem{Field: "money", Opt: aorm.Eq, Val: 100.15})
	where1 = append(where1, aorm.WhereItem{Field: "name", Opt: aorm.Like, Val: []string{"%", "li", "%"}})

	err := aorm.Use(db).Debug(false).Table("person").WhereArr(where1).GetMany(&listByWhere)
	if err != nil {
		panic(name + "testWhere" + "found err")
	}
}

func testJoin(name string, db *sql.DB) {
	var list2 []ArticleVO
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "o.type", Opt: aorm.Eq, Val: 0})
	where2 = append(where2, aorm.WhereItem{Field: "p.age", Opt: aorm.In, Val: []int{18, 20}})
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
	var where []aorm.WhereItem
	where = append(where, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
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

	var where3 []aorm.WhereItem
	where3 = append(where3, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})

	var having []aorm.WhereItem
	having = append(having, aorm.WhereItem{Field: "age_count", Opt: aorm.Gt, Val: 4})

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
	var where []aorm.WhereItem
	where = append(where, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	err := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where).
		OrderBy("age", aorm.Desc).
		GetMany(&listByOrder)
	if err != nil {
		panic(name + "testOrderBy" + "found err")
	}
}

func testLimit(name string, db *sql.DB) {
	var list3 []Person
	var where1 []aorm.WhereItem
	where1 = append(where1, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	err1 := aorm.Use(db).Debug(false).
		Table("person").
		WhereArr(where1).
		Limit(50, 10).
		GetMany(&list3)
	if err1 != nil {
		panic(name + "testLimit" + "found err")
	}

	var list4 []Person
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
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
	err := aorm.Use(db).Debug(false).LockForUpdate(true).Where(&Person{Id: aorm.IntFrom(id)}).GetOne(&itemByLock)
	if err != nil {
		panic(name + "testLock" + "found err")
	}
}

func testIncrement(name string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Increment("age", 1)
	if err != nil {
		panic(name + "testIncrement" + "found err")
	}
}

func testDecrement(name string, db *sql.DB, id int64) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Decrement("age", 2)
	if err != nil {
		panic(name + "testDecrement" + "found err")
	}
}

func testValue(dbName string, db *sql.DB, id int64) {

	var name string
	errName := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Value("name", &name)
	if errName != nil {
		panic(dbName + "testValue" + "found err")
	}

	var age int64
	errAge := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Value("age", &age)
	if errAge != nil {
		panic(dbName + "testValue" + "found err")
	}

	var money float32
	errMoney := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Value("money", &money)
	if errMoney != nil {
		panic(dbName + "testValue" + "found err")
	}

	var test float64
	errTest := aorm.Use(db).Debug(false).Where(&Person{Id: aorm.IntFrom(id)}).Value("test", &test)
	if errTest != nil {
		panic(dbName + "testValue" + "found err")
	}
}

func testPluck(name string, db *sql.DB) {

	var nameList []string
	errNameList := aorm.Use(db).Debug(false).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("name", &nameList)
	if errNameList != nil {
		panic(name + "testPluck" + "found err")
	}

	var ageList []int64
	errAgeList := aorm.Use(db).Debug(false).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("age", &ageList)
	if errAgeList != nil {
		panic(name + "testPluck" + "found err:" + errAgeList.Error())
	}

	var moneyList []float32
	errMoneyList := aorm.Use(db).Debug(false).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("money", &moneyList)
	if errMoneyList != nil {
		panic(name + "testPluck" + "found err")
	}

	var testList []float64
	errTestList := aorm.Use(db).Debug(false).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("test", &testList)
	if errTestList != nil {
		panic(name + "testPluck" + "found err")
	}
}

func testCount(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: aorm.IntFrom(18)}).Count("*")
	if err != nil {
		panic(name + "testCount" + "found err")
	}
}

func testSum(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: aorm.IntFrom(18)}).Sum("age")
	if err != nil {
		panic(name + "testSum" + "found err")
	}
}

func testAvg(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: aorm.IntFrom(18)}).Avg("age")
	if err != nil {
		panic(name + "testAvg" + "found err")
	}
}

func testMin(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: aorm.IntFrom(18)}).Min("age")
	if err != nil {
		panic(name + "testMin" + "found err")
	}
}

func testMax(name string, db *sql.DB) {
	_, err := aorm.Use(db).Debug(false).Where(&Person{Age: aorm.IntFrom(18)}).Max("age")
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
		Name: aorm.StringFrom("Alice"),
	})

	if errInsert != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	_, errCount := aorm.Use(tx).Debug(false).Where(&Person{
		Id: aorm.IntFrom(id),
	}).Count("*")
	if errCount != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	var person Person
	errPerson := aorm.Use(tx).Debug(false).Where(&Person{
		Id: aorm.IntFrom(id),
	}).GetOne(&person)
	if errPerson != nil {
		tx.Rollback()
		panic(name + "testTransaction" + "found err")
		return
	}

	_, errUpdate := aorm.Use(tx).Debug(false).Where(&Person{
		Id: aorm.IntFrom(id),
	}).Update(&Person{
		Name: aorm.StringFrom("Bob"),
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
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "o.type", Opt: aorm.Eq, Val: 0})
	where2 = append(where2, aorm.WhereItem{Field: "p.age", Opt: aorm.In, Val: []int{18, 20}})
	err := aorm.Use(db).Debug(false).
		Table("article o").
		LeftJoin("person p", aorm.Ul("p.id=o.personId")).
		Select("o.*").
		Select(aorm.Ul("p.name as personName")).
		WhereArr(where2).
		GetMany(&list2)
	if err != nil {
		panic(name + "testHelper" + "found err")
	}
}
