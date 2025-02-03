package chore

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	auth "donetick.com/core/internal/authorization"
	chModel "donetick.com/core/internal/chore/model"
	chRepo "donetick.com/core/internal/chore/repo"
	lRepo "donetick.com/core/internal/label/repo"
	"donetick.com/core/internal/notifier"
	nRepo "donetick.com/core/internal/notifier/repo"
	nps "donetick.com/core/internal/notifier/service"
	"donetick.com/core/logging"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type ThingTrigger struct {
	ID           int    `json:"thingID" binding:"required"`
	TriggerState string `json:"triggerState" binding:"required"`
	Condition    string `json:"condition"`
}

type LabelReq struct {
	LabelID int `json:"id" binding:"required"`
}

type ChoreReq struct {
	Name                 string                        `json:"name" binding:"required"`
	FrequencyType        chModel.FrequencyType         `json:"frequencyType"`
	ID                   int                           `json:"id"`
	DueDate              string                        `json:"dueDate"`
	IsRolling            bool                          `json:"isRolling"`
	IsActive             bool                          `json:"isActive"`
	Frequency            int                           `json:"frequency"`
	FrequencyMetadata    *chModel.FrequencyMetadata    `json:"frequencyMetadata"`
	Notification         bool                          `json:"notification"`
	NotificationMetadata *chModel.NotificationMetadata `json:"notificationMetadata"`
	Labels               *[]LabelReq                   `json:"labels"`
}
type Handler struct {
	choreRepo *chRepo.ChoreRepository
	notifier  *notifier.Notifier
	nPlanner  *nps.NotificationPlanner
	nRepo     *nRepo.NotificationRepository
	lRepo     *lRepo.LabelRepository
}

func NewHandler(cr *chRepo.ChoreRepository, nt *notifier.Notifier,
	np *nps.NotificationPlanner, nRepo *nRepo.NotificationRepository, lRepo *lRepo.LabelRepository) *Handler {
	return &Handler{
		choreRepo: cr,
		notifier:  nt,
		nPlanner:  np,
		nRepo:     nRepo,
		lRepo:     lRepo,
	}
}

func (h *Handler) getChores(c *gin.Context) {
	u, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	chores, err := h.choreRepo.GetChores(c, u.ID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chores",
		})
		return
	}

	c.JSON(200, gin.H{
		"res": chores,
	})
}

func (h *Handler) getChore(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	chore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}

	if currentUser.ID != chore.CreatedBy {
		c.JSON(403, gin.H{
			"error": "You are not allowed to view this chore",
		})
		return
	}

	c.JSON(200, gin.H{
		"res": chore,
	})
}

func (h *Handler) createChore(c *gin.Context) {
	logger := logging.FromContext(c)
	currentUser, ok := auth.CurrentUser(c)

	logger.Debug("Create chore", "currentUser", currentUser)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	var choreReq ChoreReq
	if err := c.ShouldBindJSON(&choreReq); err != nil {
		log.Print(err)
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}

	var dueDate *time.Time

	if choreReq.DueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, choreReq.DueDate)
		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid date",
			})
			return
		}

	}

	freqencyMetadataBytes, err := json.Marshal(choreReq.FrequencyMetadata)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error marshalling frequency metadata",
		})
		return
	}
	stringFrequencyMetadata := string(freqencyMetadataBytes)

	notificationMetadataBytes, err := json.Marshal(choreReq.NotificationMetadata)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error marshalling notification metadata",
		})
		return
	}
	stringNotificationMetadata := string(notificationMetadataBytes)

	createdChore := &chModel.Chore{
		Name:                 choreReq.Name,
		FrequencyType:        choreReq.FrequencyType,
		Frequency:            choreReq.Frequency,
		FrequencyMetadata:    &stringFrequencyMetadata,
		NextDueDate:          dueDate,
		CreatedBy:            currentUser.ID,
		IsRolling:            choreReq.IsRolling,
		IsActive:             true,
		Notification:         choreReq.Notification,
		NotificationMetadata: &stringNotificationMetadata,
		CreatedAt:            time.Now().UTC(),
	}
	id, err := h.choreRepo.CreateChore(c, createdChore)
	createdChore.ID = id

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error creating chore",
		})
		return
	}

	labels := make([]int, len(*choreReq.Labels))
	for i, label := range *choreReq.Labels {
		labels[i] = int(label.LabelID)
	}
	if err := h.lRepo.AssignLabelsToChore(c, createdChore.ID, currentUser.ID, labels, []int{}); err != nil {
		c.JSON(500, gin.H{
			"error": "Error adding labels",
		})
		return
	}

	go func() {
		h.nPlanner.GenerateNotifications(c, createdChore)
	}()

	c.JSON(200, gin.H{
		"res": id,
	})
}

