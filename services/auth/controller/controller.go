package controller

import (
	model "auth-service/models/user"
	"auth-service/service"
	"auth-service/utils"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authService service.IAuthService
}

type IAuthController interface {
	Register(c *fiber.Ctx) error
	RegisterVerify(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	ForgotPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
}

func NewAuthController(authService service.IAuthService) IAuthController {
	return &AuthController{
		authService: authService,
	}
}

func (authController *AuthController) ResetPassword(c *fiber.Ctx) error {

	switch c.Method() {
	
	case fiber.MethodGet:
		token := c.Query("token")
		if err := authController.authService.ResetPasswordGet(token); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}

		temp := utils.ResetPasswordTemplate(token)

		return c.Type("html").SendString(temp)

	case fiber.MethodPost:
		token := c.Query("token")
		var reqbody model.ResetPasswordRequest

		validate := validator.New()

		if err := c.BodyParser(&reqbody); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}

		if err := validate.Struct(reqbody); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}

		if err := authController.authService.ResetPasswordPost(token, reqbody.Password); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		return c.Status(200).JSON(fiber.Map{
			"message": "reset password successful",
		})

	default:
		return c.SendStatus(fiber.StatusMethodNotAllowed)
	}

}

func (authController *AuthController) ForgotPassword(c *fiber.Ctx) error {

	var reqbody model.ForgotPasswordRequest

	validate := validator.New()

	if err := c.BodyParser(&reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := authController.authService.ForgotPassword(reqbody.Email); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "mail sent to confirm account",
	})

}

func (authController *AuthController) Login(c *fiber.Ctx) error {

	var reqbody model.LoginRequest

	validate := validator.New()

	if err := c.BodyParser(&reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	user, err := authController.authService.GetUserBySingleField("email", reqbody.Email)

	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	if !user.IsVerified {
		return fiber.NewError(http.StatusBadRequest, "user not verified")
	}

	if err := utils.CheckPassword(reqbody.Password, user.Password); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	token, err := utils.GenerateJwtToken(user.ID, user.Role, user.Email)

	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.Status(200).JSON(fiber.Map{
		"message":      "Login successful",
		"access_token": token,
	})

}

func (authController *AuthController) RegisterVerify(c *fiber.Ctx) error {
	token := c.Query("token")

	if err := authController.authService.RegisterVerify(token); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "verify succesful",
	})
}

func (authController *AuthController) Register(c *fiber.Ctx) error {

	var reqbody model.RegisterRequest

	validate := validator.New()

	if err := c.BodyParser(&reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(reqbody); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	hashedPassword, err := utils.HashPassword(reqbody.Password)

	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	reqbody.Password = hashedPassword

	if err := authController.authService.Register(reqbody); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "mail sent to confirm account",
	})
}
