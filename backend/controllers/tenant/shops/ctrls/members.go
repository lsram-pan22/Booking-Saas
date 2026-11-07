package ctrls

import (
	"booking-saas/controllers/tenant/shops/usecases"
	"booking-saas/middleware"
	"github.com/gin-gonic/gin"
)

type MembersController struct{ Usecase *usecases.MembersUsecase }

func NewMembersController(u *usecases.MembersUsecase) *MembersController {
	return &MembersController{Usecase: u}
}
func (ctrl *MembersController) List(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	data, err := ctrl.Usecase.List(c.Request.Context(), claims.UserID, c.Param("shopID"))
	if err != nil {
		manageError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
