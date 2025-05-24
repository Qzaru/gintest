package system

import "time"

// PaymentMethodInfo 利用可能な支払い方法情報
type PaymentMethodInfo struct {
	MethodID    int     `json:"method_id"`
	MethodCode  string  `json:"method_code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// CheckoutInfoResponseV1_1 クーポン・ポイント情報取得APIのルートレスポンス (修正版)
type CheckoutInfoResponseV1_1 struct {
	AvailableCoupons     []AvailableCouponInfo `json:"available_coupons"`      // 利用可能なクーポンリスト
	UserPoints           *UserPointInfo        `json:"user_points,omitempty"`  // 保有ポイント情報
	CurrentCheckoutState *CurrentCheckoutState `json:"current_checkout_state"` // ★現在のチェックアウト状態
}

// AvailableCouponInfo 利用可能なクーポン情報
type AvailableCouponInfo struct {
	CouponID     uint64  `json:"coupon_id"`
	CouponCode   string  `json:"coupon_code"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	DiscountText string  `json:"discount_text"` // 例: "10% OFF (最大2,000円引)", "500円引き"
}

// UserPointInfo ユーザー保有ポイント情報
type UserPointInfo struct {
	AvailablePoints int `json:"available_points"`
}

// CurrentCheckoutState 現在のチェックアウト状態を表すDTO
type CurrentCheckoutState struct {
	CartSubtotalAmountFormatted   string             `json:"cart_subtotal_formatted"`          // カート商品小計 (割引前、表示用)
	AppliedCouponInfo             *AppliedCouponInfo `json:"applied_coupon_info,omitempty"`    // 適用中クーポン情報 (Nullable)
	CouponDiscountAmountFormatted string             `json:"coupon_discount_amount_formatted"` // クーポン割引額 (表示用)
	UsedPoints                    int                `json:"used_points"`                      // 利用ポイント数
	PointsDiscountAmountFormatted string             `json:"points_discount_amount_formatted"` // ポイント割引額 (表示用)
	ShippingFeeFormatted          string             `json:"shipping_fee_formatted"`           // 送料 (表示用、別途計算の場合あり)
	TotalAmountFormatted          string             `json:"total_amount_formatted"`           // ★最終支払総額 (表示用)

	// 内部計算用の数値も保持 (JSONには含めないか、開発用に含めるかは選択)
	CartSubtotalAmount   float64 `json:"-"`
	CouponDiscountAmount float64 `json:"-"`
	PointsDiscountAmount float64 `json:"-"`
	ShippingFee          float64 `json:"-"`
	TotalAmount          float64 `json:"-"`
}

// AppliedCouponInfo 適用中クーポン情報
type AppliedCouponInfo struct {
	CouponID                uint64  `json:"coupon_id"`
	CouponCode              string  `json:"coupon_code"`
	Name                    string  `json:"name"`
	DiscountAmount          float64 `json:"discount_amount"`           // この注文での実際の割引額 (計算用)
	FormattedDiscountAmount string  `json:"formatted_discount_amount"` // 表示用
}

// ApplyCouponRequest クーポン適用APIのリクエストボディ
type ApplyCouponRequest struct {
	CouponCode string `json:"coupon_code" binding:"required"`
}

// UsePointsRequest ポイント利用APIのリクエストボディ
type UsePointsRequest struct {
	PointsToUse int `json:"points_to_use" binding:"required,min=1"`
}

type CheckoutSession struct {
	UserID               uint      `gorm:"column:user_id;primaryKey;not null" json:"user_id"`                                                 // 用户ID（主键）
	CartSubtotal         float64   `gorm:"column:cart_subtotal;type:decimal(12,2);not null;default:0" json:"cart_subtotal"`                   // 购物车小计（未折扣）
	AppliedCouponID      *uint64   `gorm:"column:applied_coupon_id;type:decimal(20,0);default:null" json:"applied_coupon_id"`                 // 适用中优惠券ID（可为空）
	CouponDiscountAmount float64   `gorm:"column:coupon_discount_amount;type:decimal(10,2);not null;default:0" json:"coupon_discount_amount"` // 优惠券折扣额
	UsedPoints           int       `gorm:"column:used_points;type:uint;not null;default:0" json:"used_points"`                                // 使用的积分数
	PointsDiscountAmount float64   `gorm:"column:points_discount_amount;type:decimal(10,2);not null;default:0" json:"points_discount_amount"` // 积分折扣额
	ShippingFee          float64   `gorm:"column:shipping_fee;type:decimal(10,2);not null;default:0" json:"shipping_fee"`                     // 运费
	TotalAmount          float64   `gorm:"column:total_amount;type:decimal(12,2);not null;default:0" json:"total_amount"`                     // 最终支付总额
	LastUpdatedAt        time.Time `gorm:"column:last_updated_at;autoUpdateTime" json:"last_updated_at"`                                      // 最后更新时间
}

