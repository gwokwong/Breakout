package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/crypto/bcrypt"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// User 用户控制器
var User = &userController{}

type userController struct{}

// Login 用户登录
func (c *userController) Login(r *ghttp.Request) {
	var req model.LoginReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	var user model.User
	if err := service.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "手机号或密码错误",
			"data":    nil,
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "手机号或密码错误",
			"data":    nil,
		})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "账号已被禁用",
			"data":    nil,
		})
		return
	}

	// 生成令牌
	token, err := service.GenerateToken(user.ID, user.TeamID, user.Role)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "生成令牌失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "登录成功",
		"data": model.UserInfoResp{
			ID:       user.ID,
			TeamID:   user.TeamID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Phone:    user.Phone,
			Role:     user.Role,
			Token:    token,
		},
	})
}

// Register 用户注册
func (c *userController) Register(r *ghttp.Request) {
	var req model.RegisterReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查手机号是否已存在
	var count int64
	service.DB.Model(&model.User{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "该手机号已被注册",
			"data":    nil,
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "密码加密失败",
			"data":    nil,
		})
		return
	}

	// 创建用户
	user := model.User{
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Role:     1, // 默认创建者为所有者角色
		Status:   1,
	}

	if err := service.DB.Create(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "注册失败",
			"data":    nil,
		})
		return
	}

	// 创建默认团队
	team := model.Team{
		Name:    req.Nickname + "的团队",
		OwnerID: user.ID,
	}
	if err := service.DB.Create(&team).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "创建团队失败",
			"data":    nil,
		})
		return
	}

	// 更新用户团队ID
	user.TeamID = team.ID
	service.DB.Save(&user)

	// 生成令牌
	token, err := service.GenerateToken(user.ID, user.TeamID, user.Role)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "生成令牌失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "注册成功",
		"data": model.UserInfoResp{
			ID:       user.ID,
			TeamID:   user.TeamID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Phone:    user.Phone,
			Role:     user.Role,
			Token:    token,
		},
	})
}

// WxLogin 微信登录
func (c *userController) WxLogin(r *ghttp.Request) {
	var req model.WxLoginReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 调用微信 code2session 接口
	appid := g.Cfg().MustGet(r.Context(), "wechat.appid").String()
	secret := g.Cfg().MustGet(r.Context(), "wechat.secret").String()

	code2sessionURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appid, secret, req.Code,
	)

	resp, err := http.Get(code2sessionURL)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "微信接口调用失败",
			"data":    nil,
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "读取微信响应失败",
			"data":    nil,
		})
		return
	}

	var wxResult struct {
		Openid     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &wxResult); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "解析微信响应失败",
			"data":    nil,
		})
		return
	}

	if wxResult.ErrCode != 0 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "微信登录失败: " + wxResult.ErrMsg,
			"data":    nil,
		})
		return
	}

	// 根据 openid 查找或创建用户
	var user model.User
	if err := service.DB.Where("wx_openid = ?", wxResult.Openid).First(&user).Error; err != nil {
		// 新用户：创建默认用户和团队
		nickname := "微信用户"
		user = model.User{
			Nickname: nickname,
			Role:     1,
			Status:   1,
		}
		user.WxOpenid = &wxResult.Openid

		if err := service.DB.Create(&user).Error; err != nil {
			r.Response.WriteJson(g.Map{
				"code":    http.StatusInternalServerError,
				"message": "创建用户失败",
				"data":    nil,
			})
			return
		}

		// 创建默认团队
		team := model.Team{
			Name:    nickname + "的团队",
			OwnerID: user.ID,
		}
		if err := service.DB.Create(&team).Error; err != nil {
			r.Response.WriteJson(g.Map{
				"code":    http.StatusInternalServerError,
				"message": "创建团队失败",
				"data":    nil,
			})
			return
		}

		user.TeamID = team.ID
		service.DB.Save(&user)
	}

	// 检查用户状态
	if user.Status != 1 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "账号已被禁用",
			"data":    nil,
		})
		return
	}

	// 生成令牌
	token, err := service.GenerateToken(user.ID, user.TeamID, user.Role)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "生成令牌失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "登录成功",
		"data": model.UserInfoResp{
			ID:       user.ID,
			TeamID:   user.TeamID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Phone:    user.Phone,
			Role:     user.Role,
			Token:    token,
		},
	})
}

// Info 获取用户信息
func (c *userController) Info(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var user model.User
	if err := service.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取用户信息失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": model.UserInfoResp{
			ID:       user.ID,
			TeamID:   user.TeamID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Phone:    user.Phone,
			Role:     user.Role,
		},
	})
}

// UpdateProfile 更新用户信息
func (c *userController) UpdateProfile(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var data map[string]interface{}
	if err := r.Parse(&data); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 不允许修改的字段
	delete(data, "id")
	delete(data, "phone")
	delete(data, "password")
	delete(data, "role")
	delete(data, "status")
	delete(data, "team_id")
	delete(data, "wx_openid")

	if err := service.DB.Model(&model.User{}).Where("id = ?", userID).Updates(data).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "更新成功",
		"data":    nil,
	})
}