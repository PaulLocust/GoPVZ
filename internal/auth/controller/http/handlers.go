package http

import (
	"GoPVZ/internal/auth/usecase"
	"GoPVZ/internal/auth/validation"
	"GoPVZ/internal/dto"
	"GoPVZ/pkg/pkgValidator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// DummyLogin godoc
// @Summary      Получение тестового токена
// @Description  Генерирует токен без проверки пароля (для тестирования)
// @Tags         Domain auth
// @Accept       json
// @Produce      json
// @Param        input  body      dto.PostDummyLoginJSONBody  true  "Роль пользователя"
// @Success      200    {object}  dto.TokenResponse
// @Failure      400    {object}  dto.Error
// @Failure      500    {object}  dto.Error
// @Router       /dummyLogin [post]
func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req dto.PostDummyLoginJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidInput.Error()})
		return
	}

	validator := validation.NewDummyLoginValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	token, err := h.uc.DummyLogin(c, string(req.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.TokenResponse{
		Token: token,
	})
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя по email и паролю
// @Tags         Domain auth
// @Accept       json
// @Produce      json
// @Param        input  body      dto.PostRegisterJSONBody  true  "Данные для регистрации"
// @Success      201    {object}  dto.User
// @Failure      400    {object}  dto.Error
// @Failure      500    {object}  dto.Error
// @Router       /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.PostRegisterJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidInput.Error()})
		return
	}

	validator := validation.NewRegisterValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	user, err := h.uc.Register(c, req.Email, req.Password, string(req.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.User{Id: &user.ID, Email: req.Email, Role: dto.UserRole(user.Role)})
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю
// @Tags         Domain auth
// @Accept       json
// @Produce      json
// @Param        input  body      dto.PostLoginJSONBody  true  "Данные для входа"
// @Success      200    {object}  dto.TokenResponse
// @Failure      400    {object}  dto.Error
// @Failure      500    {object}  dto.Error
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.PostLoginJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidInput.Error()})
		return
	}

	validator := validation.NewLoginValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	token, err := h.uc.Login(c, string(req.Email), req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.TokenResponse{
		Token: token,
	})
}