func (CheckoutSession) TableName() string {
	return "checkout_sessions"
}

type UserCoupon struct {
	ID         uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`         // 保有クーポンID
	UserID     uint       `gorm:"column:user_id;not null" json:"user_id"`               // 用户ID
	CouponID   uint64     `gorm:"column:coupon_id;not null" json:"coupon_id"`           // クーポンID
	IsUsed     bool       `gorm:"column:is_used;not null;default:false" json:"is_used"` // 使用済みフラグ
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at"`
	OrderId    uint64     `json:"order_id"`
	ObtainedAt time.Time  `gorm:"column:obtained_at;not null;default:CURRENT_TIMESTAMP" json:"obtained_at"` // 取得日時
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}

type UserPoints struct {
	UserID          uint      `gorm:"column:user_id;primaryKey" json:"user_id"`                           // 用户ID
	AvailablePoints int       `gorm:"column:available_points;not null;default:0" json:"available_points"` // 可用ポイント
	LastUpdatedAt   time.Time `gorm:"column:last_updated_at;autoUpdateTime" json:"last_updated_at"`       // 最終更新日時
}

func (UserPoints) TableName() string {
	return "user_points"
}

//用于pay.js的api
type PaymentRequest struct {
	Token  string `json:"token" binding:"required"`  // Pay.js 生成的 token
	Amount int64  `json:"amount" binding:"required"` // 支付金额（单位：日元）
}

// === 注文確定API用 ===

// CreateOrderRequest 注文確定APIのリクエストボディ
type CreateOrderRequest struct {
	PayjpToken *string `json:"payjp_token" binding:"required"` // クレジットカード支払いの場合は必須
	Notes      *string `json:"notes,omitempty"`
}

// OrderInfo 作成された注文情報のDTO
type OrderInfo struct {
	OrderID              uint64  `json:"order_id"`
	OrderCode            string  `json:"order_code"`
	OrderStatus          string  `json:"order_status"`
	TotalAmount          float64 `json:"total_amount"`           // 支払総額 (計算用数値)
	FormattedTotalAmount string  `json:"formatted_total_amount"` // 表示用支払総額
	OrderedAtFormatted   string  `json:"ordered_at_formatted"`   // 表示用注文日時
}

