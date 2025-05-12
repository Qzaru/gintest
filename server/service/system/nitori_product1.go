package system

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/products"
	"github.com/pkg/errors"
)

type ProductsService struct{}

var ProductsServiceApp = new(ProductsService)
var productsredis = products.NewProductsRedisStore()

type Result1 struct {
	ProductCode     string  `json:"product_code,omitempty"`
	Name            string  `gorm:"column:products_name"`
	Description     *string `json:"description,omitempty"`
	IsTaxable       bool    `json:"is_taxable"`
	MetaTitle       *string `json:"meta_title,omitempty"`
	MetaDescription *string `json:"meta_description,omitempty"`
	//	TargetSKUInfo      *TargetSKUInfo       `json:"target_sku_info,omitempty"` // 対象SKU情報 (Nullable)
	SkuID string `json:"sku_id"`
	//Price        *PriceInfo      `json:"price"`
	Amount                  float64  `json:"amount"` //这是一个需要判断是否在促销期内的促销价格，否则为初始价格
	FormattedAmount         string   `json:"formatted_amount"`
	Type                    string   `json:"type"`
	TypeName                string   `json:"type_name"`
	OriginalAmount          *float64 `json:"original_amount,omitempty"` //这是一个通常价格
	FormattedOriginalAmount *string  `json:"formatted_original_amount,omitempty"`
	//StockInfo    *StockInfo      `json:"stock_info"`
	Status           string  `json:"status"`
	StatusText       string  `json:"status_text"`
	DeliveryEstimate *string `json:"delivery_estimate,omitempty"`
	//PrimaryImage *ImageInfo      `json:"primary_image,omitempty"`
	ImageID int     `json:"image_id"`
	URL     string  `json:"url"`
	AltText *string `json:"alt_text,omitempty"`
	//Attributes   []AttributeInfo `json:"attributes"`
	AttributeName  string `json:"attribute_name"`
	AttributeValue string `json:"attribute_value"`

	CategoriesID        int    `gorm:"column:categories_id"`
	CategoriesName      string `gorm:"column:categories_name"`
	CategoriesLevel     int    `gorm:"column:categories_level"`
	CategoriesParentID  *int   `gorm:"column:categories_parent_id"`
	SameLevelCategories string `gorm:"column:same_level_categories"`

	//Brand              *BrandInfo           `json:"brand,omitempty"`
	BrandID   int    `json:"brandid,omitempty"`
	BrandName string `json:"brandname,omitempty"`
}
type Result2 struct {
	AttributeID   int    `json:"attribute_id"`
	AttributeName string `json:"attribute_name"`
	AttributeCode string `json:"attribute_code"`
	DisplayType   string `json:"display_type"` // 'image', 'text', etc.
	//Options       []VariantOption `json:"options"`
	OptionID    int     `json:"option_id"`
	OptionValue string  `json:"option_value"`
	OptionCode  string  `json:"option_code"`
	ImageURL    *string `json:"image_url,omitempty"`     // Swatch画像等
	LinkedSkuID string  `json:"linked_sku_id,omitempty"` // 関連SKU (任意)
}

//先处理VariantOptions 和Attributes   []AttributeInfo以外的内容,考虑后续通过结构体map嵌进去

