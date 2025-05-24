package system

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	payjp "github.com/payjp/payjp-go/v1"
	"gorm.io/gorm"
)

// @author: [piexlmax](https://github.com/piexlmax)
// @function: OrderPay
// @description: 在订单中使用积分
// @param: auth model.ApplyCouponRequest
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) OrderPay(userID uint) (orderInfo system.OrderInfo, err error) {
	//调用payjp的api完成支付
	var gettotalamount system.CheckoutSession
	err = global.GVA_DB.Table("checkout_sessions").
		Where("checkout_sessions.user_id = ?", userID).
		First(&gettotalamount).Error
	pay := payjp.New("sk_test_17d54a098cd910376999acdc", nil)

	p := payjp.Token{
		Number:   "4242424242424242",
		ExpMonth: "2",
		ExpYear:  "2099",
		CVC:      "123",
	}
	p.Name = "pay taro"
	token, err := pay.Token.Create(p)

	//var tokenToCharge string = "tok_visa"
	charge, err := pay.Charge.Create(int(gettotalamount.TotalAmount), payjp.Charge{
		// 現在はjpyのみサポート
		Currency:  "jpy",
		CardToken: token.ID,
		Capture:   true,
	})
	if err != nil {
		fmt.Println("测试支付结果:", charge)
		return system.OrderInfo{}, err
	}
	//cp, err := productsService.CreatePayment(userID, true)

	//从checkoutsession表的数据生成order表
	var cs system.CheckoutSession
	err = global.GVA_DB.Table("checkout_sessions").
		Where("user_id = ?", userID).
		Order("user_id ASC").
		Limit(1).
		Find(&cs).Error
	now := time.Now()
	nowstr := now.Format("2006-01-02 15:04:05")
	userIDstr := strconv.FormatUint(uint64(userID), 10)
	ordercode := userIDstr + nowstr
	err = global.GVA_DB.Create(&system.Order{
		// 	ID                 :,
		OrderCode:            ordercode,
		UserID:               uint64(userID),
		OrderStatus:          "payment_confirmed",
		SubtotalAmount:       cs.CartSubtotal,
		CouponID:             cs.AppliedCouponID,
		CouponDiscountAmount: cs.CouponDiscountAmount,
		PointsUsed:           uint(cs.UsedPoints),
		PointsDiscountAmount: cs.PointsDiscountAmount,
		ShippingFee:          cs.ShippingFee,
		TotalAmount:          cs.TotalAmount,
		PayjpChargeID:        &charge.ID,
		OrderedAt:            now,
		PaidAt:               &charge.CreatedAt,
		//CancelledAt:          now,
		//Notes: charge.Metadata,
		// CreatedAt           :,
		// UpdatedAt           :,
	}).Error
	if err != nil {
		return system.OrderInfo{}, err
	}
	//查出orderid
	var findorderid system.Order
	err = global.GVA_DB.Table("orders").
		Where("order_code = ?", ordercode).
		Order("order_code ASC").
		Limit(1).
		Find(&findorderid).Error
	//用user_cart_items中的数据生成order_items表的数据
	var oi []system.OrderItem
	err = global.GVA_DB.Table("user_cart_items").
		Select(`
        products.id as product_id,
        user_cart_items.sku_id as sku_id,
        products.name as product_name,
        product_skus.sku_code as sku_code,
        user_cart_items.quantity as quantity,
        min(prices.price) as unit_price,
        user_cart_items.quantity * min(prices.price) as subtotal_price`).
		Joins(`join product_skus on product_skus.id = user_cart_items.sku_id`).
		Joins(`join products on products.id = product_skus.product_id`).
		Joins(`join prices on prices.sku_id = user_cart_items.sku_id`).
		Where(`
        user_cart_items.user_id = ? AND (
            (prices.start_date IS NOT NULL AND prices.end_date IS NOT NULL AND now() BETWEEN prices.start_date AND prices.end_date)
            OR
            (prices.start_date IS NULL OR prices.end_date IS NULL)
        )`, "101").
		Group(`
        products.id,
        user_cart_items.sku_id,
        products.name,
        product_skus.sku_code,
        user_cart_items.quantity`).
		Find(&oi).Error
	fmt.Println("oi的查询结果是：", oi)
	//把前面查到的orderid遍历塞进去
	for i, _ := range oi {
		oi[i].OrderID = findorderid.ID
	}
	if err = global.GVA_DB.Create(&oi).Error; err != nil {
		return system.OrderInfo{}, err
	} //这里失败了
	//修改数据，库存-、coupon标记使用、point-、usercartitems表删除、checkoutsession表删除，注意设置回滚
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		//oi中已查出订单商品数量，遍历它去修改库存
		var oldinventory system.Inventory
		var fi []system.Inventory
		//得先查出库存比对单个库存地点的库存够不够，分成够和不够两种操作
		//目前一个skuid对应1或2个库存地址，之后可能会更多
		for _, k := range oi {
			err = global.GVA_DB.Table("inventory").
				Where("sku_id = ?", k.SkuID).
				Order("location_id ASC").
				Find(&fi).Error
			//比较第一个仓库的库存，够
			if k.Quantity <= fi[0].Quantity {
				if err = global.GVA_DB.Model(&oldinventory).
					Where("sku_id = ? and location_id = ?", k.SkuID, 1).
					Updates(system.Inventory{
						Quantity: fi[0].Quantity - k.Quantity,
					}).Error; err != nil {
					return err
				}
			} else if k.Quantity > fi[0].Quantity && len(fi) > 1 { //如果第一个不够，就把第一个先清零，差的从第二个里扣
				nokoru := k.Quantity - fi[0].Quantity
				if err = global.GVA_DB.Model(&oldinventory).
					Where("sku_id = ? and location_id = ?", k.SkuID, 1).
					Updates(system.Inventory{
						Quantity: 0,
					}).Error; err != nil {
					return err
				}
				if err = global.GVA_DB.Model(&oldinventory).
					Where("sku_id = ? and location_id = ?", k.SkuID, 101).
					Updates(system.Inventory{
						Quantity: fi[1].Quantity - nokoru,
					}).Error; err != nil {
					return err
				}
			}
		}
		//处理coupon
		var userCoupon system.UserCoupon
		useat := time.Now()
		updates2 := map[string]interface{}{
			"is_used":  1,
			"used_at":  &useat,
			"order_id": oi[0].OrderID,
		}
		if err = global.GVA_DB.Model(&userCoupon).
			Where("user_id = ? AND coupon_id = ?", userID, cs.AppliedCouponID).
			Updates(updates2).Error; err != nil {
			return err
		}
		//处理point
		var userPoints system.UserPoints
		var olduserPoints system.UserPoints
		err = global.GVA_DB.Model(&userPoints).
			Where("user_id = ?", userID).
			First(&olduserPoints).Error
		if err = global.GVA_DB.Model(&userPoints).
			Where("user_id = ?", userID).
			Update("AvailablePoints", olduserPoints.AvailablePoints-cs.UsedPoints).Error; err != nil {
			return err
		}
		//处理checkoutsession
		if err = global.GVA_DB.Where("user_id = ? ", userID).Delete(&system.CheckoutSession{}).Error; err != nil {
			return err
		}
		//处理cart
		if err = global.GVA_DB.Where("user_id = ? ", userID).Delete(&system.UserCartItems{}).Error; err != nil {
			return err
		}
		// 如果没错，返回nil，事务会提交
		return nil
	})

	if err != nil {
		// 事务失败，已经自动回滚了
		return system.OrderInfo{}, err
	}
	// 事务成功

	//更新并返回orderinfo
	var orderfinal system.Order
	err = global.GVA_DB.Table("orders").
		Where("order_code = ?", ordercode).
		First(&orderfinal).Error
	formattedTotalAmount := humanize.Commaf(orderfinal.TotalAmount) + "円"
	orderat := orderfinal.OrderedAt
	orderedAtFormatted := orderat.Format("2006-01-02 15:04:05")
	orderInfo = system.OrderInfo{
		OrderID:              orderfinal.ID,
		OrderCode:            orderfinal.OrderCode,
		OrderStatus:          orderfinal.OrderStatus,
		TotalAmount:          orderfinal.TotalAmount,
		FormattedTotalAmount: formattedTotalAmount,
		OrderedAtFormatted:   orderedAtFormatted,
	}
	return orderInfo, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看已下单信息
