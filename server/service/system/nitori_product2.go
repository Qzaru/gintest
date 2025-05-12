package system

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

//type ProductsService struct{}

// var ProductsServiceApp = new(ProductsService)
// var productsredis = products.NewProductsRedisStore()

// @author: [granty1](https://github.com/granty1)
// @function: GetProductQA
// @description: 获取商品Q&A信息
// @param product_code path  string true "商品code"
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示多少条，默认10" default(10)
// @param sort  query  int false   "排序1.newest 2.oldest 3.most_helpful，默认1" default(1)
// @return: reviewListResponse,err
func (productsService *ProductsService) GetProductQA(productCode string, page int, limit int, sort int) (qaListResponse system.QAListResponse, err error) {
	type ResultQA struct {
		// QAList     []QAInfo       `json:"qa_list"`    // Q&Aリスト
		// Question *QuestionInfo `json:"question"`
		QuestionID          int64  `gorm:"column:questionid"`    // 質問ID
		QuestionText        string `gorm:"column:question_text"` // 質問本文
		QCreatedAtFormatted string `gorm:"column:qcreated_at"`   // 表示用投稿日時
		//Answer   *AnswerInfo   `json:"answer,omitempty"` // 回答情報 (回答がない場合は null)
		AnswerID            int64  `gorm:"column:answersid"`     // 回答ID
		AnswererName        string `gorm:"column:answerer_name"` // 回答者表示名
		AnswerText          string `gorm:"column:answer_text"`   // 回答本文
		HelpfulCount        int    `gorm:"column:helpful_count"` // 参考になった数
		ACreatedAtFormatted string `gorm:"column:acreated_at"`   // 表示用回答日時

		//Pagination PaginationInfo `json:"pagination"` // ページネーション情報
		CurrentPage int `json:"current_page"`
		Limit       int `json:"limit"`
		TotalCount  int `gorm:"column:total_count"`
		TotalPages  int `json:"total_pages"`
	}

	offset := (page - 1) * limit
	// var sortfield string
	// var order string
	var orderby string
	switch sort {
	case 1: //最新
		orderby = "product_questions.created_at DESC"
	case 2: //最早
		orderby = "product_questions.created_at ASC"
	case 3: //最多点赞
		orderby = "helpful_count DESC,product_questions.created_at DESC"
	default:
		orderby = "product_questions.created_at DESC"
	}
	// 执行查询
	var resultqas []ResultQA
	var totalCount int
	err = global.GVA_DB.Raw(`
		SELECT SQL_CALC_FOUND_ROWS DISTINCT 
		product_questions.id AS questionid,
		product_questions.question_text AS question_text,
		product_questions.created_at AS qcreated_at,
		question_answers.id AS answersid,
		question_answers.answerer_name AS answerer_name,
		question_answers.answer_text AS answer_text,
		COUNT(user_answer_helpful_votes.user_id) AS helpful_count,
		question_answers.created_at AS acreated_at
		FROM products
		LEFT JOIN product_questions ON product_questions.product_id = products.id
		LEFT JOIN question_answers ON question_answers.question_id = product_questions.id
		LEFT JOIN user_answer_helpful_votes ON user_answer_helpful_votes.answer_id = question_answers.id
		WHERE products.product_code = ? AND product_questions.status = ? AND question_answers.status = ?
		GROUP BY 
		product_questions.id,
		product_questions.question_text,
		product_questions.created_at,
		question_answers.id,
		question_answers.answerer_name,
		question_answers.answer_text,
		question_answers.created_at
		ORDER BY `+orderby+`
		LIMIT ? OFFSET ?`, productCode, "approved", "approved", limit, offset).Scan(&resultqas).Error

	// 获取总行数
	err = global.GVA_DB.Raw("SELECT FOUND_ROWS()").Scan(&totalCount).Error

	if len(resultqas) == 0 {
		//没有查到信息，检测商品是否存在
		var results4 []Result1
		err = global.GVA_DB.Table("products").
			Select("products.product_code as product_code").
			Where("products.product_code=?", productCode).
			Scan(&results4).Error
		if len(results4) > 0 {
			paginationInfo := system.PaginationInfo{
				CurrentPage: 1,
				Limit:       limit,
				TotalCount:  0,
				TotalPages:  1,
			}
			qaListResponse = system.QAListResponse{
				Pagination: paginationInfo,
			}
			return qaListResponse, err
		} else {
			//productsredis.Set(key1, string(resultStr))
			err = errors.New("商品不存在")
			return qaListResponse, err
		}
	}
	// var questionInfo system.QuestionInfo
	// var answerInfo system.AnswerInfo
	totalPages := math.Ceil(float64(totalCount) / float64(limit))
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  totalCount,
		TotalPages:  int(totalPages),
	}
	var qaInfos []system.QAInfo
	layout := time.RFC3339
	for _, k := range resultqas {
		t, _ := time.Parse(layout, k.QCreatedAtFormatted)
		qCreatedAtFormatted := t.Format("2006年01月02日")
		s, _ := time.Parse(layout, k.ACreatedAtFormatted)
		aCreatedAtFormatted := s.Format("2006年01月02日")

		questionInfo := system.QuestionInfo{
			ID:                 k.QuestionID,
			QuestionText:       k.QuestionText,
			CreatedAtFormatted: qCreatedAtFormatted,
		}
		answerInfo := system.AnswerInfo{
			ID:                 k.AnswerID,
			AnswererName:       k.AnswererName,
			AnswerText:         k.AnswerText,
			HelpfulCount:       k.HelpfulCount,
			CreatedAtFormatted: aCreatedAtFormatted,
		}
		qaInfo := system.QAInfo{
			Question: &questionInfo,
			Answer:   &answerInfo,
		}
		qaInfos = append(qaInfos, qaInfo)
	}

	qaListResponse = system.QAListResponse{
		QAList:     qaInfos,
		Pagination: pagination,
	}
	return qaListResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @function: GetProductSkuImages
