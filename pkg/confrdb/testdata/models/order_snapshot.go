package models

import "github.com/xoctopus/sqlx/pkg/types"

// OrderSnapshot 订单快照
// +genx:model
// @model TableName=t_order_snapshot
// @model Register=Catalog
// @model pk=ID
// @model uidx=ui_order_id;OrderID
// @model idx=i_product_name;ProductName
// @model idx=i_created_at;CreatedAt
type OrderSnapshot struct {
	types.Serial

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
