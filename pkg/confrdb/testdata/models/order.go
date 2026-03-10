package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
	"github.com/xoctopus/sqlx/pkg/types/sqltime"
)

// Order 订单
// +genx:model
// @attr TableName=t_order
// @attr Register=Catalog
// @def pk ID
// @def u_idx ui_order_id  OrderID
// @def idx   i_status     Status
// @def idx   i_created_at CreatedAt
// @def idx   i_updated_at UpdatedAt
type Order struct {
	types.AutoIncID

	RelUser
	RelOrder
	OrderData

	types.CreationModificationTime
}

type OrderID uint64

type RelOrder struct {
	// @rel Order.OrderID
	OrderID OrderID `db:"order_id"`
}

type OrderData struct {
	// OrderNo 订单编号
	OrderNo string `db:"order_no,width=64"`
	// Amount 订单金额
	Amount types.Decimal `db:"amount,width=22,precision=4"`
	// Currency 结算币种
	Currency enums.Currency `db:"currency"`
	// PaidAt 订单支付时间
	PaidAt sqltime.Timestamp `db:"paid_at,default=0"`
	// CanceledAt 订单取消时间
	CanceledAt sqltime.Timestamp `db:"canceled_at,default=0"`
	// Status 订单状态
	Status enums.OrderStatus `db:"status"`
}
