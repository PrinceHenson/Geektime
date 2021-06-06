### 问：
在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？

### 答:
sql.ErrNoRows不需要作为error再抛出，其它的error需要抛出。
sql.ErrNoRows是一种特殊情况，当查询不到符合条件的记录时返回的一个结果，我认为这应该不算作错误，因为数据库当中可能就是没有符合条件的记录。
我的做法是dao层如果拿到了ErrNoRows, 那么两个返回值中第一个数据结果也返回nil，再由上一层根据业务决定应该返回404之类的或者不做处理。
这样上一层也不需要知道ErrNoRows这个预定义的类型。

```golang
func getUser(id int) (*student, error){
	rows := pool.QueryRow("select `sid`, `sname`, `age` from `stu5` where `sid` = ?", id)

	stu := student{}
	if err := rows.Scan(&stu.sid, &stu.sName, &stu.age); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Get student from db failed")
	}
	return &stu, nil
}
```