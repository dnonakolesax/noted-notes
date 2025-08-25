package services

type BlockRepository interface {
	Save(name string, text string) error
}

type BlockService struct {
	repo BlockRepository 
}

func (bs *BlockService) Save(name string, text string) error {
	err := bs.repo.Save(name, text)

	if err != nil {
		return err
	}

	return nil
}

func NewBlockService(repo BlockRepository) *BlockService {
	return &BlockService{
		repo: repo,
	}
}
