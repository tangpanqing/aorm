[English](https://github.com/tangpanqing/aorm) | [简体中文](https://github.com/tangpanqing/aorm/blob/next/README_zh.md)

# Aorm
Aorm是一个基于go语言的数据库操作库。

给个 ⭐ 吧，如果这个项目帮助到你

## 🌟 特性
- [x] 支持 使用结构体(对象)操作数据库
- [x] 支持 MySQL,MsSQL,Postgres,Sqlite3 数据库
- [x] 支持 空值查询或修改
- [x] 支持 数据迁移
- [X] 支持 链式操作
## 🌟 预览
```go
    //定义人员结构体
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
	
    //定义文章结构体
    type Article struct {
        Id          null.Int    `aorm:"primary;auto_increment" json:"id"`
        Type        null.Int    `aorm:"index;comment:类型" json:"type"`
        PersonId    null.Int    `aorm:"comment:人员Id" json:"personId"`
        ArticleBody null.String `aorm:"type:text;comment:文章内容" json:"articleBody"`
    }

    //实例化结构体
    var person = Person{}
    var article = Article{}
	
    //保存实例
    aorm.Store(&person, &article)

    //连接数据库
    db, _ := aorm.Open(driver.Mysql, "root:root@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local")

    //迁移数据结构
    aorm.Migrator(db).AutoMigrate(&person, &article)
    
    //插入一个人员
    personId, _ := aorm.Db(db).Insert(&Person{
        Name:       null.StringFrom("Alice"),
        Sex:        null.BoolFrom(true),
        Age:        null.IntFrom(18),
        Type:       null.IntFrom(0),
        CreateTime: null.TimeFrom(time.Now()),
        Money:      null.FloatFrom(1),
        Test:       null.FloatFrom(2),
    })
    
    //插入一个文章
    articleId, _ := aorm.Db(db).Insert(&Article{
        Type:        null.IntFrom(0),
        PersonId:    null.IntFrom(personId),
        ArticleBody: null.StringFrom("文章内容"),
    })
    
    //获取一条记录
    var personItem Person
    err := aorm.Db(db).Table(&person).WhereEq(&person.Id, personId).OrderBy(&person.Id, builder.Desc).GetOne(&personItem)
    if err != nil {
        fmt.Println(err.Error())
    }
    
    //联合查询
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
    
    //带别名的联合查询
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
```
## 🌟 如何使用
- [导入](#导入)
  - [Mysql](#mysql)
  - [Sqlite](#sqlite)
  - [MSSQL](#mssql)
  - [Postgres](#postgres)
- [定义数据结构](#定义数据结构)
- [保存数据结构](#保存数据结构)
- [连接数据库](#连接数据库)   
- [自动迁移](#自动迁移)
- [基本增删改查](#基本增删改查)   
  - [增加一条记录](#增加一条记录)
  - [增加多条记录](#增加多条记录)
  - [获取一条记录](#获取一条记录)
  - [获取多条记录](#获取多条记录)
  - [更新记录](#更新记录)
  - [删除记录](#删除记录)
- [高级查询](#高级查询)
  - [查询指定表](#查询指定表)
  - [查询指定字段](#查询指定字段)
  - [查询条件](#查询条件)
  - [查询条件相关操作](#查询条件相关操作)
  - [联合查询](#联合查询)
  - [分组查询](#分组查询)
  - [筛选](#筛选)
  - [排序](#排序)
  - [分页查询](#分页查询)
  - [悲观锁](#悲观锁)
  - [自增操作](#自增操作)
  - [自减操作](#自减操作)
  - [查询某字段的值](#查询某字段的值)
  - [查询某列的值](#查询某列的值)
  - [是否存在](#是否存在)
- [聚合查询](#聚合查询)
  - [Count](#count)
  - [Sum](#sum)
  - [AVG](#avg)
  - [Min](#min)
  - [Max](#max)
- [原始SQL](#原始SQL)
- [事务操作](#事务操作)
- [清空表数据](#清空表数据)

### 导入
```go
    import (
        _ "github.com/go-sql-driver/mysql" 
        "github.com/tangpanqing/aorm"
    )
```

`github.com/tangpanqing/aorm` 对于sql操作的包装，使其更易用    

`github.com/go-sql-driver/mysql` mysql数据库的驱动包，如果你使用其他数据库，需要更改这里    

你可以通过下面的命令下载他们    

```shell
go get -u github.com/tangpanqing/aorm
```

#### Mysql
如果你使用 `Mysql` 数据库, 你或许可以使用这个驱动 `github.com/go-sql-driver/mysql`
```shell
go get -u github.com/go-sql-driver/mysql
```

#### Sqlite
如果你使用 `Sqlite` 数据库, 你或许可以使用这个驱动 `github.com/mattn/go-sqlite3`
```shell
go get -u github.com/mattn/go-sqlite3
```

#### Mssql
如果你使用 `Mssql` 数据库, 你或许可以使用这个驱动 `github.com/denisenkom/go-mssqldb`
```shell
go get -u github.com/denisenkom/go-mssqldb
```

#### Postgres
如果你使用 `Postgres` 数据库, 你或许可以使用这个驱动 `github.com/lib/pq`
```shell
go get -u github.com/lib/pq
```

### 定义数据结构
在操作数据库之前，你应该定义数据结构，如下
```go
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


    //修改默认表名
    func (p *Person) TableName() string {
        return "erp_person"
    }
    
    //可以定义该函数来设置表信息
    func (p *Person) TableOpinion() map[string]string {
        return map[string]string{
          "ENGINE":  "InnoDB",
          "COMMENT": "人员表",
        }
    }
```
首先请注意一些类型,例如 `null.Int`, `null.String`, `null.Bool`, `null.Float`, `null.Time`, 这是对 `sql.NUll*` 结构的一种包装

其次请注意 `aorm:` 标签, 这将被用于数据迁移，将数据表字段与结构体属性相对应, 以下信息你需要知道

| 关键字            | 关键字的值  | 描述              | 例子                 |
|----------------|--------|-----------------|--------------------|
| column         | string | 重新设置某列名称        | column:person_type |
| primary        | none   | 设置某列为主键索引       | primary            |
| unique         | none   | 设置某列为唯一索引       | unique             |
| index          | none   | 设置某列为普通索引       | index              |
| auto_increment | none   | 设置某列可自增         | auto_increment     |
| not null       | none   | 设置某列可空          | not null           |
| type           | string | 设置某列的数据类型       | type:double        |
| size           | int    | 设置某列的数据长度或者显示长度 | size:100           |
| comment        | string | 设置某列的备注信息       | comment:名字         |
| default        | string | 设置某列的默认值        | default:2          |


### 保存数据结构
在操作数据库之前，你应该保存数据结构，如下
```go
    var person = Person{}

    aorm.Store(&person)
```
通过 `Store` 方法, `Person`结构体的基本信息(例如表名，列名)将会被保存起来，便于之后的查询操作

### 连接数据库
使用 `aorm.Open` 方法, 你可以连接数据库, 接下来你应该`ping`数据库，保证可用
```go
    //替换这些数据库参数
    username := "root"
    password := "root"
    hostname := "localhost"
    port := "3306"
    dbname := "database_name"
    
    //连接数据库
    db, err := aorm.Open(driver.Mysql, username+":"+password+"@tcp("+hostname+":"+port+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    //设置最大的有效连接数
    db.SetMaxOpenConns(5)
	
    //设置调试模式
    db.SetDebugMode(false)
```
如果你使用其他数据库，请使用不同的数据库连接，如下
```go
//if content sqlite3 database
sqlite3Content, sqlite3Err := aorm.Open(driver.Sqlite3, "test.db")

//if content postgres database
psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "root", "postgres")
postgresContent, postgresErr := aorm.Open(driver.Postgres, psqlInfo)

//if content mssql database
mssqlInfo := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", "localhost", "database_name", "sa", "root", 1433)
mssqlContent, mssqlErr := aorm.Open(driver.Mssql, mssqlInfo)
```

### 自动迁移
使用 `AutoMigrate` 方法, 表名将是结构体名字的下划线形式，如`person`
```go
    aorm.Migrator(db).AutoMigrate(&person, &article, &student)
```
使用 `Migrate` 方法, 你可以使用其他的表名
```go
    aorm.Migrator(db).Migrate("person_1", &Person{})
```
使用 `ShowCreateTable` 方法, 你可以获得创建表的sql语句
```go
    showCreate := aorm.Migrator(db).ShowCreateTable("person")
    fmt.Println(showCreate)
```
如下
```sql
    CREATE TABLE `person` (
        `id` int NOT NULL AUTO_INCREMENT,
        `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '名字',
        `sex` tinyint DEFAULT NULL COMMENT '性别',
        `age` int DEFAULT NULL COMMENT '年龄',
        `type` int DEFAULT NULL COMMENT '类型',
        `create_time` datetime DEFAULT NULL COMMENT '创建时间',
        `money` float DEFAULT NULL COMMENT '金额',
        `article_body` text COLLATE utf8mb4_general_ci COMMENT '文章内容',
        `test` double DEFAULT NULL COMMENT '测试',
        PRIMARY KEY (`id`),
        KEY `idx_person_sex` (`sex`),
        KEY `idx_person_age` (`age`),
        KEY `idx_person_type` (`type`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='人员表'
```

### 基本增删改查

#### 增加一条记录
使用 `Insert` 方法, 你可以增加一条记录
```go
    id, errInsert := aorm.Db(db).Insert(&Person{
        Name:       null.StringFrom("Alice"),
        Sex:        null.BoolFrom(false),
        Age:        null.IntFrom(18),
        Type:       null.IntFrom(0),
        CreateTime: null.TimeFrom(time.Now()),
        Money:      null.FloatFrom(100.15987654321),
        Test:       null.FloatFrom(200.15987654321987654321),
    })
    if errInsert != nil {
        fmt.Println(errInsert)
    }
    fmt.Println(id)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    INSERT INTO person (name,sex,age,type,create_time,money,test) VALUES (?,?,?,?,?,?,?)
    Alice false 18 0 2022-12-07 10:10:26.1450773 +0800 CST m=+0.031808801 100.15987654321 200.15987654321987
```

#### 增加多条记录
使用 `InsertBatch` 方法, 你可以增加多条记录
```go
    var batch []*Person
    batch = append(batch, &Person{
        Name:       null.StringFrom("Alice"),
        Sex:        null.BoolFrom(false),
        Age:        null.IntFrom(18),
        Type:       null.IntFrom(0),
        CreateTime: null.TimeFrom(time.Now()),
        Money:      null.FloatFrom(100.15987654321),
        Test:       null.FloatFrom(200.15987654321987654321),
    })
    
    batch = append(batch, &Person{
        Name:       null.StringFrom("Bob"),
        Sex:        null.BoolFrom(true),
        Age:        null.IntFrom(18),
        Type:       null.IntFrom(0),
        CreateTime: null.TimeFrom(time.Now()),
        Money:      null.FloatFrom(100.15987654321),
        Test:       null.FloatFrom(200.15987654321987654321),
    })
    
    count, errInsertBatch := aorm.Db(db).InsertBatch(&batch)
    if errInsertBatch != nil {
        fmt.Println(errInsertBatch)
    }
    fmt.Println(count)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    INSERT INTO person (name,sex,age,type,create_time,money,test) VALUES (?,?,?,?,?,?,?),(?,?,?,?,?,?,?)
    Alice false 18 0 2022-12-16 15:28:49.3907587 +0800 CST m=+0.022987201 100.15987654321 200.15987654321987 Bob true 18 0 2022-12-16 15:28:49.3907587 +0800 CST m=+0.022987201 100.15987654321 200.15987654321987
```

#### 获取一条记录
使用 `GetOne` 方法, 你可以获取一条记录
```go
    var personItem Person
    errFind := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).GetOne(&personItem)
    if errFind != nil {
        fmt.Println(errFind)
    }
    fmt.Println(person)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.id = ? Limit ?,?
    1 0 1
```

#### 获取多条记录
使用 `GetMany` 方法, 你可以获取多条记录
```go
    var list []Person
    errSelect := aorm.Db(db).Table(&person).WhereEq(&person.Type, 0).GetMany(&list)
    if errSelect != nil {
        fmt.Println(errSelect)
    }
    for i := 0; i < len(list); i++ {
        fmt.Println(list[i])
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.type = ?
    0
```

#### 更新记录
使用 `Update` 方法, 你可以更新记录
```go
    countUpdate, errUpdate := aorm.Db(db).WhereEq(&person.Id, id).Update(&Person{Name: null.StringFrom("Bob")})
    if errUpdate != nil {
        fmt.Println(errUpdate)
    }
    fmt.Println(countUpdate)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    UPDATE person SET name=? WHERE id = ?
    Bob 1
```

#### 删除记录
使用 `Delete` 方法, 你可以删除记录
```go
    countDelete, errDelete := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Delete()
    if errDelete != nil {
        fmt.Println(errDelete)
    }
    fmt.Println(countDelete)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    DELETE FROM person WHERE person.id = ?
    1
```

### 高级查询
#### 查询指定表
使用 `Table` 方法, 你可以在查询时指定表名
```go
    _, err := aorm.Db(db).Table("person_1").Insert(&Person{Name: null.StringFrom("Cherry")})
    if err != nil {
        panic(db.DriverName() + " testTable " + "found err:" + err.Error())
    }
    
    _, err2 := aorm.Db(db).Table(&person).Insert(&Person{Name: null.StringFrom("Cherry")})
    if err2 != nil {
        panic(db.DriverName() + " testTable " + "found err:" + err2.Error())
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    INSERT INTO person_1 (name) VALUES (?)
    Cherry

    INSERT INTO person (name) VALUES (?)
    Cherry                            
```
#### 查询指定字段
使用 `Select` 方法, 你可以在查询时指定字段
```go
    var listByFiled []Person
    aorm.Db(db).Table(&person).Select(&person.Name).Select(&person.Age).WhereEq(&person.Age, 18).GetMany(&listByFiled)
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT person.name,person.age FROM person WHERE person.age = ?
    18
```
#### 查询条件
使用 `WhereArr` 方法, 你可以在查询时添加更多查询条件
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.type = ? AND person.age IN (?,?) AND person.money BETWEEN (?) AND (?) AND CONCAT(person.money,'') = ? AND person.name LIKE concat('%',?,'%')
    0 18 20 100.1 200.9 100.15 li
```
#### 查询条件相关操作
还有更多的查询操作，你需要知道

| 操作名                | 等同于         |
|--------------------|-------------|
| builder.Eq         | =           |
| builder.Ne         | !=          |
| builder.Gt         | \>          |
| builder.Ge         | \>=         |
| builder.Lt         | \<          |
| builder.Le         | \<=         |
| builder.In         | In          |
| builder.NotIn      | Not In      |
| builder.Like       | LIKE        |
| builder.NotLike    | Not Like    |
| builder.Between    | Between     |
| builder.NotBetween | Not Between |

#### 联合查询
使用 `LeftJoin` 方法, 你可以使用联合查询
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT article.*,person.name as person_name FROM article LEFT JOIN person ON person.id=article.person_id WHERE article.type = ? AND article.age IN (?,?)
    0 18 20
```
其他的联合查询方法还有 `RightJoin`, `Join`
#### 分组查询
使用 `GroupBy` 方法, 你可以进行分组查询
```go
    type PersonAge struct {
        Age         null.Int
        AgeCount    null.Int
    }

    var personAge = PersonAge{}
    aorm.Store(&personAge)

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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT person.age,count(person.age) as age_count FROM person WHERE person.type = ? GROUP BY person.age Limit ?,?
    0 0 1
```
#### 筛选
使用 `HavingArr` 以及 `Having` 方法, 你可以对分组查询的结果进行筛选
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT person.age,count(person.age) as age_count FROM person WHERE person.type = ? GROUP BY person.age Having age_count > ?
    0 4
```
#### 排序
使用 `OrderBy` 方法, 你可以对查询结果进行排序
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.type = ? Order BY person.age DESC
    0
                                               
    SELECT * FROM person o WHERE o.type = ? Order BY o.age DESC
    0                                          
```
#### 分页查询
使用 `Limit` 或者 `Page` 方法, 你可以进行分页查询
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.type = ? Order BY person.id DESC Limit ?,?
    0 50 10
```
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.type = ? Order BY person.id Limit ?,?
    0 20 10
```
#### 悲观锁
使用 `LockForUpdate` 方法, 你可以在查询时候锁住某些记录，禁止他们被修改
```go
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE person.id = ? Order BY person.id Limit ?,?  FOR UPDATE
    2 0 1
```

#### 自增操作
使用 `Increment` 方法, 你可以直接操作某字段增加数值
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Increment(&person.Age, 1)
    if err != nil {
        panic(db.DriverName() + " testIncrement " + "found err:" + err.Error())
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    UPDATE person SET age=age+? WHERE id = ?
    1 2
```

#### 自减操作
使用 `Decrement` 方法, 你可以直接操作某字段减少数值
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).Decrement(&person.Age, 2)
    if err != nil {
        panic(db.DriverName() + "testDecrement" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    UPDATE person SET age=age-? WHERE id = ?
    2 2
```

#### 查询某字段的值
使用 `Value` 方法, 你可以直接获取到某字段的值。
```go
    var name string
    errName := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Id, id).Value(&person.Name, &name)
    if errName != nil {
        panic(db.DriverName() + "testValue" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT person.name FROM person WHERE person.id = ? Order BY person.id Limit ?,?
    2 0 1
```
打印结果为 `Alice`

#### 查询某列的值
使用 `Pluck` 方法, 你可以直接查询某列的值。
```go
    var nameList []string
    errNameList := aorm.Db(db).Table(&person).OrderBy(&person.Id, builder.Desc).WhereEq(&person.Type, 0).Limit(0, 3).Pluck(&person.Name, &nameList)
    if errNameList != nil {
        panic(db.DriverName() + "testPluck" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT person.name FROM person WHERE person.type = ? Order BY person.id Limit ?,?
    0 0 5
```

#### 是否存在
```go
    exists, err := aorm.Db(db).Table(&person).WhereEq(&person.Id, id).OrderBy(&person.Id, builder.Desc).Exists()
    if err != nil {
        panic(db.DriverName() + " testExists " + "found err:" + err.Error())
    }
    return exists
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT 1 as c FROM person WHERE person.id = ? Order BY person.id Limit ?,? 
    36 0 1
```
另外, 你可以使用 `DoesntExists` 方法 如果你想知道某条记录是否不存在

### 聚合查询
#### Count
使用 `Count` 方法, 你可以查询出记录总数量
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Count("*")
    if err != nil {
        panic(db.DriverName() + "testCount" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT count(*) as c FROM person WHERE person.age = ?
    18
```
 
#### Sum
使用 `Sum` 方法, 你可以查询出符合条件的某字段之和
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Sum(&person.Age)
    if err != nil {
        panic(db.DriverName() + "testSum" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT sum(person.age) as c FROM person WHERE person.age = ?
    18
```
 
#### Avg
使用 `Avg` 方法, 你可以查询出符合条件的某字段平均值
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Avg(&person.Age)
    if err != nil {
        panic(db.DriverName() + "testAvg" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT avg(age) as c FROM person WHERE person.age = ?
    18
```
 
#### Min
使用 `Min` 方法, 你可以查询出符合条件的某字段最小值
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Min(&person.Age)
    if err != nil {
        panic(db.DriverName() + "testMin" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT min(person.age) as c FROM person WHERE person.age = ?
    18
```

#### Max
使用 `Max` 方法, 你可以查询出符合条件的某字段最大值
```go
    _, err := aorm.Db(db).Table(&person).WhereEq(&person.Age, 18).Max(&person.Age)
    if err != nil {
        panic(db.DriverName() + "testMax" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT max(person.age) as c FROM person WHERE person.age = ?
    18
```

### 原始SQL
```go
    var list []Person
    err1 := aorm.Db(db).RawSql("SELECT * FROM person WHERE id=? AND type=?", id2, 0).GetMany(&list)
    if err1 != nil {
        panic(err1)
    }
    fmt.Println(list)
    
    _, err := aorm.Db(db).RawSql("UPDATE person SET name = ? WHERE id=?", "Bob2", id2).Exec()
    if err != nil {
        panic(db.DriverName() + "testRawSql" + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    SELECT * FROM person WHERE id=? AND type=?
    9 0
                         
    UPDATE person SET name = ? WHERE id=?
    Bob2 9
```

### 事务操作
使用 `db` 的 `Begin` 方法, 开始一个事务    
然后使用 `Commit` 方法提交事务，`Rollback` 方法回滚事务
```go
    tx, _ := db.Begin()
    
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
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    INSERT INTO person (name) VALUES (?)
    Alice
                              
    SELECT Count(*) AS c FROM person  WHERE person.id = ?
    6

    SELECT * FROM person  WHERE person.id = ? ORDER BY person.id DESC Limit ?,? 
    6 0 1
                                              
    UPDATE person SET name=? WHERE id = ?
    Bob 6
```

### 清空表数据
使用 `Truncate` 方法, 你可以很方便的清空一个表    
```go
    _, err := aorm.Db(db).Table(&person).Truncate()
    if err != nil {
        panic(db.DriverName() + " testTruncate " + "found err")
    }
```
上述代码运行后得到的SQL预处理语句以及相关参数如下
```sql
    TRUNCATE TABLE person
```

## 基准测试
https://github.com/tangpanqing/orm-benchmark

## 作者

👤 **tangpanqing**

* Twitter: [@tangpanqing](https://twitter.com/tangpanqing)
* Github: [@tangpanqing](https://github.com/tangpanqing)
* Wechat: tangpanqing    
  ![wechat](./wechat.jpg)
## 希望能得到你的支持
给个 ⭐ 吧，如果这个项目帮助到你