type Order struct {
	ID                   uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	OrderCode            string     `gorm:"column:order_code;size:50;unique;not null"`
	UserID               uint64     `gorm:"column:user_id;not null"`
	OrderStatus          string     `gorm:"column:order_status;type:enum('pending_payment','payment_confirmed','processing','shipped','delivered','cancelled','refunded');default:'pending_payment';not null"`
	SubtotalAmount       float64    `gorm:"column:subtotal_amount;type:decimal(12,2);not null"`
	CouponID             *uint64    `gorm:"column:coupon_id"` // 可空
	CouponDiscountAmount float64    `gorm:"column:coupon_discount_amount;type:decimal(10,2);default:0"`
	PointsUsed           uint       `gorm:"column:points_used;not null"`
	PointsDiscountAmount float64    `gorm:"column:points_discount_amount;type:decimal(10,2);default:0"`
	ShippingFee          float64    `gorm:"column:shipping_fee;type:decimal(10,2);default:0"`
	TotalAmount          float64    `gorm:"column:total_amount;type:decimal(12,2);not null"`
	PayjpChargeID        *string    `gorm:"column:payjp_charge_id;size:255;unique"` // 可空
	OrderedAt            time.Time  `gorm:"column:ordered_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	PaidAt               *time.Time `gorm:"column:paid_at;type:datetime"`      // 可空
	CancelledAt          *time.Time `gorm:"column:cancelled_at;type:datetime"` // 可空
	Notes                *string    `gorm:"column:notes;type:text"`            // 可空
	CreatedAt            time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID       uint64    `gorm:"column:order_id;not null"`
	ProductID     string    `gorm:"column:product_id;size:36;not null"`
	SkuID         string    `gorm:"column:sku_id;size:36;not null"`
	ProductName   string    `gorm:"column:product_name;size:255;not null"`
	SkuCode       *string   `gorm:"column:sku_code;size:150"` // 可空
	Quantity      int       `gorm:"column:quantity;not null"`
	UnitPrice     float64   `gorm:"column:unit_price;type:decimal(12,2);not null"`
	SubtotalPrice float64   `gorm:"column:subtotal_price;type:decimal(12,2);not null"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type PaymentTransaction struct {
	ID              uint64  `gorm:"column:id;primaryKey;autoIncrement"`                                                               // 交易ID
	OrderID         uint64  `gorm:"column:order_id;not null"`                                                                         // 关联订单ID
	PaymentGateway  string  `gorm:"column:payment_gateway;size:50;not null;default:'payjp'"`                                          // 支付网关
	TransactionID   string  `gorm:"column:transaction_id;size:255;unique;not null"`                                                   // 第三方交易ID
	TransactionType string  `gorm:"column:transaction_type;type:enum('charge','refund','capture');not null"`                          // 交易类型
	Amount          float64 `gorm:"column:amount;type:decimal(12,2);not null"`                                                        // 金额
	Currency        string  `gorm:"column:currency;size:3;not null;default:'JPY'"`                                                    // 货币单位
	Status          string  `gorm:"column:status;type:enum('succeeded','pending','failed','expired','captured','refunded');not null"` // 状态
	RawResponse     *string `gorm:"column:raw_response;type:text"`                                                                    // 原始响应内容，可能为空
	CreatedAt       string  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`                               // 创建时间
	UpdatedAt       string  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`   // 更新时间
}

func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

// OrderListResponse 注文履歴一覧APIのルートレスポンス
type OrderListResponse struct {
	Orders     []OrderSummaryInfo `json:"orders"`
	Pagination PaginationInfo     `json:"pagination"`
}

// OrderSummaryInfo 注文履歴一覧の各注文概要
type OrderSummaryInfo struct {
	OrderID                uint64  `json:"order_id"`
	OrderCode              string  `json:"order_code"`
	OrderedAtFormatted     string  `json:"ordered_at_formatted"`
	TotalAmountFormatted   string  `json:"total_amount_formatted"`
	OrderStatus            string  `json:"order_status"`                       // 注文ステータスコード
	OrderStatusDisplay     string  `json:"order_status_display"`               // 注文ステータス表示名
	RepresentativeImageURL *string `json:"representative_image_url,omitempty"` // 代表商品画像
}

// OrderDetailResponse 注文詳細取得APIのルートレスポンス
type OrderDetailResponse struct {
	OrderInfo  *OrderHeaderInfo      `json:"order_info"`
	OrderItems []OrderDetailItemInfo `json:"order_items"`
}

// OrderHeaderInfo 注文詳細のヘッダ情報
type OrderHeaderInfo struct {
	OrderID                       uint64  `json:"order_id"`
	OrderCode                     string  `json:"order_code"`
	OrderedAtFormatted            string  `json:"ordered_at_formatted"`
	OrderStatus                   string  `json:"order_status"`
	OrderStatusDisplay            string  `json:"order_status_display"`
	SubtotalAmountFormatted       string  `json:"subtotal_amount_formatted"` // 商品小計 (割引適用後)
	CouponDiscountAmountFormatted string  `json:"coupon_discount_amount_formatted"`
	PointsDiscountAmountFormatted string  `json:"points_discount_amount_formatted"`
	ShippingFeeFormatted          string  `json:"shipping_fee_formatted"` // 固定送料
	TotalAmountFormatted          string  `json:"total_amount_formatted"`
	AppliedCouponName             *string `json:"applied_coupon_name,omitempty"` // 適用クーポン名
	PointsUsed                    int     `json:"points_used"`
	PaymentMethodName             string  `json:"payment_method_name"`       // 固定: "クレジットカード"
	PayjpChargeID                 *string `json:"payjp_charge_id,omitempty"` // Pay.jp取引ID
	Notes                         *string `json:"notes,omitempty"`
	// (配送先情報は含めない)
}

// OrderDetailItemInfo 注文詳細の商品明細情報
type OrderDetailItemInfo struct {
	SkuID              string          `json:"sku_id"`
	ProductID          string          `json:"product_id"`
	ProductName        string          `json:"product_name"` // 省略後
	SkuCode            *string         `json:"sku_code,omitempty"`
	Quantity           int             `json:"quantity"`
	UnitPriceFormatted string          `json:"unit_price_formatted"` // 単価 (表示用)
	SubtotalFormatted  string          `json:"subtotal_formatted"`   // 小計 (表示用)
	PrimaryImageURL    *string         `json:"primary_image_url,omitempty"`
	Attributes         []AttributeInfo `json:"attributes"`
}
