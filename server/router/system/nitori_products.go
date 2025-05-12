package system

import (
	"github.com/gin-gonic/gin"
)

type ProductsRouter struct{}

func (s *ProductsRouter) InitProductsRouterPublic(Router *gin.RouterGroup) {
	productsRouterPublic := Router.Group("/api/v1")
	{
		productsRouterPublic.GET("/products/:product_code", productsApi.GetProductInfo)
		productsRouterPublic.GET("/products/:product_code/reviews", productsApi.GetProductReviews)
		productsRouterPublic.GET("/products/:product_code/qa", productsApi.GetProductQA)
		productsRouterPublic.GET("/products/images/:sku_id", productsApi.GetProductSkuImages)
		productsRouterPublic.GET("/products/:product_code/related", productsApi.GetRelateProducts)
		productsRouterPublic.GET("/products/:product_code/coordinates", productsApi.GetCoordinateSet)
		productsRouterPublic.GET("/payments/methods", productsApi.GetPaymentMethods)

	}

}
func (s *ProductsRouter) InitProductsRouterPrivate(Router *gin.RouterGroup) {
	productsRouterPrivate := Router.Group("/api/v2")
	{
		productsRouterPrivate.GET("/favorites/skus", productsApi.GetFavorites)
		productsRouterPrivate.POST("/favorites/skus/:sku_id ", productsApi.SetFavorites)
		productsRouterPrivate.DELETE("/favorites/skus/:sku_id", productsApi.DeleteFavorites)
		productsRouterPrivate.GET("/history/viewed-skus", productsApi.GetViewHistory)
		productsRouterPrivate.POST("/history/viewed-skus", productsApi.SetViewHistory)
		productsRouterPrivate.GET("/cart", productsApi.GetCart)
		productsRouterPrivate.POST("/cart/items", productsApi.SetCartSku)
		productsRouterPrivate.PUT("/cart/items/:sku_id", productsApi.SetCartSkuQuantity)
		productsRouterPrivate.DELETE("/cart/items/:sku_id", productsApi.DeleteCartItem)
		productsRouterPrivate.GET("/shipping-addresses", productsApi.GetAddress)
		productsRouterPrivate.POST("/shipping-addresses", productsApi.SetAddress)
		productsRouterPrivate.PUT("/shipping-addresses/:address_id", productsApi.UpdateAddress)
		productsRouterPrivate.DELETE("/shipping-addresses/:address_id", productsApi.DeleteAddress)
		productsRouterPrivate.GET("/orders/checkout/info", productsApi.GetCheckoutInfo)
		productsRouterPrivate.POST("/orders/checkout/apply-coupon", productsApi.SetCoupon)
		productsRouterPrivate.DELETE("/orders/checkout/apply-coupon", productsApi.DeleteCoupon)
		productsRouterPrivate.POST("/orders/checkout/use-points", productsApi.UsePoints)
		productsRouterPrivate.DELETE("/orders/checkout/use-points", productsApi.UnUsePoints)
	}
}
