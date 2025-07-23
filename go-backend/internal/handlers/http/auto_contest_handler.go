package http

import (
        "net/http"

        "esports-fantasy-backend/internal/services"

        "github.com/gin-gonic/gin"
)

type AutoContestHandler struct {
        autoContestService *services.AutoContestService
}

func NewAutoContestHandler(autoContestService *services.AutoContestService) *AutoContestHandler {
        return &AutoContestHandler{
                autoContestService: autoContestService,
        }
}

// GetSchedulerStatus gets the status of auto contest management scheduler
func (h *AutoContestHandler) GetSchedulerStatus(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        status := h.autoContestService.GetSchedulerStatus()

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Scheduler status retrieved successfully",
                "data":    status,
        })
}

// ForceDistributePrizes manually triggers prize distribution for a contest
func (h *AutoContestHandler) ForceDistributePrizes(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        contestID := c.Param("contestId")
        if contestID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Contest ID is required",
                })
                return
        }

        if err := h.autoContestService.ForceDistributePrizes(contestID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to distribute prizes",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Prizes distributed successfully",
                "data": gin.H{
                        "contest_id": contestID,
                        "action":     "FORCE_PRIZE_DISTRIBUTION",
                },
        })
}

// ForceLockContest manually locks a contest
func (h *AutoContestHandler) ForceLockContest(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        contestID := c.Param("contestId")
        if contestID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Contest ID is required",
                })
                return
        }

        if err := h.autoContestService.ForceLockContest(contestID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to lock contest",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Contest locked successfully",
                "data": gin.H{
                        "contest_id": contestID,
                        "action":     "FORCE_CONTEST_LOCK",
                },
        })
}