// @author: [granty1](https://github.com/granty1)
// @function: GetProductInfo
// @description: 根据id获取商品信息
// @param product_code path  string true "商品code"
// @param skuid  query  string false   "商品skuid"
// @return: productInfoRes,err
func (productsService *ProductsService) GetProductInfo(productCode string, sku_id string) (productInfoRes system.ProductInfoResponse, err error) {
	var products system.Products
	var results1 []Result1
	var results2 []Result2
	//将searchSkuId默认为product_code对应的DefaultSKUID，直到有skuid输入
	err = global.GVA_DB.Where("product_code = ?", productCode).Find(&products).Error
	searchSkuId := products.DefaultSkuID
	if sku_id != "" {
		searchSkuId = &sku_id
	}
	if searchSkuId == nil {
		err = errors.New("总之有错误")
		return productInfoRes, err
	}

	fmt.Println(searchSkuId)
	//抽取数据
	//判断redis中有没有，如检测空值则进入sql查询
	key1 := *searchSkuId + "1"
	redisStr1 := productsredis.Get(key1, false)
	if redisStr1 == "" {
		err = global.GVA_DB.Table("product_skus").
			Select(`DISTINCT attributes.sort_order AS paixu,
		    products.product_code AS product_code, 
            products.name AS products_name, 
            products.description AS description, 
            products.is_taxable AS is_taxable, 
            products.meta_title AS meta_title, 
            products.meta_description AS meta_description, 
            product_skus.id AS sku_id, 
            MAX(CASE WHEN prices.price_type_id = 1 THEN prices.price END) AS amount, 
			MAX(CASE WHEN prices.price_type_id = 1 THEN price_types.type_code END) AS type, 
			MAX(CASE WHEN prices.price_type_id = 1 THEN price_types.name END) AS type_name, 
			MAX(CASE WHEN prices.price_type_id = 2 AND NOW() BETWEEN prices.start_date AND prices.end_date THEN prices.price END) AS original_amount,
			attributes.name AS attribute_name,
            COALESCE(attribute_options.value, sku_values.value_string, sku_values.value_number, sku_values.value_boolean) AS attribute_value,
            c1.id AS categories_id,
            c1.name AS categories_name,
            c1.level AS categories_level,
            c1.parent_id AS categories_parent_id,
			GROUP_CONCAT(DISTINCT c2.name SEPARATOR ',')as same_level_categories `).
			Joins("JOIN products ON product_skus.product_id = products.id").
			Joins("JOIN prices ON prices.sku_id = product_skus.id").
			Joins("JOIN inventory ON inventory.sku_id = product_skus.id").
			Joins("JOIN sku_availability ON sku_availability.sku_id = product_skus.id").
			Joins("JOIN categories c1 ON c1.id = products.category_id").
			Joins("JOIN categories c2 ON c1.parent_id = c2.parent_id").
			Joins("JOIN price_types ON price_types.id = prices.price_type_id").
			Joins("JOIN inventory_locations ON inventory_locations.id = inventory.location_id").
			Joins("LEFT JOIN sku_values ON sku_values.sku_id = product_skus.id").
			Joins("LEFT JOIN attributes ON attributes.id = sku_values.attribute_id").
			Joins("LEFT JOIN attribute_options ON attribute_options.id = sku_values.option_id").
			Where("product_skus.id = ?", searchSkuId).
			Where("(prices.price_type_id = ? OR (prices.price_type_id = ? AND NOW() BETWEEN prices.start_date AND prices.end_date) AND c2.id<>c1.id)", 1, 2).
			Group(`products.product_code, products.name, products.description, products.is_taxable, products.meta_title, 
		products.meta_description, product_skus.id,
		 c1.id, c1.name, c1.level, c1.parent_id, attributes.name,
		 attribute_options.value, 
        sku_values.value_string, 
        sku_values.value_number, 
        sku_values.value_boolean,
		attributes.sort_order`).
			Order("attributes.sort_order ASC").
			Scan(&results1).Error
		fmt.Println("查询结果是:", results1)
		//结果非空，则将结果存入redis
		if len(results1) != 0 {
			resultStr, _ := json.Marshal(results1) //先不管err了，面倒
			productsredis.Set(key1, string(resultStr))
			//如果没查出结果，即无效的skuid，报错返回
		} else {
			err = errors.New("总之有错误")
			fmt.Println(err)
			return productInfoRes, err
		}
	} else { //如果redis有值，绑定到结构体
		err = json.Unmarshal([]byte(redisStr1), &results1)
	}

	//测试用
	//fmt.Println("results1=", results1)
	//结果的格式处理
	key2 := *searchSkuId + "2"
	redisStr2 := productsredis.Get(key2, false)
	if redisStr1 == "" {
		err = global.GVA_DB.Table("product_skus").
			Select(`DISTINCT category_attributes.attribute_id AS attribute_id,
            attributes.name AS attribute_name,
            attributes.attribute_code AS attribute_code,
            attributes.input_type AS display_type,
            attribute_options.id AS option_id,
            attribute_options.value AS option_value,
            attribute_options.option_code AS option_code,
            sku_values.sku_id AS linked_sku_id`).
			Joins("LEFT JOIN products ON product_skus.product_id = products.id").
			Joins("LEFT JOIN categories ON products.category_id = categories.id").
			Joins("LEFT JOIN category_attributes ON categories.id = category_attributes.category_id").
			Joins("LEFT JOIN attributes ON category_attributes.attribute_id = attributes.id").
			Joins("LEFT JOIN attribute_options ON attributes.id = attribute_options.attribute_id").
			Joins("JOIN product_skus e1 ON product_skus.product_id = e1.product_id").
			Joins("JOIN sku_values ON e1.id = sku_values.sku_id AND sku_values.option_id = attribute_options.id").
			Joins("JOIN inventory ON inventory.sku_id = sku_values.sku_id").
			Where("product_skus.id = ? AND category_attributes.is_variant_attribute = ?", searchSkuId, 1).
			Where("inventory.quantity - inventory.reserved_quantity > ?", 0).
			Order("attribute_id ASC, option_id ASC").
			Scan(&results2).Error
		if len(results2) != 0 {
			resultStr2, _ := json.Marshal(results2) //先不管err了，面倒
			productsredis.Set(key2, string(resultStr2))
			//如果没查出结果，即无效的skuid，报错返回
		} //else {
		// 	err = errors.New("总之有错误")
		// 	return productInfoRes, err
		// }
	} else { //如果redis有值，绑定到结构体
		err = json.Unmarshal([]byte(redisStr2), &results2)
	}
	//测试用
	//fmt.Println("results2=", results2)
	//Option对应LinkedSkuIDs的映射
	OptionLinkedSkuIDs := make(map[int][]string)
	for _, k := range results2 {
		OptionLinkedSkuIDs[k.OptionID] = append(OptionLinkedSkuIDs[k.OptionID], k.LinkedSkuID)
	}
	//以上把linkedskuid放进切片是成功了的，但是后面没有去重复
	//Attributename对应的VariantOption构造体的映射,使AttributeName对应system.VariantOption
	AttributenameOptions := make(map[string][]system.VariantOption)
	optionNameMap := make(map[int]bool)
	for _, k := range results2 {
		if !optionNameMap[k.OptionID] {
			AttributenameOptions[k.AttributeName] = append(AttributenameOptions[k.AttributeName], system.VariantOption{
				OptionID:    k.OptionID,
				OptionValue: k.OptionValue,
				OptionCode:  k.OptionCode,
				//	ImageURL:     k.ImageURL,
				LinkedSkuIDs: OptionLinkedSkuIDs[k.OptionID],
			})
			optionNameMap[k.OptionID] = true
		}
	}

	//遍历results2转换成variantOptionGroups
	var variantOptions []system.VariantOptionGroup
	attributeNameMap := make(map[string]bool) //检查attributeName是否重复
	for _, k := range results2 {
		if !attributeNameMap[k.AttributeName] {
			variantOptions = append(variantOptions, system.VariantOptionGroup{
				AttributeName: k.AttributeName,
				AttributeCode: k.AttributeCode,
				DisplayType:   k.DisplayType,
				Options:       AttributenameOptions[k.AttributeName], //映射对应上了，但每次遍历都添加了相应的system.VariantOption，应在添加前检查AttributeName是否重复
			})
			attributeNameMap[k.AttributeName] = true
		}
	}
	//fmt.Println("检查是否去重", variantOptions)
	//去重
	uniqueMap := make(map[string]bool)
	var newvariantOptions []system.VariantOptionGroup
	for _, p := range variantOptions {
		if !uniqueMap[p.AttributeName] {
			newvariantOptions = append(newvariantOptions, p)
			uniqueMap[p.AttributeName] = true
		}
	}
	//测试用
	//fmt.Println("按钮结果", newvariantOptions)
	brandInfo := system.BrandInfo{
		ID:   results1[0].BrandID,
		Name: results1[0].BrandName,
	}
	sameLevelCategories := strings.Split(results1[0].SameLevelCategories, ",")
	categoryInfo := system.CategoryInfo{
		ID:                  results1[0].CategoriesID,
		Name:                results1[0].CategoriesName,
		Level:               results1[0].CategoriesLevel,
		ParentID:            results1[0].CategoriesParentID,
		SameLevelCategories: sameLevelCategories,
	}

	// imageInfo := system.ImageInfo{
	// 	ID:      results1[0].ImageID,
	// 	URL:     results1[0].URL,
	// 	AltText: results1[0].AltText,
	// }

	// stockInfo := system.StockInfo{
	// 	Status:     "待施工",
	// 	StatusText: "待施工",
	//DeliveryEstimate  :nill,
	//}

	// x := fmt.Sprintf("%,.0f", results1[0].Amount) // 格式化为 '2,990'
	// results1[0].FormattedAmount = x + "円"
	// if results1[0].OriginalAmount != nil {
	// 	y := fmt.Sprintf("%,.0f", *results1[0].OriginalAmount)
	// 	z := y[:len(x)-3] + "円"
	// 	results1[0].FormattedOriginalAmount = &z
	// }
	//名称和描述格式化
	if utf8.RuneCountInString(results1[0].Name) > 20 {
		runes := []rune(results1[0].Name)
		results1[0].Name = string(runes[:20]) + "..."
	}
	if utf8.RuneCountInString(*results1[0].Description) > 20 {
		x := *results1[0].Description
		runes := []rune(x)
		y := string(runes[:20]) + "..."
		results1[0].Description = &y
	}

	//价格转换
	if results1[0].OriginalAmount != nil {
		x := results1[0].Amount
		results1[0].Amount = *results1[0].OriginalAmount
		results1[0].Type = "sale"
		results1[0].TypeName = "セール価格"
		results1[0].OriginalAmount = &x
	}

	results1[0].FormattedAmount = humanize.Commaf(results1[0].Amount) + "円"
	if results1[0].OriginalAmount != nil {
		x := humanize.Commaf(*results1[0].OriginalAmount) + "円"
		results1[0].FormattedOriginalAmount = &x
	}

	priceInfo := system.PriceInfo{
		Amount:                  results1[0].Amount,
		FormattedAmount:         results1[0].FormattedAmount,
		Type:                    results1[0].Type,
		TypeName:                results1[0].TypeName,
		OriginalAmount:          results1[0].OriginalAmount,
		FormattedOriginalAmount: results1[0].FormattedOriginalAmount,
	}
	var attributeInfos []system.AttributeInfo
	for _, k := range results1 {
		attributeInfos = append(attributeInfos, system.AttributeInfo{
			//AttributeID:   k.AttributeID,
			AttributeName:  k.AttributeName,
			AttributeValue: k.AttributeValue,
		})
	}
	targetSKUInfo := system.TargetSKUInfo{
		SkuID: results1[0].SkuID,
		Price: &priceInfo,
		//StockInfo:    &stockInfo,
		//PrimaryImage: &imageInfo,
		Attributes: attributeInfos,
	}
	//然后塞进ProductInfoResponse，只返回一个实例
	productInfoRes = system.ProductInfoResponse{
		ProductCode:     results1[0].ProductCode,
		Name:            results1[0].Name,
		Description:     results1[0].Description,
		IsTaxable:       results1[0].IsTaxable,
		MetaTitle:       results1[0].MetaTitle,
		MetaDescription: results1[0].MetaDescription,
		TargetSKUInfo:   &targetSKUInfo,
		VariantOptions:  newvariantOptions,
		Category:        &categoryInfo,
		Brand:           &brandInfo,
	}
	return productInfoRes, err
}

