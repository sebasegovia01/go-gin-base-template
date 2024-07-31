package services

import (
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/sebasegovia01/base-template-go-gin/repositories"
)

type ATMService struct {
	repo *repositories.ATMRepository
}

func NewATMService(repo *repositories.ATMRepository) *ATMService {
	return &ATMService{repo: repo}
}

func (s *ATMService) Create(atm models.ATM) (*models.ATM, error) {
	return s.repo.Create(atm)
}

func (s *ATMService) GetAll() ([]models.ATM, error) {
	return s.repo.GetAll()
}

func (s *ATMService) GetByID(id int) (*models.ATM, error) {
	return s.repo.GetByID(id)
}

func (s *ATMService) Update(atm models.ATM) (*models.ATM, error) {
	return s.repo.Update(atm)
}

func (s *ATMService) Delete(id int) error {
	return s.repo.Delete(id)
}