// @description: 获取商品图片信息
// @param sku_id path  string true "商品skuid"
// @return: skuImageInfo,err
func (productsService *ProductsService) GetProductSkuImages(skuid string) (skuImageInfo []system.SKUImageInfo, err error) {
	//var skuImageInfo []system.SKUImageInfo

	err = global.GVA_DB.Table("sku_images").
		Select("id, main_image_url, thumbnail_url, alt_text, sort_order").
		Where("sku_id = ?", skuid).
		Scan(&skuImageInfo).Error
	return skuImageInfo, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取商品收藏信息
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @param sort  query  int false   "排序1.最新开始 2.最早开始 默认1" default(1)
// @return: favoriteSKUListResponse,err
func (productsService *ProductsService) GetFavorites(userID uint, page int, limit int, sort int) (favoriteSKUListResponse system.FavoriteSKUListResponse, err error) {
	//测试用
	//userID = 101
	type ResultsFavorite struct {
		//Favorites  []FavoriteSKUInfo `json:"favorites"`  // お気に入りSKUリスト
		SkuID       string `gorm:"column:sku_id"`        // SKU ID
		ProductID   string `gorm:"column:products_id"`   // 商品ID
		ProductName string `gorm:"column:products_name"` // 商品名 (省略後)
		ProductCode string `gorm:"column:product_code"`  // 商品コード
		Status      string `gorm:"column:status"`        //新加的，为了区分商品状态
		//Price             *PriceInfo      `json:"price"`                        // 現在の表示価格情報 (Nullable)
		Amount                  float64 `gorm:"column:amount"` //这是一个需要判断是否在促销期内的促销价格，否则为初始价格
		FormattedAmount         string
		Type                    string   `gorm:"column:type"`
		TypeName                string   `gorm:"column:type_name"`
		OriginalAmount          *float64 `gorm:"column:original_amount"` //这是一个通常价格
		FormattedOriginalAmount *string
		//PrimaryImage      *ImageInfo      `json:"primary_image,omitempty"`    // 代表画像 (Nullable)
		ID      int     `gorm:"column:image_id"`
		URL     string  `gorm:"column:image_url"`
		AltText *string `gorm:"column:alt_text"`

		Attributes       []system.AttributeInfo `json:"attributes"` // 対象SKUの属性リスト
		AttributeName    string
		AttributeValue   string
		NameValue        string `gorm:"column:name_value"`
		AddedAtFormatted string `gorm:"column:added_at_formatted"` // お気に入り追加日時 (表示用)
		//Pagination PaginationInfo    `json:"pagination"` // ページネーション情報
		CurrentPage int `gorm:"column:current_page"`
		Limit       int `gorm:"column:limit"`
		TotalCount  int `gorm:"column:total_count"`
		TotalPages  int `gorm:"column:total_pages"`
	}
	var resultsFavorite []ResultsFavorite

	offset := (page - 1) * limit
	var orderby string
	switch sort {
	case 1: //最新
		orderby = "user_favorite_skus.created_at DESC"
	case 2: //最早
		orderby = "user_favorite_skus.created_at ASC"
	default:
		orderby = "user_favorite_skus.created_at DESC"
	}
	err = global.GVA_DB.Raw(`
SELECT SQL_CALC_FOUND_ROWS DISTINCT 
    product_skus.id AS sku_id, 
    products.id AS products_id, 
    products.name AS products_name, 
    products.product_code AS product_code, 
	product_skus.status AS status,
    MAX(CASE WHEN prices.price_type_id = 1 THEN prices.price END) AS amount,
    MAX(CASE WHEN prices.price_type_id = 1 THEN price_types.type_code END) AS type,
    MAX(CASE WHEN prices.price_type_id = 1 THEN price_types.name END) AS type_name,
    MAX(CASE WHEN prices.price_type_id = 2 AND NOW() BETWEEN prices.start_date AND prices.end_date THEN prices.price END) AS original_amount,
    (
        SELECT sku_images.id
        FROM sku_images 
        WHERE sku_images.sku_id = product_skus.id 
        ORDER BY sku_images.id ASC
        LIMIT 1
    ) AS image_id,
    (
        SELECT sku_images.main_image_url
        FROM sku_images 
        WHERE sku_images.sku_id = product_skus.id 
        ORDER BY sku_images.id ASC
        LIMIT 1
    ) AS image_url,
    (
        SELECT sku_images.alt_text
        FROM sku_images 
        WHERE sku_images.sku_id = product_skus.id 
        ORDER BY sku_images.id ASC
        LIMIT 1
    ) AS alt_text,
    GROUP_CONCAT(DISTINCT CONCAT (attributes.name,'-',
        COALESCE(attribute_options.value, sku_values.value_string, sku_values.value_number, sku_values.value_boolean))
        ORDER BY attributes.sort_order
    ) AS name_value,
    user_favorite_skus.created_at AS added_at_formatted
FROM user_favorite_skus
JOIN product_skus ON user_favorite_skus.sku_id=product_skus.id
JOIN products ON product_skus.product_id = products.id
LEFT JOIN sku_images ON sku_images.sku_id=product_skus.id
JOIN prices ON prices.sku_id = product_skus.id
JOIN inventory ON inventory.sku_id = product_skus.id
JOIN sku_availability ON sku_availability.sku_id = product_skus.id
JOIN categories ON categories.id = products.category_id
JOIN price_types ON price_types.id = prices.price_type_id
JOIN inventory_locations ON inventory_locations.id = inventory.location_id
LEFT JOIN sku_values ON sku_values.sku_id = product_skus.id
LEFT JOIN attributes ON attributes.id = sku_values.attribute_id
LEFT JOIN attribute_options ON attribute_options.id = sku_values.option_id
WHERE user_favorite_skus.user_id = ?
  AND (
        prices.price_type_id = 1
        OR (
            prices.price_type_id = 2 
            AND NOW() BETWEEN prices.start_date AND prices.end_date
        )
      )
GROUP BY 
    products.product_code,
    products.name,
    products.id,
    product_skus.id,
    user_favorite_skus.created_at
	ORDER BY `+orderby+`
		LIMIT ? OFFSET ?`, userID, limit, offset).Scan(&resultsFavorite).Error
	// 获取总行数
	var totalCount int
	err = global.GVA_DB.Raw("SELECT FOUND_ROWS()").Scan(&totalCount).Error
	totalPages := math.Ceil(float64(totalCount) / float64(limit))
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  totalCount,
		TotalPages:  int(totalPages),
	}
	var favorites []system.FavoriteSKUInfo
	for _, k := range resultsFavorite {
		parts := strings.Split(k.NameValue, ",")

		// 2. 创建 AttributeInfo 切片
		var attributes []system.AttributeInfo

		// 3. 拆分每个元素并存入结构体
		for _, part := range parts {
			part = strings.TrimSpace(part)
			kv := strings.SplitN(part, "-", 2)
			if len(kv) == 2 {
				attr := system.AttributeInfo{
					AttributeName:  strings.TrimSpace(kv[0]),
					AttributeValue: strings.TrimSpace(kv[1]),
				}
				attributes = append(attributes, attr)
			}
		}
		primaryImage := system.ImageInfo{
			ID:      k.ID,
			URL:     k.URL,
			AltText: k.AltText,
		}
		if utf8.RuneCountInString(k.ProductName) > 20 {
			runes := []rune(k.ProductName)
			k.ProductName = string(runes[:20]) + "..."
		}

		//价格转换
		if k.OriginalAmount != nil {
			x := k.Amount
			k.Amount = *k.OriginalAmount
			k.Type = "sale"
			k.TypeName = "セール価格"
			k.OriginalAmount = &x
		}

		k.FormattedAmount = humanize.Commaf(k.Amount) + "円"
		if k.OriginalAmount != nil {
			x := humanize.Commaf(*k.OriginalAmount) + "円"
			k.FormattedOriginalAmount = &x
		}

		priceInfo := system.PriceInfo{
			Amount:                  k.Amount,
			FormattedAmount:         k.FormattedAmount,
			Type:                    k.Type,
			TypeName:                k.TypeName,
			OriginalAmount:          k.OriginalAmount,
			FormattedOriginalAmount: k.FormattedOriginalAmount,
		}
		layout := time.RFC3339
		t, _ := time.Parse(layout, k.AddedAtFormatted)
		time := t.Format("2006年01月02日")
		favorites = append(favorites, system.FavoriteSKUInfo{
			SkuID:            k.SkuID,
			ProductID:        k.ProductID,
			ProductName:      k.ProductName,
			ProductCode:      k.ProductCode,
			Status:           k.Status,
			Price:            &priceInfo,
			PrimaryImage:     &primaryImage,
			Attributes:       attributes,
			AddedAtFormatted: time,
		})
	}

	favoriteSKUListResponse = system.FavoriteSKUListResponse{
		Favorites:  favorites,
		Pagination: pagination,
	}
	return favoriteSKUListResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 设置商品收藏信息
// @param shuid  query  string false   "要收藏的商品sku"
// @return: favoriteSKUListResponse,err
func (productsService *ProductsService) SetFavorites(userID uint, skuid string) (err error) {
	//如已经存在，返回错误
	var uf []system.UserFavoriteSkus
	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).First(&uf).Error // 判断用户名是否注册
	if len(uf) != 0 {
		return errors.New("已收藏")
	}
	err = global.GVA_DB.Create(&system.UserFavoriteSkus{
		UserId:    userID,
		SkuId:     skuid,
		CreatedAt: time.Now(),
	}).Error
	return
}