// @author: [granty1](https://github.com/granty1)
// @function: GetProductReviews
// @description: 获取商品评论信息
// @param product_code path  string true "商品code"
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示多少条，默认10" default(10)
// @param sort  query  int false   "排序1.newest 2.oldest 3.highest_rating 4.lowest_rating 5.most_helpful，默认1" default(1)
// @param rating  query  int false   "只看几星评价"
// @return: reviewListResponse,err
func (productsService *ProductsService) GetProductReviews(productCode string, page int, limit int, sort int, rating int) (reviewListResponse system.ReviewListResponse, err error) {
	offset := (page - 1) * limit
	// var sortfield string
	// var order string
	var orderby string
	switch sort {
	case 1: //最新
		// sortfield = "product_reviews.created_at"
		// order = "DESC"
		orderby = "product_reviews.created_at DESC"
	case 2: //最早
		// sortfield = "product_reviews.created_at"
		// order = "ASC"
		orderby = "product_reviews.created_at ASC"
	case 3: //最高rating开始，相同rating按最新时间顺
		// sortfield = "product_reviews.rating,product_reviews.created_at"
		// order = "DESC"
		orderby = "product_reviews.rating DESC,product_reviews.created_at DESC"
	case 4: //最低rating开始，相同rating按最新时间顺
		// sortfield = "product_reviews.rating"
		// order = "ASC"
		orderby = "product_reviews.rating ASC,product_reviews.created_at DESC"
	case 5: //点赞最多
		// sortfield = "helpful_count"
		// order = "DESC"
		orderby = "helpful_count DESC,product_reviews.created_at DESC"
	default:
		// sortfield = "product_reviews.created_at"
		// order = "DESC"
		orderby = "product_reviews.created_at DESC"
	}
	type Result3 struct {
		//Summary    *ReviewSummary `json:"summary"`    // レビュー集計情報 (Nullable: 商品にレビューがない場合)
		AverageRating float64 `json:"average_rating"`        // 平均評価
		ReviewCount   int     `json:"review_count"`          // 承認済みレビュー総数
		Rating1Count  int     `gorm:"column:rating_1_count"` // `json:"rating_1_count"` // 星1の数
		Rating2Count  int     `gorm:"column:rating_2_count"` //`json:"rating_2_count"` // 星2の数
		Rating3Count  int     `gorm:"column:rating_3_count"` //`json:"rating_3_count"` // 星3の数
		Rating4Count  int     `gorm:"column:rating_4_count"` //`json:"rating_4_count"` // 星4の数
		Rating5Count  int     `gorm:"column:rating_5_count"` //`json:"rating_5_count"` // 星5の数

		//Reviews    []ReviewInfo   `json:"reviews"`    // レビューリスト
		ID                 int     `json:"id"`              // レビューID
		Nickname           string  `json:"nickname"`        // ニックネーム
		Rating             int     `json:"rating"`          // 評価 (1-5)
		Title              *string `json:"title,omitempty"` // タイトル (Nullable)
		Comment            string  `json:"comment"`         // 本文
		CreatedAtFormatted string  `gorm:"column:created_at"`
		ImageUrls          string  `json:"image_urls,omitempty"` // ★添付画像URLリスト (画像がない場合は空配列 or 省略)
		HelpfulCount       int     `json:"helpful_count"`        // ★参考になった数

		//Pagination PaginationInfo `json:"pagination"`
		CurrentPage int `json:"current_page"`
		Limit       int `json:"limit"`
		TotalCount  int `json:"total_count"`
		TotalPages  int `json:"total_pages"`
	}

	var results3 []Result3
	if rating == 0 {
		err = global.GVA_DB.Table("products").
			Select(`
        review_summaries.average_rating AS average_rating,
        review_summaries.review_count AS review_count,
        review_summaries.rating_1_count AS rating_1_count,
        review_summaries.rating_2_count AS rating_2_count,
        review_summaries.rating_3_count AS rating_3_count,
        review_summaries.rating_4_count AS rating_4_count,
        review_summaries.rating_5_count AS rating_5_count,
        product_reviews.id AS id,
        product_reviews.nickname AS nickname,
        product_reviews.rating AS rating,
        product_reviews.title AS title,
        product_reviews.comment AS comment,
        product_reviews.created_at AS created_at,
        GROUP_CONCAT(DISTINCT review_images.image_url SEPARATOR ',') AS image_urls,
        COUNT(user_review_helpful_votes.user_id) AS helpful_count
    `).
			Joins("RIGHT JOIN product_reviews ON products.id = product_reviews.product_id").
			Joins("JOIN review_summaries ON review_summaries.product_id = product_reviews.product_id").
			Joins("LEFT JOIN review_images ON review_images.review_id = product_reviews.id").
			Joins("LEFT JOIN user_review_helpful_votes ON user_review_helpful_votes.review_id = product_reviews.id").
			Where("products.product_code = ? AND product_reviews.status = ?", productCode, "approved").
			Group(`
        product_reviews.id,
        product_reviews.nickname,
        product_reviews.rating,
        product_reviews.title,
        product_reviews.comment,
        product_reviews.created_at,
        review_summaries.average_rating,
        review_summaries.review_count,
        review_summaries.rating_1_count,
        review_summaries.rating_2_count,
        review_summaries.rating_3_count,
        review_summaries.rating_4_count,
        review_summaries.rating_5_count
    `).Order(orderby).
			Limit(limit).
			Offset(offset). // 添加偏移量以支持分页
			Scan(&results3).Error
	} else {
		err = global.GVA_DB.Table("products").
			Select(`
        review_summaries.average_rating AS average_rating,
        review_summaries.review_count AS review_count,
        review_summaries.rating_1_count AS rating_1_count,
        review_summaries.rating_2_count AS rating_2_count,
        review_summaries.rating_3_count AS rating_3_count,
        review_summaries.rating_4_count AS rating_4_count,
        review_summaries.rating_5_count AS rating_5_count,
        product_reviews.id AS id,
        product_reviews.nickname AS nickname,
        product_reviews.rating AS rating,
        product_reviews.title AS title,
        product_reviews.comment AS comment,
        product_reviews.created_at AS created_at,
        GROUP_CONCAT(DISTINCT review_images.image_url SEPARATOR ',') AS image_urls,
        COUNT(user_review_helpful_votes.user_id) AS helpful_count
    `).
			Joins("RIGHT JOIN product_reviews ON products.id = product_reviews.product_id").
			Joins("JOIN review_summaries ON review_summaries.product_id = product_reviews.product_id").
			Joins("LEFT JOIN review_images ON review_images.review_id = product_reviews.id").
			Joins("LEFT JOIN user_review_helpful_votes ON user_review_helpful_votes.review_id = product_reviews.id").
			Where("products.product_code = ? AND product_reviews.status = ?", productCode, "approved").
			Where("product_reviews.rating = ?", rating).
			Group(`
        product_reviews.id,
        product_reviews.nickname,
        product_reviews.rating,
        product_reviews.title,
        product_reviews.comment,
        product_reviews.created_at,
        review_summaries.average_rating,
        review_summaries.review_count,
        review_summaries.rating_1_count,
        review_summaries.rating_2_count,
        review_summaries.rating_3_count,
        review_summaries.rating_4_count,
        review_summaries.rating_5_count
    `).Order(orderby).
			Limit(limit).
			Offset(offset). // 添加偏移量以支持分页
			Scan(&results3).Error
	}

	if len(results3) == 0 {
		//没有查到信息，检测商品是否存在
		var results4 []Result1
		err = global.GVA_DB.Table("products").
			Select("products.product_code as product_code").
			Where("products.product_code=?", productCode).
			Scan(&results4).Error
		if len(results4) > 0 {
			//该商品存在，那么给综合评价赋值
			summary := system.ReviewSummary{
				AverageRating: 0.00,
				ReviewCount:   0,
				Rating1Count:  0,
				Rating2Count:  0,
				Rating3Count:  0,
				Rating4Count:  0,
				Rating5Count:  0,
			}
			paginationInfo := system.PaginationInfo{
				CurrentPage: 1,
				Limit:       limit,
				TotalCount:  0,
				TotalPages:  1,
			}
			reviewListResponse = system.ReviewListResponse{
				Summary: &summary,
				//Reviews:    reviewInfos,
				Pagination: paginationInfo,
			}
			return reviewListResponse, err
		} else {
			//productsredis.Set(key1, string(resultStr))
			err = errors.New("商品不存在")
			return reviewListResponse, err
		}
	}

	//测试用
	//fmt.Println("查询原始结果", results3)

	var reviewInfos []system.ReviewInfo

	//rating默认是0或者我输入的1-5的一个数字

	//格式化时间
	layout := time.RFC3339
	//fmt.Println("初始化前的时间是：", results3[0].CreatedAtFormatted)
	for _, k := range results3 {
		imageUrlsSlice := strings.Split(k.ImageUrls, ",")
		t, _ := time.Parse(layout, k.CreatedAtFormatted)
		createdAtFormatted := t.Format("2006年01月02日")
		reviewInfos = append(reviewInfos, system.ReviewInfo{
			ID:                 k.ID,
			Nickname:           k.Nickname,
			Rating:             k.Rating,
			Title:              k.Title,
			Comment:            k.Comment,
			CreatedAtFormatted: createdAtFormatted,
			ImageUrls:          imageUrlsSlice,
			HelpfulCount:       k.HelpfulCount,
		})
	}
	//totalCount := len(reviewInfos)
	summary := system.ReviewSummary{
		AverageRating: results3[0].AverageRating,
		ReviewCount:   results3[0].ReviewCount,
		Rating1Count:  results3[0].Rating1Count,
		Rating2Count:  results3[0].Rating2Count,
		Rating3Count:  results3[0].Rating3Count,
		Rating4Count:  results3[0].Rating4Count,
		Rating5Count:  results3[0].Rating5Count,
	}
	totalPages := math.Ceil(float64(results3[0].ReviewCount) / float64(limit))
	paginationInfo := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  results3[0].ReviewCount,
		TotalPages:  int(totalPages),
		//这不对
	}
	reviewListResponse = system.ReviewListResponse{
		Summary:    &summary,
		Reviews:    reviewInfos,
		Pagination: paginationInfo,
	}
	return reviewListResponse, err
}
