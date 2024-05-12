package routes

import (
	"inventory/internal/models"
	"inventory/internal/services"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetOrders(c *gin.Context, o services.OrderServicable) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(err)
		return
	}

	orders, err := o.FindByProductID(uint(productID))
	if err != nil {
		c.Error(err)
		c.Redirect(302, "/products")
		return
	}

	c.HTML(http.StatusOK, "orders.html", gin.H{
		"Orders": orders,
		"States": models.OrderStates,
	})
}

func PostOrders(c *gin.Context, o services.OrderServicable) {
	defer c.Redirect(http.StatusFound, c.Request.URL.Path)

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(err)
		return
	}

	quantity, err := strconv.ParseUint(c.PostForm("quantity"), 10, 32)
	if err != nil {
		c.Error(err)
		return
	}

	state, err := strconv.ParseUint(c.PostForm("state"), 10, 32)
	if err != nil {
		c.Error(err)
		return
	}

	id, err := strconv.Atoi(c.PostForm("id"))
	if err != nil {
		id = -1
	}

	if id <= 0 {
		order := models.Order{ProductID: uint(productID), Quantity: uint(quantity), State: models.OrderState(state)}
		err = o.Create(&order)
		if err != nil {
			c.Error(err)
		}
	} else {
		order := models.Order{ID: uint(id), ProductID: uint(productID), Quantity: uint(quantity), State: models.OrderState(state)}
		err = o.Update(&order)
		if err != nil {
			c.Error(err)
		}
	}
}

func DeleteOrders(c *gin.Context, o services.OrderServicable) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		return
	}

	id, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		c.Error(err)
		return
	}

	err = o.Delete(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
}
