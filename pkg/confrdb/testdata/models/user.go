package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
)

// User 用户
// +genx:model
// @attr TableName=t_user
// @attr Register=Catalog
// @def pk ID
// @def u_idx ui_user_id   UserID
// @def idx   ui_username  Username
// @def idx   i_status     Status
// @def idx   i_created_at CreatedAt
type User struct {
	types.AutoIncID

	RelUser
	UserData

	types.OperationDatetime
}

type UserID uint64

type RelUser struct {
	// @rel User.UserID
	UserID UserID `db:"user_id"`
}

type UserData struct {
	// Username 用户名
	Username string `db:"username,width=127"`
	// Email 邮箱
	Email string `db:"email,width=127"`
	// Phone 电话
	Phone string `db:"phone,width=32"`
	// Status 用户状态
	Status enums.UserStatus `db:"status"`
}
