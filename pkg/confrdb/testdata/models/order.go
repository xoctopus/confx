package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
	"github.com/xoctopus/sqlx/pkg/types/sqltime"
)

// Order 订单
// +genx:model
// @model TableName=t_order
// @model Register=Catalog
// @model pk=ID
// @model uidx=ui_order_id;OrderID
// @model idx=i_status;Status
// @model idx=i_created_at;CreatedAt
// @model idx=i_updated_at;UpdatedAt
type Order struct {
	types.Serial

	RelUser
	RelOrder
	RelOrderNo
	OrderMeta
	OrderState

	types.CreationModificationTime
}

type OrderID uint64

type RelOrder struct {
	// @model rel=Order.OrderID
	OrderID OrderID `db:"order_id"`
}

type RelOrderNo struct {
	// OrderNo 订单编号
	// @model rel=Order.OrderNo
	OrderNo string `db:"order_no,width=64" json:"orderNO"`
}

type OrderMeta struct {
	// Amount 订单金额
	Amount types.Decimal `db:"amount,width=22,precision=4" json:"amount"`
	// Currency 结算币种
	Currency enums.Currency `db:"currency" json:"currency"`
}

type OrderState struct {
	// PaidAt 订单支付时间
	PaidAt sqltime.Timestamp `db:"paid_at,default=0" json:"paidAt"`
	// CanceledAt 订单取消时间
	CanceledAt sqltime.Timestamp `db:"canceled_at,default=0" json:"canceledAt"`
	// Status 订单状态
	Status enums.OrderStatus `db:"status" json:"status"`
}
