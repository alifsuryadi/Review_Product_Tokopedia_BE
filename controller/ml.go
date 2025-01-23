package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"ulascan-be/dto"
	"ulascan-be/service"
	"ulascan-be/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	MLController interface {
		GetSentimentAnalysisAndSummarization(ctx *gin.Context)
		GetSentimentAnalysisAndSummarizationAsGuest(ctx *gin.Context)
	}

	mlController struct {
		tokopediaService service.TokopediaService
		modelService     service.ModelService
		geminiService    service.GeminiService
		historyService   service.HistoryService
	}
)

func NewMLController(
	ts service.TokopediaService,
	ms service.ModelService,
	gs service.GeminiService,
	hs service.HistoryService,
) MLController {
	return &mlController{
		tokopediaService: ts,
		modelService:     ms,
		geminiService:    gs,
		historyService:   hs,
	}
}

// GetSentimentAnalysisAndSummarizationAsGuest godoc
// @Summary Get product analysis as guest
// @Description Get product analysis form url link as guest.
// @Tags Analysis
// @Accept json
// @Produce json
// @Param product_url query string true "Tokopedia Product Link"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/ml/guest/analysis [get]
func (c *mlController) GetSentimentAnalysisAndSummarizationAsGuest(ctx *gin.Context) {
	productUrl := ctx.Query("product_url")
	if productUrl == "" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrProductUrlMissing.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	parsedUrl, err := url.Parse(productUrl)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if parsedUrl.Host == "tokopedia.link" {
		expandedUrl, err := expandUrl(productUrl)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
		productUrl = expandedUrl

		parsedUrl, err = url.Parse(productUrl)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
	}

	// Validate that the URL is from tokopedia.com
	if parsedUrl.Host != "www.tokopedia.com" && parsedUrl.Host != "tokopedia.com" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrNotTokopediaUrls.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	pathParts := strings.Split(parsedUrl.Path, "/")
	if len(pathParts) < 3 {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrProductUrlWrongFormat.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	productReq := dto.GetProductRequest{
		ShopDomain: pathParts[1],
		ProductKey: pathParts[2],
		ProductUrl: "https://www.tokopedia.com/" + pathParts[1] + "/" + pathParts[2],
	}

	product, err := c.tokopediaService.GetProduct(ctx, productReq)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PRODUCT_ID, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// fmt.Println("=== PRODUCT ID ===")
	// fmt.Println(product)

	reviewsReq := dto.GetReviewsRequest{
		ProductUrl: productReq.ProductUrl,
		ProductId:  product.ProductId,
	}

	reviews, err := c.tokopediaService.GetReviews(ctx, reviewsReq)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// fmt.Println("=== REVIEWS ===")
	// fmt.Println(reviews)

	// Assuming `reviews` is a slice of a struct with `Message` and `Rating` fields
	statements := make([]string, len(reviews))
	ratingSum := 0.0

	for i, review := range reviews {
		statements[i] = review.Message
		ratingSum += float64(review.Rating)
	}

	var ratingAvg float64
	if len(reviews) > 0 {
		ratingAvg = ratingSum / float64(len(reviews))
	} else {
		ratingAvg = 0.0 // or handle the case where there are no reviews
	}

	predictReq := dto.PredictRequest{
		Statements: statements,
	}

	// fmt.Println("=== PREDICT REQ ===")
	// fmt.Println(predictReq)

	var builder strings.Builder
	for _, review := range reviews {
		builder.WriteString(review.Message)
		builder.WriteString("\n")
	}
	concatenatedMessage := builder.String()

	var shopAvatar string
	var predictResult dto.PredictResponse
	var analyzeResult dto.AnalyzeResponse
	var summarizeResult string

	var wg sync.WaitGroup
	// var shopAvatarErr, predictErr error
	var shopAvatarErr, predictErr, analyzeErr, summarizeErr error

	wg.Add(4)

	go func() {
		defer wg.Done()
		shopAvatar, shopAvatarErr = c.tokopediaService.GetShopAvatar(ctx, productReq.ShopDomain)
	}()

	go func() {
		defer wg.Done()
		predictResult, predictErr = c.modelService.Predict(ctx, predictReq)
	}()

	go func() {
		defer wg.Done()
		analyzeResult, analyzeErr = c.geminiService.Analyze(ctx, concatenatedMessage)
	}()

	go func() {
		defer wg.Done()
		summarizeResult, err = c.geminiService.Summarize(ctx, concatenatedMessage)
	}()

	wg.Wait()

	if shopAvatarErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_SHOP_AVATAR, shopAvatarErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if predictErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PREDICT, predictErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if summarizeErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ANALYZE, summarizeErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	if analyzeErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ANALYZE, analyzeErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_REVIEWS, dto.MLResult{
		ProductName:        product.ProductName,
		ProductDescription: product.ProductDescription,
		Rating:             len(reviews),
		Ulasan:             predictResult.CountNegative + predictResult.CountPositive,
		Bintang:            ratingAvg,
		ImageUrls:          product.ImageUrls,
		ShopName:           product.ShopName,
		ShopAvatar:         shopAvatar,
		CountNegative:      predictResult.CountNegative,
		CountPositive:      predictResult.CountPositive,
		Packaging:          analyzeResult.Packaging,
		Delivery:           analyzeResult.Delivery,
		AdminResponse:      analyzeResult.AdminResponse,
		ProductCondition:   analyzeResult.ProductCondition,
		Summary:            summarizeResult,
	})
	ctx.JSON(http.StatusOK, res)
}