func (h *Handler) editChore(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	var choreReq ChoreReq
	if err := c.ShouldBindJSON(&choreReq); err != nil {
		log.Print(err)
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}

	var dueDate *time.Time

	if choreReq.DueDate != "" {
		rawDueDate, err := time.Parse(time.RFC3339, choreReq.DueDate)
		rawDueDate = rawDueDate.UTC()
		dueDate = &rawDueDate
		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid date",
			})
			return
		}

	}

	oldChore, err := h.choreRepo.GetChore(c, choreReq.ID)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}
	if currentUser.ID != oldChore.CreatedBy {
		c.JSON(403, gin.H{
			"error": "You are not allowed to edit this chore",
		})
		return
	}
	freqencyMetadataBytes, err := json.Marshal(choreReq.FrequencyMetadata)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error marshalling frequency metadata",
		})
		return
	}

	stringFrequencyMetadata := string(freqencyMetadataBytes)

	notificationMetadataBytes, err := json.Marshal(choreReq.NotificationMetadata)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error marshalling notification metadata",
		})
		return
	}
	stringNotificationMetadata := string(notificationMetadataBytes)

	// Create a map to store the existing labels for quick lookup
	oldLabelsMap := make(map[int]struct{})
	for _, oldLabel := range *oldChore.Labels {
		oldLabelsMap[oldLabel.ID] = struct{}{}
	}
	newLabelMap := make(map[int]struct{})
	for _, newLabel := range *choreReq.Labels {
		newLabelMap[newLabel.LabelID] = struct{}{}
	}
	// check what labels need to be added and what labels need to be deleted:
	labelsToAdd := make([]int, 0)
	labelsToBeRemoved := make([]int, 0)

	for _, label := range *choreReq.Labels {
		if _, ok := oldLabelsMap[label.LabelID]; !ok {
			labelsToAdd = append(labelsToAdd, label.LabelID)
		}
	}
	for _, oldLabel := range *oldChore.Labels {
		if _, ok := newLabelMap[oldLabel.ID]; !ok {
			labelsToBeRemoved = append(labelsToBeRemoved, oldLabel.ID)
		}
	}

	if err := h.lRepo.AssignLabelsToChore(c, choreReq.ID, currentUser.ID, labelsToAdd, labelsToBeRemoved); err != nil {
		c.JSON(500, gin.H{
			"error": "Error adding labels",
		})
		return
	}

	updatedChore := &chModel.Chore{
		ID:                   choreReq.ID,
		Name:                 choreReq.Name,
		FrequencyType:        choreReq.FrequencyType,
		Frequency:            choreReq.Frequency,
		FrequencyMetadata:    &stringFrequencyMetadata,
		NextDueDate:          dueDate,
		CreatedBy:            currentUser.ID,
		IsRolling:            choreReq.IsRolling,
		IsActive:             choreReq.IsActive,
		Notification:         choreReq.Notification,
		NotificationMetadata: &stringNotificationMetadata,
		CreatedAt:            oldChore.CreatedAt,
	}
	if err := h.choreRepo.UpsertChore(c, updatedChore); err != nil {
		c.JSON(500, gin.H{
			"error": "Error adding chore",
		})
		return
	}

	go func() {
		h.nPlanner.GenerateNotifications(c, updatedChore)
	}()

	c.JSON(200, gin.H{
		"message": "Chore added successfully",
	})
}

func (h *Handler) deleteChore(c *gin.Context) {
	// logger := logging.FromContext(c)
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	// check if the user is the owner of the chore before deleting
	if err := h.choreRepo.IsChoreOwner(c, id, currentUser.ID); err != nil {
		c.JSON(403, gin.H{
			"error": "You are not allowed to delete this chore",
		})
		return
	}

	if err := h.choreRepo.DeleteChore(c, id); err != nil {
		c.JSON(500, gin.H{
			"error": "Error deleting chore",
		})
		return
	}
	h.nRepo.DeleteAllChoreNotifications(id)

	c.JSON(200, gin.H{
		"message": "Chore deleted successfully",
	})
}

func (h *Handler) skipChore(c *gin.Context) {
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	chore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}
	nextDueDate, err := scheduleNextDueDate(chore, chore.NextDueDate.UTC())
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error scheduling next due date",
		})
		return
	}

	if err := h.choreRepo.CompleteChore(c, chore, currentUser.ID, nextDueDate, nil); err != nil {
		c.JSON(500, gin.H{
			"error": "Error completing chore",
		})
		return
	}
	updatedChore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}

	c.JSON(200, gin.H{
		"res": updatedChore,
	})
}

