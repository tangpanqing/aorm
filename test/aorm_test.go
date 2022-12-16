package test

import (
	"database/sql"
	"fmt"
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
	PersonName  aorm.Int    `aorm:"comment:人员名称" json:"personName"`
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

func TestAll(t *testing.T) {
	db := testConnect()
	defer db.Close()

	testMigrate(db)

	testShowCreateTable(db)

	id := testInsert(db)
	testInsertBatch(db)

	testGetOne(db, id)
	testGetMany(db)
	testUpdate(db, id)
	testDelete(db, id)

	id2 := testInsert(db)
	testTable(db)
	testSelect(db)
	return
	testWhere(db)
	testJoin(db)
	testGroupBy(db)
	testHaving(db)
	testOrderBy(db)
	testLimit(db)
	testLock(db, id2)

	testIncrement(db, id2)
	testDecrement(db, id2)

	testValue(db, id2)

	testPluck(db)

	testCount(db)
	testSum(db)
	testAvg(db)
	testMin(db)
	testMax(db)

	testQuery(db)
	testExec(db)

	testTransaction(db)
	testTruncate(db)
	testHelper(db)
}

func testConnect() *sql.DB {
	fmt.Println("--- testConnect ---")

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

func testMigrate(db *sql.DB) {
	fmt.Println("--- testMigrate ---")

	//AutoMigrate
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").AutoMigrate(&Person{})
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "文章").AutoMigrate(&Article{})

	//Migrate
	aorm.Use(db).Opinion("ENGINE", "InnoDB").Opinion("COMMENT", "人员表").Migrate("person_1", &Person{})
}

func testShowCreateTable(db *sql.DB) {
	fmt.Println("--- testShowCreateTable ---")

	showCreate := aorm.Use(db).ShowCreateTable("person")
	fmt.Println(showCreate)
}

func testInsert(db *sql.DB) int64 {
	fmt.Println("--- testInsert ---")

	id, errInsert := aorm.Use(db).Debug(true).Insert(&Person{
		Name:       aorm.StringFrom("Alice"),
		Sex:        aorm.BoolFrom(false),
		Age:        aorm.IntFrom(18),
		Type:       aorm.IntFrom(0),
		CreateTime: aorm.TimeFrom(time.Now()),
		Money:      aorm.FloatFrom(100.15987654321),
		Test:       aorm.FloatFrom(200.15987654321987654321),
	})
	if errInsert != nil {
		fmt.Println(errInsert)
	}
	fmt.Println(id)

	return id
}

func testInsertBatch(db *sql.DB) int64 {
	fmt.Println("--- testInsertBatch ---")

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

	count, errInsertBatch := aorm.Use(db).Debug(true).InsertBatch(&batch)
	if errInsertBatch != nil {
		fmt.Println(errInsertBatch)
	}
	fmt.Println(count)

	return count
}

func testGetOne(db *sql.DB, id int64) {
	fmt.Println("--- testGetOne ---")

	var person Person
	errFind := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).GetOne(&person)
	if errFind != nil {
		fmt.Println(errFind)
	}
	fmt.Println(person)
}

func testGetMany(db *sql.DB) {
	fmt.Println("--- testGetMany ---")

	var list []Person
	errSelect := aorm.Use(db).Debug(true).Where(&Person{Type: aorm.IntFrom(0)}).GetMany(&list)
	if errSelect != nil {
		fmt.Println(errSelect)
	}
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i])
	}
}

func testUpdate(db *sql.DB, id int64) {
	fmt.Println("--- testUpdate ---")

	countUpdate, errUpdate := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Update(&Person{Name: aorm.StringFrom("Bob")})
	if errUpdate != nil {
		fmt.Println(errUpdate)
	}
	fmt.Println(countUpdate)
}

func testDelete(db *sql.DB, id int64) {
	fmt.Println("--- testDelete ---")

	countDelete, errDelete := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Delete()
	if errDelete != nil {
		fmt.Println(errDelete)
	}
	fmt.Println(countDelete)
}

func testTable(db *sql.DB) {
	fmt.Println("--- testTable ---")

	aorm.Use(db).Debug(true).Table("person_1").Insert(&Person{Name: aorm.StringFrom("Cherry")})
}

func testSelect(db *sql.DB) {
	fmt.Println("--- testSelect ---")

	var listByFiled []Person
	aorm.Use(db).Debug(true).Select("name,age").Where(&Person{Age: aorm.IntFrom(18)}).GetMany(&listByFiled)

	sub := aorm.Use(db).Table("test_table").SelectCount("test_name", "test_name_count")
	aorm.Use(db).Debug(true).SelectExp(sub, "test_name_count_new").Select("name,age").Where(&Person{Age: aorm.IntFrom(18)}).GetMany(&listByFiled)
}

