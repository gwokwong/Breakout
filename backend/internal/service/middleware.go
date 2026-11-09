package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware 中间件实例
var Middleware = &middlewareService{}

type middlewareService struct{}

// InitMiddleware 初始化中间件
func InitMiddleware() {
	// 初始化JWT配置
	jwtSecret = []byte(g.Cfg().MustGet(context.Background(), "jwt.secret").String())
	jwtExpire = g.Cfg().MustGet(context.Background(), "jwt.expire").Int()
}

var (
	jwtSecret []byte
	jwtExpire int
)

// JwtClaims JWT自定义声明
type JwtClaims struct {
	UserID uint64 `json:"user_id"`
	TeamID uint64 `json:"team_id"`
	Role   int    `json:"role"`
	jwt.RegisteredClaims
}

// Auth JWT认证中间件
func (s *middlewareService) Auth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
			"data":    nil,
		})
		r.Exit()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusUnauthorized,
			"message": "授权格式错误",
			"data":    nil,
		})
		r.Exit()
		return
	}

	tokenString := parts[1]
	claims := &JwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusUnauthorized,
			"message": "授权已过期或无效",
			"data":    nil,
		})
		r.Exit()
		return
	}

	// 将用户信息存入上下文
	r.SetCtxVar("user_id", claims.UserID)
	r.SetCtxVar("team_id", claims.TeamID)
	r.SetCtxVar("user_role", claims.Role)

	r.Middleware.Next()
}

// ResponseHandler 响应处理中间件
func (s *middlewareService) ResponseHandler(r *ghttp.Request) {
	r.Middleware.Next()

	// 如果已经有响应内容，不处理
	if r.Response.BufferLength() > 0 {
		return
	}

	// 统一响应格式
	err := r.GetError()
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	data := r.GetHandlerResponse()
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    data,
	})
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID, teamID uint64, role int) (string, error) {
	claims := JwtClaims{
		UserID: userID,
		TeamID: teamID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "breakout",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}