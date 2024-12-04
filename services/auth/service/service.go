package service

import (
	model "auth-service/models/user"
	"auth-service/utils"
	"context"
	"fmt"
	"time"
	"auth-service/events/publisher"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
	publish *publisher.Producer
}

type IAuthService interface {
	Register(data model.RegisterRequest) error
	RegisterVerify(token string) error
	GetUserBySingleField(field string, value string) (*model.User, error)
	ForgotPassword(email string) error
	ResetPasswordGet(token string) error
	ResetPasswordPost(token string, password string) error
	DeleteUserById(ctx context.Context, id string) error
}

func NewAuthService(db *gorm.DB, publish *publisher.Producer) IAuthService {
	return &AuthService{
		db: db,
		publish: publish,
	}
}

func (authService *AuthService) DeleteUserById(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	if err := authService.db.Delete(model.User{ID : parsedID}).Error; err != nil {
		return err
	}
	return nil
}

func (authService *AuthService) ResetPasswordPost(token string, password string) error {
	
	var user model.User
	
	if err := authService.db.Where("reset_password_token = ?", token).First(&user).Error; err != nil {
		return err
	}

	if user.ResetPasswordExpire.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		return err
	}
	
	user.ResetPasswordExpire = time.Now().Add(-time.Hour)
	user.ResetPasswordToken = ""
	user.Password = hashedPassword

	if err := authService.db.Save(&user).Error; err != nil {
		return err
	} 

	return nil
}

func (authService *AuthService) ResetPasswordGet(token string) error {

	user, err := authService.GetUserBySingleField("reset_password_token", token)

	if err != nil {
		return err
	}

	if user.ResetPasswordExpire.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	return nil
}

func (authService *AuthService) ForgotPassword(email string) error {

	user ,err := authService.GetUserByEmail(email)

	if err != nil {
		return err
	}

	if !user.IsVerified {
		return fmt.Errorf("user not verified")
	}

	user.ResetPasswordToken = utils.GenerateSecureToken()

	user.ResetPasswordExpire = time.Now().Add(72 * time.Hour)
	
	if err := authService.db.Save(&user).Error; err != nil {
		return err
	}

	htmlTemp := utils.ForgotPasswordTempalte(user.ResetPasswordToken)

	if err := authService.publish.Publisher(htmlTemp, email); err != nil {
		return err
	}

	return nil
}

func (authService *AuthService) GetUserBySingleField(field string, value string) (*model.User, error) {

	var user model.User

	if err := authService.db.Where(fmt.Sprintf("%s = ?", field), value).First(&user).Error; err != nil {
		return  &model.User{}, err
	}

	return &user, nil
}

func (authService *AuthService) GetUserByEmail(email string) (*model.User, error) {

	return authService.GetUserBySingleField("email" ,email)
	
}

func (authService *AuthService) Register(data model.RegisterRequest) error {
	
	token := utils.GenerateSecureToken()

	user := model.User{
		ID: uuid.New(),
		Username: data.Username,
		Email: data.Email,
		Password: data.Password,
		VerificationToken: token,
	}
	
	if err := authService.db.Create(&user) ; err.Error != nil {
		return err.Error
	}
	
	htmlTemp := utils.RegisterVerifyTempalte(token)

	if err := authService.publish.Publisher(htmlTemp, data.Email); err != nil {
		return err
	}
	
	return nil
}

func (authService *AuthService) RegisterVerify(token string) error {
	
	var user model.User

	if err := authService.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		return  err
	}
	
	if err := authService.db.Model(&user).Update("is_verified", true).Error; err != nil {
        return err
    }

	return nil
}