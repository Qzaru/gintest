package system

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductsApi struct{}

// GetProductInfo
// @Tags      Products
// @Summary   查看商品基础信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetProductInfo
// @description: 根据id获取商品信息
// @param product_code path  string true "商品code"
// @param skuid  query  string false   "商品skuid"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看商品基础信息"
// @Router    /api/v1/products/{product_code} [get]
func (s *ProductsApi) GetProductInfo(c *gin.Context) {
	//var products system.Products
	//fmt.Println("Request path:", c.Request.URL.Path)
	var productSkus system.ProductSku
	productCode := c.Param("product_code")
	for _, char := range productCode {
		if !unicode.IsDigit(char) {
			response.FailWithDetailed2(response.ERRORBADREQUEST, "不正な商品識別子です。", "INVALID_PARAMETER", c)
			return
		}
	}

	// 检查提取到的参数是否有效
	if productCode == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithMessage("商品代码不能为空", c)
		return
	}

	err := c.ShouldBindQuery(&productSkus)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// err = utils.Verify(productSkus, utils.IdVerify)
	// if err != nil {
	// 	response.FailWithMessage(err.Error(), c)
	// 	return
	// }
	//fmt.Println(productCode, productSkus.ID)
	reGetProductInfo, err := productsService.GetProductInfo(productCode, productSkus.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORBADREQUEST, "不正なSKU ID形式です。", "INVALID_PARAMETER", c)
		return
	}
	response.OkWithDetailed(gin.H{"reGetProductInfo": reGetProductInfo}, "查询成功", c)
}

// GetProductReviews
// @Tags      Products
// @Summary   查看商品评论信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetProductReviews
// @description: 根据product_code获取商品评论信息
// @param product_code path  string true "商品code"
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @param sort  query  int false   "排序1.最新开始 2.最早开始 3.评价从高到低 4.评价从低到高 5.点赞数从高到低，默认1" default(1)
// @param rating  query  int false   "只看几星评价"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看商品评论信息"
// @Router    /api/v1/products/{product_code}/reviews [get]
func (s *ProductsApi) GetProductReviews(c *gin.Context) {
	var reviewsRequests system.ReviewsRequests
	productCode := c.Param("product_code")
	for _, char := range productCode {
		if !unicode.IsDigit(char) {
			response.FailWithDetailed2(response.ERRORBADREQUEST, "不正な商品識別子です。", "INVALID_PARAMETER", c)
			return
		}
	}
	if productCode == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithDetailed2(response.ERRORNOTFOUND, "商品が見つかりません。", "NOT_FOUND", c)
		return
	}
	err := c.ShouldBindQuery(&reviewsRequests)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	reviewListResponse, err := productsService.GetProductReviews(productCode, reviewsRequests.Page, reviewsRequests.Limit, reviewsRequests.Sort, reviewsRequests.Rating)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, "商品が見つかりません。", "NOT_FOUND", c)
		return
	}
	response.OkWithDetailed(gin.H{"reviewListResponse": reviewListResponse}, "", c)
}

// GetProductQA
// @Tags      Products
// @Summary   查看商品Q&A信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetProductQA
// @description: 根据product_code获取商品Q&A信息
// @param product_code path  string true "商品code"
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @param sort  query  int false   "排序1.最新开始 2.最早开始 3.点赞数从高到低，默认1" default(1)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看商品QA信息"
// @Router    /api/v1/products/{product_code}/qa [get]
func (s *ProductsApi) GetProductQA(c *gin.Context) {
	var qaRequests system.QARequests
	productCode := c.Param("product_code")
	for _, char := range productCode {
		if !unicode.IsDigit(char) {
			response.FailWithDetailed2(response.ERRORBADREQUEST, response.ERRORMESSAGE_WRONGPRODUCTCODE, response.ERRORCODE_INVALID_PARAMETER, c)
			return
		}
	}
	if productCode == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	err := c.ShouldBindQuery(&qaRequests)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	qaListResponse, err := productsService.GetProductQA(productCode, qaRequests.Page, qaRequests.Limit, qaRequests.Sort)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"qaListResponse": qaListResponse}, "OK", c)
}

