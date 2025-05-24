package system

// CampaignTeaserInfo キャンペーン一覧表示用の簡易情報DTO
type CampaignTeaserInfo struct {
	CampaignId         uint64  `json:"campaign_id"`
	CampaignCode       string  `json:"campaign_code"`
	Name               string  `json:"name"`
	Catchphrase        *string `json:"catchphrase,omitempty"`
	BannerImageUrl     *string `json:"banner_image_url,omitempty"`
	StartDateFormatted string  `json:"start_date_formatted"`   // 表示用開始日時
	EndDateFormatted   string  `json:"end_date_formatted"`     // 表示用終了日時
	CampaignUrl        *string `json:"campaign_url,omitempty"` // キャンペーン詳細ページへの相対パスなど
}

// CampaignListResponse キャンペーン一覧APIのルートレスポンス
type CampaignListResponse struct {
	Campaigns  []CampaignTeaserInfo `json:"campaigns"`
	Pagination *PaginationInfo      `json:"pagination,omitempty"` // ページネーションが必要な場合
}

// CampaignDetailResponse キャンペーン詳細取得APIのルートレスポンス
type CampaignDetailResponse struct {
	CampaignInfo   *CampaignFullInfo     `json:"campaign_info"`
	TargetProducts []SearchedProductInfo `json:"target_products"` // 商品検索APIのDTOを流用可能
	Pagination     PaginationInfo        `json:"pagination"`      // 対象商品リストのページネーション
}

// CampaignFullInfo キャンペーン詳細情報DTO
type CampaignFullInfo struct {
	CampaignId         uint64  `json:"campaign_id"`
	CampaignCode       string  `json:"campaign_code"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	MainVisualUrl      *string `json:"main_visual_url,omitempty"`
	StartDateFormatted string  `json:"start_date_formatted"`
	EndDateFormatted   string  `json:"end_date_formatted"`
	TargetType         string  `json:"target_type"` // 'all_products', 'specific_categories', 'specific_products'
	// TargetCategoriesDescription *string `json:"target_categories_description,omitempty"` // 対象カテゴリの説明 (例: "ベッド・マットレス、ソファ")
	// ... その他必要な情報 ...
}
