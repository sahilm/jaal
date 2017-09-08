package ssh

import (
	"context"
	"net"

	"github.com/sahilm/jaal/jaal"
	"github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type contextKey struct {
	name string
}

var (
	ContextKeyUser = &contextKey{"user"}

	ContextKeyPassword = &contextKey{"password"}

	ContextKeySessionID = &contextKey{"session-id"}

	ContextKeyPermissions = &contextKey{"permissions"}

	ContextKeyClientVersion = &contextKey{"client-version"}

	ContextKeyServerVersion = &contextKey{"server-version"}

	ContextKeyLocalAddr = &contextKey{"local-addr"}

	ContextKeyRemoteAddr = &contextKey{"remote-addr"}
)

type Context interface {
	context.Context

	User() string

	Password() string

	SessionID() string

	ClientVersion() string

	ServerVersion() string

	RemoteAddr() net.Addr

	LocalAddr() net.Addr

	Permissions() *gossh.Permissions

	SetValue(key, value interface{})
}

type sshContext struct {
	context.Context
}

func newContext() (*sshContext, context.CancelFunc) {
	innerCtx, cancel := context.WithCancel(context.Background())
	ctx := &sshContext{innerCtx}
	perms := &gossh.Permissions{}
	ctx.SetValue(ContextKeyPermissions, perms)
	return ctx, cancel
}

func (ctx *sshContext) applyConnMetadata(cm gossh.ConnMetadata, logger *logrus.Logger) {
	if ctx.Value(ContextKeySessionID) != nil {
		return
	}
	sha, err := jaal.ShortSHA256(cm.SessionID())
	if err != nil {
		logrus.Fatal("error computing sha of correlation id", err)
	}

	ctx.SetValue(ContextKeySessionID, sha)
	ctx.SetValue(ContextKeyClientVersion, string(cm.ClientVersion()))
	ctx.SetValue(ContextKeyServerVersion, string(cm.ServerVersion()))
	ctx.SetValue(ContextKeyUser, cm.User())
	ctx.SetValue(ContextKeyLocalAddr, cm.LocalAddr())
	ctx.SetValue(ContextKeyRemoteAddr, cm.RemoteAddr())
}

func (ctx *sshContext) applyPassword(password string) {
	ctx.SetValue(ContextKeyPassword, password)
}

func (ctx *sshContext) SetValue(key, value interface{}) {
	ctx.Context = context.WithValue(ctx.Context, key, value)
}

func (ctx *sshContext) User() string {
	return ctx.Value(ContextKeyUser).(string)
}

func (ctx *sshContext) Password() string {
	return ctx.Value(ContextKeyPassword).(string)
}

func (ctx *sshContext) SessionID() string {
	return ctx.Value(ContextKeySessionID).(string)
}

func (ctx *sshContext) ClientVersion() string {
	return ctx.Value(ContextKeyClientVersion).(string)
}

func (ctx *sshContext) ServerVersion() string {
	return ctx.Value(ContextKeyServerVersion).(string)
}

func (ctx *sshContext) RemoteAddr() net.Addr {
	return ctx.Value(ContextKeyRemoteAddr).(net.Addr)
}

func (ctx *sshContext) LocalAddr() net.Addr {
	return ctx.Value(ContextKeyLocalAddr).(net.Addr)
}

func (ctx *sshContext) Permissions() *gossh.Permissions {
	return ctx.Value(ContextKeyPermissions).(*gossh.Permissions)
}
