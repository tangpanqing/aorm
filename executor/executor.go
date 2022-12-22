package executor

import (
	"github.com/tangpanqing/aorm/model"
)

// ExpItem 将某子语句重命名为某字段
type ExpItem struct {
	Executor  **Executor
	FieldName string
}

// Executor 查询记录所需要的条件
type Executor struct {
	//数据库操作连接
	LinkCommon model.LinkCommon

	//查询参数
	tableName       string
	selectList      []string
	selectExpList   []*ExpItem
	groupList       []string
	whereList       []WhereItem
	joinList        []string
	havingList      []WhereItem
	orderList       []string
	offset          int
	pageSize        int
	isDebug         bool
	isLockForUpdate bool

	//sql与参数
	sql       string
	paramList []interface{}

	//驱动名字
	driverName string
}
