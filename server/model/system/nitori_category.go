package system

// === カテゴリツリー取得API用 ===

// CategoryTreeNode カテゴリツリーの各ノード情報
type CategoryTreeNode struct {
	CategoryID string              `json:"category_id"` // 数値IDを文字列として返すか、数値型のままかはAPI設計による
	Name       string              `json:"name"`
	Level      int                 `json:"level"`
	Slug       string              `json:"slug,omitempty"`     // カテゴリコードやURL用スラッグ
	Children   []*CategoryTreeNode `json:"children,omitempty"` // サブカテゴリのリスト (再帰)
}

// === 特定カテゴリ情報取得API用 ===

// CategoryDetailResponse 特定カテゴリ情報APIのルートレスポンス
type CategoryDetailResponse struct {
	CategoryInfo            *CategoryInfoBasic        `json:"category_info"`                        // 対象カテゴリ自身の情報
	Breadcrumbs             []BreadcrumbItem          `json:"breadcrumbs"`                          // パンくずリスト
	SubCategories           []SubCategoryInfo         `json:"sub_categories,omitempty"`             // 直下のサブカテゴリリスト
	FilterableAttributes    []FilterableAttributeInfo `json:"filterable_attributes,omitempty"`      // 絞り込み可能属性リスト (オプション)
	TotalProductsInCategory *int                      `json:"total_products_in_category,omitempty"` // このカテゴリの商品総数 (オプション)
}

// CategoryInfoBasic カテゴリの基本情報
type CategoryInfoBasic struct {
	CategoryId  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Level       int     `json:"level"`
	ParentId    *string `json:"parent_id,omitempty"` // 親カテゴリID (Nullable)
	Slug        *string `json:"slug,omitempty"`
}

// BreadcrumbItem パンくずリストの各アイテム
type BreadcrumbItem struct {
	CategoryId string `json:"category_id"`
	Name       string `json:"name"`
	//Slug       string `json:"slug,omitempty"`
	Level int `json:"level"`
}

// SubCategoryInfo サブカテゴリ情報
type SubCategoryInfo struct {
	CategoryId string `json:"category_id"`
	Name       string `json:"name"`
	//Slug         string `json:"slug,omitempty"`
	ProductCount *int `json:"product_count,omitempty"` // このサブカテゴリの商品数 (オプション)
}

// FilterableAttributeInfo 絞り込み可能な属性情報 (商品検索APIのFacetAttributeGroupと類似)
type FilterableAttributeInfo struct {
	AttributeID   int                    `json:"attribute_id"`
	AttributeName string                 `json:"attribute_name"`
	AttributeCode string                 `json:"attribute_code"`
	DisplayType   string                 `json:"display_type"` // 'image', 'text', etc.
	Options       []FilterableOptionItem `json:"options"`
}

// FilterableOptionItem 属性の選択肢情報 (商品検索APIのFacetOptionItemと類似)
type FilterableOptionItem struct {
	OptionID     int    `json:"option_id"`
	OptionValue  string `json:"option_value"`
	ProductCount *int   `json:"product_count,omitempty"` // この選択肢に該当する商品数 (オプション)
	// SwatchImageURL *string `json:"swatch_image_url,omitempty"` // 必要なら
}
