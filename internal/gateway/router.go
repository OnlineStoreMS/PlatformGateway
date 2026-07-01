package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"platformgateway/internal/config"
	"platformgateway/internal/jwt"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	validator := jwt.NewValidator(cfg.JWT.Secret)
	pimBase, _ := url.Parse(config.StripTrailingSlash(cfg.Upstreams.ProductCore))
	iamBase, _ := url.Parse(config.StripTrailingSlash(cfg.Upstreams.UserCore))

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors(cfg))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "platform-gateway",
			"upstreams": gin.H{
				"iam": iamBase.String(),
				"pim": pimBase.String(),
			},
		})
	})

	pimProxy := reverseProxy(pimBase)
	iamProxy := reverseProxy(iamBase)

	// Gin 同一层级只能有一个 catch-all；admin 走 PIM，其余 /api/v1/* 走 IAM（NoRoute）
	r.Any("/api/v1/admin/*path", pimAuth(cfg, validator), pimProxy)
	r.Any("/api/v1/auth/login", iamProxy)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			iamProxy(c)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
	})

	return r
}

func cors(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if cfg.CORS.Allows(origin) && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func pimAuth(cfg *config.Config, v *jwt.Validator) gin.HandlerFunc {
	if !cfg.JWT.ValidatePIM {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
			return
		}
		claims, err := v.Parse(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录已过期"})
			return
		}
		c.Request.Header.Set("X-Tenant-ID", formatUint(claims.TenantID))
		c.Request.Header.Set("X-User-ID", formatUint(claims.UserID))
		c.Next()
	}
}

func formatUint(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func reverseProxy(target *url.URL) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
