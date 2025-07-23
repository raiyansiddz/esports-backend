package services

import (
	"encoding/json"
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type ContestTemplateService interface {
	CreateTemplate(req *CreateContestTemplateRequest) (*models.ContestTemplate, error)
	GetAllTemplates() ([]models.ContestTemplate, error)
	GetActiveTemplates() ([]models.ContestTemplate, error)
	GetTemplatesByGame(gameID uuid.UUID) ([]models.ContestTemplate, error)
	GetVIPTemplates() ([]models.ContestTemplate, error)
	GetTemplateByID(id uuid.UUID) (*models.ContestTemplate, error)
	UpdateTemplate(id uuid.UUID, req *CreateContestTemplateRequest) error
	ToggleTemplateStatus(id uuid.UUID) error
	DeleteTemplate(id uuid.UUID) error
	CreateContestFromTemplate(templateID, matchID uuid.UUID) (*models.Contest, error)
}

type CreateContestTemplateRequest struct {
	Name           string                 `json:"name" binding:"required"`
	GameID         uuid.UUID              `json:"game_id" binding:"required"`
	EntryFee       float64                `json:"entry_fee" binding:"required"`
	PrizeStructure map[string]interface{} `json:"prize_structure" binding:"required"`
	MaxEntries     int                    `json:"max_entries" binding:"required"`
	IsVIP          bool                   `json:"is_vip"`
}

type contestTemplateService struct {
	templateRepo repository.ContestTemplateRepository
	contestRepo  repository.ContestRepository
	gameRepo     repository.GameRepository
	config       *config.Config
}

func NewContestTemplateService(
	templateRepo repository.ContestTemplateRepository,
	contestRepo repository.ContestRepository,
	gameRepo repository.GameRepository,
	config *config.Config,
) ContestTemplateService {
	return &contestTemplateService{
		templateRepo: templateRepo,
		contestRepo:  contestRepo,
		gameRepo:     gameRepo,
		config:       config,
	}
}

func (s *contestTemplateService) CreateTemplate(req *CreateContestTemplateRequest) (*models.ContestTemplate, error) {
	// Verify game exists
	_, err := s.gameRepo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Convert prize structure to JSON
	prizeJSON, err := json.Marshal(req.PrizeStructure)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize prize structure: %w", err)
	}

	template := &models.ContestTemplate{
		Name:           req.Name,
		GameID:         req.GameID,
		EntryFee:       req.EntryFee,
		PrizeStructure: string(prizeJSON),
		MaxEntries:     req.MaxEntries,
		IsVIP:          req.IsVIP,
		IsActive:       true,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (s *contestTemplateService) GetAllTemplates() ([]models.ContestTemplate, error) {
	return s.templateRepo.GetAll()
}

func (s *contestTemplateService) GetActiveTemplates() ([]models.ContestTemplate, error) {
	return s.templateRepo.GetActive()
}

func (s *contestTemplateService) GetTemplatesByGame(gameID uuid.UUID) ([]models.ContestTemplate, error) {
	return s.templateRepo.GetByGameID(gameID)
}

func (s *contestTemplateService) GetVIPTemplates() ([]models.ContestTemplate, error) {
	return s.templateRepo.GetVIPTemplates()
}

func (s *contestTemplateService) GetTemplateByID(id uuid.UUID) (*models.ContestTemplate, error) {
	return s.templateRepo.GetByID(id)
}

func (s *contestTemplateService) UpdateTemplate(id uuid.UUID, req *CreateContestTemplateRequest) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Verify game exists if changed
	if template.GameID != req.GameID {
		_, err := s.gameRepo.GetByID(req.GameID)
		if err != nil {
			return fmt.Errorf("game not found: %w", err)
		}
	}

	// Convert prize structure to JSON
	prizeJSON, err := json.Marshal(req.PrizeStructure)
	if err != nil {
		return fmt.Errorf("failed to serialize prize structure: %w", err)
	}

	template.Name = req.Name
	template.GameID = req.GameID
	template.EntryFee = req.EntryFee
	template.PrizeStructure = string(prizeJSON)
	template.MaxEntries = req.MaxEntries
	template.IsVIP = req.IsVIP

	if err := s.templateRepo.Update(template); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

func (s *contestTemplateService) ToggleTemplateStatus(id uuid.UUID) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	template.IsActive = !template.IsActive

	if err := s.templateRepo.Update(template); err != nil {
		return fmt.Errorf("failed to update template status: %w", err)
	}

	return nil
}

func (s *contestTemplateService) DeleteTemplate(id uuid.UUID) error {
	return s.templateRepo.Delete(id)
}

func (s *contestTemplateService) CreateContestFromTemplate(templateID, matchID uuid.UUID) (*models.Contest, error) {
	// Get template
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create contest from template
	contest := &models.Contest{
		MatchID:        matchID,
		Name:           template.Name,
		EntryFee:       template.EntryFee,
		PrizePool:      template.PrizeStructure,
		MaxEntries:     template.MaxEntries,
		CurrentEntries: 0,
		IsPrivate:      template.IsVIP, // VIP templates create private contests
		Status:         "open",
	}

	if err := s.contestRepo.Create(contest); err != nil {
		return nil, fmt.Errorf("failed to create contest from template: %w", err)
	}

	return contest, nil
}