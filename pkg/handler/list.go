package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/balamuteon/todo_restapi/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) createList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}
	defer h.invalidateListCache(userId)

	var input todo.TodoList
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.TodoList.Create(userId, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type getAllListsResponse struct {
	Data []todo.TodoList `json:"data"`
}

func (h *Handler) getAllLists(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:lists", userId)
	cacheValue, err := h.cache.Get(ctx, cacheKey)
	if err == nil {
		var lists []todo.TodoList
		if err := json.Unmarshal([]byte(cacheValue), &lists); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, getAllListsResponse{
			Data: lists,
		})
		logrus.Debug("got from cache")
		return
	}

	lists, err := h.services.TodoList.GetAll(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.cache.Set(ctx, cacheKey, lists, cache.СacheTTL)

	c.JSON(http.StatusOK, getAllListsResponse{
		Data: lists,
	})
}

func (h *Handler) getListById(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var list todo.TodoList
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:lists:%d", userId, id)
	cacheValue, err := h.cache.Get(ctx, cacheKey)
	if err == nil {
		if err := json.Unmarshal([]byte(cacheValue), &list); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, list)
		logrus.Debug("got from cache")
		return
	}

	list, err = h.services.TodoList.GetById(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.cache.Set(ctx, cacheKey, list, cache.СacheTTL)

	c.JSON(http.StatusOK, list)
}

func (h *Handler) updateList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}
	defer h.invalidateListCache(userId)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var input todo.UpdateListInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.TodoList.Update(userId, id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}
	defer h.invalidateListCache(userId)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.TodoList.Delete(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "ok",
	})
}

func (h *Handler) invalidateListCache(userId int) {
	ctx := context.Background()
	cachePattern := fmt.Sprintf("user:%d:lists*", userId)
	h.cache.Delete(ctx, cachePattern)
	logrus.Debug("cache invalidated")
}