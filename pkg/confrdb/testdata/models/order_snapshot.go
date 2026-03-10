package models

import "github.com/xoctopus/sqlx/pkg/types"

// OrderSnapshot 订单快照
// +genx:model
// @attr TableName=t_order_snapshot
// @attr Register=Catalog
// @def pk ID
// @def u_idx ui_order_id    OrderID
// @def idx   i_product_name ProductName
// @def idx   i_created_at   CreatedAt
type OrderSnapshot struct {
	types.AutoIncID

	RelOrder
	RelProduct
	OrderSnapshotData

	types.CreationTime
}

type OrderSnapshotData struct {
	// ProductSKU 产品SKU
	ProductSKU string `db:"product_sku,width=64"`
	// ProductName 产品名称 Product.Name
	ProductName string `db:"product_name,width=256"`
	// Price 产品单价 Product.Price
	Price types.Decimal `db:"price,width=22,precision=4"`
	// Quantity 订单产品数量
	Quantity int64 `db:"quantity"`
	// Subtotal 订单金额
	Subtotal types.Decimal `db:"subtotal,width=22,precision=4"`
}
