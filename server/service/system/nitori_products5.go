package system

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

// @author: [granty1](https://github.com/granty1)
// @description: 查看目录树
// @return: cartResponse,err
func (productsService *ProductsService) GetCategoryTree(depth int) (categoryTreeNode []*system.CategoryTreeNode, err error) {
	type CategoryRecord struct {
		ID       string
		ParentId string
		Name     string
		Level    int
	}

	var records []CategoryRecord

	// 执行递归SQL
	query := fmt.Sprintf(`
WITH RECURSIVE category_tree AS (
    SELECT id, parent_id, name, 0 AS level
    FROM categories
    WHERE parent_id IS NULL OR parent_id = 0
    UNION ALL
    SELECT c.id, c.parent_id, c.name, ct.level + 1
    FROM categories c
    JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT * FROM category_tree
WHERE level <= %d
ORDER BY level, parent_id, id;
`, depth)

	err = global.GVA_DB.Raw(query).Scan(&records).Error
	if err != nil {
		return
	}
	//fmt.Println("records:", records)
	nodeMap := make(map[string]*system.CategoryTreeNode)
	var roots []*system.CategoryTreeNode
	for _, rec := range records {
		node := new(system.CategoryTreeNode) // 使用 new 关键字分配新的内存
		node.CategoryID = rec.ID
		node.Name = rec.Name
		node.Level = rec.Level
		node.Children = []*system.CategoryTreeNode{}

		nodeMap[rec.ID] = node
	}

	// 构建树
	for _, rec := range records {
		node := nodeMap[rec.ID]
		if rec.ParentId == "" {
			// 根节点
			roots = append(roots, node)
		} else {
			// 非根节点，挂到父节点的Children
			parent, ok := nodeMap[rec.ParentId]
			if ok {
				parent.Children = append(parent.Children, node)
			} else {
				return
			}
		}
	}
	return roots, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看目录详情
// @return: cartResponse,err
func (productsService *ProductsService) GetCategoryDetails(categoryID string) (categoryDetailResponse system.CategoryDetailResponse, err error) {
	type ResultCategoryDetail struct {
		//CategoryInfo            *CategoryInfoBasic        `json:"category_info"`                        // 対象カテゴリ自身の情報
		ACategoryId string  `json:"a_category_id"`
		AName       string  `json:"a_name"`
		Description *string `json:"description,omitempty"`
		ALevel      int     `json:"a_level"`
		ParentId    *string `json:"parent_id,omitempty"` // 親カテゴリID (Nullable)
		//Breadcrumbs             []BreadcrumbItem          `json:"breadcrumbs"`                          // パンくずリスト
		BCategoryId string `json:"b_category_id"`
		BName       string `json:"b_name"`
		BLevel      int    `json:"b_level"`
		//SubCategories           []SubCategoryInfo         `json:"sub_categories,omitempty"`             // 直下のサブカテゴリリスト
		SCategoryId   string `json:"s_category_id"`
		SName         string `json:"s_name"`
		SProductCount *int   `json:"s_product_count,omitempty"` // このサブカテゴリの商品数 (オプション)
		//FilterableAttributes    []FilterableAttributeInfo `json:"filterable_attributes,omitempty"`      // 絞り込み可能属性リスト (オプション)
		AttributeId   int    `json:"attribute_id"`
		AttributeName string `json:"attribute_name"`
		AttributeCode string `json:"attribute_code"`
		DisplayType   string `json:"display_type"` // 'image', 'text', etc.
		Attributes    string
		Options       string
		//Options       []FilterableOptionItem `json:"options"`
		OptionId                int    `json:"option_id"`
		OptionValue             string `json:"option_value"`
		ProductCount            int    `json:"product_count,omitempty"`              // この選択肢に該当する商品数 (オプション)
		TotalProductsInCategory *int   `json:"total_products_in_category,omitempty"` // このカテゴリの商品総数 (オプション)

	}
	var result1 []ResultCategoryDetail
	var result2 []ResultCategoryDetail

	err = global.GVA_DB.Raw(`
WITH RECURSIVE subcategories AS (
    SELECT categories.id FROM categories WHERE id = ?
    UNION ALL
    SELECT c.id FROM categories c
    JOIN subcategories s ON c.parent_id = s.id
)
SELECT sku_values.option_id as option_id,
       COUNT(DISTINCT products.id) as product_count
FROM subcategories
JOIN products ON subcategories.id = products.category_id
LEFT JOIN product_skus ON products.id = product_skus.product_id
LEFT JOIN sku_values ON product_skus.id = sku_values.sku_id
LEFT JOIN attribute_options ON sku_values.option_id = attribute_options.id
LEFT JOIN attributes ON attribute_options.attribute_id = attributes.id
WHERE attributes.is_filterable = 1
GROUP BY sku_values.option_id
`, categoryID).Scan(&result2).Error
	if err != nil {
		return
	}
	optionproductcount := make(map[int]int)
	for _, k := range result2 {
		if k.OptionId != 0 {
			optionproductcount[k.OptionId] = k.ProductCount
		}
	}

	err = global.GVA_DB.Table("categories").
		Select(`categories.id as a_category_id,
            categories.name as a_name,
            categories.description as description,
            categories.level as a_level,
            categories.parent_id as parent_id,
            CONCAT(att.id,'-',att.name,'-',att.attribute_code,'-',att.input_type) as attributes,
            GROUP_CONCAT(DISTINCT CONCAT(attribute_options.id,'-',attribute_options.value) ORDER BY att.sort_order) as options,
			count(distinct products.id) as total_products_in_category`).
		Joins(`LEFT JOIN category_attributes ON category_attributes.category_id=categories.id`).
		Joins(`LEFT JOIN (SELECT * FROM attributes WHERE attributes.is_filterable='1') att ON att.id=category_attributes.attribute_id`).
		Joins(`LEFT JOIN attribute_options ON attribute_options.attribute_id=att.id`).
		Joins(`join products on categories.id=products.category_id`).
		Where("categories.id = ?", categoryID).
		Group(`categories.id, categories.name, categories.description, categories.level, categories.parent_id, att.id`).
		Scan(&result1).Error
	if err != nil {
		return
	}
	if len(result1) == 0 {
		err = global.GVA_DB.Table("categories").
			Select(`categories.id as a_category_id, categories.name as a_name, categories.description as description, categories.level as a_level, categories.parent_id as parent_id, COUNT(DISTINCT products.id) as total_products_in_category`).
			Joins(`LEFT JOIN products ON categories.id = products.category_id`).
			Where("categories.id = ?", categoryID).
			Group("categories.id, categories.name, categories.description, categories.level, categories.parent_id").
			Scan(&result1).Error
		if err != nil {
			return
		}
	}

	filterableAttributeInfo := []system.FilterableAttributeInfo{}
	for _, k := range result1 {
		//var aint int
		//var attributefindoption map[int][]system.FilterableOptionItem
		//attributefindoption = make(map[int][]system.FilterableOptionItem)
		attrs := []system.FilterableOptionItem{}
		if k.Options != "" {
			parts := strings.Split(k.Options, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				kv := strings.SplitN(part, "-", 2)
				optionID, _ := strconv.Atoi(strings.TrimSpace(kv[0]))
				x := optionproductcount[optionID]
				if len(kv) == 2 {
					attrs = append(attrs, system.FilterableOptionItem{
						OptionID:     optionID,
						OptionValue:  strings.TrimSpace(kv[1]),
						ProductCount: &x,
					})
					//attributefindoption[aint] = append(attributefindoption[aint], attr)
				}
			}
		}
		if k.Attributes != "" {
			parts2 := strings.Split(k.Attributes, "-")
			attributeID, _ := strconv.Atoi(strings.TrimSpace(parts2[0]))
			filterableAttributeInfo = append(filterableAttributeInfo, system.FilterableAttributeInfo{
				AttributeID:   attributeID,
				AttributeName: parts2[1],
				AttributeCode: parts2[2],
				DisplayType:   parts2[3],
				Options:       attrs,
			})
		}
	}
	categoryInfo := system.CategoryInfoBasic{
		CategoryId:  result1[0].ACategoryId,
		Name:        result1[0].AName,
		Description: result1[0].Description,
		Level:       result1[0].ALevel,
		ParentId:    result1[0].ParentId,
	}
	breadcrumbItem := []system.BreadcrumbItem{}
	err = global.GVA_DB.Raw(`
WITH RECURSIVE parent_path AS (
    SELECT id as category_id, name, level, parent_id
    FROM categories
    WHERE id = ?
    UNION ALL
    SELECT c.id, c.name, c.level, c.parent_id
    FROM categories c
    JOIN parent_path p ON c.id = p.parent_id
)
SELECT * FROM parent_path WHERE category_id != ? ORDER BY level ASC;
`, categoryID, categoryID).Scan(&breadcrumbItem).Error
	if err != nil {
		return
	}
	subCategoryInfo := []system.SubCategoryInfo{}
	err = global.GVA_DB.Raw(`
WITH RECURSIVE descendants AS (
    SELECT 
        c.id AS parent_id,
        c.id AS descendant_id,
        c.name,
        0 AS level
    FROM categories c
    WHERE c.parent_id = ?
    UNION ALL
    SELECT 
        d.parent_id,
        c.id,
        c.name,
        d.level + 1
    FROM categories c
    JOIN descendants d ON c.parent_id = d.descendant_id
)
SELECT 
    descendants.parent_id as category_id,
    categories.name,
    COUNT(DISTINCT products.id) as product_count
FROM descendants
LEFT JOIN products ON products.category_id = descendants.descendant_id
JOIN categories ON categories.id = descendants.parent_id
GROUP BY descendants.parent_id, categories.name
`, categoryID).Scan(&subCategoryInfo).Error
	if err != nil {
		return
	}
	totalcount := 0
	for _, k := range subCategoryInfo {
		totalcount += *k.ProductCount
	}
	if result1[0].ALevel == 2 {
		totalcount = *result1[0].TotalProductsInCategory
	}
	categoryDetailResponse = system.CategoryDetailResponse{
		CategoryInfo:            &categoryInfo,
		Breadcrumbs:             breadcrumbItem,
		SubCategories:           subCategoryInfo,
		FilterableAttributes:    filterableAttributeInfo,
		TotalProductsInCategory: &totalcount,
	}
	return categoryDetailResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取促销活动
// @return: cartResponse,err
func (productsService *ProductsService) GetCampaigns(page int, limit int, status int) (campaignListResponse system.CampaignListResponse, err error) {
	var result1 []system.CampaignTeaserInfo
	var totalcount int64
	offset := (page - 1) * limit
	switch status {
	case 1:
		err = global.GVA_DB.Table("campaigns").
			Select(`id as campaign_id,
            campaign_code,
            name,
            catchphrase,
            banner_image_url,
            start_date as start_date_formatted,
            end_date as end_date_formatted,
            main_visual_url as campaign_url`).
			Where("is_active=1 AND NOW() BETWEEN start_date AND end_date").
			Order("start_date ASC").
			Limit(limit).Offset(offset).
			Find(&result1).Error
		err = global.GVA_DB.Table("campaigns").
			Where("is_active=1 AND NOW() BETWEEN start_date AND end_date").
			Order("start_date ASC").
			Count(&totalcount).Error
	case 2:
		err = global.GVA_DB.Table("campaigns").
			Select(`id as campaign_id,
            campaign_code,
            name,
            catchphrase,
            banner_image_url,
            start_date as start_date_formatted,
            end_date as end_date_formatted,
            main_visual_url as campaign_url`).
			Where("is_active=1 AND start_date > NOW()").
			Order("start_date ASC").
			Limit(limit).Offset(offset).
			Find(&result1).Error
		err = global.GVA_DB.Table("campaigns").
			Where("is_active=1 AND start_date > NOW()").
			Order("start_date ASC").
			Count(&totalcount).Error
	case 3:
		err = global.GVA_DB.Table("campaigns").
			Select(`id as campaign_id,
            campaign_code,
            name,
            catchphrase,
            banner_image_url,
            start_date as start_date_formatted,
            end_date as end_date_formatted,
            main_visual_url as campaign_url`).
			Where("is_active = ?", "1").
			Order("start_date ASC").
			Limit(limit).Offset(offset).
			Find(&result1).Error
		err = global.GVA_DB.Table("campaigns").
			Where("is_active = ?", "1").
			Order("start_date ASC").
			Count(&totalcount).Error
	}
	layout := time.RFC3339
	for i, _ := range result1 {
		sd, _ := time.Parse(layout, result1[i].StartDateFormatted)
		result1[i].StartDateFormatted = sd.Format("2006年01月02日  15:04:05")
		ed, _ := time.Parse(layout, result1[i].EndDateFormatted)
		result1[i].EndDateFormatted = ed.Format("2006年01月02日  15:04:05")
	}

	totalPages := math.Ceil(float64(totalcount) / float64(limit))
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  int(totalcount),
		TotalPages:  int(totalPages),
	}
	campaignListResponse = system.CampaignListResponse{
		Campaigns:  result1,
		Pagination: &pagination,
	}
	return campaignListResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取促销活动
// @return: cartResponse,err
func (productsService *ProductsService) GetCampaignDetail(campaign_id string, page int, limit int, sort int) (campaignDetailResponse system.CampaignDetailResponse, err error) {
	type Resultcampaign struct {
		//CampaignInfo   *CampaignFullInfo     `json:"campaign_info"`
		CampaignId         uint64  `json:"campaign_id"`
		CampaignCode       string  `json:"campaign_code"`
		Name               string  `json:"name"`
		Description        *string `json:"description,omitempty"`
		MainVisualUrl      *string `json:"main_visual_url,omitempty"`
		StartDateFormatted string  `json:"start_date_formatted"`
		EndDateFormatted   string  `json:"end_date_formatted"`
		TargetType         string  `json:"target_type"`
		//TargetProducts []SearchedProductInfo `json:"target_products"` // 商品検索APIのDTOを流用可能
		ProductId           string `json:"product_id"`
		ProductCode         string `json:"product_code,omitempty"`
		ProductName         string `json:"product_name"` // 省略後
		PriceRangeFormatted string `json:"price_range_formatted"`
		IsOnSale            bool   `json:"is_on_sale"`
		SalePrice           string
		//ReviewSummary       *ReviewSummaryInfo `json:"review_summary,omitempty"`
		AverageRating float64 `json:"average_rating"` // 平均評価
		ReviewCount   int     `json:"review_count"`   // レビュー件数

		ThumbnailImageUrl *string `json:"thumbnail_image_url,omitempty"`
		CategoryName      string  `json:"category_name"`
		//Pagination     PaginationInfo        `json:"pagination"`      // 対象商品リストのページネーション
	}
	var result []Resultcampaign
	var result2 []Resultcampaign
	var totalcount int64
	offset := (page - 1) * limit
	var orderby string
	switch sort {
	case 1:
		orderby = "min(prices.price) ASC"
	case 2:
		orderby = "products.created_at DESC"
	}
	err = global.GVA_DB.Table("campaigns").
		Select(`campaigns.id as campaign_id,
            campaigns.campaign_code as campaign_code,
            campaigns.name as name,
            campaigns.description as description,
            campaigns.main_visual_url as main_visual_url,
            campaigns.start_date as start_date_formatted,
            campaigns.end_date as end_date_formatted,
            campaigns.target_type as target_type,
            products.id as product_id,
            products.product_code as product_code,
            products.name as product_name,
            case
                when min(prices.price) < max(prices.price) then CONCAT(format(min(prices.price),0),'~',format(max(prices.price),0),'円')
                when min(prices.price) = max(prices.price) then CONCAT(format(min(prices.price),0),'円')
            end as price_range_formatted,
            GROUP_CONCAT(DISTINCT sale_price SEPARATOR ',') as sale_price,
            review_summaries.average_rating,
            review_summaries.review_count,
            si.thumbnail_url as thumbnail_image_url,
            categories.name as category_name`).
		Joins("JOIN campaign_products ON campaign_products.campaign_id=campaigns.id").
		Joins("JOIN products ON campaign_products.product_id=products.id").
		Joins("JOIN product_skus ON product_skus.product_id=products.id").
		Joins("JOIN (SELECT * FROM sku_images WHERE sort_order=0) si ON si.sku_id=products.default_sku_id").
		Joins("JOIN prices ON prices.sku_id=product_skus.id").
		Joins("JOIN (SELECT pr1.sku_id, pr1.price, CASE WHEN NOW() BETWEEN pr2.start_date AND pr2.end_date THEN pr2.price ELSE NULL END as sale_price FROM prices pr1 JOIN prices pr2 ON pr1.sku_id=pr2.sku_id WHERE pr1.price_type_id=1) pr3 ON pr3.sku_id=product_skus.id").
		Joins("JOIN review_summaries ON review_summaries.product_id=products.id").
		Joins("JOIN categories ON categories.id=products.category_id").
		Where("campaigns.id = ?", campaign_id).
		Group(`campaigns.id, campaigns.campaign_code, campaigns.name, campaigns.description, campaigns.main_visual_url, campaigns.start_date, campaigns.end_date, campaigns.target_type, products.id, products.product_code, products.name, review_summaries.average_rating, review_summaries.review_count, categories.name, si.thumbnail_url`).
		Order(orderby).
		Limit(limit).
		Offset(offset).
		Scan(&result).Error
	err = global.GVA_DB.Raw(`
select
    campaigns.id as campaign_id,
    campaigns.campaign_code as campaign_code,
    campaigns.name as name,
    campaigns.description as description,
    campaigns.main_visual_url as main_visual_url,
    campaigns.start_date as start_date_formatted,
    campaigns.end_date as end_date_formatted,
    campaigns.target_type as target_type,
    products.id as product_id,
    products.product_code as product_code,
    products.name as product_name,
    case
        when min(prices.price)<max(prices.price) then CONCAT(format(min(prices.price),0),'~',format(max(prices.price),0),'円')
        when min(prices.price)=max(prices.price) then concat(format(min(prices.price),0),'円')
    end as price_range_formatted,
    GROUP_CONCAT(DISTINCT sale_price SEPARATOR ',') as sale_price,
    review_summaries.average_rating as average_rating,
    review_summaries.review_count as review_count,
    si.thumbnail_url as thumbnail_image_url,
    categories.name as category_name
from campaigns
join campaign_products on campaign_products.campaign_id=campaigns.id
join products on campaign_products.product_id=products.id
join product_skus on product_skus.product_id=products.id
join (select * from sku_images where sort_order=0) si on si.sku_id=products.default_sku_id
join prices on prices.sku_id=product_skus.id
join (
    select pr1.sku_id,
           pr1.price,
           case when NOW() BETWEEN pr2.start_date AND pr2.end_date then pr2.price else null end as sale_price
    from prices pr1
    join prices pr2 on pr1.sku_id=pr2.sku_id
    where pr1.price_type_id=1
) pr3 on pr3.sku_id=product_skus.id
join review_summaries on review_summaries.product_id=products.id
join categories on categories.id=products.category_id
where campaigns.id=?
group by
    campaigns.id,
    campaigns.campaign_code,
    campaigns.name,
    campaigns.description,
    campaigns.main_visual_url,
    campaigns.start_date,
    campaigns.end_date,
    campaigns.target_type,
    products.id,
    products.product_code,
    products.name,
    review_summaries.average_rating,
    review_summaries.review_count,
    categories.name,
    si.thumbnail_url
`, campaign_id).Find(&result2).Error
	if err != nil {
		return
	}
	totalcount = int64(len(result2))
	layout := time.RFC3339
	sd, _ := time.Parse(layout, result[0].StartDateFormatted)
	result[0].StartDateFormatted = sd.Format("2006年01月02日  15:04:05")
	ed, _ := time.Parse(layout, result[0].EndDateFormatted)
	result[0].EndDateFormatted = ed.Format("2006年01月02日  15:04:05")
	campaignInfo := system.CampaignFullInfo{
		CampaignId:         result[0].CampaignId,
		CampaignCode:       result[0].CampaignCode,
		Name:               result[0].Name,
		Description:        result[0].Description,
		MainVisualUrl:      result[0].MainVisualUrl,
		StartDateFormatted: result[0].StartDateFormatted,
		EndDateFormatted:   result[0].EndDateFormatted,
		TargetType:         result[0].TargetType,
	}
	targetProducts := []system.SearchedProductInfo{}
	//reviewSummary := system.ReviewSummaryInfo{}
	//var isonsale bool
	for _, k := range result {
		reviewSummary := system.ReviewSummaryInfo{
			AverageRating: k.AverageRating,
			ReviewCount:   k.ReviewCount,
		}
		var isonsale bool
		if k.SalePrice != "" {
			isonsale = true
		}
		targetProducts = append(targetProducts, system.SearchedProductInfo{
			ProductID:           k.ProductId,
			ProductCode:         k.ProductCode,
			ProductName:         k.ProductName,
			PriceRangeFormatted: k.PriceRangeFormatted,
			IsOnSale:            isonsale,
			ReviewSummary:       &reviewSummary,
			ThumbnailImageUrl:   k.ThumbnailImageUrl,
			CategoryName:        k.CategoryName,
		})
	}
	totalPages := math.Ceil(float64(totalcount) / float64(limit))
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  int(totalcount),
		TotalPages:  int(totalPages),
	}
	campaignDetailResponse = system.CampaignDetailResponse{
		CampaignInfo:   &campaignInfo,
		TargetProducts: targetProducts,
		Pagination:     pagination,
	}
	return campaignDetailResponse, err
}
