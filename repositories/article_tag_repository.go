package repositories

import (
	"github.com/jinzhu/gorm"
	"github.com/mlogclub/mlog/model"
	"github.com/mlogclub/simple"
)

type ArticleTagRepository struct {
}

func NewArticleTagRepository() *ArticleTagRepository {
	return &ArticleTagRepository{}
}

func (this *ArticleTagRepository) Get(db *gorm.DB, id int64) *model.ArticleTag {
	ret := &model.ArticleTag{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (this *ArticleTagRepository) Take(db *gorm.DB, where ...interface{}) *model.ArticleTag {
	ret := &model.ArticleTag{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (this *ArticleTagRepository) QueryCnd(db *gorm.DB, cnd *simple.QueryCnd) (list []model.ArticleTag, err error) {
	err = cnd.DoQuery(db).Find(&list).Error
	return
}

func (this *ArticleTagRepository) Query(db *gorm.DB, queries *simple.ParamQueries) (list []model.ArticleTag, paging *simple.Paging) {
	queries.StartQuery(db).Find(&list)
	queries.StartCount(db).Model(&model.ArticleTag{}).Count(&queries.Paging.Total)
	paging = queries.Paging
	return
}

func (this *ArticleTagRepository) Create(db *gorm.DB, t *model.ArticleTag) (err error) {
	err = db.Create(t).Error
	return
}

func (this *ArticleTagRepository) Update(db *gorm.DB, t *model.ArticleTag) (err error) {
	err = db.Save(t).Error
	return
}

func (this *ArticleTagRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.ArticleTag{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (this *ArticleTagRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.ArticleTag{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (this *ArticleTagRepository) Delete(db *gorm.DB, id int64) {
	db.Model(&model.ArticleTag{}).Delete("id", id)
}

func (this *ArticleTagRepository) GetUnique(db *gorm.DB, articleId, tagId int64) *model.ArticleTag {
	ret := &model.ArticleTag{}
	if err := db.First(ret, "article_id = ? and tag_id = ?", articleId, tagId).Error; err != nil {
		return nil
	}
	return ret
}

func (this *ArticleTagRepository) GetByArticleId(db *gorm.DB, articleId int64) ([]model.ArticleTag, error) {
	return this.QueryCnd(db, simple.NewQueryCnd("article_id = ?", articleId))
}
