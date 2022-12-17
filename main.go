package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

var (
	errItemNotFound = "item not found"
)

// Item is a model represents the inventory item.
type Item struct {
	ID        int     `json:"ID"`
	Name      string  `json:"name" binding:"required,alpha"`
	Quantity  int     `json:"quantity" binding:"required"`
	UnitPrice float64 `json:"unitPrice" binding:"required"`
}

// Pagination for fetching the items in a pagination way using query param.
type Pagination struct {
	Page int `form:"page" binding:"min=0"`
	Size int `form:"size" binding:"min=0"`
}

// ID is a model represents item id path parameter.
type ID struct {
	ID int `uri:"id" binding:"required"`
}

// inventory as map for quicker retrieval using id.
var inventory = make(map[int]Item)

// inventoryList as array to persist the order and to use pagination.
var inventoryList = make([]Item, 0)

func main() {
	// Server creation.
	server := gin.Default()

	// Route to add new item to the inventory.
	server.Handle(http.MethodPost, "/item", func(ctx *gin.Context) {
		var item Item
		if err := ctx.ShouldBindJSON(&item); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}
		id := len(inventory) + 1
		item.ID = id
		inventory[id] = item
		inventoryList = append(inventoryList, item)

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "successfully added",
			"id":      id,
		})
	})

	// Route to fetch all items from the inventory.
	server.Handle(http.MethodGet, "/item", func(ctx *gin.Context) {
		var pagination Pagination
		if err := ctx.ShouldBindQuery(&pagination); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}

		if pagination.Page == 0 && pagination.Size == 0 {
			ctx.JSON(http.StatusOK, inventoryList)
			return
		}

		if pagination.Size == 0 {
			pagination.Size = 1
		}
		if pagination.Page == 0 {
			pagination.Page = 1
		}

		start := (pagination.Page - 1) * pagination.Size
		end := pagination.Page * pagination.Size

		if !(len(inventoryList) > start) {
			ctx.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{
				"message": "page/size limit exceeded",
			})
			return
		}
		if len(inventoryList) >= end {
			ctx.JSON(http.StatusOK, inventoryList[start:end])
			return
		}
		ctx.JSON(http.StatusOK, inventoryList[start:])
		return
	})

	// Route to fetch only one item using id in path parameter.
	server.Handle(http.MethodGet, "/item/:id", func(ctx *gin.Context) {
		var id ID
		if err := ctx.ShouldBindUri(&id); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}

		item, isExists := inventory[id.ID]
		if !isExists {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": errItemNotFound,
			})
			return
		}

		ctx.JSON(http.StatusOK, item)
	})

	// Route to delete item in the inventory for given id in the path parameter.
	server.Handle(http.MethodDelete, "/item/:id", func(ctx *gin.Context) {
		var id ID
		if err := ctx.ShouldBindUri(&id); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}

		item, isExists := inventory[id.ID]
		if !isExists {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": errItemNotFound,
			})
			return
		}

		var itemPosition int
		for i, invItem := range inventoryList {
			if invItem.ID == id.ID {
				itemPosition = i
			}
		}

		// Deleting from the array
		inventoryList = append(inventoryList[:itemPosition], inventoryList[itemPosition+1:]...)

		// Deleting from the map
		delete(inventory, item.ID)
		ctx.JSON(http.StatusOK, gin.H{
			"message": "deleted",
		})
	})

	// Route to update item in the inventory for given id in the path parameter and item data from json body.
	server.Handle(http.MethodPatch, "/item/:id", func(ctx *gin.Context) {
		var id ID
		if err := ctx.ShouldBindUri(&id); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}

		var item Item
		if err := ctx.ShouldBindJSON(&item); err != nil {
			ctx.JSON(http.StatusBadRequest, customValidationError(err))
			return
		}

		_, isExists := inventory[id.ID]
		if !isExists {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": errItemNotFound,
			})
			return
		}

		// Updating the id as item might not have the id in json body.
		item.ID = id.ID

		// Updating the item in the inventory list by matching id.
		for i, invItem := range inventoryList {
			if invItem.ID == id.ID {
				inventoryList[i] = item
				break
			}
		}

		// Updating the map with the same object.
		inventory[id.ID] = item
		ctx.JSON(http.StatusOK, gin.H{
			"message": "updated",
		})
	})

	// Route to download the items as csv file.
	server.Handle(http.MethodGet, "/item/csv", func(ctx *gin.Context) {
		content := "id,name,quantity,unit price\n"
		for _, item := range inventoryList {
			content = content + fmt.Sprintf("%d,%s,%d,%f\n", item.ID, item.Name, item.Quantity, item.UnitPrice)
		}

		ctx.Set("Content-Disposition", "attachment; filename=inventory.csv")
		ctx.Data(http.StatusOK, "text/csv", []byte(content))
	})

	// Running the server.
	err := server.Run()
	if err != nil {
		log.Println(err)
	}
}

// customErrors is a error map to map valid error message for validate field bindings.
var customErrors = map[string]error{
	"Name.required":      errors.New("is required"),
	"Name.alpha":         errors.New("should only contain alphabets"),
	"Quantity.required":  errors.New("is required"),
	"UnitPrice.required": errors.New("is required"),
	"Page.min":           errors.New("should not be less than 1"),
	"Size.min":           errors.New("should not be less than 1"),
	"ID.required":        errors.New("is required"),
}

// customValidationError converts bindings validation error to readable error.
func customValidationError(err error) []map[string]string {
	errs := make([]map[string]string, 0)
	switch err := err.(type) {
	case validator.ValidationErrors:
		for _, e := range err {
			errorMap := make(map[string]string)

			key := e.Field() + "." + e.Tag()

			if v, ok := customErrors[key]; ok {
				errorMap[e.Field()] = v.Error()
			} else {
				errorMap[e.Field()] = fmt.Sprintf("custom message is not available: %v", err)
			}
			errs = append(errs, errorMap)
		}
		return errs
	case *json.UnmarshalTypeError:
		e := err
		errs = append(errs, map[string]string{e.Field: fmt.Sprintf("%v can not be a %v", e.Field, e.Value)})
		return errs
	}
	errs = append(errs, map[string]string{"unknown": fmt.Sprintf("unsupported custom error for: %v", err)})
	return errs
}