func (h *Handler) updateDueDate(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}

	type DueDateReq struct {
		DueDate string `json:"dueDate" binding:"required"`
	}

	var dueDateReq DueDateReq
	if err := c.ShouldBindJSON(&dueDateReq); err != nil {
		log.Print(err)
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}

	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	rawDueDate, err := time.Parse(time.RFC3339, dueDateReq.DueDate)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid date",
		})
		return
	}
	dueDate := rawDueDate.UTC()
	chore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}

	if currentUser.ID != chore.CreatedBy {
		c.JSON(403, gin.H{
			"error": "You are not allowed to update this chore",
		})
	}

	chore.NextDueDate = &dueDate
	if err := h.choreRepo.UpsertChore(c, chore); err != nil {
		c.JSON(500, gin.H{
			"error": "Error updating due date",
		})
		return
	}

	c.JSON(200, gin.H{
		"res": chore,
	})
}

func (h *Handler) completeChore(c *gin.Context) {
	type CompleteChoreReq struct {
		Note string `json:"note"`
	}
	var req CompleteChoreReq
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}
	completeChoreID := c.Param("id")
	var completedDate time.Time
	rawCompletedDate := c.Query("completedDate")
	if rawCompletedDate == "" {
		completedDate = time.Now().UTC()
	} else {
		var err error
		completedDate, err = time.Parse(time.RFC3339, rawCompletedDate)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid date",
			})
			return
		}
	}

	_ = c.ShouldBind(&req)

	id, err := strconv.Atoi(completeChoreID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	chore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}

	var nextDueDate *time.Time
	if chore.FrequencyType == "adaptive" {
		history, err := h.choreRepo.GetChoreHistoryWithLimit(c, chore.ID, 5)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Error getting chore history",
			})
			return
		}
		nextDueDate, err = scheduleAdaptiveNextDueDate(chore, completedDate, history)
		if err != nil {
			log.Printf("Error scheduling next due date: %s", err)
			c.JSON(500, gin.H{
				"error": "Error scheduling next due date",
			})
			return
		}

	} else {
		nextDueDate, err = scheduleNextDueDate(chore, completedDate)
		if err != nil {
			log.Printf("Error scheduling next due date: %s", err)
			c.JSON(500, gin.H{
				"error": "Error scheduling next due date",
			})
			return
		}
	}

	if err := h.choreRepo.CompleteChore(c, chore, currentUser.ID, nextDueDate, &completedDate); err != nil {
		c.JSON(500, gin.H{
			"error": "Error completing chore",
		})
		return
	}
	updatedChore, err := h.choreRepo.GetChore(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore",
		})
		return
	}

	h.nPlanner.GenerateNotifications(c, updatedChore)

	c.JSON(200, gin.H{
		"res": updatedChore,
	})
}

func (h *Handler) GetChoreHistory(c *gin.Context) {
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	choreHistory, err := h.choreRepo.GetChoreHistory(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore history",
		})
		return
	}

	c.JSON(200, gin.H{
		"res": choreHistory,
	})
}

func (h *Handler) GetChoreDetail(c *gin.Context) {
	currentUser, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(500, gin.H{
			"error": "Error getting current user",
		})
		return
	}
	rawID := c.Param("id")
	id, err := strconv.Atoi(rawID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	detailed, err := h.choreRepo.GetChoreDetailByID(c, id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Error getting chore history",
		})
		return
	}

	if currentUser.ID != detailed.CreatedBy {
		c.JSON(403, gin.H{
			"error": "You are not allowed to view this chore",
		})
	}

	c.JSON(200, gin.H{
		"res": detailed,
	})
}

func Routes(router *gin.Engine, h *Handler, auth *jwt.GinJWTMiddleware) {
	choresRoutes := router.Group("api/v1/chores")
	choresRoutes.Use(auth.MiddlewareFunc())
	{
		choresRoutes.GET("/", h.getChores)
		choresRoutes.PUT("/", h.editChore)
		choresRoutes.POST("/", h.createChore)
		choresRoutes.GET("/:id", h.getChore)
		choresRoutes.GET("/:id/details", h.GetChoreDetail)
		choresRoutes.GET("/:id/history", h.GetChoreHistory)
		choresRoutes.POST("/:id/do", h.completeChore)
		choresRoutes.POST("/:id/skip", h.skipChore)
		choresRoutes.PUT("/:id/dueDate", h.updateDueDate)
		choresRoutes.DELETE("/:id", h.deleteChore)
	}
}
