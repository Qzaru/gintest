package system

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// @author: [granty1](https://github.com/granty1)
// @description: 查看购物车
// @return: cartResponse,err
func (productsService *ProductsService) GetCart(userID uint) (cartResponse system.CartResponse, err error) {
	type ResultsCart struct {
		//Items                []CartItemInfo `json:"items"`                  // カート内商品リスト
		SkuID       string `json:"sku_id"`                 // SKU ID
		ProductID   string `json:"product_id"`             // 商品ID
		ProductName string `json:"product_name"`           // 商品名 (省略後)
		ProductCode string `json:"product_code,omitempty"` // 商品コード
		Quantity    int    `json:"quantity"`               // カート内の数量
		//Price             *PriceInfo      `json:"price"`                   // 現在の単価情報 (Nullable)
		Amount                  float64  `json:"amount"` //这是一个需要判断是否在促销期内的促销价格，否则为初始价格
		FormattedAmount         string   `json:"formatted_amount"`
		Type                    string   `json:"type"`
		TypeName                string   `json:"type_name"`
		OriginalAmount          *float64 `json:"original_amount,omitempty"` //这是一个通常价格
		FormattedOriginalAmount *string  `json:"formatted_original_amount,omitempty"`
		ValidSalePrice          *float64

		SubtotalFormatted string `json:"subtotal_formatted"` // 小計 (表示用文字列 例: "7,980円")
		//PrimaryImage      *ImageInfo      `json:"primary_image,omitempty"` // サムネイル画像推奨 (Nullable)
		ID      int     `gorm:"column:image_id"`
		URL     string  `json:"url"`
		AltText *string `json:"alt_text,omitempty"`

		//Attributes        []AttributeInfo `json:"attributes"`              // 対象SKUの属性リスト
		AttributeName  string `json:"attribute_name"`
		AttributeValue string `json:"attribute_value"`
		NameValue      string
		StockStatus    string `json:"stock_status"` // 在庫状況コード ('available', 'low_stock', 'out_of_stock')

		TotalItemsCount      int     `json:"total_items_count"`      // カート内総商品点数 (数量の合計)
		TotalAmount          float64 `json:"total_amount"`           // 合計金額 (計算用数値)
		TotalAmountFormatted string  `json:"total_amount_formatted"` // 合計金額 (表示用文字列 例: "55,880円")
	}
	var resultsCart []ResultsCart
	err = global.GVA_DB.Table("user_cart_items").
		Select(`DISTINCT
        user_cart_items.sku_id as sku_id,
        products.id as product_id,
        products.name as product_name,
        products.product_code as product_code,
        user_cart_items.quantity as quantity,
        pr3.price as amount,
        price_types.type_code as type,
        price_types.name as type_name,
        GROUP_CONCAT(DISTINCT pr3.sale_price SEPARATOR ',') as valid_sale_price,
        (
            SELECT sku_images.id
            FROM sku_images
            WHERE sku_images.sku_id = user_cart_items.sku_id
            ORDER BY sku_images.id ASC
            LIMIT 1
        ) as image_id,
        (
            SELECT sku_images.thumbnail_url
            FROM sku_images
            WHERE sku_images.sku_id = user_cart_items.sku_id
            ORDER BY sku_images.id ASC
            LIMIT 1
        ) as url,
        (
            SELECT sku_images.alt_text
            FROM sku_images
            WHERE sku_images.sku_id = user_cart_items.sku_id
            ORDER BY sku_images.id ASC
            LIMIT 1
        ) as alt_text,
        GROUP_CONCAT(DISTINCT CONCAT(attributes.name, "-", COALESCE(attribute_options.value, sku_values.value_string, sku_values.value_number, sku_values.value_boolean)) ORDER BY attributes.sort_order) as name_value,
        CASE
            WHEN SUM(inventory.quantity - inventory.reserved_quantity) = 0 THEN 'out_of_stock'
            WHEN SUM(inventory.quantity - inventory.reserved_quantity) > 0 AND SUM(inventory.quantity - inventory.reserved_quantity) < 50 THEN 'low_stock'
            WHEN SUM(inventory.quantity - inventory.reserved_quantity) > 50 THEN 'available'
        END as stock_status
    `).
		Joins("JOIN product_skus ON user_cart_items.sku_id = product_skus.id").
		Joins("JOIN products ON product_skus.product_id = products.id").
		Joins("JOIN prices pr1 ON pr1.sku_id = user_cart_items.sku_id").
		Joins(`JOIN (
        SELECT pr1.sku_id,
               pr1.price,
               pr1.price_type_id,
               CASE WHEN NOW() BETWEEN pr2.start_date AND pr2.end_date THEN pr2.price ELSE NULL END as sale_price
        FROM prices pr1
        JOIN prices pr2 ON pr1.sku_id = pr2.sku_id
        WHERE pr1.price_type_id=1
    ) pr3 ON pr3.sku_id = product_skus.id`).
		Joins("JOIN price_types ON price_types.id = pr3.price_type_id").
		Joins("LEFT JOIN sku_values ON sku_values.sku_id = user_cart_items.sku_id").
		Joins("LEFT JOIN attributes ON attributes.id = sku_values.attribute_id").
		Joins("LEFT JOIN attribute_options ON attribute_options.id = sku_values.option_id").
		Joins("JOIN inventory ON inventory.sku_id = user_cart_items.sku_id").
		Where("user_cart_items.user_id = ?", userID).
		Group("user_cart_items.sku_id, products.id, products.name, products.product_code, user_cart_items.quantity, pr3.price, price_types.type_code, price_types.name").
		// 如果需要限制条数，加入 Limit 和 Offset
		// Limit(10).Offset(0).
		Find(&resultsCart).Error

	var totalItemsCount int
	var totalAmount float64
	var items []system.CartItemInfo

	fmt.Println("查询的结果是：", resultsCart)
	for _, k := range resultsCart {
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
		//有优惠价则调整
		if k.ValidSalePrice != nil {
			k.OriginalAmount = &k.Amount
			k.Amount = *k.ValidSalePrice
			k.Type = "sale"
			k.TypeName = "セール価格"
		}
		//计算小计价格
		x := k.Amount * float64(k.Quantity)
		//格式化价格
		k.SubtotalFormatted = humanize.Commaf(x) + "円"
		k.FormattedAmount = humanize.Commaf(k.Amount) + "円"
		if k.OriginalAmount != nil {
			y := humanize.Commaf(*k.OriginalAmount) + "円"
			k.FormattedOriginalAmount = &y
		}
		price := system.PriceInfo{
			Amount:                  k.Amount,
			FormattedAmount:         k.FormattedAmount,
			Type:                    k.Type,
			TypeName:                k.TypeName,
			OriginalAmount:          k.OriginalAmount,
			FormattedOriginalAmount: k.FormattedOriginalAmount,
		}
		primaryImage := system.ImageInfo{
			ID:      k.ID,
			URL:     k.URL,
			AltText: k.AltText,
		}

		items = append(items, system.CartItemInfo{
			SkuID:             k.SkuID,
			ProductID:         k.ProductID,
			ProductName:       k.ProductName,
			ProductCode:       k.ProductCode,
			Quantity:          k.Quantity,
			Price:             &price,
			SubtotalFormatted: k.SubtotalFormatted,
			PrimaryImage:      &primaryImage,
			Attributes:        attributes,
			StockStatus:       k.StockStatus,
		})

		totalItemsCount += k.Quantity
		totalAmount += k.Amount * float64(k.Quantity)
	}
	totalAmountFormatted := humanize.Commaf(totalAmount) + "円"
	cartResponse = system.CartResponse{
		Items:                items,
		TotalItemsCount:      totalItemsCount,
		TotalAmount:          totalAmount,
		TotalAmountFormatted: totalAmountFormatted,
	}
	return cartResponse, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 添加商品到购物车
// @param sku_id  query  string true   "添加的商品"
// @param quantity  query  int false   "添加数量，默认1"  default(1)
// @return: favoriteSKUListResponse,err
func (productsService *ProductsService) SetCartSku(userID uint, skuid string, quantity int) (err error) {
	var uc system.UserCartItems
	var quantityresult int64
	err = global.GVA_DB.Table("inventory").
		Select("SUM(inventory.quantity - inventory.reserved_quantity)as quantityresult").
		Where("inventory.sku_id = ?", skuid).
		Scan(&quantityresult).Error

	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).First(&uc).Error
	if err == nil {
		if quantityresult < int64((uc.Quantity + quantity)) {
			return errors.New("库存不足")
		} else {
			err = global.GVA_DB.Model(&uc).Where("user_id = ? AND sku_id= ?", userID, skuid).Updates(system.UserCartItems{Quantity: uc.Quantity + quantity, UpdatedAt: time.Now()}).Error
		}
	} else if err == gorm.ErrRecordNotFound {
		err = global.GVA_DB.Create(&system.UserCartItems{
			UserId:    userID,
			SkuId:     skuid,
			Quantity:  quantity,
			AddedAt:   time.Now(),
			UpdatedAt: time.Now(),
		}).Error
	}
	return
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: SetCartSkuQuantity
// @description: 更改购物车商品数量
// @param: auth model.UserCartItems
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) SetCartSkuQuantity(uci system.UserCartItemsQuantityRes, skuid string, userID uint) (userCartItems system.UserCartItems, err error) {
	var oldquantity system.UserCartItems
	//需确认购物车内是否存在该商品，以及库存
	var quantityresult int64
	err = global.GVA_DB.Table("inventory").
		Select("SUM(inventory.quantity - inventory.reserved_quantity)as quantityresult").
		Where("inventory.sku_id = ?", skuid).
		Scan(&quantityresult).Error
	if err != nil || quantityresult < int64(uci.Quantity) {
		return system.UserCartItems{}, errors.New("库存错误")
	}
	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).First(&oldquantity).Error
	if err != nil {
		global.GVA_LOG.Debug(err.Error())
		return system.UserCartItems{}, errors.New("查询购物车数据失败")
	}
	q := uci.Quantity
	newuci := system.UserCartItems{
		Quantity:  q,
		UpdatedAt: time.Now(),
	}
	err = global.GVA_DB.Model(&oldquantity).Where("user_id = ? AND sku_id= ?", userID, skuid).Updates(&newuci).Error
	return newuci, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 删除商品收藏信息
