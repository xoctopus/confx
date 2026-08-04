package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
)

// User 用户
// +genx:model
// @model TableName=t_user
// @model Register=Catalog
// @model pk=ID
// @model uidx=ui_user_id;UserID
// @model idx=ui_username,BTREE;Username
// @model idx=i_status;Status,NULLS,FIRST
// @model idx=i_created_at;CreatedAt
type User struct {
	types.Serial

	RelUser
	UserMeta
	UserState

	types.OperationDatetime
}

type UserID uint64

type RelUser struct {
	// @model rel=User.UserID
	UserID UserID `db:"user_id"`
}

type UserMeta struct {
	// Username 用户名
	Username string `db:"username,width=127"`
	// Email 邮箱
	Email string `db:"email,width=127"`
	// Phone 电话
	Phone string `db:"phone,width=32"`
}

type UserState struct {
	// Status 用户状态
	Status enums.UserStatus `db:"status"`
}
