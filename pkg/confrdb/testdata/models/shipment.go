package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
	"github.com/xoctopus/sqlx/pkg/types/sqltime"
)

// Shipment 物流
// +genx:model
// @attr TableName=t_shipment
// @attr Register=Catalog
// @def pk ID
// @def u_idx ui_order_id    OrderID
// @def u_idx ui_tracking_no TrackingNo
// @def idx   i_carrier      Carrier
// @def idx   i_status       Status
// @def idx   i_shipped_at   ShippedAt
// @def idx   i_delivered_at DeliveredAt
type Shipment struct {
	types.AutoIncID

	RelOrder
	ShipmentData

	types.CreationModificationTime
}

type ShipmentData struct {
	// Carrier 物流运营商
	Carrier string `db:"carrier,width=64"`
	// TrackingNo 物流单号
	TrackingNo string `db:"tracking_no,width=128"`
	// Status 物流状态
	Status enums.ShipmentStatus `db:"status"`
	// ShippedAt 开始运输时间
	ShippedAt sqltime.Timestamp `db:"shipped_at"`
	// DeliveredAt 抵达时间
	DeliveredAt sqltime.Timestamp `db:"delivered_at"`
}
