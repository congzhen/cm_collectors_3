package main

import (
	"cm_collectors_server/core"
	"cm_collectors_server/internal/selfupdate"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func initUpdater() {
	host := core.Config.System.ServerHost
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	if host == "::" {
		host = "::1"
	}
	m, e := selfupdate.New(selfupdate.Hooks{
		ConfigURL: core.GetUpdateSoftConfigURL, Backup: backupBeforeUpdate,
		Protected: func() []string {
			paths := []string{core.Config.System.FilePath, core.Config.System.LogFilePath, core.Config.Sqlite3.Path, core.Config.AutoBackup.BackupPath, core.Config.Jwt.PrivateKeyPath, core.Config.Jwt.PublicKeyPath, "config.yaml"}
			result := []string{}
			for _, p := range paths {
				if p != "" {
					if a, e := filepath.Abs(p); e == nil {
						result = append(result, a)
					}
				}
			}
			return result
		},
		HealthURL: "http://" + net.JoinHostPort(host, strconv.Itoa(core.Config.System.Port)) + "/api/health",
		Shutdown: func() {
			// Maintenance has already drained all API work and background jobs.
			// Always exit after draining: the general shutdown handler may return on timeout.
			if server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := server.Shutdown(ctx); err != nil {
					core.LogErr(err)
					_ = server.Close()
				}
				cancel()
			}
			if db, e := core.DBS().DB(); e == nil {
				_ = db.Close()
			}
			if os.Getenv("CM_UPDATE_SUPERVISED") == "1" {
				os.Exit(selfupdate.UpdateExitCode)
			}
			os.Exit(0)
		},
	})
	if e != nil {
		core.LogErr(fmt.Errorf("初始化更新服务失败: %w", e))
		return
	}
	selfupdate.Default = m
}
func backupBeforeUpdate(dir string) error {
	if core.Config.System.Database == "mysql" {
		return fmt.Errorf("MySQL 自动备份尚未配置，已停止安装；请备份数据库后手动更新")
	}
	name := filepath.Join(dir, "database-"+strconv.FormatInt(core.TimeNow().UnixNano(), 10)+".sqlite")
	if e := core.DBS().Exec("VACUUM INTO '" + strings.ReplaceAll(filepath.ToSlash(name), "'", "''") + "'").Error; e != nil {
		return fmt.Errorf("数据库备份失败: %w", e)
	}
	in, e := os.Open("config.yaml")
	if e != nil {
		return fmt.Errorf("配置备份失败: %w", e)
	}
	defer in.Close()
	out, e := os.OpenFile(filepath.Join(dir, "config.yaml"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = io.Copy(out, in)
	if e == nil {
		e = out.Sync()
	}
	ce := out.Close()
	if e != nil {
		return e
	}
	return ce
}