func testWhere(db *sql.DB) {
	fmt.Println("--- testWhere ---")

	//Where
	var listByWhere []Person

	var where1 []aorm.WhereItem
	where1 = append(where1, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	where1 = append(where1, aorm.WhereItem{Field: "age", Opt: aorm.In, Val: []int{18, 20}})
	where1 = append(where1, aorm.WhereItem{Field: "money", Opt: aorm.Between, Val: []float64{100.1, 200.9}})
	where1 = append(where1, aorm.WhereItem{Field: "money", Opt: aorm.Eq, Val: 100.15})
	where1 = append(where1, aorm.WhereItem{Field: "name", Opt: aorm.Like, Val: []string{"%", "li", "%"}})

	aorm.Use(db).Debug(true).Table("person").WhereArr(where1).GetMany(&listByWhere)
	for i := 0; i < len(listByWhere); i++ {
		fmt.Println(listByWhere[i])
	}
}

func testJoin(db *sql.DB) {
	fmt.Println("--- testJoin ---")

	var list2 []ArticleVO
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "o.type", Opt: aorm.Eq, Val: 0})
	where2 = append(where2, aorm.WhereItem{Field: "p.age", Opt: aorm.In, Val: []int{18, 20}})
	aorm.Use(db).Debug(true).
		Table("article o").
		LeftJoin("person p", "p.id=o.person_id").
		Select("o.*").
		Select("p.name as person_name").
		WhereArr(where2).
		GetMany(&list2)
}

func testGroupBy(db *sql.DB) {
	fmt.Println("--- testGroupBy ---")

	var personAge PersonAge
	var where []aorm.WhereItem
	where = append(where, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	aorm.Use(db).Debug(true).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where).
		GetOne(&personAge)
	fmt.Println(personAge)
}

func testHaving(db *sql.DB) {
	fmt.Println("--- testHaving ---")

	var listByHaving []PersonAge

	var where3 []aorm.WhereItem
	where3 = append(where3, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})

	var having []aorm.WhereItem
	having = append(having, aorm.WhereItem{Field: "age_count", Opt: aorm.Gt, Val: 4})

	err := aorm.Use(db).Debug(true).
		Table("person").
		Select("age").
		Select("count(age) as age_count").
		GroupBy("age").
		WhereArr(where3).
		HavingArr(having).
		GetMany(&listByHaving)
	if err != nil {
		panic(err)
	}
	fmt.Println(listByHaving)
}

func testOrderBy(db *sql.DB) {
	fmt.Println("--- testOrderBy ---")

	var listByOrder []Person
	var where []aorm.WhereItem
	where = append(where, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	err := aorm.Use(db).Debug(true).
		Table("person").
		WhereArr(where).
		OrderBy("age", aorm.Desc).
		GetMany(&listByOrder)
	if err != nil {
		panic(err)
	}
	fmt.Println(listByOrder)
}

func testLimit(db *sql.DB) {
	fmt.Println("--- testLimit ---")

	var list3 []Person
	var where1 []aorm.WhereItem
	where1 = append(where1, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	err1 := aorm.Use(db).Debug(true).
		Table("person").
		WhereArr(where1).
		Limit(50, 10).
		GetMany(&list3)
	if err1 != nil {
		panic(err1)
	}
	fmt.Println(list3)

	var list4 []Person
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "type", Opt: aorm.Eq, Val: 0})
	err := aorm.Use(db).Debug(true).
		Table("person").
		WhereArr(where2).
		Page(3, 10).
		GetMany(&list4)
	if err != nil {
		panic(err)
	}
	fmt.Println(list4)
}

func testLock(db *sql.DB, id int64) {
	fmt.Println("--- testLock ---")

	var itemByLock Person
	err := aorm.Use(db).Debug(true).LockForUpdate(true).Where(&Person{Id: aorm.IntFrom(id)}).GetOne(&itemByLock)
	if err != nil {
		panic(err)
	}
	fmt.Println(itemByLock)
}

