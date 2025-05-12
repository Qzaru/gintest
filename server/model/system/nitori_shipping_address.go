package system

// ShippingAddressListResponse 配送先住所リストAPIのルートレスポンス
type ShippingAddressListResponse struct {
	Addresses []ShippingAddressInfo `json:"addresses"`
}

// ShippingAddressInfo 配送先住所情報のDTO
type ShippingAddressInfo struct {
	AddressID     uint64  `json:"address_id"`
	PostalCode    string  `json:"postal_code"`
	Prefecture    string  `json:"prefecture"`
	City          string  `json:"city"`
	AddressLine1  string  `json:"address_line1"`
	AddressLine2  *string `json:"address_line2,omitempty"`
	RecipientName string  `json:"recipient_name"`
	PhoneNumber   string  `json:"phone_number"`
	IsDefault     bool    `json:"is_default"`
}

// ShippingAddressInput 配送先住所の追加/編集時の入力DTO
type ShippingAddressInput struct {
	PostalCode    string  `json:"postal_code" binding:"required,max=10"`
	Prefecture    string  `json:"prefecture" binding:"required,max=50"`
	City          string  `json:"city" binding:"required,max=100"`
	AddressLine1  string  `json:"address_line1" binding:"required,max=255"`
	AddressLine2  *string `json:"address_line2,omitempty" binding:"max=255"`
	RecipientName string  `json:"recipient_name" binding:"required,max=100"`
	PhoneNumber   string  `json:"phone_number" binding:"required,max=20"`
	IsDefault     bool    `json:"is_default"`
}

type UserShippingAddress struct {
	Id            int64
	UserId        int64
	PostalCode    string  `json:"postal_code"`
	Prefecture    string  `json:"prefecture"`
	City          string  `json:"city"`
	AddressLine1  string  `json:"address_line1"`
	AddressLine2  *string `json:"address_line2,omitempty"`
	RecipientName string  `json:"recipient_name"`
	PhoneNumber   string  `json:"phone_number"`
	IsDefault     bool    `json:"is_default"`
}

func (UserShippingAddress) TableName() string {
	return "user_shipping_addresses"
}