// @author: [granty1](https://github.com/granty1)
// @description: 删除商品收藏信息
// @param userid  query  int true   "当前用户"
// @param skuid query  int true  "要删除的商品"
// @return: err
func (productsService *ProductsService) DeleteFavorites(userID uint, skuid string) (err error) {
	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).Delete(&system.UserFavoriteSkus{}).Error
	return err
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取关联商品
// @param product_code path  string true "商品code"
// @param limit  query  int false   "每页显示几条，默认6" default(6)
// @return: relatedProductListResponse,err
func (productsService *ProductsService) GetRelateProducts(productCode string, limit int) (relatedProductListResponse system.RelatedProductListResponse, err error) {

	type ResultRelated struct {
		ProductID           string `json:"product_id"`             // 関連商品のID
		ProductCode         string `json:"product_code,omitempty"` // 関連商品のコード
		ProductName         string `json:"product_name"`           // 関連商品の名称 (省略後)
		PriceRangeFormatted string `json:"price_range_formatted"`  // ★価格帯文字列 (例: "2,990～3,990円")
		ValidSalePrice      string `json:"valid_sale_price"`
		IsOnSale            bool   `json:"is_on_sale"`
		// ★値下げフラグ
		//ReviewSummary       *ReviewSummaryInfo `json:"review_summary,omitempty"`   // ★レビュー集計情報 (Nullable)
		AverageRating float64 `json:"average_rating"` // 平均評価
		ReviewCount   int     `json:"review_count"`   // レビュー件数

		ThumbnailImageURL *string `json:"thumbnail_image_url,omitempty"` // 代表SKUのサムネイル画像URL (Nullable)
	}
	var resultsrelated []ResultRelated
	err = global.GVA_DB.Raw(`
        SELECT DISTINCT 
            p2.id AS product_id,
            p2.product_code AS product_code,
            p2.name AS product_name,
            (
                SELECT sku_images.thumbnail_url
                FROM sku_images 
                WHERE sku_images.sku_id = p2.default_sku_id 
                ORDER BY sku_images.id ASC
                LIMIT 1
            ) AS thumbnail_image_url,
            review_summaries.average_rating AS average_rating,
            review_summaries.review_count AS review_count,
            GROUP_CONCAT(DISTINCT pr3.price SEPARATOR ',') as price_range_formatted,
            GROUP_CONCAT(DISTINCT sale_price SEPARATOR ',') as valid_sale_price,
            p2.updated_at
        FROM products p1
        JOIN products p2 ON p1.category_id = p2.category_id
        JOIN review_summaries ON review_summaries.product_id = p2.id
        RIGHT JOIN product_skus ON product_skus.product_id = p2.id
        JOIN prices pr1 ON pr1.sku_id = product_skus.id
        JOIN (
            SELECT DISTINCT pr1.sku_id,
                pr1.price,
                CASE WHEN NOW() BETWEEN pr2.start_date AND pr2.end_date THEN pr2.price ELSE NULL END AS sale_price
            FROM prices pr1
            JOIN prices pr2 ON pr1.sku_id = pr2.sku_id
            WHERE pr1.price_type_id = 1 
        ) pr3 ON pr3.sku_id = product_skus.id
        WHERE p1.product_code = ? AND p2.product_code <> ? 
        GROUP BY
            p2.id,
            p2.product_code,
            p2.name,
            review_summaries.average_rating,
            review_summaries.review_count,
            p2.updated_at
        ORDER BY
            valid_sale_price DESC,
            p2.updated_at DESC;
    `,

		productCode, productCode).Limit(limit).Scan(&resultsrelated).Error
	//测试用
	fmt.Println("原始查询结果是：", resultsrelated)
	relatedProducts := []system.RelatedProductInfoV1_1{}
	for _, k := range resultsrelated {
		//取出价格，格式化
		priceslice := strings.Split(k.PriceRangeFormatted, ",")
		minPrice, _ := Min(priceslice)
		maxPrice, _ := Max(priceslice)
		//检查是否有有效优惠价
		if k.ValidSalePrice != "" {
			validSalePrice, _ := strconv.ParseFloat(k.ValidSalePrice, 64)
			if int(validSalePrice) < minPrice {
				minPrice = int(validSalePrice)
			}
			//标记是否减价
			k.IsOnSale = true
		}

		if minPrice != maxPrice {
			k.PriceRangeFormatted = humanize.Commaf(float64(minPrice)) + "~" + humanize.Commaf(float64(maxPrice)) + "円"
		} else if minPrice == maxPrice {
			k.PriceRangeFormatted = humanize.Commaf(float64(maxPrice)) + "円"
		}

		reviewSummaryInfo := system.ReviewSummaryInfo{
			AverageRating: k.AverageRating,
			ReviewCount:   k.ReviewCount,
		}
		relatedProductInfoV1_1 := system.RelatedProductInfoV1_1{
			ProductID:           k.ProductID,
			ProductCode:         k.ProductCode,
			ProductName:         k.ProductName,
			PriceRangeFormatted: k.PriceRangeFormatted,
			IsOnSale:            k.IsOnSale,
			ReviewSummary:       &reviewSummaryInfo,
			ThumbnailImageURL:   k.ThumbnailImageURL,
		}
		relatedProducts = append(relatedProducts, relatedProductInfoV1_1)
	}
	relatedProductListResponse = system.RelatedProductListResponse{
		RelatedProducts: relatedProducts,
	}

	return relatedProductListResponse, err
}

