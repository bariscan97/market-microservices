package controller

import (
	
	model "inventory-service/models/product"
	"inventory-service/service"
	"inventory-service/utils"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("isvalid", model.ValidName)
}

type ProductController struct {
	productService service.IProductService
}

type IProductController interface {
	CreateProduct(c *gin.Context)
	GetAllProducts(c *gin.Context)
	UpdateProductById(c *gin.Context)
	GetCategories(c *gin.Context)
	GetProductsByCategory(c *gin.Context)
	DeleteProductById(c *gin.Context)
	GetProductById(c *gin.Context)
}

func NewProductController(productService service.IProductService) IProductController {
	return &ProductController{
		productService : productService,
	}
}

func (productController *ProductController) GetProductById(c *gin.Context) {

	id := c.Param("id")
	
	parsedID, err := uuid.Parse(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return 
	}

	product, err := productController.productService.GetProductById(parsedID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"data" : product,
	})
}

func (productController *ProductController) DeleteProductById(c *gin.Context) {
	
	id := c.Param("id")
	
	parsedID, err := uuid.Parse(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return 
	}

	if err := productController.productService.DeleteProductById(parsedID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return 
	}

	c.JSON(200, gin.H{
		"message" : "successful",
	})
}

func (productController *ProductController) GetProductsByCategory(c *gin.Context) {
	
	page := c.Query("page")

	if page == "" {
		page = "0"
	}

	_, err := strconv.Atoi(page)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	cat := c.Param("category")


	products, err := productController.productService.GetProductsByCategory(cat, page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"data" : products,
	})
	
}

func (productController *ProductController) GetCategories(c *gin.Context) {
	
	page := c.Query("page")

	if page == "" {
		page = "0"
	}

	_, err := strconv.Atoi(page)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	categories, err := productController.productService.GetCategories(page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" :err.Error(),
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : categories,
	})
}

func (productController *ProductController) UpdateProductById(c *gin.Context) {
	
	var	fields model.UpdateProduct

	id := c.Param("id")
	
	parsedID, err := uuid.Parse(id)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return 
	}

	file, err := c.FormFile("file")

	if err == nil {
		
		c.SaveUploadedFile(file, "uploads/"+file.Filename)
		
		img_url, err := utils.UploadToCloudinary(file)
		
		if err != nil {
			c.JSON(400, gin.H{
				"error" : err.Error(),
			})
		}
		
		fields.Image_url = &img_url
	}

	if err := c.ShouldBind(&fields); err != nil {
		c.JSON(400, err.Error())
		return
	}

	if fields.Name != "" {
	
		slug := utils.Slugify(fields.Name)
		
		fields.Slug = &slug
	}
	
	if err := validate.Struct(fields); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	

	if err := productController.productService.UpdateProductById(
		parsedID, 
		utils.StructToMap(fields),
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error" : err.Error(),
			})
			return 
	}

	c.JSON(200, gin.H{
		"message" :"successful",
	})
}

func (productController *ProductController) CreateProduct(c *gin.Context) {

	var product model.CreateProdcut

	file, err := c.FormFile("file")
	
	if err == nil {
		
		c.SaveUploadedFile(file, "uploads/"+file.Filename)

		img_url, err := utils.UploadToCloudinary(file)

		if err != nil {
			c.JSON(400, gin.H{
				"error" : err.Error(),
			})
		}

		product.Image_url = img_url
		
	}

	if err := c.ShouldBind(&product); err != nil {
		c.JSON(400, err.Error())
		return
	}
	
	if err := validate.Struct(product); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	slug := utils.Slugify(product.Name)
		
	product.Slug = slug
	
	ID, err := productController.productService.CreateProduct(&product)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return 
	}

	c.JSON(http.StatusCreated, gin.H{
		"id" :ID,
	})
}

func (productController *ProductController) GetAllProducts(c *gin.Context) {

	page := c.Query("page")

	if page == "" {
		page = "0"
	}

	_, err := strconv.Atoi(page)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	products, err := productController.productService.GetAllProducts(nil,page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"data" : products,
	})
}