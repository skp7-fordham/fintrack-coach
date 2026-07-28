package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

type categoryRepository interface {
	CreateCategory(ctx context.Context, input domain.CreateCategoryInput) (*domain.Category, error)
	ListCategories(ctx context.Context, filter domain.ListCategoriesFilter) ([]domain.Category, error)
	UpdateCategory(ctx context.Context, input domain.UpdateCategoryInput) (*domain.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID string) error
}

type CategoryService struct {
	repo categoryRepository
}

func NewCategoryService(repo categoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(
	ctx context.Context,
	userID string,
	req dto.CreateCategoryRequest,
) (*domain.Category, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	name, err := validateCategoryName(req.Name)
	if err != nil {
		return nil, err
	}
	categoryType, err := validateCategoryType(req.CategoryType)
	if err != nil {
		return nil, err
	}
	color, err := validateOptionalColor(req.Color)
	if err != nil {
		return nil, err
	}
	icon, err := validateOptionalIcon(req.Icon)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateCategory(ctx, domain.CreateCategoryInput{
		UserID:       userID,
		Name:         name,
		CategoryType: categoryType,
		Color:        color,
		Icon:         icon,
	})
}

func (s *CategoryService) ListCategories(
	ctx context.Context,
	userID string,
	typeFilter string,
) ([]domain.Category, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	filter := domain.ListCategoriesFilter{UserID: userID}
	if typeFilter = strings.TrimSpace(typeFilter); typeFilter != "" {
		categoryType, err := validateCategoryType(typeFilter)
		if err != nil {
			return nil, err
		}
		filter.CategoryType = &categoryType
	}

	return s.repo.ListCategories(ctx, filter)
}

func (s *CategoryService) UpdateCategory(
	ctx context.Context,
	userID, categoryID string,
	req dto.UpdateCategoryRequest,
) (*domain.Category, error) {
	userID = strings.TrimSpace(userID)
	categoryID = strings.TrimSpace(categoryID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(categoryID) {
		return nil, &domain.ValidationError{Message: "category id must be a valid UUID"}
	}

	if req.Name == nil && req.CategoryType == nil && req.Color == nil && req.Icon == nil {
		return nil, &domain.ValidationError{Message: "at least one field is required"}
	}

	input := domain.UpdateCategoryInput{
		UserID:     userID,
		CategoryID: categoryID,
	}

	if req.Name != nil {
		name, err := validateCategoryName(*req.Name)
		if err != nil {
			return nil, err
		}
		input.Name = &name
	}
	if req.CategoryType != nil {
		categoryType, err := validateCategoryType(*req.CategoryType)
		if err != nil {
			return nil, err
		}
		input.CategoryType = &categoryType
	}
	if req.Color != nil {
		color, err := validateOptionalColor(req.Color)
		if err != nil {
			return nil, err
		}
		input.Color = color
	}
	if req.Icon != nil {
		icon, err := validateOptionalIcon(req.Icon)
		if err != nil {
			return nil, err
		}
		input.Icon = icon
	}

	return s.repo.UpdateCategory(ctx, input)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	userID = strings.TrimSpace(userID)
	categoryID = strings.TrimSpace(categoryID)
	if !isValidUUID(userID) {
		return &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(categoryID) {
		return &domain.ValidationError{Message: "category id must be a valid UUID"}
	}
	return s.repo.DeleteCategory(ctx, userID, categoryID)
}

func validateCategoryName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", &domain.ValidationError{Message: "name is required"}
	}
	if utf8.RuneCountInString(name) > maxCategoryNameLength {
		return "", &domain.ValidationError{Message: "name must be at most 100 characters"}
	}
	return name, nil
}

func validateCategoryType(raw string) (string, error) {
	categoryType := strings.TrimSpace(raw)
	switch categoryType {
	case "income", "expense":
		return categoryType, nil
	default:
		return "", &domain.ValidationError{Message: "category_type must be income or expense"}
	}
}

func validateOptionalColor(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	color := strings.TrimSpace(*raw)
	if color == "" {
		return nil, &domain.ValidationError{Message: "color must use #RRGGBB format"}
	}
	if !hexColorPattern.MatchString(color) {
		return nil, &domain.ValidationError{Message: "color must use #RRGGBB format"}
	}
	normalized := strings.ToUpper(color)
	return &normalized, nil
}

func validateOptionalIcon(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	icon := strings.TrimSpace(*raw)
	if icon == "" {
		return nil, &domain.ValidationError{Message: "icon must not be empty"}
	}
	if utf8.RuneCountInString(icon) > maxCategoryIconLength {
		return nil, &domain.ValidationError{Message: "icon must be at most 50 characters"}
	}
	return &icon, nil
}