// 写两个关于求string切片中最大最小值的函数以便调用
func Max(slice []string) (int, error) {
	if len(slice) == 0 {
		return -1, fmt.Errorf("切片不能为空")
	}

	maxValue := slice[0]
	maxNum, err := strconv.ParseFloat(maxValue, 64)
	if err != nil {
		return -1, err
	}

	for _, str := range slice[1:] {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return -1, err
		}
		if num > maxNum {
			maxNum = num
			//maxValue = str
		}
	}
	return int(maxNum), nil
}

func Min(slice []string) (int, error) {
	if len(slice) == 0 {
		return -1, fmt.Errorf("切片不能为空")
	}

	minValue := slice[0]
	minNum, err := strconv.ParseFloat(minValue, 64)
	if err != nil {
		return -1, err
	}

	for _, str := range slice[1:] {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return -1, err
		}
		if num < minNum {
			minNum = num
			//minValue = str
		}
	}
	return int(minNum), nil
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取关联商品搭配
// @param product_code path  string true "商品code"
// @param limit  query  int false   "每页显示几条，默认6" default(6)
// @return: coordinateSetTeaserListResponse,err
func (productsService *ProductsService) GetCoordinateSet(productCode string, limit int) (coordinateSetTeaserListResponse system.CoordinateSetTeaserListResponse, err error) {
	var coordinateSetTeaserInfo []system.CoordinateSetTeaserInfo
	err = global.GVA_DB.Table("products").
		Select("coordinate_sets.id as set_id, coordinate_sets.theme_image_url as set_theme_image_url, coordinate_sets.contributor_nickname as contributor_nickname, coordinate_sets.contributor_avatar_url as contributor_avatar_url, coordinate_sets.contributor_store_name as contributor_store_name").
		Joins("JOIN coordinate_set_items ON products.id = coordinate_set_items.product_id").
		Joins("JOIN coordinate_sets ON coordinate_set_items.coordinate_set_id = coordinate_sets.id").
		Where("products.product_code = ?", productCode).
		Order("coordinate_sets.sort_order ASC").
		Limit(limit).
		Scan(&coordinateSetTeaserInfo).Error
	coordinateSetTeaserListResponse = system.CoordinateSetTeaserListResponse{
		Coordinates: coordinateSetTeaserInfo,
	}
	return coordinateSetTeaserListResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 获取关联商品
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @return: relatedProductListResponse,err
func (productsService *ProductsService) GetViewHistory(userID uint, page int, limit int) (viewedSKUListResponseV1_1 system.ViewedSKUListResponseV1_1, err error) {
	type ResultsViewed struct {
		//History    []ViewedSKUInfoV1_1 `json:"history"`    // 閲覧履歴SKUリスト
		SkuID               string `json:"sku_id"`                 // SKU ID
		ProductID           string `json:"product_id"`             // 商品ID
		ProductName         string `json:"product_name"`           // 商品名 (省略後)
		ProductCode         string `json:"product_code,omitempty"` // 商品コード
		PriceRangeFormatted string `json:"price_range_formatted"`  // ★ 商品の価格帯文字列
		ValidSalePrice      string
		//PrimaryImage        *ImageInfo         `json:"primary_image,omitempty"`  // サムネイル画像推奨 (Nullable)
		ID      int     `json:"image_id"`
		URL     string  `json:"url"`
		AltText *string `json:"alt_text,omitempty"`
		//ReviewSummary       *ReviewSummaryInfo `json:"review_summary,omitempty"` // ★ 商品のレビュー集計情報 (Nullable)
		AverageRating float64 `json:"average_rating"` // 平均評価
		ReviewCount   int     `json:"review_count"`   // レビュー件数

		ViewedAtFormatted string `json:"viewed_at_formatted"` // 最終閲覧日時 (表示用)

		//Pagination PaginationInfo      `json:"pagination"` // ページネーション情報
		CurrentPage int `json:"current_page"`
		Limit       int `json:"limit"`
		TotalCount  int `json:"total_count"`
		TotalPages  int `json:"total_pages"`
	}
	offset := (page - 1) * limit
	var resultsViewed []ResultsViewed
	err = global.GVA_DB.Table("user_viewed_skus").
		Select(`
SQL_CALC_FOUND_ROWS distinct
user_viewed_skus.sku_id as sku_id,
 products.id  as  product_id,
products.name	as	product_name,
products.product_code  as	product_code,
 GROUP_CONCAT(DISTINCT pr3.price SEPARATOR ',') as price_range_formatted,
GROUP_CONCAT(DISTINCT sale_price SEPARATOR ',') as valid_sale_price,


 (
    SELECT sku_images.id
    FROM sku_images 
    WHERE sku_images.sku_id = user_viewed_skus.sku_id
    ORDER BY sku_images.id ASC
    LIMIT 1
  ) as  image_id,
(
    SELECT sku_images.thumbnail_url
    FROM sku_images 
    WHERE sku_images.sku_id = user_viewed_skus.sku_id
    ORDER BY sku_images.id ASC
    LIMIT 1
  ) as	url,
(
    SELECT sku_images.alt_text
    FROM sku_images 
    WHERE sku_images.sku_id = user_viewed_skus.sku_id
    ORDER BY sku_images.id ASC
    LIMIT 1
  )  as   alt_text,
review_summaries.average_rating AS average_rating,
review_summaries.review_count AS review_count,
		
user_viewed_skus.viewed_at	as	ViewedAtFormatted`).
		Joins("JOIN product_skus product_skus1 ON product_skus1.id = user_viewed_skus.sku_id").
		Joins("JOIN product_skus product_skus2 ON product_skus1.product_id = product_skus2.product_id").
		Joins("JOIN products ON product_skus1.product_id = products.id").
		Joins("LEFT JOIN review_summaries ON review_summaries.product_id = products.id").
		Joins("JOIN prices pr1 ON pr1.sku_id = product_skus2.id").
		Joins(`JOIN (
        SELECT distinct pr1.sku_id,pr1.price,
               CASE WHEN NOW() BETWEEN pr2.start_date AND pr2.end_date THEN pr2.price ELSE NULL END as sale_price
        FROM prices pr1
        JOIN prices pr2 ON pr1.sku_id=pr2.sku_id
        WHERE pr1.price_type_id=1
    ) pr3 ON pr3.sku_id=product_skus2.id`).
		Where("user_viewed_skus.user_id = ?", userID).
		Group("user_viewed_skus.sku_id, products.id, products.name, products.product_code, review_summaries.average_rating, review_summaries.review_count, user_viewed_skus.viewed_at").
		Offset(offset).Limit(limit).
		Find(&resultsViewed).Error
	fmt.Println("查询结果是：", resultsViewed)
	var totalCount int
	err = global.GVA_DB.Raw("SELECT FOUND_ROWS()").Scan(&totalCount).Error
	totalPages := math.Ceil(float64(totalCount) / float64(limit))
	history := []system.ViewedSKUInfoV1_1{}
	for _, k := range resultsViewed {
		//取出价格，格式化
		priceslice := strings.Split(k.PriceRangeFormatted, ",")
		minPrice, _ := Min(priceslice)
		maxPrice, _ := Max(priceslice)
		//检查是否有有效优惠价
		if k.ValidSalePrice != "" {
			validSalePrice, _ := strconv.ParseFloat(k.ValidSalePrice, 64)
			if int(validSalePrice) < minPrice {
				minPrice = int(validSalePrice)
			}

		}

		if minPrice != maxPrice {
			k.PriceRangeFormatted = humanize.Commaf(float64(minPrice)) + "~" + humanize.Commaf(float64(maxPrice)) + "円"
		} else if minPrice == maxPrice {
			k.PriceRangeFormatted = humanize.Commaf(float64(maxPrice)) + "円"
		}

		reviewSummaryInfo := system.ReviewSummaryInfo{
			AverageRating: k.AverageRating,
			ReviewCount:   k.ReviewCount,
		}
		imageInfo := system.ImageInfo{
			ID:      k.ID,
			URL:     k.URL,
			AltText: &k.URL,
		}
		layout := time.RFC3339
		t, _ := time.Parse(layout, k.ViewedAtFormatted)
		time := t.Format("2006年01月02日")
		viewedSKUInfoV1_1 := system.ViewedSKUInfoV1_1{
			SkuID:               k.SkuID,
			ProductID:           k.ProductID,
			ProductName:         k.ProductName,
			ProductCode:         k.ProductCode,
			PriceRangeFormatted: k.PriceRangeFormatted,
			PrimaryImage:        &imageInfo,
			ReviewSummary:       &reviewSummaryInfo,
			ViewedAtFormatted:   time,
		}
		history = append(history, viewedSKUInfoV1_1)
	}
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  totalCount,
		TotalPages:  int(totalPages),
	}
	viewedSKUListResponseV1_1 = system.ViewedSKUListResponseV1_1{
		History:    history,
		Pagination: pagination,
	}

	return viewedSKUListResponseV1_1, err

}

// @author: [granty1](https://github.com/granty1)
// @description: 添加商品浏览记录
// @param shuid  query  string false   "要添加的浏览记录的sku"
// @return: favoriteSKUListResponse,err
func (productsService *ProductsService) SetViewHistory(userID uint, skuid string) (err error) {
	//如已经存在，返回错误
	var uv system.UserViewedSkus
	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).First(&uv).Error // 判断用户名是否注册
	if err == nil {
		err = global.GVA_DB.Model(&uv).Where("user_id = ? AND sku_id= ?", userID, skuid).Updates(system.UserViewedSkus{ViewedAt: time.Now()}).Error

	} else if err == gorm.ErrRecordNotFound {
		err = global.GVA_DB.Create(&system.UserViewedSkus{
			UserId:   userID,
			SkuId:    skuid,
			ViewedAt: time.Now(),
		}).Error
	}
	return
}