// GetProductSkuImages
// @Tags      Products
// @Summary   查看商品图片信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetProductSkuImages
// @description: 根据skuid获取商品图片信息
// @param sku_id path  string true "商品skuid"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看商品图片信息"
// @Router    /api/v1/products/images/{sku_id} [get]
func (s *ProductsApi) GetProductSkuImages(c *gin.Context) {
	skuid := c.Param("sku_id")
	if skuid == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	skuImageInfo, err := productsService.GetProductSkuImages(skuid)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"skuImageInfo": skuImageInfo}, "OK", c)
}

// GetFavorites
// @Tags      Products
// @Summary   查看商品收藏信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Description: 根据用户id获取商品收藏信息
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @param sort  query  int false   "排序1.最新开始 2.最早开始 默认1" default(1)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看商品收藏信息"
// @Router    /api/v2/favorites/skus [get]
func (s *ProductsApi) GetFavorites(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	getFavorites, err := productsService.GetFavorites(userID, uf.Page, uf.Limit, uf.Sort)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getFavorites": getFavorites}, "OK", c)
}

// SetFavorites
// @Tags      Products
// @Summary   设置商品收藏信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: SetFavorites
// @description: 在用户名下设置商品收藏信息
// @param sku_id  query  string true   "准备收藏的商品sku"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "进行商品收藏"
// @Router    /api/v2/favorites/skus/{sku_id} [post]
func (s *ProductsApi) SetFavorites(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	err = productsService.SetFavorites(userID, uf.SkuId)
	if err != nil {
		if err.Error() == "已收藏" {
			global.GVA_LOG.Error("已收藏!", zap.Error(err))
			response.OkWithMessage(response.MESSAGE_FAVORITED, c)
			return
		} else {
			global.GVA_LOG.Error("收藏失败!", zap.Error(err))
			response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
			return
		}
	}
	response.OkWithMessage("收藏成功", c)
}

// DeleteFavorites
// @Tags      Products
// @Summary   取消商品收藏
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: DeleteFavorites
// @param sku_id  query  string true   "被取消收藏的商品sku"
// @description: 删除商品收藏信息
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "删除商品收藏信息"
// @Router    /api/v2/favorites/skus/{sku_id} [delete]
func (s *ProductsApi) DeleteFavorites(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	err = productsService.DeleteFavorites(userID, uf.SkuId)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_FAVORITENOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithMessage("删除成功", c)

}

// GetRelateProducts
// @Tags      Products
// @Summary   获取商品关联商品
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetRelateProducts
// @description: 根据product_code获取关联商品
// @param product_code path  string true "商品code"
// @param limit  query  int false   "每页显示几条，默认6" default(6)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取关联商品"
// @Router    /api/v1/products/{product_code}/related [get]
func (s *ProductsApi) GetRelateProducts(c *gin.Context) {
	var qa system.QARequests
	productCode := c.Param("product_code")
	for _, char := range productCode {
		if !unicode.IsDigit(char) {
			response.FailWithDetailed2(response.ERRORBADREQUEST, response.ERRORMESSAGE_WRONGPRODUCTCODE, response.ERRORCODE_INVALID_PARAMETER, c)
			return
		}
	}
	if productCode == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	err := c.ShouldBindQuery(&qa)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	relatedProductListResponse, err := productsService.GetRelateProducts(productCode, qa.Limit)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"relatedProductListResponse": relatedProductListResponse}, "OK", c)
}

