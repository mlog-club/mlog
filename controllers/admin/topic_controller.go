package admin

import (
	"github.com/kataras/iris"
	"github.com/mlogclub/mlog/controllers/render"
	"github.com/mlogclub/mlog/model"
	"github.com/mlogclub/mlog/services"
	"github.com/mlogclub/simple"
	"strconv"
)

type TopicController struct {
	Ctx          iris.Context
	TopicService *services.TopicService
}

func (this *TopicController) GetBy(id int64) *simple.JsonResult {
	t := this.TopicService.Get(id)
	if t == nil {
		return simple.ErrorMsg("Not found, id=" + strconv.FormatInt(id, 10))
	}
	return simple.JsonData(t)
}

func (this *TopicController) AnyList() *simple.JsonResult {
	list, paging := this.TopicService.Query(simple.NewParamQueries(this.Ctx).PageAuto().Desc("id"))

	var results []map[string]interface{}
	for _, topic := range list {
		builder := simple.NewRspBuilderExcludes(topic, "content")

		// 用户
		builder = builder.Put("user", render.BuildUserDefaultIfNull(topic.UserId))

		// 简介
		mr := simple.Markdown(topic.Content)
		builder.Put("summary", mr.SummaryText)

		// 标签
		tags := this.TopicService.GetTopicTags(topic.Id)
		builder.Put("tags", render.BuildTags(tags))

		results = append(results, builder.Build())
	}

	return simple.JsonData(&simple.PageResult{Results: results, Page: paging})

}

func (this *TopicController) PostCreate() *simple.JsonResult {
	t := &model.Topic{}
	err := this.Ctx.ReadForm(t)
	if err != nil {
		return simple.ErrorMsg(err.Error())
	}

	err = this.TopicService.Create(t)
	if err != nil {
		return simple.ErrorMsg(err.Error())
	}
	return simple.JsonData(t)
}

func (this *TopicController) PostUpdate() *simple.JsonResult {
	id, err := simple.FormValueInt64(this.Ctx, "id")
	if err != nil {
		return simple.ErrorMsg(err.Error())
	}
	t := this.TopicService.Get(id)
	if t == nil {
		return simple.ErrorMsg("entity not found")
	}

	err = this.Ctx.ReadForm(t)
	if err != nil {
		return simple.ErrorMsg(err.Error())
	}

	err = this.TopicService.Update(t)
	if err != nil {
		return simple.ErrorMsg(err.Error())
	}
	return simple.JsonData(t)
}
