package service

import (
	"context"

	"ulascan-be/constants"
	"ulascan-be/dto"
	"ulascan-be/entity"
	"ulascan-be/repository"
	"ulascan-be/utils"
)

type (
	UserService interface {
		RegisterUser(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
		GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
		GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error)
		Verify(ctx context.Context, req dto.UserLoginRequest) (dto.UserLoginResponse, error)
	}

	userService struct {
		userRepo   repository.UserRepository
		jwtService JWTService
	}
)

func NewUserService(userRepo repository.UserRepository, jwtService JWTService) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *userService) RegisterUser(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error) {
	_, flag, _ := s.userRepo.CheckEmail(ctx, nil, req.Email)
	if flag {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	user := entity.User{
		Name:     req.Name,
		Role:     constants.ENUM_ROLE_USER,
		Email:    req.Email,
		Password: req.Password,
	}

	userReg, err := s.userRepo.RegisterUser(ctx, nil, user)
	if err != nil {
		return dto.UserResponse{}, dto.ErrCreateUser
	}

	return dto.UserResponse{
		ID:    userReg.ID.String(),
		Name:  userReg.Name,
		Role:  userReg.Role,
		Email: userReg.Email,
	}, nil
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepo.GetUserById(ctx, nil, userId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserById
	}

	return dto.UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Role:  user.Role,
		Email: user.Email,
	}, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error) {
	emails, err := s.userRepo.GetUserByEmail(ctx, nil, email)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserByEmail
	}

	return dto.UserResponse{
		ID:    emails.ID.String(),
		Name:  emails.Name,
		Role:  emails.Role,
		Email: emails.Email,
	}, nil
}

func (s *userService) Verify(ctx context.Context, req dto.UserLoginRequest) (dto.UserLoginResponse, error) {
	check, flag, err := s.userRepo.CheckEmail(ctx, nil, req.Email)
	if err != nil || !flag {
		return dto.UserLoginResponse{}, dto.ErrEmailNotFound
	}

	checkPassword, err := utils.PasswordCompare(check.Password, []byte(req.Password))
	if err != nil || !checkPassword {
		return dto.UserLoginResponse{}, dto.ErrPasswordNotMatch
	}

	token := s.jwtService.GenerateToken(check.ID.String(), check.Role)

	return dto.UserLoginResponse{
		Token: token,
		Role:  check.Role,
	}, nil
}
