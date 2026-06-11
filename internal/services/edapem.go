package services

import (
	"context"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
)

type EdapemService struct {
	EdapemRepository interfaces.IEdapemRepository
}

func NewEdapemService(edapemRepository interfaces.IEdapemRepository) *EdapemService {
	return &EdapemService{
		EdapemRepository: edapemRepository,
	}
}

func (s *EdapemService) GetTotalDapemByType(ctx context.Context, startDate, endDate, dapemType string) (models.EdapemSummaryRow, error) {
	return s.EdapemRepository.GetTotalDapemByType(ctx, startDate, endDate, dapemType)
}
