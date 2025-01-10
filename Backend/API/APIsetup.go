package api

import (
	"fmt"
	requestandreceivedatafrommongodb "financialscraper/RequestAndReceiveDataFromMongoDB"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes configures all the routes for the API
func SetupRoutes(r *gin.Engine, mongoClient *mongo.Client) {
	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api")
	{
		// Financial statements endpoint
		fmt.Println("Registering route: GET /api/:cik")
		api.GET("/:cik", requestandreceivedatafrommongodb.HandleGetFinancialStatements(mongoClient))
	}
}
