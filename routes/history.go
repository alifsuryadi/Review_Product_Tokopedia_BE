package routes

import (
	"review_product_tokopedia_be/controller"
	"review_product_tokopedia_be/middleware"
	"review_product_tokopedia_be/service"

	"github.com/gin-gonic/gin"
)

func History(route *gin.RouterGroup, historyController controller.HistoryController, jwtService service.JWTService) {
	routes := route.Group("/history")
	{
		routes.GET("", middleware.Authenticate(jwtService), historyController.GetHistories)
		routes.GET("/:id", middleware.Authenticate(jwtService), historyController.GetHistory)
	}
}
