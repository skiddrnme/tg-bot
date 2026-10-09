package service_test

import (
	"context"
	
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct {
	telegramID  int64
	githubLogin string
	err         error

	// Что записывать (для проверки вызовов)
	getTelegramIDCalls []string


}

type MockNotifierService struct{
	mock.Mock
}

func (m *mockUserRepo) GetTelegramID(ctx context.Context, login string) (int64, error){
	m.getTelegramIDCalls = append(m.getTelegramIDCalls, login)
	return m.telegramID, m.err
}

func (m *mockUserRepo) GetGitHubLogin(ctx context.Context, login string) (string, error){
	return m.githubLogin, m.err
}

func (m *mockUserRepo) Save(ctx context.Context, id int64, login string) error{
	return m.err
	
}


