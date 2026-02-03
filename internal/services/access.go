package services

import (
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/dnonakolesax/noted-notes/internal/validator"
)

type AccessRepo interface {
	GetAll(fileID string) ([]model.Access, error)
	Get(fileID string, userID string, byBlock bool) (string, error)
	Grant(fileID string, userID string, level string) error
	Update(fileID string, userID string, level string) error
	Revoke(fileID string, userID string) error
}

type AccessService struct {
	aRepo        AccessRepo
	aclValidator *validator.ACLValidator
}

func NewAccessService(repo AccessRepo) *AccessService {
	return &AccessService{aRepo: repo, aclValidator: validator.NewACL()}
}

func (as *AccessService) Get(fileID string, userID string, byBlock bool) (string, error) {
	rights, err := as.aRepo.Get(fileID, userID, byBlock)

	if err != nil {
		return "", err
	}

	return rights, nil
}

func (as *AccessService) Grant(fileID string, userID string, level string) error {
	err := as.aRepo.Grant(fileID, userID, level)

	if !as.aclValidator.Validate(level) {
		return fmt.Errorf("access is not valid!")
	}

	if err != nil {
		return err
	}

	return nil
}

func (as *AccessService) Update(fileID string, userID string, level string) error {
	err := as.aRepo.Update(fileID, userID, level)

	if !as.aclValidator.Validate(level) {
		return fmt.Errorf("access is not valid!")
	}
	
	if err != nil {
		return err
	}

	return nil
}

func (as *AccessService) Revoke(fileID string, userID string) error {
	err := as.aRepo.Revoke(fileID, userID)

	if err != nil {
		return err
	}

	return nil
}

func (as *AccessService) GetAll(filedID string) ([]model.Access, error) {
	list, err := as.aRepo.GetAll(filedID)

	if err != nil {
		return []model.Access{}, err
	}

	return list, nil
}