// GetCoordinateSet
// @Tags      Products
// @Summary   获取商品关联搭配
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetCoordinateSet
// @description: 根据product_code获取关联商品搭配
// @param product_code path  string true "商品code"
// @param limit  query  int false   "每页显示几条，默认6" default(6)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取关联商品搭配"
// @Router    /api/v1/products/{product_code}/coordinates [get]
func (s *ProductsApi) GetCoordinateSet(c *gin.Context) {
	var qa system.QARequests
	productCode := c.Param("product_code")
	for _, char := range productCode {
		if !unicode.IsDigit(char) {
			response.FailWithDetailed2(response.ERRORBADREQUEST, response.ERRORMESSAGE_WRONGPRODUCTCODE, response.ERRORCODE_INVALID_PARAMETER, c)
			return
		}
	}
	if productCode == "" {
		// 返回错误响应（商品代码不能为空）
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	err := c.ShouldBindQuery(&qa)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	coordinateSetTeaserListResponse, err := productsService.GetCoordinateSet(productCode, qa.Limit)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"coordinateSetTeaserListResponse": coordinateSetTeaserListResponse}, "OK", c)
}

// GetViewHistory
// @Tags      Products
// @Summary   获取商品查看记录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: GetViewHistory
// @description: 根据userid获取商品查看记录
// @param page  query  int false   "查看第几页,默认1" default(1)
// @param limit  query  int false   "每页显示几条，默认10" default(10)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取商品查看记录"
// @Router    /api/v2/history/viewed-skus [get]
func (s *ProductsApi) GetViewHistory(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	getViewHistory, err := productsService.GetViewHistory(userID, uf.Page, uf.Limit)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getViewHistory": getViewHistory}, "OK", c)
}

