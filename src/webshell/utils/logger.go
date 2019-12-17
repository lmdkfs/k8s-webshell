package utils

import (
	"path"
	"time"

	"webshell/webshell/config"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)


var Logger *logrus.Logger
var cfg = config.NewConfig()
type loggerEntryWithfields interface {
	WithFields(fields logrus.Fields) *logrus.Entry
}

func init() {
	Logger = logrus.New()
	// 显示代码行号
	Logger.SetReportCaller(true)
	// 禁止终端输出
	//src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	//if err!= nil{
	//	fmt.Println("err", err)
	//}
	//Logger.Out = src


	LogPath := path.Join(cfg.Log.LogPath, cfg.Log.LogName)

	logWriter, err := rotatelogs.New(
		LogPath+".%Y-%m-%d-%H-%M.log",
		rotatelogs.WithLinkName(LogPath),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	if err != nil {
		Logger.Errorf("config local file system logger err. %v", err)
	}

	writeMap := lfshook.WriterMap{
		logrus.DebugLevel: logWriter,
		logrus.InfoLevel:  logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.FatalLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}
	lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{})
	Logger.AddHook(lfHook)

}



func GinRus(logger loggerEntryWithfields, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestPath := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := logger.WithFields(logrus.Fields{
			"serviceName": "k8s-webshell",
			"short_message":"k8s-webshell",
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       requestPath,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())

		} else {
			entry.Info()
		}
	}
}
