package services

import (
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/google/uuid"
)

type AddressService struct {
	Repo repository.PgSQLRepository
}

func SetupAddressService (repo repository.PgSQLRepository) *AddressService {
	return &AddressService{
		Repo: repo,
	}
}

func (s *AddressService) CreateAddress(address *models.UserAddress) error {

	if address == nil {
		return apperror.New(constant.BADREQUEST, "Address data is nil", nil)
	}

	if err := s.Repo.Insert(address); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to create address for user", err)
	}

	return nil
}

func (s *AddressService) GetUserAddresses(userID string) ([]models.UserAddress, error) {

	uId, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}
	
	var addresses []models.UserAddress
	if err := s.Repo.FindAllWhere(&addresses, "user_id = ?", uId); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch addresses", err)
	}

	return addresses, nil
}

func (s *AddressService) UpdateAddress(addressID string, fields map[string]interface{}) error {

	aId, err := uuid.Parse(addressID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	var address models.UserAddress
	if err := s.Repo.FindById(&address, aId); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Address not found",
			err,
		)
	}

	if err := s.Repo.UpdateByFields(&models.UserAddress{}, aId, fields); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to update address",
			err,
		)
	}

	return nil
}

func (s *AddressService) DeleteAddress(addressID string) error {

	aId, err := uuid.Parse(addressID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	var address models.UserAddress
	if err := s.Repo.FindById(&address, aId); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Address not found",
			err,
		)
	}

	if err := s.Repo.Delete(&address, aId); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to delete address",
			err,
		)
	}

	return nil
}