// GetSentimentAnalysisAndSummarization godoc
// @Summary Get product analysis
// @Description Get product analysis form url link.
// @Tags Analysis
// @Accept json
// @Produce json
// @Param product_url query string true "Tokopedia Product Link"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Security BearerAuth
// @Router /api/ml/analysis [get]
func (c *mlController) GetSentimentAnalysisAndSummarization(ctx *gin.Context) {
	productUrl := ctx.Query("product_url")
	if productUrl == "" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrProductUrlMissing.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	parsedUrl, err := url.Parse(productUrl)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if parsedUrl.Host == "tokopedia.link" {
		expandedUrl, err := expandUrl(productUrl)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
		productUrl = expandedUrl

		parsedUrl, err = url.Parse(productUrl)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
	}

	// Validate that the URL is from tokopedia.com
	if parsedUrl.Host != "www.tokopedia.com" && parsedUrl.Host != "tokopedia.com" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrNotTokopediaUrls.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	pathParts := strings.Split(parsedUrl.Path, "/")
	if len(pathParts) < 3 {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, dto.ErrProductUrlWrongFormat.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	productReq := dto.GetProductRequest{
		ShopDomain: pathParts[1],
		ProductKey: pathParts[2],
		ProductUrl: "https://www.tokopedia.com/" + pathParts[1] + "/" + pathParts[2],
	}

	product, err := c.tokopediaService.GetProduct(ctx, productReq)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PRODUCT_ID, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// fmt.Println("=== PRODUCT ID ===")
	// fmt.Println(product)

	reviewsReq := dto.GetReviewsRequest{
		ProductUrl: productReq.ProductUrl,
		ProductId:  product.ProductId,
	}

	reviews, err := c.tokopediaService.GetReviews(ctx, reviewsReq)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_REVIEWS, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// fmt.Println("=== REVIEWS ===")
	// fmt.Println(reviews)

	// Assuming `reviews` is a slice of a struct with `Message` and `Rating` fields
	statements := make([]string, len(reviews))
	ratingSum := 0.0

	for i, review := range reviews {
		statements[i] = review.Message
		ratingSum += float64(review.Rating)
	}

	var ratingAvg float64
	if len(reviews) > 0 {
		ratingAvg = ratingSum / float64(len(reviews))
	} else {
		ratingAvg = 0.0 // or handle the case where there are no reviews
	}

	predictReq := dto.PredictRequest{
		Statements: statements,
	}

	// fmt.Println("=== PREDICT REQ ===")
	// fmt.Println(predictReq)

	var builder strings.Builder
	for _, review := range reviews {
		builder.WriteString(review.Message)
		builder.WriteString("\n")
	}
	concatenatedMessage := builder.String()

	var shopAvatar string
	var predictResult dto.PredictResponse
	var analyzeResult dto.AnalyzeResponse
	var summarizeResult string

	var wg sync.WaitGroup
	// var shopAvatarErr, predictErr error
	var shopAvatarErr, predictErr, analyzeErr, summarizeErr error

	wg.Add(4)

	go func() {
		defer wg.Done()
		shopAvatar, shopAvatarErr = c.tokopediaService.GetShopAvatar(ctx, productReq.ShopDomain)
	}()

	go func() {
		defer wg.Done()
		predictResult, predictErr = c.modelService.Predict(ctx, predictReq)
	}()

	go func() {
		defer wg.Done()
		analyzeResult, analyzeErr = c.geminiService.Analyze(ctx, concatenatedMessage)
	}()

	go func() {
		defer wg.Done()
		summarizeResult, err = c.geminiService.Summarize(ctx, concatenatedMessage)
	}()

	wg.Wait()

	if shopAvatarErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_SHOP_AVATAR, shopAvatarErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if predictErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PREDICT, predictErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if summarizeErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ANALYZE, summarizeErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	if analyzeErr != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ANALYZE, analyzeErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userID, exists := ctx.Get("user_id")

	if !exists {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, analyzeErr.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_HISTORY, "Invalid user ID type", nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_HISTORY, "Invalid user ID format", nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	history := dto.HistoryCreateRequest{
		UserID:           userUUID,
		ProductID:        product.ProductId,
		Rating:           len(reviews),
		Ulasan:           predictResult.CountNegative + predictResult.CountPositive,
		Bintang:          ratingAvg,
		URL:              productReq.ProductUrl,
		ProductName:      product.ProductName,
		CountPositive:    predictResult.CountPositive,
		CountNegative:    predictResult.CountNegative,
		Packaging:        analyzeResult.Packaging,
		Delivery:         analyzeResult.Delivery,
		AdminResponse:    analyzeResult.AdminResponse,
		ProductCondition: analyzeResult.ProductCondition,
		Summary:          summarizeResult,
	}
	_, err = c.historyService.CreateHistory(ctx, history)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_HISTORY, err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_REVIEWS, dto.MLResult{
		ProductName:        product.ProductName,
		ProductDescription: product.ProductDescription,
		Rating:             len(reviews),
		Ulasan:             predictResult.CountNegative + predictResult.CountPositive,
		Bintang:            ratingAvg,
		ImageUrls:          product.ImageUrls,
		ShopName:           product.ShopName,
		ShopAvatar:         shopAvatar,
		CountNegative:      predictResult.CountNegative,
		CountPositive:      predictResult.CountPositive,
		Packaging:          analyzeResult.Packaging,
		Delivery:           analyzeResult.Delivery,
		AdminResponse:      analyzeResult.AdminResponse,
		ProductCondition:   analyzeResult.ProductCondition,
		Summary:            summarizeResult,
	})
	ctx.JSON(http.StatusOK, res)
}

func expandUrl(shortUrl string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Prevent the default redirect behavior to handle it manually
			return http.ErrUseLastResponse
		},
	}

	// First request
	req1, err := http.NewRequest("GET", shortUrl, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	// Set headers for the first request
	req1.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1")

	// Make the first request
	resp1, err := client.Do(req1)
	if err != nil {
		return "", fmt.Errorf("error making first request: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status code from first request: %d", resp1.StatusCode)
	}

	newUrl1 := resp1.Header.Get("Location")
	if newUrl1 == "" {
		return "", errors.New("empty Location header in first response")
	}

	// Second request
	req2, err := http.NewRequest("GET", newUrl1, nil)
	if err != nil {
		return "", fmt.Errorf("error creating second request: %v", err)
	}
	// Set headers for the second request
	req2.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1")

	// Make the second request
	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("error making second request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status code from second request: %d", resp2.StatusCode)
	}

	finalUrl := resp2.Header.Get("Location")
	if finalUrl == "" {
		return "", errors.New("empty Location header in second response")
	}

	return finalUrl, nil
}
