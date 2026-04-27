package handler

import (
	"fmt"
	"ming/internal/config"
	"ming/internal/game"
	"ming/sdk/consts"
	"strconv"
)

// EnsureAppGame 校验请求是否属于本应用（业务规则，由 handler 层统一把关）。
func EnsureAppGame(ctx game.Context) bool {
	code, err := strconv.ParseInt(config.Cfg.Application.GameCode, 10, 32)
	if err != nil {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "server GameCode config invalid")
		return false
	}
	if ctx.GameId() != int32(code) {
		ctx.ErrResp(consts.ErrorCode_RequestErr, fmt.Sprintf("gameId not match: %d, expect %d", ctx.GameId(), code))
		return false
	}
	return true
}
