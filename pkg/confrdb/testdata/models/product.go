package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
)

// Product 商品
// +genx:model
// @attr TableName=t_product
// @attr Register=Catalog
// @def pk ID
// @def u_idx ui_product_id  ProductID,DeletedAt
// @def idx   i_product_name Name
// @def idx   i_status       Status
// @def idx   i_updated_at   UpdatedAt
type Product struct {
	types.AutoIncID

	RelProduct
	ProductData

	types.CreationModificationDeletionTime
}

type ProductID int64

type RelProduct struct {
	// @rel Product.ProductID
	ProductID ProductID `db:"product_id"`
}

type ProductData struct {
	// SKU 库存标签
	SKU string `db:"sku"`
	// Name 产品名称
	Name string `db:"name,width=256"`
	// Description 产品描述
	Description string `db:"description"`
	// Price 单价
	Price types.Decimal `db:"price,width=22,precision=4"`
	// Currency 货币
	Currency enums.Currency `db:"currency"`
	// Status 产品销售状态
	Status enums.ProductStatus `db:"status"`
}
