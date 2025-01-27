package routes

import (
	"review_product_tokopedia_be/controller"
	"review_product_tokopedia_be/middleware"
	"review_product_tokopedia_be/service"

	"github.com/gin-gonic/gin"
)

func ML(route *gin.RouterGroup, mlController controller.MLController, jwtService service.JWTService) {
	routes := route.Group("/ml")
	{
		routes.GET("/guest/analysis", mlController.GetSentimentAnalysisAndSummarizationAsGuest)
		routes.GET("/analysis", middleware.Authenticate(jwtService), mlController.GetSentimentAnalysisAndSummarization)
	}
}
