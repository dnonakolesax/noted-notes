package services

type AccessRepo interface {
	Get(fileID string, userID string, byBlock bool) (string, error)
}

type AccessService struct {
	aRepo AccessRepo
}

func NewAccessService(repo AccessRepo) *AccessService {
	return &AccessService{aRepo: repo}
}

func (as *AccessService) Get(fileID string, userID string, byBlock bool) (string, error) {
	rights, err := as.aRepo.Get(fileID, userID, byBlock)

	if err != nil {
		return "", err
	}

	return rights, nil
}
