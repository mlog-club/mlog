package api

import (
	"math/rand"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple"

	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/services"
	"bbs-go/services/cache"
)

type TopicController struct {
	Ctx iris.Context
}

// // 同步帖子相关计数
// func (this *TopicController) GetSynccount() *simple.JsonResult {
// 	go func() {
// 		services.TopicService.Scan(func(topics []model.Topic) bool {
// 			for _, topic := range topics {
// 				commentCount := services.CommentService.Count(model.EntityTypeTopic, topic.Id)
// 				likeCount := services.TopicLikeService.Count(topic.Id)
// 				_ = services.TopicService.Updates(topic.Id, map[string]interface{}{
// 					"comment_count": commentCount,
// 					"like_count":    likeCount,
// 				})
// 			}
// 			return true
// 		})
// 	}()
// 	return simple.JsonSuccess()
// }

// 节点
func (this *TopicController) GetNodes() *simple.JsonResult {
	nodes := services.TopicNodeService.GetNodes()
	return simple.JsonData(render.BuildNodes(nodes))
}

// 节点信息
func (this *TopicController) GetNode() *simple.JsonResult {
	nodeId := simple.FormValueInt64Default(this.Ctx, "nodeId", 0)
	node := services.TopicNodeService.Get(nodeId)
	return simple.JsonData(render.BuildNode(node))
}

// 发表帖子
func (this *TopicController) PostCreate() *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}
	nodeId := simple.FormValueInt64Default(this.Ctx, "nodeId", 0)
	title := strings.TrimSpace(simple.FormValue(this.Ctx, "title"))
	content := strings.TrimSpace(simple.FormValue(this.Ctx, "content"))
	tags := simple.FormValueStringArray(this.Ctx, "tags")
	topic, err := services.TopicService.Publish(user.Id, nodeId, tags, title, content)
	if err != nil {
		return simple.JsonError(err)
	}
	return simple.JsonData(render.BuildSimpleTopic(topic))
}

// 编辑时获取详情
func (this *TopicController) GetEditBy(topicId int64) *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}

	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != model.StatusOk {
		return simple.JsonErrorMsg("话题不存在或已被删除")
	}
	if topic.UserId != user.Id {
		return simple.JsonErrorMsg("无权限")
	}

	tags := services.TopicService.GetTopicTags(topicId)
	var tagNames []string
	if len(tags) > 0 {
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
	}

	return simple.NewEmptyRspBuilder().
		Put("topicId", topic.Id).
		Put("nodeId", topic.NodeId).
		Put("title", topic.Title).
		Put("content", topic.Content).
		Put("tags", tagNames).
		JsonResult()
}

// 编辑帖子
func (this *TopicController) PostEditBy(topicId int64) *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}

	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != model.StatusOk {
		return simple.JsonErrorMsg("话题不存在或已被删除")
	}
	if topic.UserId != user.Id {
		return simple.JsonErrorMsg("无权限")
	}

	nodeId := simple.FormValueInt64Default(this.Ctx, "nodeId", 0)
	title := strings.TrimSpace(simple.FormValue(this.Ctx, "title"))
	content := strings.TrimSpace(simple.FormValue(this.Ctx, "content"))
	tags := simple.FormValueStringArray(this.Ctx, "tags")

	err := services.TopicService.Edit(topicId, nodeId, tags, title, content)
	if err != nil {
		return simple.JsonError(err)
	}
	return simple.JsonData(render.BuildSimpleTopic(topic))
}

// 删除帖子
func (this *TopicController) PostDeleteBy(topicId int64) *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}
	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != model.StatusOk {
		return simple.JsonSuccess()
	}
	if topic.UserId != user.Id {
		return simple.JsonErrorMsg("无权限")
	}
	err := services.TopicService.Delete(topicId)
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	return simple.JsonSuccess()
}

// 帖子详情
func (this *TopicController) GetBy(topicId int64) *simple.JsonResult {
	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != model.StatusOk {
		return simple.JsonErrorMsg("主题不存在")
	}
	services.TopicService.IncrViewCount(topicId) // 增加浏览量
	return simple.JsonData(render.BuildTopic(topic))
}

// 点赞
func (this *TopicController) GetLikeBy(topicId int64) *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}
	err := services.TopicLikeService.Like(user.Id, topicId)
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	return simple.JsonSuccess()
}