// SetViewHistory
// @Tags      Products
// @Summary   添加商品浏览记录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: SetFavorites
// @description: 在用户名下添加商品浏览记录
// @param sku_id  query  string true   "浏览的商品"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "进行商品收藏"
// @Router    /api/v2/history/viewed-skus [post]
func (s *ProductsApi) SetViewHistory(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	err = productsService.SetViewHistory(userID, uf.SkuId)
	if err != nil {
		global.GVA_LOG.Error("添加失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithMessage("添加浏览记录成功", c)
}

// GetCart
// @Tags      Products
// @Summary   查看购物车
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Description: 根据用户id查看购物车
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看购物车"
// @Router    /api/v2/cart [get]
func (s *ProductsApi) GetCart(c *gin.Context) {
	// var uf systemReq.UserFavorites
	// err := c.ShouldBindQuery(&uf)
	// if err != nil {
	// 	response.FailWithMessage(err.Error(), c)
	// 	return
	// }
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	getCart, err := productsService.GetCart(userID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getCart": getCart}, "OK", c)
}

// SetCartSku
// @Tags      Products
// @Summary   在购物车添加商品
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: SetFavorites
// @description: 在购物车添加商品（可指定数量）
// @param sku_id  query  string true   "添加的商品"
// @param quantity  query  int true   "添加数量，默认且至少为1"  default(1)
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "在购物车添加商品"
// @Router    /api/v2/cart/items [post]
func (s *ProductsApi) SetCartSku(c *gin.Context) {
	var uf systemReq.UserFavorites
	err := c.ShouldBindQuery(&uf)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)

	err = productsService.SetCartSku(userID, uf.SkuId, uf.Quantity)
	if err != nil {
		global.GVA_LOG.Error("添加失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithMessage("添加至购物车成功", c)
}

// SetCartSkuQuantity
// @Tags      Products
// @Summary   更新购物车商品数量
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @param sku_id  path  string true   "被修改数量的商品"
// @Param     data  body      system.UserCartItemsQuantityRes         true  "填数量即可"
// @Success   200   {object}  response.Response{data=system.UserCartItemsQuantityRes,msg=string}  "更新购物车商品数量"
// @Router    /api/v2/cart/items/{sku_id} [put]
func (s *ProductsApi) SetCartSkuQuantity(c *gin.Context) {
	var uci system.UserCartItemsQuantityRes
	err := c.ShouldBindJSON(&uci)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sku_id := c.Param("sku_id")
	userID := utils.GetUserID(c)
	setCartSkuQuantity, err := productsService.SetCartSkuQuantity(uci, sku_id, userID)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(setCartSkuQuantity, "更新成功", c)
}

// DeleteCartItem
// @Tags      Products
// @Summary   删除购物车商品
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: DeleteCartItem
// @param sku_id  path  string true   "删除购物车商品"
// @description: 删除购物车商品
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "删除购物车商品"
// @Router    /api/v2/cart/items/{sku_id} [delete]
func (s *ProductsApi) DeleteCartItem(c *gin.Context) {
	sku_id := c.Param("sku_id")
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数

	err := productsService.DeleteCartItem(userID, sku_id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_FAVORITENOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithMessage("删除成功", c)

}

// GetAddress
// @Tags      Products
// @Summary   查看配送地址
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Description: 根据用户id查看配送地址
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看配送地址"
// @Router    /api/v2/shipping-addresses [get]
func (s *ProductsApi) GetAddress(c *gin.Context) {
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)

	getAddress, err := productsService.GetAddress(userID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getAddress": getAddress}, "OK", c)
}

// @Tags      Products
// @Summary   添加配送地址
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body  system.ShippingAddressInput  true  "添加新的配送地址"
// @Success   200   {object} response.Response{data=system.ShippingAddressInput,msg=string} "添加新的配送地址"
// @Router    /api/v2/shipping-addresses [post]
func (s *ProductsApi) SetAddress(c *gin.Context) {
	var sai system.ShippingAddressInput
	err := c.ShouldBindJSON(&sai)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	setAddress, err := productsService.SetAddress(sai, userID)
	if err != nil {
		global.GVA_LOG.Error("添加配送地址失败!", zap.Error(err))
		response.FailWithMessage("添加配送地址失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(setAddress, "添加配送地址成功", c)
}

// UpdateAddress
// @Tags      Products
// @Summary   更新配送地址
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @param address_id  path  string true   "被更新的配送地址id"
// @Param     data  body      system.ShippingAddressInput         true "更新配送地址"
// @Success   200   {object}  response.Response{data=system.ShippingAddressInput,msg=string}  "更新配送地址"
// @Router    /api/v2/shipping-addresses/{address_id} [put]
func (s *ProductsApi) UpdateAddress(c *gin.Context) {
	var sai system.ShippingAddressInput
	err := c.ShouldBindJSON(&sai)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	x := c.Param("address_id")
	address_id, _ := strconv.ParseUint(x, 10, 64)
	userID := utils.GetUserID(c)
	updateAddress, err := productsService.UpdateAddress(sai, address_id, userID)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(updateAddress, "更新成功", c)
}

// DeleteAddress
// @Tags      Products
// @Summary   删除配送地址
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @function: DeleteAddress
// @param address_id  path  string true   "删除配送地址"
// @description: 删除配送地址
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "删除配送地址"
// @Router    /api/v2/shipping-addresses/{address_id} [delete]
func (s *ProductsApi) DeleteAddress(c *gin.Context) {
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//使用GetFavorites函数
	x := c.Param("address_id")
	address_id, _ := strconv.ParseUint(x, 10, 64)
	err := productsService.DeleteAddress(userID, address_id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_FAVORITENOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithMessage("删除成功", c)

}

// GetPaymentMethods
// @Tags      Products
// @Summary   查看支付方式
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Description: 查看支付方式
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "查看支付方式"
// @Router    /api/v1/payments/methods [get]
func (s *ProductsApi) GetPaymentMethods(c *gin.Context) {
	getPaymentMethods, err := productsService.GetPaymentMethods()
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getPaymentMethods": getPaymentMethods}, "OK", c)
}

// GetCheckoutInfo
// @Tags      Products
// @Summary   获取订单信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Description: 获取订单信息
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "获取订单信息"
// @Router    /api/v2/orders/checkout/info [get]
func (s *ProductsApi) GetCheckoutInfo(c *gin.Context) {
	//var cr system.CartResponse
	// err := c.ShouldBindJSON(&cr)
	// if err != nil {
	// 	response.FailWithMessage(err.Error(), c)
	// 	return
	// }
	userID := utils.GetUserID(c)
	//测试用
	fmt.Println("提取到的用户id是", userID)
	//totalAmount:=cr.TotalAmount
	getCheckoutInfo, err := productsService.GetCheckoutInfo(userID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithDetailed2(response.ERRORNOTFOUND, response.ERRORMESSAGE_NOTFOUND, response.ERRORCODE_NOT_FOUND, c)
		return
	}
	response.OkWithDetailed2(gin.H{"getCheckoutInfo": getCheckoutInfo}, "OK", c)
}

// @Tags      Products
// @Summary   在订单中使用优惠券
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body  system.ApplyCouponRequest  true  "在订单中使用优惠券"
// @Success   200   {object} response.Response{data=system.ShippingAddressInput,msg=string} "在订单中使用优惠券"
// @Router    /api/v2/orders/checkout/apply-coupon [post]
func (s *ProductsApi) SetCoupon(c *gin.Context) {
	var acr system.ApplyCouponRequest
	err := c.ShouldBindJSON(&acr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	setCoupon, err := productsService.SetCoupon(acr.CouponCode, userID)
	if err != nil {
		global.GVA_LOG.Error("使用优惠券失败!", zap.Error(err))
		response.FailWithMessage("使用优惠券失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(setCoupon, "使用优惠券成功", c)
}

// @Tags      Products
// @Summary   在订单中取消使用优惠券
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body  system.ApplyCouponRequest  true  "在订单中取消使用优惠券"
// @Success   200   {object} response.Response{data=system.ShippingAddressInput,msg=string} "在订单中取消使用优惠券"
// @Router    /api/v2/orders/checkout/apply-coupon [delete]
func (s *ProductsApi) DeleteCoupon(c *gin.Context) {
	var acr system.ApplyCouponRequest
	err := c.ShouldBindJSON(&acr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	deleteCoupon, err := productsService.DeleteCoupon(acr.CouponCode, userID)
	if err != nil {
		global.GVA_LOG.Error("撤销优惠券失败!", zap.Error(err))
		response.FailWithMessage("撤销优惠券失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(deleteCoupon, "撤销优惠券成功", c)
}

// @Tags      Products
// @Summary   在订单中使用积分
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body  system.UsePointsRequest  true  "在订单中使用积分"
// @Success   200   {object} response.Response{data=system.ShippingAddressInput,msg=string} "在订单中使用积分"
// @Router    /api/v2/orders/checkout/use-points [post]
func (s *ProductsApi) UsePoints(c *gin.Context) {
	var upr system.UsePointsRequest
	err := c.ShouldBindJSON(&upr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	usePoints, err := productsService.UsePoints(upr.PointsToUse, userID)
	if err != nil {
		global.GVA_LOG.Error("使用积分失败!", zap.Error(err))
		response.FailWithMessage("使用积分失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(usePoints, "使用积分成功", c)
}

// @Tags      Products
// @Summary   在订单中取消使用积分
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body  system.ApplyCouponRequest  true  "在订单中取消使用积分"
// @Success   200   {object} response.Response{data=system.ShippingAddressInput,msg=string} "在订单中取消使用积分"
// @Router    /api/v2/orders/checkout/use-points [delete]
func (s *ProductsApi) UnUsePoints(c *gin.Context) {
	var upr system.UsePointsRequest
	err := c.ShouldBindJSON(&upr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	unUsePoints, err := productsService.UnUsePoints(upr.PointsToUse, userID)
	if err != nil {
		global.GVA_LOG.Error("撤销积分失败!", zap.Error(err))
		response.FailWithMessage("撤销积分失败"+err.Error(), c)
		return
	}
	response.OkWithDetailed(unUsePoints, "撤销积分成功", c)
}
