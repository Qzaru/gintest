package system

// ProductSearchResponse 商品検索APIのルートレスポンス
type ProductSearchResponse struct {
	Products   []SearchedProductInfo `json:"products"`         // 検索結果の商品リスト
	Pagination PaginationInfo        `json:"pagination"`       // ページネーション情報
	Facets     *SearchFacets         `json:"facets,omitempty"` // ファセット情報 (オプション)
}

// SearchedProductInfo 検索結果の各商品情報
type SearchedProductInfo struct {
	ProductID           string             `json:"product_id"`
	ProductCode         string             `json:"product_code,omitempty"`
	ProductName         string             `json:"product_name"` // 省略後
	PriceRangeFormatted string             `json:"price_range_formatted"`
	IsOnSale            bool               `json:"is_on_sale"`
	ReviewSummary       *ReviewSummaryInfo `json:"review_summary,omitempty"`
	ThumbnailImageUrl   *string            `json:"thumbnail_image_url,omitempty"`
	CategoryName        string             `json:"category_name"` // 商品が属するカテゴリ名
}

// SearchFacets ファセット検索のための情報
type SearchFacets struct {
	Categories []FacetCategoryItem `json:"categories,omitempty"`
	//	Brands      []FacetBrandItem      `json:"brands,omitempty"`       // ブランドファセットが必要な場合
	PriceRanges []FacetPriceRangeItem `json:"price_ranges,omitempty"` // 価格帯ファセット
	Attributes  []FacetAttributeGroup `json:"attributes,omitempty"`   // 属性ファセット
}

// FacetCategoryItem カテゴリファセットのアイテム
type FacetCategoryItem struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	ProductCount int    `json:"product_count"` // このカテゴリに該当する商品数
}

// FacetBrandItem ブランドファセットのアイテム
type FacetBrandItem struct {
	BrandID      int    `json:"brand_id"`
	BrandName    string `json:"brand_name"`
	ProductCount int    `json:"product_count"`
}

// FacetPriceRangeItem 価格帯ファセットのアイテム
type FacetPriceRangeItem struct {
	RangeLabel   string `json:"range_label"` // 例: "1,000円～2,999円"
	MinPrice     int    `json:"min_price"`
	MaxPrice     int    `json:"max_price"`
	ProductCount int    `json:"product_count"`
}

// FacetAttributeGroup 属性ファセットのグループ
type FacetAttributeGroup struct {
	AttributeID   int               `json:"attribute_id"`
	AttributeName string            `json:"attribute_name"`
	Options       []FacetOptionItem `json:"options"`
}

// FacetOptionItem 属性ファセットの選択肢アイテム
type FacetOptionItem struct {
	OptionID     int    `json:"option_id"`
	OptionValue  string `json:"option_value"`
	ProductCount int    `json:"product_count"`
}
