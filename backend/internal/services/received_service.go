type receivedService struct {
	repo  repository.ReceivedRepository
	cache *cache.Cache
}

func NewReceivedService(repo repository.ReceivedRepository, cache *cache.Cache) ReceivedService {
	return &receivedService{
		repo:  repo,
		cache: cache,
	}
}

func (s *receivedService) GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error) {
	// Implementation would go here
	return s.repo.GetFiltered(ctx, filters, limit, offset)
}

func (s *receivedService) GetByID(ctx context.Context, id int) (*models.ReceivedItem, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("received:%d", id)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if item, ok := cached.(*models.ReceivedItem); ok {
			return item, nil
		}
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache.Set(cacheKey, item)
	return item, nil
}

func (s *receivedService) Create(ctx context.Context, item *models.ReceivedItem) error {
	err := s.repo.Create(ctx, item)
	if err != nil {
		return err
	}

	// Invalidate relevant caches
	s.cache.Delete("received:all")
	s.cache.Delete("analytics:dashboard")
	return nil
}

func (s *receivedService) Update(ctx context.Context, item *models.ReceivedItem) error {
	err := s.repo.Update(ctx, item)
	if err != nil {
		return err
	}

	// Invalidate caches
	cacheKey := fmt.Sprintf("received:%d", item.ID)
	s.cache.Delete(cacheKey)
	s.cache.Delete("received:all")
	s.cache.Delete("analytics:dashboard")
	return nil
}

func (s *receivedService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	cacheKey := fmt.Sprintf("received:%d", id)
	s.cache.Delete(cacheKey)
	s.cache.Delete("received:all")
	return nil
}

func (s *receivedService) GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error) {
	return s.repo.GetByWorkOrder(ctx, workOrder)
}

func (s *receivedService) UpdateStatus(ctx context.Context, id int, status string) error {
	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Invalidate caches
	cacheKey := fmt.Sprintf("received:%d", id)
	s.cache.Delete(cacheKey)
	return nil
}

func (s *receivedService) GetPendingInspection(ctx context.Context) ([]models.ReceivedItem, error) {
	cacheKey := "received:pending_inspection"
	if cached, exists := s.cache.Get(cacheKey); exists {
		if items, ok := cached.([]models.ReceivedItem); ok {
			return items, nil
		}
	}

	items, err := s.repo.GetPendingInspection(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(cacheKey, items)
	return items, nil
}