func testIncrement(db *sql.DB, id int64) {
	fmt.Println("--- testIncrement ---")

	count, err := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Increment("age", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}

func testDecrement(db *sql.DB, id int64) {
	fmt.Println("--- testDecrement ---")

	count, err := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Decrement("age", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}

func testValue(db *sql.DB, id int64) {
	fmt.Println("--- testValue ---")

	var name string
	errName := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Value("name", &name)
	if errName != nil {
		panic(errName)
	}
	fmt.Println(name)

	var age int64
	errAge := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Value("age", &age)
	if errAge != nil {
		panic(errAge)
	}
	fmt.Println(age)

	var money float32
	errMoney := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Value("money", &money)
	if errMoney != nil {
		panic(errMoney)
	}
	fmt.Println(money)

	var test float64
	errTest := aorm.Use(db).Debug(true).Where(&Person{Id: aorm.IntFrom(id)}).Value("test", &test)
	if errTest != nil {
		panic(errTest)
	}
	fmt.Println(test)
}

func testPluck(db *sql.DB) {
	fmt.Println("--- testPluck ---")

	var nameList []string
	errNameList := aorm.Use(db).Debug(true).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("name", &nameList)
	if errNameList != nil {
		panic(errNameList)
	}
	for i := 0; i < len(nameList); i++ {
		fmt.Println(nameList[i])
	}

	var ageList []int64
	errAgeList := aorm.Use(db).Debug(true).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("age", &ageList)
	if errAgeList != nil {
		panic(errAgeList)
	}
	for i := 0; i < len(ageList); i++ {
		fmt.Println(ageList[i])
	}

	var moneyList []float32
	errMoneyList := aorm.Use(db).Debug(true).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("money", &moneyList)
	if errMoneyList != nil {
		panic(errMoneyList)
	}
	for i := 0; i < len(moneyList); i++ {
		fmt.Println(moneyList[i])
	}

	var testList []float64
	errTestList := aorm.Use(db).Debug(true).Where(&Person{Type: aorm.IntFrom(0)}).Limit(0, 3).Pluck("test", &testList)
	if errTestList != nil {
		panic(errTestList)
	}
	for i := 0; i < len(testList); i++ {
		fmt.Println(testList[i])
	}
}

func testCount(db *sql.DB) {
	fmt.Println("--- testCount ---")

	count, err := aorm.Use(db).Debug(true).Where(&Person{Age: aorm.IntFrom(18)}).Count("*")
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}

func testSum(db *sql.DB) {
	fmt.Println("--- testSum ---")

	sum, err := aorm.Use(db).Debug(true).Where(&Person{Age: aorm.IntFrom(18)}).Sum("age")
	if err != nil {
		panic(err)
	}
	fmt.Println(sum)
}

func testAvg(db *sql.DB) {
	fmt.Println("--- testAvg ---")

	avg, err := aorm.Use(db).Debug(true).Where(&Person{Age: aorm.IntFrom(18)}).Avg("age")
	if err != nil {
		panic(err)
	}
	fmt.Println(avg)
}

func testMin(db *sql.DB) {
	fmt.Println("--- testMin ---")

	min, err := aorm.Use(db).Debug(true).Where(&Person{Age: aorm.IntFrom(18)}).Min("age")
	if err != nil {
		panic(err)
	}
	fmt.Println(min)
}

func testMax(db *sql.DB) {
	fmt.Println("--- testMax ---")

	max, err := aorm.Use(db).Debug(true).Where(&Person{Age: aorm.IntFrom(18)}).Max("age")
	if err != nil {
		panic(err)
	}
	fmt.Println(max)
}

func testQuery(db *sql.DB) {
	fmt.Println("--- testQuery ---")

	resQuery, err := aorm.Use(db).Debug(true).Query("SELECT * FROM person WHERE id=? AND type=?", 1, 3)
	if err != nil {
		panic(err)
	}

	fmt.Println(resQuery)
}

func testExec(db *sql.DB) {
	fmt.Println("--- testExec ---")

	resExec, err := aorm.Use(db).Debug(true).Exec("UPDATE person SET name = ? WHERE id=?", "Bob", 3)
	if err != nil {
		panic(err)
	}
	fmt.Println(resExec.RowsAffected())
}

func testTransaction(db *sql.DB) {
	fmt.Println("--- testTransaction ---")

	tx, _ := db.Begin()

	id, errInsert := aorm.Use(tx).Debug(true).Insert(&Person{
		Name: aorm.StringFrom("Alice"),
	})

	if errInsert != nil {
		fmt.Println(errInsert)
		tx.Rollback()
		return
	}

	_, errCount := aorm.Use(tx).Debug(true).Where(&Person{
		Id: aorm.IntFrom(id),
	}).Count("*")
	if errCount != nil {
		fmt.Println(errCount)
		tx.Rollback()
		return
	}

	var person Person
	errPerson := aorm.Use(tx).Debug(true).Where(&Person{
		Id: aorm.IntFrom(id),
	}).GetOne(&person)
	if errPerson != nil {
		fmt.Println(errPerson)
		tx.Rollback()
		return
	}

	countUpdate, errUpdate := aorm.Use(tx).Debug(true).Where(&Person{
		Id: aorm.IntFrom(id),
	}).Update(&Person{
		Name: aorm.StringFrom("Bob"),
	})

	if errUpdate != nil {
		fmt.Println(errUpdate)
		tx.Rollback()
		return
	}

	fmt.Println(countUpdate)
	tx.Commit()
}

func testTruncate(db *sql.DB) {
	fmt.Println("--- testTruncate ---")

	count, err := aorm.Use(db).Debug(true).Table("person").Truncate()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}

func testHelper(db *sql.DB) {
	fmt.Println("--- testHelper ---")

	var list2 []ArticleVO
	var where2 []aorm.WhereItem
	where2 = append(where2, aorm.WhereItem{Field: "o.type", Opt: aorm.Eq, Val: 0})
	where2 = append(where2, aorm.WhereItem{Field: "p.age", Opt: aorm.In, Val: []int{18, 20}})
	aorm.Use(db).Debug(true).
		Table("article o").
		LeftJoin("person p", aorm.Ul("p.id=o.personId")).
		Select("o.*").
		Select(aorm.Ul("p.name as personName")).
		WhereArr(where2).
		GetMany(&list2)
}
