package system

// CartResponse カート内容取得APIのルートレスポンス
type CartResponse struct {
	Items                []CartItemInfo `json:"items"`                  // カート内商品リスト
	TotalItemsCount      int            `json:"total_items_count"`      // カート内総商品点数 (数量の合計)
	TotalAmount          float64        `json:"total_amount"`           // 合計金額 (計算用数値)
	TotalAmountFormatted string         `json:"total_amount_formatted"` // 合計金額 (表示用文字列 例: "55,880円")
}

// CartItemInfo カート内の個々の商品情報
type CartItemInfo struct {
	SkuID             string          `json:"sku_id"`                  // SKU ID
	ProductID         string          `json:"product_id"`              // 商品ID
	ProductName       string          `json:"product_name"`            // 商品名 (省略後)
	ProductCode       string          `json:"product_code,omitempty"`  // 商品コード
	Quantity          int             `json:"quantity"`                // カート内の数量
	Price             *PriceInfo      `json:"price"`                   // 現在の単価情報 (Nullable)
	SubtotalFormatted string          `json:"subtotal_formatted"`      // 小計 (表示用文字列 例: "7,980円")
	PrimaryImage      *ImageInfo      `json:"primary_image,omitempty"` // サムネイル画像推奨 (Nullable)
	Attributes        []AttributeInfo `json:"attributes"`              // 対象SKUの属性リスト
	StockStatus       string          `json:"stock_status"`            // 在庫状況コード ('available', 'low_stock', 'out_of_stock')
}

// --- 以下のDTOは他のAPIと共通化可能 ---

// PriceInfo 価格情報 (計算用数値と表示用文字列を含む)
// type PriceInfo struct {
//     Amount                 float64  `json:"amount"`                          // 計算用単価
//     FormattedAmount        string   `json:"formatted_amount"`                // 表示用単価文字列
//     Type                   string   `json:"type"`
//     TypeName               string   `json:"type_name"`
//     OriginalAmount         *float64 `json:"original_amount,omitempty"`       // セール時元単価(計算用)
//     FormattedOriginalAmount *string  `json:"formatted_original_amount,omitempty"` // セール時元単価(表示用)
// }

// ImageInfo 画像情報
// type ImageInfo struct {
//     ID      int     `json:"id"`
//     URL     string  `json:"url"` // サムネイルURL想定
//     AltText *string `json:"alt_text,omitempty"`
// }

// AttributeInfo 対象SKUの属性情報
// type AttributeInfo struct {
//     AttributeID   int     `json:"attribute_id"`
//     AttributeName string  `json:"attribute_name"`
//     OptionID      *int    `json:"option_id,omitempty"`
//     OptionValue   *string `json:"option_value,omitempty"`
//     ValueString   *string `json:"value_string,omitempty"`
//     // ... 他の value_xxx 型
// }

// ErrorResponse エラーレスポンス構造 (共通)
// type ErrorResponse struct {
// 	Error ErrorDetail `json:"error"`
// }
// type ErrorDetail struct {
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// 	Target  string `json:"target,omitempty"`
// }
