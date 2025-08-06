package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

    "github.com/iliyamo/Library-Management-System/internal/utils"
	"github.com/labstack/echo/v4"
)

type bucketState struct {
	Tokens     float64 `json:"tokens"`
	LastRefill int64   `json:"last_refill"`
}

const (
	bucketCapacity     = 20
	tokenRefillRate    = 1.0 / 3.0 // 1 token every 3 seconds
	tokenRefillSeconds = 3
)

// RateLimit محدودسازی درخواست‌ها با الگوریتم Token Bucket برای هر کاربر
func RateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDRaw := c.Get("user_id")
		if userIDRaw == nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "شناسه کاربر یافت نشد"})
		}
		userID := fmt.Sprintf("%v", userIDRaw)
		redisKey := fmt.Sprintf("rate:user:%s", userID)

		now := time.Now().Unix()
		state := bucketState{
			Tokens:     bucketCapacity,
			LastRefill: now,
		}

		cached, err := utils.GetCache(redisKey)
		if err == nil {
			_ = json.Unmarshal([]byte(cached), &state)

			duration := now - state.LastRefill
			newTokens := float64(duration) * tokenRefillRate
			state.Tokens = min(bucketCapacity, state.Tokens+newTokens)
			state.LastRefill = now
		}

		if state.Tokens < 1 {
			expire := int(tokenRefillSeconds * float64(bucketCapacity))
			_ = utils.SetCache(redisKey, cached, time.Duration(expire)*time.Second)
			return c.JSON(http.StatusTooManyRequests, echo.Map{"error": "تعداد درخواست بیش از حد مجاز است"})
		}

		state.Tokens -= 1
		data, _ := json.Marshal(state)
		_ = utils.SetCache(redisKey, string(data), time.Duration(bucketCapacity*tokenRefillSeconds)*time.Second)

		return next(c)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