// @param sku_id  query  string true   "删除购物车商品"
// @return: err
func (productsService *ProductsService) DeleteCartItem(userID uint, skuid string) (err error) {
	err = global.GVA_DB.Where("user_id = ? AND sku_id= ?", userID, skuid).Delete(&system.UserCartItems{}).Error
	return err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看配送地址
// @return: cartResponse,err
func (productsService *ProductsService) GetAddress(userID uint) (shippingAddressListResponse system.ShippingAddressListResponse, err error) {
	var resultsaddress []system.ShippingAddressInfo
	err = global.GVA_DB.Table("user_shipping_addresses").
		Select(`user_shipping_addresses.id as address_id,
            user_shipping_addresses.postal_code as postal_code,
            user_shipping_addresses.prefecture as prefecture,
            user_shipping_addresses.city as city,
            user_shipping_addresses.address_line1 as address_line1,
            user_shipping_addresses.address_line2 as address_line2,
            user_shipping_addresses.recipient_name as recipient_name,
            user_shipping_addresses.phone_number as phone_number,
            user_shipping_addresses.is_default as is_default`).
		Where("user_shipping_addresses.user_id = ?", userID).
		Order("is_default DESC, updated_at DESC").
		Find(&resultsaddress).Error
	shippingAddressListResponse = system.ShippingAddressListResponse{
		Addresses: resultsaddress,
	}
	return shippingAddressListResponse, err
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: SetAddress
// @description: 添加配送地址
// @param: sai model.ShippingAddressInput
// @return: shippingAddressInputRes system.ShippingAddressInput, err error
func (productsService *ProductsService) SetAddress(sai system.ShippingAddressInput, userID uint) (shippingAddressInputRes system.ShippingAddressInput, err error) {
	//もし is_default=true であれば、そのユーザーの他の住所の is_default を false に更新する処理（DBアクセス）を実行する
	var oldaddress system.UserShippingAddress
	fmt.Println("sai.IsDefault:", sai.IsDefault)
	if sai.IsDefault {
		x := int64(userID)
		err = global.GVA_DB.Where("user_id = ? AND is_default= ?", x, 1).First(&oldaddress).Error
		err = global.GVA_DB.Model(&oldaddress).Where("user_id = ? AND is_default= ?", x, 1).Update("IsDefault", false).Error
	}
	err = global.GVA_DB.Create(&system.UserShippingAddress{
		UserId:        int64(userID),
		PostalCode:    sai.PostalCode,
		Prefecture:    sai.Prefecture,
		City:          sai.City,
		AddressLine1:  sai.AddressLine1,
		AddressLine2:  sai.AddressLine2,
		RecipientName: sai.RecipientName,
		PhoneNumber:   sai.PhoneNumber,
		IsDefault:     sai.IsDefault,
	}).Error
	return shippingAddressInputRes, err

}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: UpdateAddress
// @description: 更新配送地址
// @param: auth model.UserCartItems
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) UpdateAddress(sai system.ShippingAddressInput, address_id uint64, userID uint) (shippingAddressInfo system.ShippingAddressInfo, err error) {
	//DBアクセス: user_shipping_addresses テーブルを検索し、address_id と user_id で対象住所が存在するか確認。存在しない、またはユーザーが異なる場合はエラー。
	//もし is_default=true であり、かつその住所が元々デフォルトでなかった場合、他の住所の is_default を false に更新。

	var oldaddress1 system.UserShippingAddress
	var oldaddress2 system.UserShippingAddress
	//var olddefaultaddress system.UserShippingAddress

	if sai.IsDefault {
		err = global.GVA_DB.Where("user_id = ? AND is_default= ?", int64(userID), 1).First(&oldaddress1).Error
		err = global.GVA_DB.Model(&oldaddress1).Where("user_id = ? AND is_default= ?", int64(userID), 1).Update("IsDefault", false).Error
	}
	newaddressinfo := system.UserShippingAddress{
		//Id           :,
		UserId:        int64(userID),
		PostalCode:    sai.PostalCode,
		Prefecture:    sai.Prefecture,
		City:          sai.City,
		AddressLine1:  sai.AddressLine1,
		AddressLine2:  sai.AddressLine2,
		RecipientName: sai.RecipientName,
		PhoneNumber:   sai.PhoneNumber,
		IsDefault:     sai.IsDefault,
	}
	shippingAddressInfo = system.ShippingAddressInfo{
		AddressID:     address_id,
		PostalCode:    sai.PostalCode,
		Prefecture:    sai.Prefecture,
		City:          sai.City,
		AddressLine1:  sai.AddressLine1,
		AddressLine2:  sai.AddressLine2,
		RecipientName: sai.RecipientName,
		PhoneNumber:   sai.PhoneNumber,
		IsDefault:     sai.IsDefault,
	}
	err = global.GVA_DB.Model(&oldaddress2).Where("user_id = ? AND id= ?", int64(userID), int64(address_id)).Updates(&newaddressinfo).Error
	return shippingAddressInfo, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 删除配送地址
// @param address_id  path  string true   "删除配送地址"
// @return: err
func (productsService *ProductsService) DeleteAddress(userID uint, address_id uint64) (err error) {
	err = global.GVA_DB.Where("user_id = ? AND id= ?", int64(userID), int64(address_id)).Delete(&system.UserShippingAddress{}).Error
	return err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看支付方式
// @return: paymentMethodInfo,err
func (productsService *ProductsService) GetPaymentMethods() (paymentMethodInfo []system.PaymentMethodInfo, err error) {
	err = global.GVA_DB.Table("payment_methods").Select("id as method_id, method_code, name, description").
		Where("is_active = ?", 1).Order("sort_order").
		Find(&paymentMethodInfo).Error
	return paymentMethodInfo, err
}

// @author: [granty1](https://github.com/granty1)
// @description: 查看配送地址
// @return: cartResponse,err
func (productsService *ProductsService) GetCheckoutInfo(userID uint, onlyget bool) (checkoutInfoResponseV1_1 system.CheckoutInfoResponseV1_1, err error) {
	//购物车中的商品们视为准备购买的商品，小计存入到checkoutsession，没有就创建，有就更新小计
	//当onlyget为false时，跑以下代码，从购物车生成session；true时跳过，直接到后面代码
	if !onlyget {
		cart, err := productsService.GetCart(userID)
		var findsession system.CheckoutSession
		err = global.GVA_DB.Table("checkout_sessions").
			Where("user_id = ?", userID).
			First(&findsession).Error
		if err == gorm.ErrRecordNotFound {
			err = global.GVA_DB.Create(&system.CheckoutSession{
				UserID:       userID,
				CartSubtotal: cart.TotalAmount,
				ShippingFee:  550.00,
				TotalAmount:  cart.TotalAmount,
			}).Error
		} else if err == nil {
			updates := map[string]interface{}{
				"cart_subtotal":          cart.TotalAmount,
				"applied_coupon_id":      nil,
				"coupon_discount_amount": 0,
				"used_points":            0,
				"points_discount_amount": 0,
				"shipping_fee":           550.00,
				"total_amount":           cart.TotalAmount + 550.00,
			}
			err = global.GVA_DB.Model(&findsession).Where("user_id = ?", userID).Updates(updates).Error
		}
	}
	//以下是正常的checkoutsession中已有数据时，对数据进行抓取处理
	type ResultsCheckout struct {
		//AvailableCoupons []AvailableCouponInfo `json:"available_coupons"`        // 利用可能なクーポンリスト
		CouponID      uint64  `json:"coupon_id"`
		CouponCode    string  `json:"coupon_code"`
		Name          string  `json:"name"`
		Description   *string `json:"description,omitempty"`
		DiscountText  string  `json:"discount_text"` // 例: "10% OFF (最大2,000円引)", "500円引き"
		DiscountType  string  `json:"discount_type"`
		DiscountValue string  `json:"discount_value"`
		DiscountMax   string  `json:"discount_max"`

		//UserPoints       *UserPointInfo        `json:"user_points,omitempty"`    // 保有ポイント情報 (Nullable)
		AvailablePoints int `json:"available_points"`

		//CurrentCheckoutState
		CartSubtotalAmountFormatted string `json:"cart_subtotal_formatted"` // カート商品小計 (割引前、表示用)
		//AppliedCouponInfo             *AppliedCouponInfo `json:"applied_coupon_info,omitempty"` // 適用中クーポン情報 (Nullable)
		AcouponID               uint64  `json:"acoupon_id"`
		AcouponCode             string  `json:"acoupon_code"`
		Aname                   string  `json:"aname"`
		AdiscountType           string  `json:"adiscount_type"`
		AdiscountValue          string  `json:"adiscount_value"` //用于匹配类型计算以下实际折扣
		AdiscountMax            string  `json:"adiscount_max"`
		DiscountAmount          float64 `json:"discount_amount"`           // この注文での実際の割引額 (計算用)
		FormattedDiscountAmount string  `json:"formatted_discount_amount"` // 表示用

		CouponDiscountAmountFormatted string `json:"coupon_discount_amount_formatted"` // クーポン割引額 (表示用)
		UsedPoints                    int    `json:"used_points"`                      // 利用ポイント数
		PointsDiscountAmountFormatted string `json:"points_discount_amount_formatted"` // ポイント割引額 (表示用)
		ShippingFeeFormatted          string `json:"shipping_fee_formatted"`           // 送料 (表示用、別途計算の場合あり)
		TotalAmountFormatted          string `json:"total_amount_formatted"`           // ★最終支払総額 (表示用)

	}

	var resultsCheckout1 []ResultsCheckout
	var resultsCheckout2 ResultsCheckout
	err = global.GVA_DB.Table("user_coupons").
		Select(`user_coupons.coupon_id as coupon_id,
            coupons.coupon_code as coupon_code,
            coupons.name as name,
            coupons.description as description,
            coupons.discount_type as discount_type,
            coupons.discount_value as discount_value,
			coupons.max_discount_amount as discount_max,
            user_points.available_points as available_points`).
		Joins(`join coupons on user_coupons.coupon_id = coupons.id`).
		Joins(`join user_points on user_points.user_id = user_coupons.user_id`).
		Joins(`join checkout_sessions on user_coupons.user_id=checkout_sessions.user_id`).
		Where(`user_coupons.user_id = ?`, userID).
		Where(`user_coupons.is_used = 0`).
		Where(`now() between coupons.start_date and coupons.end_date`).
		Where(`coupons.min_purchase_amount IS NULL or coupons.min_purchase_amount <= checkout_sessions.cart_subtotal`).
		Find(&resultsCheckout1).Error
	availableCouponInfo := []system.AvailableCouponInfo{}
	if len(resultsCheckout1) > 0 {
		for _, k := range resultsCheckout1 {
			switch {
			case k.DiscountType == "percentage" && k.DiscountMax != "":
				str1 := strings.TrimSuffix(k.DiscountValue, ".00")
				str2 := strings.TrimSuffix(k.DiscountMax, ".00")
				k.DiscountText = str1 + "% OFF (最大" + str2 + "円引)"
			case k.DiscountType == "percentage" && k.DiscountMax == "":
				str1 := strings.TrimSuffix(k.DiscountValue, ".00")
				k.DiscountText = str1 + "% OFF"
			case k.DiscountType == "fixed":
				str1 := strings.TrimSuffix(k.DiscountValue, ".00")
				k.DiscountText = str1 + "円引き"
			}
			availableCouponInfo = append(availableCouponInfo, system.AvailableCouponInfo{
				CouponID:     k.CouponID,
				CouponCode:   k.CouponCode,
				Name:         k.Name,
				Description:  k.Description,
				DiscountText: k.DiscountText,
			})
		}
	}

	// userPointInfo := system.UserPointInfo{
	// 	AvailablePoints: resultsCheckout1[0].AvailablePoints,
	// }
	err = global.GVA_DB.Table("checkout_sessions").
		Select(`checkout_sessions.cart_subtotal as cart_subtotal_amount_formatted,
				coupons.id as acoupon_id,
				coupons.coupon_code as acoupon_code,
				coupons.name as aname,
				coupons.discount_type as adiscount_type,
				coupons.discount_value as adiscount_value,
				coupons.max_discount_amount  as adiscount_max,
				checkout_sessions.coupon_discount_amount as coupon_discount_amount_formatted,
				checkout_sessions.used_points as used_points,
				checkout_sessions.points_discount_amount as points_discount_amount_formatted,
				checkout_sessions.shipping_fee as shipping_fee_formatted,
				checkout_sessions.total_amount as total_amount_formatted,
				user_points.available_points as available_points`).
		Joins("left join coupons on coupons.id=checkout_sessions.applied_coupon_id").
		Joins("join user_points on user_points.user_id = checkout_sessions.user_id").
		Where("checkout_sessions.user_id = ?", userID).
		Order("checkout_sessions.user_id ASC").
		Limit(1).
		Find(&resultsCheckout2).Error

	userPointInfo := system.UserPointInfo{
		AvailablePoints: resultsCheckout2.AvailablePoints,
	}
	x, _ := strconv.ParseFloat(resultsCheckout2.CartSubtotalAmountFormatted, 64) //小计的数字类型
	// y, _ := strconv.ParseFloat(resultsCheckout2.ADiscountValue, 64)              //折扣值的数字类型
	// z, _ := strconv.ParseFloat(resultsCheckout2.ADiscountMax, 64)                //折扣上限的数字类型
	// //计算折扣
	// switch {
	// case resultsCheckout2.ADiscountType == "percentage" && resultsCheckout2.ADiscountMax != "":
	// 	resultsCheckout2.DiscountAmount = x * y / 100
	// 	if resultsCheckout2.DiscountAmount > z {
	// 		resultsCheckout2.DiscountAmount = z
	// 	}
	// case resultsCheckout2.ADiscountType == "percentage" && resultsCheckout2.ADiscountMax == "":
	// 	resultsCheckout2.DiscountAmount = x * y / 100
	// case resultsCheckout2.DiscountType == "fixed":
	// 	resultsCheckout2.DiscountAmount = x - y
	// }
	resultsCheckout2.DiscountAmount, _ = strconv.ParseFloat(resultsCheckout2.CouponDiscountAmountFormatted, 64)
	//格式化
	a, _ := strconv.ParseFloat(resultsCheckout2.CouponDiscountAmountFormatted, 64)
	b, _ := strconv.ParseFloat(resultsCheckout2.PointsDiscountAmountFormatted, 64)
	c, _ := strconv.ParseFloat(resultsCheckout2.ShippingFeeFormatted, 64)
	d, _ := strconv.ParseFloat(resultsCheckout2.TotalAmountFormatted, 64)
	resultsCheckout2.CartSubtotalAmountFormatted = humanize.Commaf(x) + "円"
	resultsCheckout2.FormattedDiscountAmount = humanize.Commaf(resultsCheckout2.DiscountAmount) + "円"
	resultsCheckout2.CouponDiscountAmountFormatted = humanize.Commaf(a) + "円"
	resultsCheckout2.PointsDiscountAmountFormatted = humanize.Commaf(b) + "円"
	resultsCheckout2.ShippingFeeFormatted = humanize.Commaf(c) + "円"
	resultsCheckout2.TotalAmountFormatted = humanize.Commaf(d) + "円"

	//这里似乎不需要手动计算，只需要查询checkoutsession的数据即可
	appliedCouponInfo := system.AppliedCouponInfo{
		CouponID:                resultsCheckout2.AcouponID,
		CouponCode:              resultsCheckout2.AcouponCode,
		Name:                    resultsCheckout2.Aname,
		DiscountAmount:          resultsCheckout2.DiscountAmount,
		FormattedDiscountAmount: resultsCheckout2.FormattedDiscountAmount,
	}
	currentCheckoutState := system.CurrentCheckoutState{
		CartSubtotalAmountFormatted:   resultsCheckout2.CartSubtotalAmountFormatted,
		AppliedCouponInfo:             &appliedCouponInfo,
		CouponDiscountAmountFormatted: resultsCheckout2.CouponDiscountAmountFormatted,
		UsedPoints:                    resultsCheckout2.UsedPoints,
		PointsDiscountAmountFormatted: resultsCheckout2.PointsDiscountAmountFormatted,
		ShippingFeeFormatted:          resultsCheckout2.ShippingFeeFormatted,
		TotalAmountFormatted:          resultsCheckout2.TotalAmountFormatted,
	}
	checkoutInfoResponseV1_1 = system.CheckoutInfoResponseV1_1{
		AvailableCoupons:     availableCouponInfo,
		UserPoints:           &userPointInfo,
		CurrentCheckoutState: &currentCheckoutState,
	}
	return checkoutInfoResponseV1_1, err
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: SetCoupon
// @description: 在订单中使用优惠券
// @param: auth model.ApplyCouponRequest
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) SetCoupon(couponCode string, userID uint) (checkoutInfoResponse system.CheckoutInfoResponseV1_1, err error) {
	//需确认用户是否有该优惠券，该优惠券是否满足使用条件
	type AppliedCouponResult struct {
		CouponID                uint64  `json:"coupon_id"`
		CouponCode              string  `json:"coupon_code"`
		Name                    string  `json:"name"`
		DiscountAmount          float64 `json:"discount_amount"`           // この注文での実際の割引額 (計算用)
		FormattedDiscountAmount string  `json:"formatted_discount_amount"` // 表示用
		DiscountType            string  `json:"discount_type"`
		DiscountValue           string  `json:"discount_value"`
		MaxDiscountAmount       string  `json:"max_discount_amount"`
		CartSubtotal            string  `json:"cart_subtotal"`
		TotalAmount             string  `json:"total_amount"`
	}
	var appliedCouponResult []AppliedCouponResult
	err = global.GVA_DB.Table("user_coupons").
		Select(`coupons.id as coupon_id,
            checkout_sessions.cart_subtotal,
            checkout_sessions.total_amount,
            coupons.discount_type,
            coupons.discount_value,
            coupons.max_discount_amount`).
		Joins("join coupons on user_coupons.coupon_id = coupons.id").
		Joins("join checkout_sessions on checkout_sessions.user_id = user_coupons.user_id").
		Where("user_coupons.user_id = ?", userID).
		Where("coupons.coupon_code = ?", couponCode).
		Where("user_coupons.is_used = ?", 0).
		Where("coupons.is_active = ?", 1).
		Where("now() BETWEEN coupons.start_date AND coupons.end_date").
		Where("(coupons.min_purchase_amount IS NULL OR coupons.min_purchase_amount <= checkout_sessions.cart_subtotal)"). // 条件
		Find(&appliedCouponResult).Error
	a := appliedCouponResult[0]
	if err == nil && len(appliedCouponResult) != 0 {
		//使用时要准确计算出折扣值（不可超过自身限制、小计上限等）

		x, _ := strconv.ParseFloat(a.TotalAmount, 64)       //原本结算金额的数字类型
		y, _ := strconv.ParseFloat(a.DiscountValue, 64)     //折扣值的数字类型
		z, _ := strconv.ParseFloat(a.MaxDiscountAmount, 64) //折扣上限的数字类型
		//计算折扣和需要被更新的最终价格
		var total_amount float64
		switch {
		case a.DiscountType == "percentage" && a.MaxDiscountAmount != "":
			a.DiscountAmount = x * y / 100
			if a.DiscountAmount > z {
				a.DiscountAmount = z
			}
			total_amount = x - a.DiscountAmount
		case a.DiscountType == "percentage" && a.MaxDiscountAmount == "":
			a.DiscountAmount = x * y / 100
			total_amount = x - a.DiscountAmount
		case a.DiscountType == "fixed":
			a.DiscountAmount = y
			total_amount = x - a.DiscountAmount
		}
		//update checkoutsession的字段
		var oldsession system.CheckoutSession
		// var userCoupon system.UserCoupon
		//更新优惠券id 折扣额 和最终价格
		err = global.GVA_DB.Model(&oldsession).Where("user_id = ?", userID).Updates(system.CheckoutSession{AppliedCouponID: &a.CouponID, CouponDiscountAmount: a.DiscountAmount, TotalAmount: total_amount}).Error
		//还需要在usercoupon里更新优惠券已使用
		// t := time.Now()
		// err = global.GVA_DB.Model(&userCoupon).Where("user_id = ? AND coupon_id= ?", userID, a.CouponID).Updates(system.UserCoupon{IsUsed: true, UsedAt: &t}).Error
		//用另一个api抓取购物车信息然后返回
		getCheckoutInfo, err := productsService.GetCheckoutInfo(userID, true)
		return getCheckoutInfo, err
	} else {
		return system.CheckoutInfoResponseV1_1{}, err
	}
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: DeleteCoupon
// @description: 在订单中取消使用优惠券
// @param: auth model.ApplyCouponRequest
// @return: authority system.CheckoutInfoResponseV1_1, err error
func (productsService *ProductsService) DeleteCoupon(couponCode string, userID uint) (checkoutInfoResponse system.CheckoutInfoResponseV1_1, err error) {
	//在checkoutsession里找到对应订单，更新为不使用coupon状态（id 折扣归零，总价复原）
	//在usercoupon里找到对应coupon，更新为未使用
	var checkoutSession system.CheckoutSession
	err = global.GVA_DB.Table("checkout_sessions").
		Select(`applied_coupon_id,
            coupon_discount_amount,
			total_amount`).
		Where("user_id = ?", userID).
		First(&checkoutSession).Error

	total_amount := checkoutSession.TotalAmount + checkoutSession.CouponDiscountAmount
	var oldsession system.CheckoutSession
	//更新优惠券id 折扣额 和最终价格
	updates := map[string]interface{}{
		"applied_coupon_id":      nil,
		"coupon_discount_amount": 0,
		"total_amount":           total_amount,
	}
	err = global.GVA_DB.Model(&oldsession).Where("user_id = ?", userID).Updates(updates).Error
	//用另一个api抓取购物车信息然后返回
	getCheckoutInfo, err := productsService.GetCheckoutInfo(userID, true)
	//将用户的优惠券恢复为未使用
	// var userCoupon system.UserCoupon
	// updates2 := map[string]interface{}{
	// 	"is_used": 0,
	// 	"used_at": nil,
	// }
	// err = global.GVA_DB.Model(&userCoupon).Where("user_id = ? AND coupon_id = ?", userID, checkoutSession.AppliedCouponID).Updates(updates2).Error
	return getCheckoutInfo, err
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: UsePoints
// @description: 在订单中使用积分
// @param: auth model.ApplyCouponRequest
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) UsePoints(pointsToUse int, userID uint) (checkoutInfoResponse system.CheckoutInfoResponseV1_1, err error) {
	//需确认用户是否有该优惠券，该优惠券是否满足使用条件
	//var checkoutSession system.CheckoutSession
	type CheckoutSession2 struct {
		UserID               uint    `gorm:"column:user_id;primaryKey;not null" json:"user_id"`                                                 // 用户ID（主键）
		CartSubtotal         float64 `gorm:"column:cart_subtotal;type:decimal(12,2);not null;default:0" json:"cart_subtotal"`                   // 购物车小计（未折扣）
		UsedPoints           int     `gorm:"column:used_points;type:uint;not null;default:0" json:"used_points"`                                // 使用的积分数
		CouponDiscountAmount float64 `gorm:"column:coupon_discount_amount;type:decimal(10,2);not null;default:0" json:"coupon_discount_amount"` // 优惠券折扣额
		PointsDiscountAmount float64 `gorm:"column:points_discount_amount;type:decimal(10,2);not null;default:0" json:"points_discount_amount"` // 积分折扣额
		TotalAmount          float64 `gorm:"column:total_amount;type:decimal(12,2);not null;default:0" json:"total_amount"`                     // 最终支付总额
		AvailablePoints      int     `json:"available_points"`
	}
	var checkoutSession CheckoutSession2
	err = global.GVA_DB.Table("checkout_sessions").
		Select(`checkout_sessions.cart_subtotal,
            checkout_sessions.coupon_discount_amount,
			checkout_sessions.used_points,
			checkout_sessions.total_amount,
			user_points.available_points`).
		Joins("JOIN user_points ON user_points.user_id=checkout_sessions.user_id").
		Where("checkout_sessions.user_id = ?", userID).
		Where("user_points.available_points > ?", pointsToUse).
		First(&checkoutSession).Error
	//如果没找到数据，可认为是积分不够
	if err == gorm.ErrRecordNotFound {
		return checkoutInfoResponse, err
	}
	//求一个积分使用上限
	maxpointtouse := checkoutSession.CartSubtotal - checkoutSession.CouponDiscountAmount - float64(checkoutSession.UsedPoints)
	//此处判断是否超过上限，这里超过直接报错返回得了
	if pointsToUse > int(maxpointtouse) {
		err = errors.New("超过积分使用上限")
		return checkoutInfoResponse, err
	}
	//没有问题则进入使用积分阶段，usedpoint增加，total_amount减少
	//update checkoutsession的字段
	var oldsession system.CheckoutSession
	var userPoints system.UserPoints
	used_points := checkoutSession.UsedPoints + pointsToUse
	total_amount := checkoutSession.TotalAmount - float64(pointsToUse)
	updates := map[string]interface{}{
		"used_points":            used_points,
		"points_discount_amount": float64(used_points),
		"total_amount":           total_amount,
	}
	//更新优惠券id 折扣额 和最终价格
	err = global.GVA_DB.Model(&oldsession).Where("user_id = ?", userID).Updates(updates).Error
	//还需要在userpoints里更新积分已使用
	err = global.GVA_DB.Model(&userPoints).Where("user_id = ?", userID).Update("AvailablePoints", checkoutSession.AvailablePoints-pointsToUse).Error
	//用另一个api抓取购物车信息然后返回
	getCheckoutInfo, err := productsService.GetCheckoutInfo(userID, true)
	return getCheckoutInfo, err
}

// @author: [piexlmax](https://github.com/piexlmax)
// @function: UsePoints
// @description: 在订单中使用积分
// @param: auth model.ApplyCouponRequest
// @return: authority system.SysAuthority, err error
func (productsService *ProductsService) UnUsePoints(pointsToUnUse int, userID uint) (checkoutInfoResponse system.CheckoutInfoResponseV1_1, err error) {
	//需确认用户是否有该优惠券，该优惠券是否满足使用条件
	//var checkoutSession system.CheckoutSession
	type CheckoutSession2 struct {
		UserID               uint    `gorm:"column:user_id;primaryKey;not null" json:"user_id"`                                                 // 用户ID（主键）
		CartSubtotal         float64 `gorm:"column:cart_subtotal;type:decimal(12,2);not null;default:0" json:"cart_subtotal"`                   // 购物车小计（未折扣）
		UsedPoints           int     `gorm:"column:used_points;type:uint;not null;default:0" json:"used_points"`                                // 使用的积分数
		CouponDiscountAmount float64 `gorm:"column:coupon_discount_amount;type:decimal(10,2);not null;default:0" json:"coupon_discount_amount"` // 优惠券折扣额
		PointsDiscountAmount float64 `gorm:"column:points_discount_amount;type:decimal(10,2);not null;default:0" json:"points_discount_amount"` // 积分折扣额
		TotalAmount          float64 `gorm:"column:total_amount;type:decimal(12,2);not null;default:0" json:"total_amount"`                     // 最终支付总额
		AvailablePoints      int     `json:"available_points"`
	}
	var checkoutSession CheckoutSession2
	err = global.GVA_DB.Table("checkout_sessions").
		Select(`checkout_sessions.cart_subtotal,
            checkout_sessions.coupon_discount_amount,
			checkout_sessions.used_points,
			checkout_sessions.total_amount,
			user_points.available_points`).
		Joins("JOIN user_points ON user_points.user_id=checkout_sessions.user_id").
		Where("checkout_sessions.user_id = ?", userID).
		Where("checkout_sessions.used_points > ?", pointsToUnUse).
		First(&checkoutSession).Error
	//如果没找到数据，可认为是超过了已使用的积分,直接报错返回
	if err == gorm.ErrRecordNotFound {
		return checkoutInfoResponse, err
	}

	//没有问题则进入取消使用积分阶段，usedpoint减少，total_amount增加
	//update checkoutsession的字段
	var oldsession system.CheckoutSession
	var userPoints system.UserPoints
	used_points := checkoutSession.UsedPoints - pointsToUnUse
	total_amount := checkoutSession.TotalAmount + float64(pointsToUnUse)
	updates := map[string]interface{}{
		"used_points":            used_points,
		"points_discount_amount": float64(used_points),
		"total_amount":           total_amount,
	}
	//更新优惠券id 折扣额 和最终价格
	err = global.GVA_DB.Model(&oldsession).Where("user_id = ?", userID).Updates(updates).Error
	//还需要在userpoints里更新积分已使用
	err = global.GVA_DB.Model(&userPoints).Where("user_id = ?", userID).Update("AvailablePoints", checkoutSession.AvailablePoints+pointsToUnUse).Error
	//用另一个api抓取购物车信息然后返回
	getCheckoutInfo, err := productsService.GetCheckoutInfo(userID, true)
	return getCheckoutInfo, err
}