// 最新帖子
func (this *TopicController) GetRecent() *simple.JsonResult {
	topics := services.TopicService.Find(simple.NewSqlCnd().Where("status = ?", model.StatusOk).Desc("id").Limit(10))
	return simple.JsonData(render.BuildSimpleTopics(topics))
}

// 用户最近的帖子
func (this *TopicController) GetUserRecent() *simple.JsonResult {
	userId, err := simple.FormValueInt64(this.Ctx, "userId")
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	topics := services.TopicService.Find(simple.NewSqlCnd().Where("user_id = ? and status = ?",
		userId, model.StatusOk).Desc("id").Limit(10))
	return simple.JsonData(render.BuildSimpleTopics(topics))
}

// 用户帖子列表
func (this *TopicController) GetUserTopics() *simple.JsonResult {
	userId, err := simple.FormValueInt64(this.Ctx, "userId")
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	page := simple.FormValueIntDefault(this.Ctx, "page", 1)

	topics, paging := services.TopicService.FindPageByCnd(simple.NewSqlCnd().
		Eq("user_id", userId).
		Eq("status", model.StatusOk).
		Page(page, 20).Desc("id"))

	return simple.JsonPageData(render.BuildSimpleTopics(topics), paging)
}

// 帖子列表
func (this *TopicController) GetTopics() *simple.JsonResult {
	page := simple.FormValueIntDefault(this.Ctx, "page", 1)

	topics, paging := services.TopicService.FindPageByCnd(simple.NewSqlCnd().
		Eq("status", model.StatusOk).
		Page(page, 20).Desc("last_comment_time"))

	return simple.JsonPageData(render.BuildSimpleTopics(topics), paging)
}

// 节点帖子列表
func (this *TopicController) GetNodeTopics() *simple.JsonResult {
	page := simple.FormValueIntDefault(this.Ctx, "page", 1)
	nodeId := simple.FormValueInt64Default(this.Ctx, "nodeId", 0)
	topics, paging := services.TopicService.FindPageByCnd(simple.NewSqlCnd().
		Eq("node_id", nodeId).
		Eq("status", model.StatusOk).
		Page(page, 20).Desc("last_comment_time"))

	return simple.JsonPageData(render.BuildSimpleTopics(topics), paging)
}

// 标签帖子列表
func (this *TopicController) GetTagTopics() *simple.JsonResult {
	page := simple.FormValueIntDefault(this.Ctx, "page", 1)
	tagId, err := simple.FormValueInt64(this.Ctx, "tagId")
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	topics, paging := services.TopicService.GetTagTopics(tagId, page)
	return simple.JsonPageData(render.BuildSimpleTopics(topics), paging)
}

// 推荐帖子
func (this *TopicController) GetRecommendTopics() *simple.JsonResult {
	page := simple.FormValueIntDefault(this.Ctx, "page", 1)
	topics, paging := services.TopicService.FindPageByCnd(simple.NewSqlCnd().
		Eq("recommend", true).
		Eq("status", model.StatusOk).
		Page(page, 20).Desc("last_comment_time"))

	return simple.JsonPageData(render.BuildSimpleTopics(topics), paging)
}

// 收藏
func (this *TopicController) GetFavoriteBy(topicId int64) *simple.JsonResult {
	user := services.UserTokenService.GetCurrent(this.Ctx)
	if user == nil {
		return simple.JsonError(simple.ErrorNotLogin)
	}
	err := services.FavoriteService.AddTopicFavorite(user.Id, topicId)
	if err != nil {
		return simple.JsonErrorMsg(err.Error())
	}
	return simple.JsonSuccess()
}

// 推荐
func (this *TopicController) GetRecommend() *simple.JsonResult {
	topics := cache.TopicCache.GetRecommendTopics()
	if topics == nil || len(topics) == 0 {
		return simple.JsonSuccess()
	} else {
		dest := make([]model.Topic, len(topics))
		perm := rand.Perm(len(topics))
		for i, v := range perm {
			dest[v] = topics[i]
		}
		end := 10
		if end > len(topics) {
			end = len(topics)
		}
		ret := dest[0:end]
		return simple.JsonData(render.BuildSimpleTopics(ret))
	}
}

// 最新话题
func (this *TopicController) GetNewest() *simple.JsonResult {
	topics := services.TopicService.Find(simple.NewSqlCnd().Eq("status", model.StatusOk).Desc("id").Limit(6))
	return simple.JsonData(render.BuildSimpleTopics(topics))
}
