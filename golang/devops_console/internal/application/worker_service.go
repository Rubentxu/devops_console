package application

import (
    "devops_console/internal/domain/worker"
)

type WorkerService struct {
    workerRepository worker.WorkerRepository
}

func NewWorkerService(workerRepository worker.WorkerRepository) *WorkerService {
    return &WorkerService{workerRepository: workerRepository}
}

func (s *WorkerService) CreateWorker(workerCreate worker.WorkerCreate) (*worker.Worker, error) {
    return s.workerRepository.Create(workerCreate)
}

func (s *WorkerService) GetAllWorkers() ([]*worker.Worker, error) {
    return s.workerRepository.GetAll()
}

func (s *WorkerService) GetWorkerByID(workerID string) (*worker.Worker, error) {
    return s.workerRepository.GetByID(workerID)
}

func (s *WorkerService) UpdateWorker(workerID string, workerUpdate worker.WorkerUpdate) (*worker.Worker, error) {
    return s.workerRepository.Update(workerID, workerUpdate)
}

func (s *WorkerService) DeleteWorker(workerID string) error {
    return s.workerRepository.Delete(workerID)
}