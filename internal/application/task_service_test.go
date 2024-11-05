package application_test

import (
	"context"
	"devops_console/internal/application"
	"devops_console/internal/domain/entities"
	"devops_console/internal/ports"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Mocking the repository and executor
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (entities.DevOpsTask, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.DevOpsTask), args.Error(1)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entities.DevOpsTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entities.DevOpsTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetAll(ctx context.Context, filters ports.TaskFilters) ([]entities.DevOpsTask, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]entities.DevOpsTask), args.Error(1)
}

type MockTaskExecutor struct {
	mock.Mock
}

func (m *MockTaskExecutor) ExecuteTask(ctx context.Context, task *entities.DevOpsTask) (string, error) {
	args := m.Called(ctx, task)
	return args.String(0), args.Error(1)
}

func (m *MockTaskExecutor) CancelTask(ctx context.Context, executionID string) error {
	args := m.Called(ctx, executionID)
	return args.Error(0)
}

func (m *MockTaskExecutor) GetTaskStatus(ctx context.Context, taskExecutionID string) (entities.TaskStatus, error) {
	args := m.Called(ctx, taskExecutionID)
	return args.Get(0).(entities.TaskStatus), args.Error(1)
}

func (m *MockTaskExecutor) SubscribeToTaskEvents(taskExecutionID string) (<-chan entities.TaskEvent, error) {
	args := m.Called(taskExecutionID)
	return args.Get(0).(<-chan entities.TaskEvent), args.Error(1)
}

// Test suite
type TaskServiceTestSuite struct {
	suite.Suite
	service    *application.TaskServiceImpl
	repository *MockTaskRepository
	executor   *MockTaskExecutor
}

func (suite *TaskServiceTestSuite) SetupTest() {
	suite.repository = new(MockTaskRepository)
	suite.executor = new(MockTaskExecutor)
	suite.service = application.NewTaskServiceImpl(suite.repository, suite.executor)
}

func (suite *TaskServiceTestSuite) TestExecuteTask_Success() {
	taskID := "integration-tests-task-id"
	task := entities.DevOpsTask{ID: taskID}
	executionID := "execution-id"

	// Mocking the repository to return a task when GetByID is called
	suite.repository.On("GetByID", mock.Anything, taskID).Return(task, nil)
	// Mocking the executor to return a successful execution ID when ExecuteTask is called
	suite.executor.On("ExecuteTask", mock.Anything, &task).Return(executionID, nil)
	// Mocking the repository to successfully update the task
	suite.repository.On("Update", mock.Anything, mock.AnythingOfType("*entities.DevOpsTask")).Return(nil)

	// Calling the ExecuteTask method and checking the results
	resultExecutionID, err := suite.service.ExecuteTask(taskID)

	// Asserting that no error occurred
	assert.NoError(suite.T(), err)
	// Asserting that the returned execution ID matches the expected ID
	assert.Equal(suite.T(), executionID, resultExecutionID)
}

func (suite *TaskServiceTestSuite) TestExecuteTask_Failure() {
	taskID := "integration-tests-task-id"
	task := entities.DevOpsTask{ID: taskID}
	_ = "execution-id"
	execErr := fmt.Errorf("execution error")

	// Mocking the repository to return a task when GetByID is called
	suite.repository.On("GetByID", mock.Anything, taskID).Return(task, nil)
	// Mocking the executor to return an error when ExecuteTask is called
	suite.executor.On("ExecuteTask", mock.Anything, &task).Return("", execErr)
	// Mocking the repository to successfully update the task
	suite.repository.On("Update", mock.Anything, &task).Return(nil)

	// Calling the ExecuteTask method and checking the results
	resultExecutionID, err := suite.service.ExecuteTask(taskID)

	// Asserting that an error occurred
	assert.Error(suite.T(), err)
	// Asserting that the returned execution ID is empty
	assert.Empty(suite.T(), resultExecutionID)
	// Asserting that the error message matches the expected error message
	assert.Equal(suite.T(), execErr.Error(), err.Error())
}

func TestTaskServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TaskServiceTestSuite))
}
