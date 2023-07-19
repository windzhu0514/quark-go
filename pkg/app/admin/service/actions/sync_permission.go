package actions

import (
	"strings"

	"github.com/gobeam/stringy"
	"github.com/quarkcms/quark-go/v2/pkg/app/admin/component/message"
	"github.com/quarkcms/quark-go/v2/pkg/app/admin/model"
	"github.com/quarkcms/quark-go/v2/pkg/app/admin/template/resource/actions"
	"github.com/quarkcms/quark-go/v2/pkg/builder"
	"github.com/quarkcms/quark-go/v2/pkg/dal/db"
	"gorm.io/gorm"
)

type SyncPermission struct {
	actions.Action
}

// 初始化
func (p *SyncPermission) Init() *SyncPermission {
	// 初始化父结构
	p.ParentInit()

	// 行为名称
	p.Name = "同步权限"

	// 执行成功后刷新的组件
	p.Reload = "table"

	// 是否具有loading，当action 的作用类型为ajax,submit时有效
	p.WithLoading = true

	// 设置展示位置
	p.SetOnlyOnIndex(true)

	// 行为类型
	p.ActionType = "ajax"

	return p
}

// 执行行为句柄
func (p *SyncPermission) Handle(ctx *builder.Context, query *gorm.DB) error {
	permissions := ctx.Engine.GetUrlPaths()
	data := []model.Permission{}

	var names []string
	var currentNames []string
	db.Client.Model(&model.Permission{}).Pluck("name", &names)
	for _, v := range permissions {
		if strings.Contains(v.Url, "/api/admin") {
			has := false
			url := strings.ReplaceAll(v.Url, "/api/admin/", "")
			url = strings.ReplaceAll(url, "/", "_") + "_" + strings.ToLower(v.Method)
			name := stringy.
				New(url).
				CamelCase("?", "")
			currentNames = append(currentNames, name)
			for _, nv := range names {
				if nv == name {
					has = true
				}
			}
			if has == false {
				permission := model.Permission{
					Name:      name,
					Method:    v.Method,
					Path:      v.Url,
					GuardName: "admin",
				}
				data = append(data, permission)
			}
		}
	}
	if len(data) == 0 {
		return ctx.JSON(200, message.Error("暂无新增权限！"))
	}

	err := query.Create(data).Error
	if err != nil {
		return ctx.JSON(200, message.Error(err.Error()))
	}

	err = db.Client.Model(&model.Permission{}).Where("name NOT IN ?", currentNames).Delete("").Error
	if err != nil {
		return ctx.JSON(200, message.Error(err.Error()))
	}

	return ctx.JSON(200, message.Success("操作成功"))
}