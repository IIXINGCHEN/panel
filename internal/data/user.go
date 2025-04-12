package data

import (
	"errors"

	"github.com/go-rat/utils/hash"
	"github.com/leonelquinteros/gotext"
	"gorm.io/gorm"

	"github.com/tnb-labs/panel/internal/biz"
)

type userRepo struct {
	t      *gotext.Locale
	db     *gorm.DB
	hasher hash.Hasher
}

func NewUserRepo(t *gotext.Locale, db *gorm.DB) biz.UserRepo {
	return &userRepo{
		t:      t,
		db:     db,
		hasher: hash.NewArgon2id(),
	}
}

func (r *userRepo) Create(username, password string) (*biz.User, error) {
	value, err := r.hasher.Make(password)
	if err != nil {
		return nil, err
	}

	user := &biz.User{
		Username: username,
		Password: value,
	}
	if err = r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) CheckPassword(username, password string) (*biz.User, error) {
	user := new(biz.User)
	if err := r.db.Where("username = ?", username).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(r.t.Get("username or password error"))
		} else {
			return nil, err
		}
	}

	if !r.hasher.Check(password, user.Password) {
		return nil, errors.New(r.t.Get("username or password error"))
	}

	return user, nil
}

func (r *userRepo) Get(id uint) (*biz.User, error) {
	user := new(biz.User)
	if err := r.db.First(user, id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) Save(user *biz.User) error {
	return r.db.Save(user).Error
}
