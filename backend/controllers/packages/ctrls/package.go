package ctrls

import (
	"booking-saas/controllers/packages/models"
	"booking-saas/controllers/packages/repositories"
	"booking-saas/controllers/packages/usecases"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type PackageController struct{ Usecase *usecases.PackageUsecase }

func NewPackageController(u *usecases.PackageUsecase) *PackageController {
	return &PackageController{Usecase: u}
}

func (ctrl *PackageController) List(c *gin.Context)    { ctrl.list(c, false) }
func (ctrl *PackageController) Catalog(c *gin.Context) { ctrl.list(c, true) }
func (ctrl *PackageController) list(c *gin.Context, activeOnly bool) {
	result, err := ctrl.Usecase.List(c.Request.Context(), activeOnly)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": result})
}
func (ctrl *PackageController) Create(c *gin.Context) { ctrl.save(c, "") }
func (ctrl *PackageController) Update(c *gin.Context) {
	id := c.Param("id")
	if !models.ValidID(id) {
		c.JSON(400, gin.H{"success": false, "message": "Invalid package ID"})
		return
	}
	ctrl.save(c, id)
}
func (ctrl *PackageController) save(c *gin.Context, id string) {
	var p models.Package
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "Invalid package data"})
		return
	}
	saved, err := ctrl.Usecase.Save(c.Request.Context(), p, id)
	if err != nil {
		respondError(c, err)
		return
	}
	status := http.StatusCreated
	if id != "" {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{"success": true, "data": saved})
}
func respondError(c *gin.Context, err error) {
	status, message := 500, "Unable to process packages"
	switch {
	case errors.Is(err, usecases.ErrInvalidPackage):
		status, message = 400, "Invalid package data"
	case errors.Is(err, repositories.ErrNotFound):
		status, message = 404, "Package not found"
	case errors.Is(err, repositories.ErrDuplicateCode):
		status, message = 409, "Package code already exists"
	default:
		log.Printf("package operation: %v", err)
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
