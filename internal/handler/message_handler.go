package handler

import (
	"net/http"
	"strconv"

	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

func (h *MessageHandler) SendMessage(c echo.Context) error {
	var req model.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	message, err := h.messageService.SendMessage(c.Request().Context(), &req, userID)
	if err != nil {
		logger.Error("Failed to send message", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to send message",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Message sent successfully",
		Data:    message,
	})
}

func (h *MessageHandler) GetMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	message, err := h.messageService.GetMessageByID(c.Request().Context(), messageID, userID)
	if err != nil {
		logger.Error("Failed to get message", logger.WithFields(map[string]interface{}{
			"message_id": messageID,
			"error":      err.Error(),
		}))
		return c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "Message not found",
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Message retrieved successfully",
		Data:    message,
	})
}

func (h *MessageHandler) GetRoomMessages(c echo.Context) error {
	roomIDStr := c.Param("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	limit := 50

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	messages, meta, err := h.messageService.GetMessages(c.Request().Context(), roomID, userID, page, limit)
	if err != nil {
		logger.Error("Failed to get room messages", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to retrieve messages",
			Error:   err.Error(),
		})
	}

	response := model.PaginatedResponse{
		APIResponse: model.APIResponse{
			Success: true,
			Message: "Messages retrieved successfully",
			Data:    messages,
		},
		Meta: *meta,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *MessageHandler) EditMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	var req model.EditMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	message, err := h.messageService.EditMessage(c.Request().Context(), messageID, &req, userID)
	if err != nil {
		logger.Error("Failed to edit message", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to edit message",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Message edited successfully",
		Data:    message,
	})
}

func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.DeleteMessage(c.Request().Context(), messageID, userID); err != nil {
		logger.Error("Failed to delete message", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to delete message",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Message deleted successfully",
	})
}

func (h *MessageHandler) ReactToMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	var req model.ReactToMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.ReactToMessage(c.Request().Context(), messageID, &req, userID); err != nil {
		logger.Error("Failed to add reaction", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to add reaction",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Reaction added successfully",
	})
}

func (h *MessageHandler) RemoveReaction(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	emoji := c.QueryParam("emoji")
	if emoji == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Emoji parameter is required",
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.RemoveReaction(c.Request().Context(), messageID, emoji, userID); err != nil {
		logger.Error("Failed to remove reaction", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to remove reaction",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Reaction removed successfully",
	})
}

func (h *MessageHandler) MarkAsRead(c echo.Context) error {
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid message ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.MarkAsRead(c.Request().Context(), messageID, userID); err != nil {
		logger.Error("Failed to mark message as read", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to mark message as read",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Message marked as read",
	})
}

func (h *MessageHandler) StartTyping(c echo.Context) error {
	roomIDStr := c.Param("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.StartTyping(c.Request().Context(), roomID, userID); err != nil {
		logger.Error("Failed to start typing", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to start typing",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Typing started",
	})
}

func (h *MessageHandler) StopTyping(c echo.Context) error {
	roomIDStr := c.Param("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.messageService.StopTyping(c.Request().Context(), roomID, userID); err != nil {
		logger.Error("Failed to stop typing", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to stop typing",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Typing stopped",
	})
}