// @return: cartResponse,err
func (productsService *ProductsService) GetOrders(userID uint, page int, limit int, status int) (orderListResponse system.OrderListResponse, err error) {
	var osi []system.OrderSummaryInfo
	var osicount []system.OrderSummaryInfo
	offset := (page - 1) * limit
	var order_status string
	switch status {
	case 1:
		order_status = "pending_payment"
	case 2:
		order_status = "delivered"
	case 3:
		order_status = "shipped"
	case 4:
		order_status = "processing"
	case 5:
		order_status = "cancelled"
	case 6:
		order_status = "all"
	}
	if order_status != "all" {
		err = global.GVA_DB.Table("orders").
			Select(`
	orders.id as order_id,
        orders.order_code,
        orders.ordered_at as ordered_at_formatted,
        orders.total_amount as total_amount_formatted,
        orders.order_status,
        (
            SELECT si.thumbnail_url
            FROM sku_images si
            JOIN order_items oi ON si.sku_id = oi.sku_id
            WHERE oi.order_id = orders.id
            ORDER BY si.sku_id ASC
            LIMIT 1
        ) as representative_image_url`).
			Where("orders.user_id = ? AND orders.order_status = ?", userID, order_status).
			Limit(limit).Offset(offset).
			Find(&osi).Error

		err = global.GVA_DB.Table("orders").
			Select(`
	orders.id as order_id,
        orders.order_code,
        orders.ordered_at as ordered_at_formatted,
        orders.total_amount as total_amount_formatted,
        orders.order_status,
        (
            SELECT si.thumbnail_url
            FROM sku_images si
            JOIN order_items oi ON si.sku_id = oi.sku_id
            WHERE oi.order_id = orders.id
            ORDER BY si.sku_id ASC
            LIMIT 1
        ) as representative_image_url`).
			Where("orders.user_id = ? AND orders.order_status = ?", userID, order_status).
			Find(&osicount).Error
	} else {
		err = global.GVA_DB.Table("orders").
			Select(`
	orders.id as order_id,
        orders.order_code,
        orders.ordered_at as ordered_at_formatted,
        orders.total_amount as total_amount_formatted,
        orders.order_status,
        (
            SELECT si.thumbnail_url
            FROM sku_images si
            JOIN order_items oi ON si.sku_id = oi.sku_id
            WHERE oi.order_id = orders.id
            ORDER BY si.sku_id ASC
            LIMIT 1
        ) as representative_image_url`).
			Where("orders.user_id = ? ", userID).
			Limit(limit).Offset(offset).
			Find(&osi).Error

		err = global.GVA_DB.Table("orders").
			Select(`
	orders.id as order_id,
        orders.order_code,
        orders.ordered_at as ordered_at_formatted,
        orders.total_amount as total_amount_formatted,
        orders.order_status,
        (
            SELECT si.thumbnail_url
            FROM sku_images si
            JOIN order_items oi ON si.sku_id = oi.sku_id
            WHERE oi.order_id = orders.id
            ORDER BY si.sku_id ASC
            LIMIT 1
        ) as representative_image_url`).
			Where("orders.user_id = ? ", userID).
			Find(&osicount).Error
	}
	layout := time.RFC3339
	for i, _ := range osi {
		switch osi[i].OrderStatus {
		case "pending_payment":
			osi[i].OrderStatusDisplay = "待支付"
		case "delivered":
			osi[i].OrderStatusDisplay = "已送达"
		case "shipped":
			osi[i].OrderStatusDisplay = "已出库"
		case "processing":
			osi[i].OrderStatusDisplay = "进行中"
		case "cancelled":
			osi[i].OrderStatusDisplay = "已取消"
		}
		//时间格式化
		t, _ := time.Parse(layout, osi[i].OrderedAtFormatted)
		osi[i].OrderedAtFormatted = t.Format("2006年01月02日  15:04:05")
		//价格格式化
		x := osi[i].TotalAmountFormatted
		y, _ := strconv.ParseFloat(x, 64)
		osi[i].TotalAmountFormatted = humanize.Commaf(y) + "円"
	}

	totalCount := len(osicount)
	totalPages := math.Ceil(float64(totalCount) / float64(limit))
	pagination := system.PaginationInfo{
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  totalCount,
		TotalPages:  int(totalPages),
	}
	orderListResponse = system.OrderListResponse{
		Orders:     osi,
		Pagination: pagination,
	}
	return orderListResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看已下单信息
// @return: cartResponse,err
func (productsService *ProductsService) GetOrderDetail(userID uint, orderId uint64) (orderDetailResponse system.OrderDetailResponse, err error) {
	var oht system.OrderHeaderInfo
	err = global.GVA_DB.Table("orders").
		Select(`orders.id AS order_id,
        orders.order_code,
        orders.ordered_at as ordered_at_formatted,
        orders.order_status,
        orders.subtotal_amount as subtotal_amount_formatted,
        orders.coupon_discount_amount as coupon_discount_amount_formatted,
        orders.points_discount_amount as points_discount_amount_formatted,
        orders.shipping_fee as shipping_fee_formatted,
        orders.total_amount as total_amount_formatted,
        coupons.name as applied_coupon_name,
        orders.points_used,
        orders.payjp_charge_id,
        orders.notes`).
		Joins("join coupons on coupons.id = orders.coupon_id").
		Where("orders.id = ?", orderId).
		Limit(1).
		Order("orders.id ASC").
		Find(&oht).Error
	switch oht.OrderStatus {
	case "pending_payment":
		oht.OrderStatusDisplay = "待支付"
	case "delivered":
		oht.OrderStatusDisplay = "已送达"
	case "shipped":
		oht.OrderStatusDisplay = "已出库"
	case "processing":
		oht.OrderStatusDisplay = "进行中"
	case "cancelled":
		oht.OrderStatusDisplay = "已取消"
	}
	//时间格式化
	layout := time.RFC3339
	t, _ := time.Parse(layout, oht.OrderedAtFormatted)
	oht.OrderedAtFormatted = t.Format("2006年01月02日  15:04:05")
	//价格格式化
	x := oht.TotalAmountFormatted
	y, _ := strconv.ParseFloat(x, 64)
	oht.TotalAmountFormatted = humanize.Commaf(y) + "円"
	x = oht.SubtotalAmountFormatted
	y, _ = strconv.ParseFloat(x, 64)
	oht.SubtotalAmountFormatted = humanize.Commaf(y) + "円"
	x = oht.CouponDiscountAmountFormatted
	y, _ = strconv.ParseFloat(x, 64)
	oht.CouponDiscountAmountFormatted = humanize.Commaf(y) + "円"
	x = oht.PointsDiscountAmountFormatted
	y, _ = strconv.ParseFloat(x, 64)
	oht.PointsDiscountAmountFormatted = humanize.Commaf(y) + "円"
	x = oht.ShippingFeeFormatted
	y, _ = strconv.ParseFloat(x, 64)
	oht.ShippingFeeFormatted = humanize.Commaf(y) + "円"

	oht.PaymentMethodName = "クレジットカード"

	type OrderDetailItemInfo2 struct {
		SkuID              string  `json:"sku_id"`
		ProductID          string  `json:"product_id"`
		ProductName        string  `json:"product_name"` // 省略後
		SkuCode            *string `json:"sku_code,omitempty"`
		Quantity           int     `json:"quantity"`
		UnitPriceFormatted string  `json:"unit_price_formatted"` // 単価 (表示用)
		SubtotalFormatted  string  `json:"subtotal_formatted"`   // 小計 (表示用)
		PrimaryImageURL    *string `json:"primary_image_url,omitempty"`
		Attributes         string  `json:"attributes"`
	}
	var odii2 []OrderDetailItemInfo2
	var odii1 []system.OrderDetailItemInfo
	err = global.GVA_DB.Table("orders").
		Select(`order_items.product_id as product_id,
				order_items.sku_id as sku_id,
				order_items.product_name as product_name,
				order_items.sku_code as sku_code,
				order_items.quantity as quantity,
				order_items.unit_price as unit_price_formatted,
				order_items.subtotal_price as subtotal_formatted,
				(
					SELECT si.thumbnail_url
					FROM sku_images si
					JOIN order_items oi ON si.sku_id = oi.sku_id
					WHERE oi.order_id = orders.id
					ORDER BY si.sku_id ASC
					LIMIT 1
				) as primary_image_url,
				GROUP_CONCAT(DISTINCT CONCAT(attributes.name, "-", 
					COALESCE(attribute_options.value, sku_values.value_string, sku_values.value_number, sku_values.value_boolean))
					ORDER BY attributes.sort_order) AS attributes`).
		Joins("join order_items on order_items.order_id = orders.id").
		Joins("left join sku_values on sku_values.sku_id = order_items.sku_id").
		Joins("left join attributes on sku_values.attribute_id = attributes.id").
		Joins("left join attribute_options on attribute_options.id = sku_values.option_id").
		Where("orders.id = ?", orderId).
		Group("order_items.product_id, order_items.sku_id, order_items.product_name, order_items.sku_code, order_items.quantity, order_items.unit_price, order_items.subtotal_price,orders.id").
		Find(&odii2).Error
	for _, k := range odii2 {
		parts := strings.Split(k.Attributes, ",")
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
		x := k.UnitPriceFormatted
		y, _ := strconv.ParseFloat(x, 64)
		k.UnitPriceFormatted = humanize.Commaf(y) + "円"
		x = k.SubtotalFormatted
		y, _ = strconv.ParseFloat(x, 64)
		k.SubtotalFormatted = humanize.Commaf(y) + "円"
		odii1 = append(odii1, system.OrderDetailItemInfo{
			SkuID:              k.SkuID,
			ProductID:          k.ProductID,
			ProductName:        k.ProductName,
			SkuCode:            k.SkuCode,
			Quantity:           k.Quantity,
			UnitPriceFormatted: k.UnitPriceFormatted,
			SubtotalFormatted:  k.SubtotalFormatted,
			PrimaryImageURL:    k.PrimaryImageURL,
			Attributes:         attributes,
		})

	}

	orderDetailResponse = system.OrderDetailResponse{
		OrderInfo:  &oht,
		OrderItems: odii1,
	}
	return orderDetailResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 搜索商品
// @return: cartResponse,err
func (productsService *ProductsService) Search(keyword string, page int, limit int, sort int) (productSearchResponse system.ProductSearchResponse, err error) {
	// size := limit
	// // 构造搜索请求
	// query := map[string]interface{}{
	// 	"from": (page - 1) * size,
	// 	"size": size,
	// 	"query": map[string]interface{}{
	// 		"match": map[string]interface{}{
	// 			"name": keyword,
	// 		},
	// 	},
	// 	// 可以加入排序等
	// }

	// var buf bytes.Buffer
	// if err = json.NewEncoder(&buf).Encode(query); err != nil {
	// 	return
	// }

	// resp, err := global.GVA_ElasticSearch.Search(
	// 	global.GVA_ElasticSearch.Search.WithContext(context.Background()),
	// 	global.GVA_ElasticSearch.Search.WithIndex("products"),
	// 	global.GVA_ElasticSearch.Search.WithBody(&buf),
	// )
	// if err != nil {
	// 	return
	// }
	// defer resp.Body.Close()

	// var result map[string]interface{}
	// json.NewDecoder(resp.Body).Decode(&result)

	// // 处理返回结果（提取命中内容）
	// hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	// var products []map[string]interface{}
	// for _, hit := range hits {
	// 	source := hit.(map[string]interface{})["_source"]
	// 	products = append(products, source.(map[string]interface{}))
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"total":    result["hits"].(map[string]interface{})["total"],
	// 	"products": products,
	// })

	type Result struct {
		//Products   []SearchedProductInfo `json:"products"`         // 検索結果の商品リスト
		ProductID           string `json:"product_id"`
		ProductCode         string `json:"product_code,omitempty"`
		ProductName         string `json:"product_name"` // 省略後
		PriceRangeFormatted string `json:"price_range_formatted"`
		IsOnSale            bool   `json:"is_on_sale"`
		SalePrice           string `json:"sale_price"`
		//ReviewSummary       *ReviewSummaryInfo `json:"review_summary,omitempty"`
		AverageRating float64 `json:"average_rating"` // 平均評価
		ReviewCount   int     `json:"review_count"`   // レビュー件数

		ThumbnailImageUrl *string `json:"thumbnail_image_url,omitempty"`
		//CategoryName        string             `json:"category_name"` // 商品が属するカテゴリ名

		//Pagination PaginationInfo        `json:"pagination"`       // ページネーション情報
		//Facets     *SearchFacets         `json:"facets,omitempty"` // ファセット情報 (オプション)
		//Categories  []FacetCategoryItem   `json:"categories,omitempty"`
		CategoryID    int    `json:"category_id"`
		CategoryName  string `json:"category_name"`
		CproductCount int    `json:"cproduct_count"` // このカテゴリに該当する商品数
		//Brands      []FacetBrandItem      `json:"brands,omitempty"`       // ブランドファセットが必要な場合
		// BrandID      int    `json:"brand_id"`
		// BrandName    string `json:"brand_name"`
		// ProductCount int    `json:"product_count"`

		//PriceRanges []FacetPriceRangeItem `json:"price_ranges,omitempty"` // 価格帯ファセット
		RangeLabel    string `json:"range_label"` // 例: "1,000円～2,999円"
		MinPrice      int    `json:"min_price"`
		MaxPrice      int    `json:"max_price"`
		RproductCount int    `json:"rproduct_count"`
		//Attributes  []FacetAttributeGroup `json:"attributes,omitempty"`   // 属性ファセット
		Attributestr  string
		AttributeID   int    `json:"attribute_id"`
		AttributeName string `json:"attribute_name"`
		//Options       []FacetOptionItem `json:"options"`
		Optionstr     string
		OptionID      int    `json:"option_id"`
		OptionValue   string `json:"option_value"`
		OproductCount int    `json:"oproduct_count"`
	}
	var result1 []Result
	var result2 []Result
	var result3 []Result

	sqlQuery := `
		select 
products.id as product_id,
products.product_code as 	product_code,
products.name as 		product_name,
case
when min(prices.price)<max(prices.price) then CONCAT(format(min(prices.price),0),'~',format(max(prices.price),0),'円')
when min(prices.price)=max(prices.price) then concat(format(min(prices.price),0),'円')
end as price_range_formatted,
GROUP_CONCAT(DISTINCT sale_price SEPARATOR ',') as sale_price,
review_summaries.average_rating as 	average_rating,
review_summaries.review_count as		review_count,

 (
            SELECT si.thumbnail_url
            FROM sku_images si
            JOIN product_skus ps ON si.sku_id = ps.id
            WHERE ps.product_id=products.id
            ORDER BY si.sku_id ASC
            LIMIT 1
        ) as	thumbnail_image_url,
	
	categories.id as	category_id,
	categories.name as category_name,
		count(distinct p1.id) as  cproduct_count,

		case 
		when min(prices.price)<=39999  then '~39,999円' 
        when min(prices.price)>39999   then '40,000円~'
        end as range_label,
	case 
		when min(prices.price)<=39999 then 0
        when min(prices.price)>39999   then 40000
        end as min_price,
	case 
		when min(prices.price)<=39999 then 39999
        when min(prices.price)>39999   then 0
        end as	max_price,
	case 
		when min(prices.price)<=39999 then (select count(distinct p2.id)from products p2 join product_skus prs1 on prs1.product_id=p2.id
    join prices pr1 on pr1.sku_id=prs1.id where pr1.price<=39999)
        when min(prices.price)>39999   then (select count(distinct p2.id)from products p2 join product_skus prs1 on prs1.product_id=p2.id
    join prices pr1 on pr1.sku_id=prs1.id where pr1.price>39999)
        end as rproduct_count
        

    
    from products
    join review_summaries on review_summaries.product_id=products.id
    join categories on categories.id=products.category_id
    join products p1 on p1.category_id=categories.id
    join product_skus on product_skus.product_id=products.id
    join prices on prices.sku_id=product_skus.id
    JOIN (select distinct  pr1.sku_id,
pr1.price,
case when NOW() BETWEEN pr2.start_date AND pr2.end_date then pr2.price 
else null
end as sale_price
from prices pr1
join prices pr2 on pr1.sku_id=pr2.sku_id
where  pr1.price_type_id=1 
)pr3 on pr3.sku_id=product_skus.id
 
    join sku_values on sku_values.sku_id=product_skus.id
    join attributes on attributes.id = sku_values.attribute_id
    join attribute_options on attribute_options.attribute_id=attributes.id 
    group by
    products.id,
    products.product_code,
    products.name,
    review_summaries.average_rating,
    review_summaries.review_count,
    categories.id,
    categories.name
		LIMIT 1000
	`

	// 使用 Raw 方法执行查询
	err = global.GVA_DB.Raw(sqlQuery).Scan(&result1).Error
	if err != nil {
		fmt.Println("result1 error:", err)
		return
	}
	for i, _ := range result1 {
		if result1[i].SalePrice != "" {
			result1[i].IsOnSale = true
		}
	}
	pinfos := []system.SearchedProductInfo{}
	categories := []system.FacetCategoryItem{}
	priceRanges := []system.FacetPriceRangeItem{}
	categoryIDMap := make(map[int]bool)
	rangeLabelMap := make(map[string]bool)

	for _, k := range result1 {
		reviewSummary := system.ReviewSummaryInfo{
			AverageRating: k.AverageRating,
			ReviewCount:   k.ReviewCount,
		}
		pinfos = append(pinfos, system.SearchedProductInfo{
			ProductID:           k.ProductID,
			ProductCode:         k.ProductCode,
			ProductName:         k.ProductName,
			PriceRangeFormatted: k.PriceRangeFormatted,
			IsOnSale:            k.IsOnSale,
			ReviewSummary:       &reviewSummary,
			ThumbnailImageUrl:   k.ThumbnailImageUrl,
			//  CategoryName
		})
		if !categoryIDMap[k.CategoryID] {
			categories = append(categories, system.FacetCategoryItem{
				CategoryID:   k.CategoryID,
				CategoryName: k.CategoryName,
				ProductCount: k.CproductCount,
			})
			categoryIDMap[k.CategoryID] = true
		}
		if !rangeLabelMap[k.RangeLabel] {
			priceRanges = append(priceRanges, system.FacetPriceRangeItem{
				RangeLabel:   k.RangeLabel,
				MinPrice:     k.MinPrice,
				MaxPrice:     k.MaxPrice,
				ProductCount: k.RproductCount,
			})
			rangeLabelMap[k.RangeLabel] = true
		}
	}

	err = global.GVA_DB.Table("products").
		Select(`attribute_options.id as option_id, COUNT(DISTINCT products.id) as product_count`).
		Joins(`LEFT JOIN product_skus ON products.id = product_skus.product_id`).
		Joins(`LEFT JOIN sku_values ON product_skus.id = sku_values.sku_id`).
		Joins(`LEFT JOIN attribute_options ON sku_values.attribute_id = attribute_options.attribute_id`).
		Group("attribute_options.id").
		Scan(&result3).Error
	if err != nil {
		return
	}
	OptionLinkedCount := make(map[int]int)
	for _, k := range result3 {
		OptionLinkedCount[k.OptionID] = k.OproductCount
	}

	err = global.GVA_DB.Table("products").
		Select(`products.id as product_id,
        CONCAT(attributes.id,'-',attributes.name) as attributestr,
        GROUP_CONCAT(DISTINCT CONCAT(attribute_options.id,'-',attribute_options.value) ORDER BY attributes.sort_order) as optionstr`).
		Joins("join review_summaries on review_summaries.product_id=products.id").
		Joins("join categories on categories.id=products.category_id").
		Joins("join product_skus on product_skus.product_id=products.id").
		Joins("join prices on prices.sku_id=product_skus.id").
		Joins(`join (select distinct pr1.sku_id, pr1.price,
        case when NOW() BETWEEN pr2.start_date AND pr2.end_date then pr2.price else null end as sale_price
        from prices pr1
        join prices pr2 on pr1.sku_id=pr2.sku_id
        where pr1.price_type_id=1) pr3 on pr3.sku_id=product_skus.id`).
		Joins("join sku_values on sku_values.sku_id=product_skus.id").
		Joins("join attributes on attributes.id = sku_values.attribute_id").
		Joins("join attribute_options on attribute_options.attribute_id=attributes.id").
		Group("products.id, attributes.id").
		Limit(1000).
		Find(&result2).Error
	if err != nil {
		fmt.Println("result2 error:", err)
		return
	}
	var attributes []system.FacetAttributeGroup
	for i, _ := range result2 {

		//var aint int
		//var attributefindoption map[int][]system.FacetOptionItem
		//attributefindoption = make(map[int][]system.FacetOptionItem)
		parts := strings.Split(result2[i].Optionstr, ",")
		var options []system.FacetOptionItem
		for _, part := range parts {
			part = strings.TrimSpace(part)
			kv := strings.SplitN(part, "-", 2)
			optionID, _ := strconv.Atoi(strings.TrimSpace(kv[0]))
			if len(kv) == 2 {
				options = append(options, system.FacetOptionItem{
					OptionID:     optionID,
					OptionValue:  strings.TrimSpace(kv[1]),
					ProductCount: OptionLinkedCount[optionID],
				})
				//attributefindoption[aint] =  options
				//还要塞进map里与属性映射
			}
		}

		astr := strings.Split(result2[i].Attributestr, "-")
		aint, _ := strconv.Atoi(strings.TrimSpace(astr[0]))
		attributes = append(attributes, system.FacetAttributeGroup{
			AttributeID:   aint,
			AttributeName: strings.TrimSpace(astr[1]),
			Options:       options,
		})
	}
	facets := system.SearchFacets{
		Categories:  categories,
		PriceRanges: priceRanges,
		Attributes:  attributes,
	}
	productSearchResponse = system.ProductSearchResponse{
		Products: pinfos,
		//Pagination :,
		Facets: &facets,
	}
	return productSearchResponse, err
}
