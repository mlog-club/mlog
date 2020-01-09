package services

import (
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple"

	"bbs-go/model"
	"bbs-go/repositories"
	"bbs-go/services/cache"
)

var UserTokenService = newUserTokenService()

func newUserTokenService() *userTokenService {
	return &userTokenService{}
}

type userTokenService struct {
}

func (this *userTokenService) Get(id int64) *model.UserToken {
	return repositories.UserTokenRepository.Get(simple.DB(), id)
}

func (this *userTokenService) Take(where ...interface{}) *model.UserToken {
	return repositories.UserTokenRepository.Take(simple.DB(), where...)
}

func (this *userTokenService) Find(cnd *simple.SqlCnd) []model.UserToken {
	return repositories.UserTokenRepository.Find(simple.DB(), cnd)
}

func (this *userTokenService) FindOne(cnd *simple.SqlCnd) *model.UserToken {
	return repositories.UserTokenRepository.FindOne(simple.DB(), cnd)
}

func (this *userTokenService) FindPageByParams(params *simple.QueryParams) (list []model.UserToken, paging *simple.Paging) {
	return repositories.UserTokenRepository.FindPageByParams(simple.DB(), params)
}

func (this *userTokenService) FindPageByCnd(cnd *simple.SqlCnd) (list []model.UserToken, paging *simple.Paging) {
	return repositories.UserTokenRepository.FindPageByCnd(simple.DB(), cnd)
}

// 获取当前登录用户的id
func (this *userTokenService) GetCurrentUserId(ctx iris.Context) int64 {
	user := this.GetCurrent(ctx)
	if user != nil {
		return user.Id
	}
	return 0
}

// 获取当前登录用户
func (this *userTokenService) GetCurrent(ctx iris.Context) *model.User {
	token := this.GetUserToken(ctx)
	userToken := cache.UserTokenCache.Get(token)
	// 没找到授权
	if userToken == nil || userToken.Status == model.StatusDeleted {
		return nil
	}
	// 授权过期
	if userToken.ExpiredAt <= simple.NowTimestamp() {
		return nil
	}
	user := cache.UserCache.Get(userToken.UserId)
	if user == nil || user.Status != model.StatusOk {
		return nil
	}
	return user
}

// 退出登录
func (this *userTokenService) Signout(ctx iris.Context) error {
	token := this.GetUserToken(ctx)
	userToken := repositories.UserTokenRepository.GetByToken(simple.DB(), token)
	if userToken == nil {
		return nil
	}
	return repositories.UserTokenRepository.UpdateColumn(simple.DB(), userToken.Id, "status", model.StatusDeleted)
}

// 从请求体中获取UserToken
func (this *userTokenService) GetUserToken(ctx iris.Context) string {
	userToken := ctx.FormValue("userToken")
	if len(userToken) > 0 {
		return userToken
	}
	return ctx.GetHeader("X-User-Token")
}

// 生成
func (this *userTokenService) Generate(userId int64) (string, error) {
	token := simple.Uuid()
	expiredAt := time.Now().Add(time.Hour * 24 * 7) // 7天后过期
	userToken := &model.UserToken{
		Token:      token,
		UserId:     userId,
		ExpiredAt:  simple.Timestamp(expiredAt),
		Status:     model.StatusOk,
		CreateTime: simple.NowTimestamp(),
	}
	err := repositories.UserTokenRepository.Create(simple.DB(), userToken)
	if err != nil {
		return "", err
	}
	return token, nil
}

// 禁用
func (this *userTokenService) Disable(token string) error {
	t := repositories.UserTokenRepository.GetByToken(simple.DB(), token)
	if t == nil {
		return nil
	}
	err := repositories.UserTokenRepository.UpdateColumn(simple.DB(), t.Id, "status", model.StatusDeleted)
	if err != nil {
		cache.UserTokenCache.Invalidate(token)
	}
	return err
}
