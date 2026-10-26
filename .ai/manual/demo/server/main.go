// Isolated documentation host. Build from cm_collectors_server's module;
// run ONLY inside .ai/manual/.local/runtime. Never embeds production html/data.
package main

import (
 "cm_collectors_server/config"
 "cm_collectors_server/core"
 "cm_collectors_server/models"
 "cm_collectors_server/processors"
 "cm_collectors_server/routers"
 "embed"
 "log"
 "net"
 "os"
 "path/filepath"
 "strings"
 "github.com/gin-gonic/gin"
 "gopkg.in/yaml.v3"
)

//go:embed html
var html embed.FS

func main() {
 cwd, _ := os.Getwd()
 if !strings.HasSuffix(filepath.ToSlash(cwd), "/.ai/manual/.local/runtime") { log.Fatal("manual runtime directory required") }
 // Validate the fresh manual configuration BEFORE core.Init opens any DB.
 raw, err := os.ReadFile("config.yaml")
 if err != nil { log.Fatal(err) }
 var conf config.Config
 if err := yaml.Unmarshal(raw, &conf); err != nil { log.Fatal(err) }
 if conf.System.Database != "sqlite3" || conf.System.Port != 18345 || conf.System.ServerHost != "127.0.0.1" { log.Fatal("isolated sqlite configuration required") }
 within := func(p string) bool {
  if p == "" { return true } // Missing optional fields receive safe relative defaults.
  absolute, err := filepath.Abs(p); if err != nil { return false }
  rel, err := filepath.Rel(cwd, absolute)
  return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
 }
 for _, p := range []string{conf.Sqlite3.Path, conf.System.FilePath, conf.System.LogFilePath, conf.Jwt.PrivateKeyPath, conf.Jwt.PublicKeyPath, conf.AutoBackup.BackupPath, conf.PerformerAvatarLibrary.CachePath, conf.Scraper.LogPath} {
  if !within(p) { log.Fatal("all manual data paths must remain inside the isolated runtime") }
 }
 if conf.Sqlite3.Path == "" || len(conf.ServerFileManagement.RootPath)==0 { log.Fatal("explicit manual database and media root required") }
 for _, p := range conf.ServerFileManagement.RootPath { if p=="" || !within(p) { log.Fatal("manual file manager cannot expose outside paths") } }
 listener, err := net.Listen("tcp", "127.0.0.1:18345")
 if err != nil { log.Fatal(err) } // Never kills an existing service.
 core.Init()
 if core.Config.System.Port != 18345 || core.Config.System.ServerHost != "127.0.0.1" { log.Fatal("isolated configuration required") }
 if err := models.DB_Init(core.DBS()); err != nil { log.Fatal(err) }
 processors.InitCronjob()
 processors.CronJobsExec{}.Stop() // Documentation host does not run scheduled jobs.
 processors.InitAutoBackup()
 processors.InitVideoMetadata()
 processors.InitVideoTranscode()
 gin.SetMode(gin.ReleaseMode)
 engine := gin.New()
 engine.Use(gin.Recovery())
 // APIs that create/edit schedules call RestartCronjob. Stop the demo scheduler
 // after those requests; production application behavior is not modified.
 engine.Use(func(c *gin.Context) { c.Next(); if strings.HasPrefix(c.Request.URL.Path,"/api/cronJobs/") { processors.CronJobsExec{}.Stop() } })
 router := routers.InitRouter(engine, &html)
 log.Println("Manual API: http://127.0.0.1:18345")
 log.Fatal(router.RunListener(listener))
}
