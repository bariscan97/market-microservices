package main

import (
	grpc_server "inventory-service/grpc/grpc-handler"
	"inventory-service/config"
	"inventory-service/controller"
	"inventory-service/repository"
	"inventory-service/service"
	"log"
	"os"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	
	err := godotenv.Load("../.env")
	
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configurationManager  := config.NewConfigurationManager()

	pool := config.GetConnectionPool(configurationManager.PostgreSqlConfig)

	productRepository := repository.NewProductRepo(pool)
	productService := service.NewProductService(productRepository)
	productController := controller.NewProductController(productService)

	grpcServer := grpc_server.NewGrpcServer(productRepository)

	
	app := gin.Default()
	
	app.MaxMultipartMemory = 8 << 20
	
	{
		app.POST("/", productController.CreateProduct)
		app.GET("/",productController.GetAllProducts)
		app.GET("/categories", productController.GetCategories)
		app.GET("/categories/:category", productController.GetProductsByCategory)
		app.GET("/:id", productController.GetProductById)
		app.PUT("/:id", productController.UpdateProductById)
		app.DELETE("/:id", productController.DeleteProductById)
	}

	go grpcServer.Run(":55002")
	go productRepository.Listen(context.Background())

	log.Fatal(app.Run(":" + os.Getenv("PORT")))
}