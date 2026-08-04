package models

import (
	"github.com/xoctopus/sqlx/example/enums"
	"github.com/xoctopus/sqlx/pkg/types"
	"github.com/xoctopus/sqlx/pkg/types/sqltime"
)

// Shipment 物流
// +genx:model
// @model TableName=t_shipment
// @model Register=Catalog
// @model pk=ID
// @model uidx=ui_order_id;OrderID
// @model uidx=ui_tracking_no;TrackingNo
// @model idx=i_carrier;Carrier
// @model idx=i_status;Status
// @model idx=i_shipped_at;ShippedAt
// @model idx=i_delivered_at;DeliveredAt
type Shipment struct {
	types.Serial

	RelOrder
	RelShipmentNo
	ShipmentMeta
	ShipmentState

	types.CreationModificationTime
}

type RelShipmentNo struct {
	// TrackingNo 物流单号
	// @model rel=Shipment.TrackingNo
	TrackingNo string `db:"tracking_no,width=128"`
}

type ShipmentMeta struct {
	// Carrier 物流运营商
	Carrier string `db:"carrier,width=64"`
}

type ShipmentState struct {
	// Status 物流状态
	Status enums.ShipmentStatus `db:"status"`
	// ShippedAt 开始运输时间
	ShippedAt sqltime.Timestamp `db:"shipped_at"`
	// DeliveredAt 抵达时间
	DeliveredAt sqltime.Timestamp `db:"delivered_at"`
}
