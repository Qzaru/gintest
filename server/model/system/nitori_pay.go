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
	ID         uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                             // 保有クーポンID
	UserID     uint       `gorm:"column:user_id;not null" json:"user_id"`                                   // 用户ID
	CouponID   uint64     `gorm:"column:coupon_id;not null" json:"coupon_id"`                               // クーポンID
	IsUsed     bool       `gorm:"column:is_used;not null;default:false" json:"is_used"`                     // 使用済みフラグ
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at"`                                            // 使用日時
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
