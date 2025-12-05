package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskService) CancelTask(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.SetupRoutes(router)
	return router
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful task creation",
			requestBody: CreateTaskRequest{
				Name:          "Test Task",
				Description:   "Test Description",
				ExecutionMode: domain.ExecutionModeImmediate,
			},
			mockSetup: func(m *MockTaskService) {
				m.On("CreateTask", mock.Anything, mock.Anything).Return(&domain.Task{
					ID:            "task-123",
					Name:          "Test Task",
					Description:   "Test Description",
					ExecutionMode: domain.ExecutionModeImmediate,
					Status:        domain.TaskStatusPending,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "task-123", resp.ID)
				assert.Equal(t, "Test Task", resp.Name)
			},
		},
		{
			name: "missing required field",
			requestBody: CreateTaskRequest{
				Description:   "Test Description",
				ExecutionMode: domain.ExecutionModeImmediate,
			},
			mockSetup:      func(m *MockTaskService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "invalid_request", resp.Error)
			},
		},
		{
			name: "invalid execution mode",
			requestBody: CreateTaskRequest{
				Name:          "Test Task",
				ExecutionMode: domain.ExecutionModeImmediate,
			},
			mockSetup: func(m *MockTaskService) {
				m.On("CreateTask", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidExecutionMode)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "invalid_execution_mode", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockSetup      func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful get task",
			taskID: "task-123",
			mockSetup: func(m *MockTaskService) {
				m.On("GetTask", mock.Anything, "task-123").Return(&domain.Task{
					ID:            "task-123",
					Name:          "Test Task",
					Status:        domain.TaskStatusPending,
					ExecutionMode: domain.ExecutionModeImmediate,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "task-123", resp.ID)
			},
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			mockSetup: func(m *MockTaskService) {
				m.On("GetTask", mock.Anything, "nonexistent").Return(nil, service.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "not_found", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/api/v1/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListTasks(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "list all tasks",
			queryParams: "",
			mockSetup: func(m *MockTaskService) {
				tasks := []*domain.Task{
					{
						ID:            "task-1",
						Name:          "Task 1",
						Status:        domain.TaskStatusPending,
						ExecutionMode: domain.ExecutionModeImmediate,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
					{
						ID:            "task-2",
						Name:          "Task 2",
						Status:        domain.TaskStatusRunning,
						ExecutionMode: domain.ExecutionModeScheduled,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				m.On("ListTasks", mock.Anything, mock.Anything).Return(tasks, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, float64(2), resp["count"])
			},
		},
		{
			name:        "list tasks with filter",
			queryParams: "?status=pending&limit=10",
			mockSetup: func(m *MockTaskService) {
				m.On("ListTasks", mock.Anything, mock.Anything).Return([]*domain.Task{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, float64(0), resp["count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/api/v1/tasks"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestCancelTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockSetup      func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful cancel",
			taskID: "task-123",
			mockSetup: func(m *MockTaskService) {
				m.On("CancelTask", mock.Anything, "task-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "task-123", resp["task_id"])
			},
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			mockSetup: func(m *MockTaskService) {
				m.On("CancelTask", mock.Anything, "nonexistent").Return(service.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "not_found", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetTaskStatus(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockSetup      func(*MockTaskService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful get status",
			taskID: "task-123",
			mockSetup: func(m *MockTaskService) {
				m.On("GetTask", mock.Anything, "task-123").Return(&domain.Task{
					ID:         "task-123",
					Status:     domain.TaskStatusRunning,
					RetryCount: 2,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "task-123", resp["task_id"])
				assert.Equal(t, "running", resp["status"])
				assert.Equal(t, float64(2), resp["retry_count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			req, _ := http.NewRequest("GET", "/api/v1/tasks/"+tt.taskID+"/status", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHealthCheck(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "task not found",
			err:            service.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:           "invalid input",
			err:            service.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_input",
		},
		{
			name:           "invalid execution mode",
			err:            service.ErrInvalidExecutionMode,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_execution_mode",
		},
		{
			name:           "invalid schedule config",
			err:            service.ErrInvalidScheduleConfig,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_schedule_config",
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handler := &Handler{}
			handler.handleError(c, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}
