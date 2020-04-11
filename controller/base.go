package controller

import (
	"airbox/model"

	"github.com/labstack/echo/v4"
)

// BaseController is responsible for the common operation of controller
type BaseController struct {
}

var base *BaseController

// GetBaseController return instance of BaseController
func GetBaseController() *BaseController {
	if base == nil {
		base = &BaseController{}
	}
	return base
}

// Auth return the authority of request
func (*BaseController) Auth(c echo.Context) *model.User {
	value := c.Get("authority")
	if value == nil {
		return nil
	}
	return value.(*model.User)
}
