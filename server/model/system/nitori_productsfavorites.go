package system

type FavoriteSKUListResponse struct {
	Favorites  []FavoriteSKUInfo `json:"favorites"`  // お気に入りSKUリスト
	Pagination PaginationInfo    `json:"pagination"` // ページネーション情報
}

// FavoriteSKUInfo 個々のお気に入りSKU情報
type FavoriteSKUInfo struct {
	SkuID            string          `json:"sku_id"`                  // SKU ID
	ProductID        string          `json:"product_id"`              // 商品ID
	ProductName      string          `json:"product_name"`            // 商品名 (省略後)
	ProductCode      string          `json:"product_code,omitempty"`  // 商品コード
	Status           string          `json:"status"`                  //新加的，商品状态显示
	Price            *PriceInfo      `json:"price"`                   // 現在の表示価格情報 (Nullable)
	PrimaryImage     *ImageInfo      `json:"primary_image,omitempty"` // 代表画像 (Nullable)
	Attributes       []AttributeInfo `json:"attributes"`              // 対象SKUの属性リスト
	AddedAtFormatted string          `json:"added_at_formatted"`      // お気に入り追加日時 (表示用)
}
