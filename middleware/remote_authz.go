package middleware

import (
	"net/http"
	"pic-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RemoteAuthz() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := "http://" + viper.GetString("auth.host") + ":" + viper.GetString("auth.port") + "/api/v1/verify/authz"

		req, err := http.NewRequest(http.MethodPost, url, nil)

		if err != nil {
			utils.SendJsonResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort()
			return
		}
		// 复制 Authorization 请求头
		authorization := c.GetHeader("Authorization")
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			utils.SendJsonResponse(c, http.StatusUnauthorized, "Unauthorized; request failed", nil)
			c.Abort()
			return
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			utils.SendJsonResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
