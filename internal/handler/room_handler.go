package handler

import (
	"net/http"
	"strconv"

	"realtime-api/internal/logger"
	"realtime-api/internal/model"
	"realtime-api/internal/service"
	"realtime-api/internal/websocket"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	roomService service.RoomService
}

func NewRoomHandler(roomService service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

func (h *RoomHandler) CreateRoom(c echo.Context) error {
	var req model.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	room, err := h.roomService.CreateRoom(c.Request().Context(), &req, userID)
	if err != nil {
		logger.Error("Failed to create room", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to create room",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Room created successfully",
		Data:    room,
	})
}

func (h *RoomHandler) GetRoom(c echo.Context) error {
	roomIDStr := c.Param("id")
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

	room, err := h.roomService.GetRoomByID(c.Request().Context(), roomID, userID)
	if err != nil {
		logger.Error("Failed to get room", logger.WithFields(map[string]interface{}{
			"room_id": roomID,
			"error":   err.Error(),
		}))
		return c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "Room not found",
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Room retrieved successfully",
		Data:    room,
	})
}

func (h *RoomHandler) ListRooms(c echo.Context) error {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")
	roomType := c.QueryParam("type")

	page := 1
	limit := 10

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

	var rooms []model.Room
	var meta *model.PaginationMeta
	var err error

	if roomType == "public" || roomType == "" {
		// List public rooms
		rooms, meta, err = h.roomService.GetPublicRooms(c.Request().Context(), page, limit)
	} else {
		// List user's rooms
		rooms, err = h.roomService.GetUserRooms(c.Request().Context(), userID)
		if err == nil {
			// Create pagination meta for user rooms
			meta = &model.PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      len(rooms),
				TotalPages: 1,
			}
		}
	}

	if err != nil {
		logger.Error("Failed to list rooms", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to retrieve rooms",
			Error:   err.Error(),
		})
	}

	response := model.PaginatedResponse{
		APIResponse: model.APIResponse{
			Success: true,
			Message: "Rooms retrieved successfully",
			Data:    rooms,
		},
		Meta: *meta,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *RoomHandler) UpdateRoom(c echo.Context) error {
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	var req model.UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	room, err := h.roomService.UpdateRoom(c.Request().Context(), roomID, &req, userID)
	if err != nil {
		logger.Error("Failed to update room", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to update room",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Room updated successfully",
		Data:    room,
	})
}

func (h *RoomHandler) DeleteRoom(c echo.Context) error {
	roomIDStr := c.Param("id")
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

	if err := h.roomService.DeleteRoom(c.Request().Context(), roomID, userID); err != nil {
		logger.Error("Failed to delete room", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to delete room",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Room deleted successfully",
	})
}

func (h *RoomHandler) JoinRoom(c echo.Context) error {
	roomIDStr := c.Param("id")
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

	if err := h.roomService.JoinRoom(c.Request().Context(), roomID, userID); err != nil {
		logger.Error("Failed to join room", logger.WithFields(map[string]interface{}{
			"room_id": roomID,
			"user_id": userID,
			"error":   err.Error(),
		}))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to join room",
			Error:   err.Error(),
		})
	}

	// Join WebSocket room
	websocket.GetHub().JoinRoom(userID, roomID)

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Successfully joined room",
	})
}

func (h *RoomHandler) LeaveRoom(c echo.Context) error {
	roomIDStr := c.Param("id")
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

	if err := h.roomService.LeaveRoom(c.Request().Context(), roomID, userID); err != nil {
		logger.Error("Failed to leave room", logger.WithFields(map[string]interface{}{
			"room_id": roomID,
			"user_id": userID,
			"error":   err.Error(),
		}))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to leave room",
			Error:   err.Error(),
		})
	}

	// Leave WebSocket room
	websocket.GetHub().LeaveRoom(userID, roomID)

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Successfully left room",
	})
}

func (h *RoomHandler) GetRoomMembers(c echo.Context) error {
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	members, err := h.roomService.GetRoomMembers(c.Request().Context(), roomID)
	if err != nil {
		logger.Error("Failed to get room members", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to retrieve room members",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Room members retrieved successfully",
		Data:    members,
	})
}

