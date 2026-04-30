package game

import (
	"errors"
	"fmt"
	"ming/internal/config"
	"ming/internal/handler/hp"
	"ming/sdk/xlog"
	"net"
	"net/http"
)

// 配置http服务
func (g *Game) configureHttpServer() {
	g.httpServer = &http.Server{
		Addr: config.ApiPort(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// websocket
			if r.URL.Path == config.WebsocketPath() {
				if err := g.wsServer.HandleRequest(w, r); err != nil {
					xlog.Error().Err(err).Msg("Failed to handle WebSocket request")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			//hello测试
			if r.URL.Path == "/hello" {
				w.Write([]byte("hello"))
				return
			} else if r.URL.Path == "/login" {
				hp.LoginHandler(w, r)
				return
			} else if r.URL.Path == "/register" {
				hp.RegisterHandler(w, r)
				return
			}
		}),
	}
}

// 启动http服务器
func (g *Game) startHttpServer() chan error {
	serverStarted := make(chan error, 1)
	go func() {
		xlog.Info().Msgf("http server started on %s", config.ApiPort())
		listener, err := net.Listen("tcp", config.ApiPort())
		if err != nil {
			serverStarted <- fmt.Errorf("http server failed to listen on %s: %w", config.ApiPort(), err)
			return
		}
		// 通知服务器已成功启动
		serverStarted <- nil
		// Serve 会一直在协程内阻塞，直到服务器关闭
		if err := g.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverStarted <- fmt.Errorf("server start failed: %w", err)
			return
		}
	}()
	return serverStarted
}
