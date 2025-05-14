package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const cacheTTL = 60 * time.Second

func (h *Handler) createList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

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

	h.invalidateCache(userId)

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
	cacheKey := fmt.Sprintf("lists:user:%d", userId)
	cachedLists, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var lists []todo.TodoList
		if err := json.Unmarshal([]byte(cachedLists), &lists); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "failed to unmarshal cached lists")
			return
		}
		// logrus.Debug("got lists from redis")
		c.JSON(http.StatusOK, getAllListsResponse{Data: lists})
		return
	} else if err != redis.Nil {
		newErrorResponse(c, http.StatusInternalServerError, "failed to get lists from redis: "+err.Error())
		return
	}

	lists, err := h.services.TodoList.GetAll(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	listsBytes, err := json.Marshal(lists)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "failed to marshal lists")
		return
	}

	err = h.redis.Set(ctx, cacheKey, listsBytes, cacheTTL).Err()
	if err != nil {
		logrus.Errorf("failed to set lists in redis: %s", err)
	}

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

	list, err := h.services.TodoList.GetById(userId, id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *Handler) updateList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

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

	h.invalidateCache(userId)

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		return
	}

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

	h.invalidateCache(userId)

	c.JSON(http.StatusOK, statusResponse{
		Status: "ok",
	})
}

func (h *Handler) invalidateCache(userId int) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("lists:user:%d", userId)
	err := h.redis.Del(ctx, cacheKey).Err()
	if err != nil {
		logrus.Errorf("failed to delete cache for user %d: %s", userId, err.Error())
	}
}