func (h *RoomHandler) AddMember(c echo.Context) error {
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	var req struct {
		UserID uuid.UUID `json:"user_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get inviter user ID from JWT token
	inviterID := uuid.New() // Placeholder

	if err := h.roomService.AddMember(c.Request().Context(), roomID, req.UserID, inviterID); err != nil {
		logger.Error("Failed to add room member", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to add member to room",
			Error:   err.Error(),
		})
	}

	// Join WebSocket room
	websocket.GetHub().JoinRoom(req.UserID, roomID)

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Member added to room successfully",
	})
}

func (h *RoomHandler) RemoveMember(c echo.Context) error {
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Get remover user ID from JWT token
	removerID := uuid.New() // Placeholder

	if err := h.roomService.RemoveMember(c.Request().Context(), roomID, userID, removerID); err != nil {
		logger.Error("Failed to remove room member", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to remove member from room",
			Error:   err.Error(),
		})
	}

	// Leave WebSocket room
	websocket.GetHub().LeaveRoom(userID, roomID)

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Member removed from room successfully",
	})
}

func (h *RoomHandler) CreateInvite(c echo.Context) error {
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid room ID format",
			Error:   err.Error(),
		})
	}

	var req model.CreateInviteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// TODO: Get inviter user ID from JWT token
	inviterID := uuid.New() // Placeholder

	invite, err := h.roomService.CreateInvite(c.Request().Context(), roomID, inviterID, &req)
	if err != nil {
		logger.Error("Failed to create room invite", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to create invite",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Room invite created successfully",
		Data:    invite,
	})
}

func (h *RoomHandler) AcceptInvite(c echo.Context) error {
	inviteCodeStr := c.Param("invite_code")

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	room, err := h.roomService.AcceptInvite(c.Request().Context(), inviteCodeStr, userID)
	if err != nil {
		logger.Error("Failed to accept room invite", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to accept invite",
			Error:   err.Error(),
		})
	}

	// Join WebSocket room
	websocket.GetHub().JoinRoom(userID, room.ID)

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Invite accepted successfully",
		Data: map[string]interface{}{
			"room": room,
		},
	})
}

func (h *RoomHandler) RejectInvite(c echo.Context) error {
	inviteCodeStr := c.Param("invite_code")

	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	if err := h.roomService.RejectInvite(c.Request().Context(), inviteCodeStr, userID); err != nil {
		logger.Error("Failed to reject room invite", logger.WithField("error", err.Error()))
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Failed to reject invite",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Invite rejected successfully",
	})
}

// ListUserChatRooms returns paginated list of user's chat rooms for chat list display
func (h *RoomHandler) ListUserChatRooms(c echo.Context) error {
	// TODO: Get user ID from JWT token
	userID := uuid.New() // Placeholder

	// Get pagination parameters
	page := 1
	limit := 20

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	rooms, meta, err := h.roomService.ListUserChatRooms(c.Request().Context(), userID, page, limit)
	if err != nil {
		logger.Error("Failed to get user chat rooms", logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to get chat rooms",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Chat rooms retrieved successfully",
		Data: map[string]interface{}{
			"rooms": rooms,
			"meta":  meta,
		},
	})
}

// CreateOrGetDirectRoom creates or gets an existing direct room between two users
func (h *RoomHandler) CreateOrGetDirectRoom(c echo.Context) error {
	// TODO: Get user ID from JWT token
	currentUserID := uuid.New() // Placeholder

	otherUserIDStr := c.Param("user_id")
	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Prevent creating direct room with self
	if currentUserID == otherUserID {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Cannot create direct room with yourself",
		})
	}

	room, err := h.roomService.CreateOrGetDirectRoom(c.Request().Context(), currentUserID, otherUserID)
	if err != nil {
		logger.Error("Failed to create or get direct room", logger.WithFields(map[string]interface{}{
			"current_user_id": currentUserID,
			"other_user_id":   otherUserID,
			"error":           err.Error(),
		}))
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "Failed to create or get direct room",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Direct room ready",
		Data:    room,
	})
}
