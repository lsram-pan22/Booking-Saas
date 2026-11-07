package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidMember = errors.New("invalid member data")
var memberPhone = regexp.MustCompile(`^(0[0-9]{8,9}|\+66[0-9]{8,9})$`)
var staffUsername = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,49}$`)

type AddMemberRepository interface {
	AddMember(context.Context, string, string, models.NewMember) (models.ShopMember, error)
}
type AddMemberUsecase struct{ Repo AddMemberRepository }

func NewAddMemberUsecase(repo AddMemberRepository) *AddMemberUsecase {
	return &AddMemberUsecase{Repo: repo}
}
func (u *AddMemberUsecase) Add(ctx context.Context, ownerID, shopID string, req models.AddMemberRequest) (models.ShopMember, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	address, err := mail.ParseAddress(req.Email)
	if !utils.ValidUserID(ownerID) || !utils.ValidUserID(shopID) || err != nil || address.Address != req.Email || len(req.Email) > 254 || (req.Mode != "new" && req.Mode != "existing") {
		return models.ShopMember{}, ErrInvalidMember
	}
	input := models.NewMember{Mode: req.Mode, Email: req.Email}
	if req.Mode == "new" {
		input.Username = strings.ToLower(strings.TrimSpace(req.Username))
		if !staffUsername.MatchString(input.Username) {
			return models.ShopMember{}, ErrInvalidMember
		}
		input.FullName = strings.TrimSpace(req.FullName)
		input.Phone = strings.NewReplacer(" ", "", "-", "", "\t", "").Replace(strings.TrimSpace(req.Phone))
		if input.FullName == "" || utf8.RuneCountInString(input.FullName) > 100 || !memberPhone.MatchString(input.Phone) || utf8.RuneCountInString(req.Password) < 12 || len(req.Password) > 72 {
			return models.ShopMember{}, ErrInvalidMember
		}
		input.PasswordHash, err = utils.HashPassword(req.Password)
		if err != nil {
			return models.ShopMember{}, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.AddMember(ctx, ownerID, shopID, input)
}
