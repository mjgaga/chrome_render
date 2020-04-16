package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var conf jsonInfo

type jsonInfo struct {
	LogrusLevel       *string `json:"logrusLevel"`
	SaveFrameJpg      *bool   `json:"-"`
	FrameJpgPath      *string `json:"frameJpgPath"`
	ForceFrameRate    *int    `json:"forceFrameRate"`
	ForceVideoBitRate *int    `json:"forceVideoBitRate"`
	ForceAudioBitRate *int    `json:"forceAudioBitRate"`
	JpegQuality       *int    `json:"jpegQuality"`
	VideoFrameGop     *int    `json:"frameGop"`
	InjectNoise       *bool   `json:"injectNoise"`
	LocalVideoPath    *string `json:"localVideoPath"`
	RtmpServer        *string `json:"rtmpServer"`
	RecordDomainAddr  *string `json:"recordDomainAddr"`
	SchedulingServer  *string `json:"schedulingServer"`
	RenderDomainUrl   *string `json:"renderDomainUrl"`
	MonitorCenterUrl  *string `json:"monitorCenterUrl"`
	VideoWidth        *int    `json:"videoWidth"`
	VideoHeight       *int    `json:"videoHeight"`
	SendReport        *bool   `json:"sendReport"`
	MaxRecordCount    *int    `json:"maxRecordCount"`
	HttpAddr          *string `json:"httpAddr"`
}

func init() {
	logLevel := flag.String("log-level", "info", "log level: debug, info, error")
	frameJpgPath := flag.String("frame-jpg", "", "a path to save page frame image")
	localVideoPath := flag.String("video-path", "./videos/", "a local dir path for save video")
	rtmpServer := flag.String("rtmp", "rtmp://127.0.0.1/test/", "rtmp server address for push stream")
	injectNoise := flag.Bool("inject-noise", false, "inject a noise audio at chrome start")
	forceFrameRate := flag.Int("frame-rate", 10, "video frame rate")
	forceVideoBitRate := flag.Int("video-bitrate", 500, "video bitrate kb/s")
	forceAudioBitRate := flag.Int("audio-bitrate", 48, "audio bitrate kb/s")
	jpegQuality := flag.Int("quality", 70, "video bitrate kb/s")
	frameGop := flag.Int("frame-gop", 100, "video bitrate kb/s")
	videoWidth := flag.Int("width", 1280, "video width")
	videoHeight := flag.Int("height", 720, "video height")
	maxRecordCount := flag.Int("max-record-count", 0, "maximum number of concurrent recording streams")
	schedulingServer := flag.String("ws", "ws://127.0.0.1:8080", "scheduling server ws address")
	recordDomainAddr := flag.String("record-domain", "http://127.0.0.1/", "record domain url")
	renderDomainUrl := flag.String("domain-url", "http://127.0.0.1/", "render page url")
	monitorCenterUrl := flag.String("monitor-url", "http://127.0.0.1:3000/", "monitor center server address")
	httpAddr := flag.String("http-addr", ":9999", "http listen on host:port")
	sendReport := flag.Bool("report", true, "send record report to scheduling server")
	showVersion := flag.Bool("v", false, "current version")
	configFile := flag.String("c", "", "read config from specified file")
	dumpConfig := flag.Bool("dump-config", false, "dump default config")
	flag.Parse()

	conf.LogrusLevel = logLevel
	conf.FrameJpgPath = frameJpgPath
	conf.ForceFrameRate = forceFrameRate
	conf.ForceVideoBitRate = forceVideoBitRate
	conf.ForceAudioBitRate = forceAudioBitRate
	conf.VideoWidth = videoWidth
	conf.VideoHeight = videoHeight
	conf.InjectNoise = injectNoise
	conf.VideoFrameGop = frameGop
	conf.LocalVideoPath = localVideoPath
	conf.RtmpServer = rtmpServer
	conf.JpegQuality = jpegQuality
	conf.SchedulingServer = schedulingServer
	conf.RenderDomainUrl = renderDomainUrl
	conf.RecordDomainAddr = recordDomainAddr
	conf.MonitorCenterUrl = monitorCenterUrl
	conf.SendReport = sendReport
	conf.MaxRecordCount = maxRecordCount
	conf.HttpAddr = httpAddr

	if *configFile != "" {
		ReadConfig(*configFile)
	}

	if *showVersion {
		fmt.Println(FullVersion())
		os.Exit(0)
	}

	if *dumpConfig {
		fmt.Println(MarshalConfig())
		os.Exit(0)
	}

	saveFrameJpg := (*conf.FrameJpgPath) != ""
	conf.SaveFrameJpg = &saveFrameJpg
	setLogCof(*conf.LogrusLevel)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)

	go func() {
		for sig := range sigs {
			logrus.Warnf("get new os signal: %s", sig)
			if sig == syscall.SIGUSR1 {
				ReadConfig(*configFile)
				saveFrameJpg := (*conf.FrameJpgPath) != ""
				conf.SaveFrameJpg = &saveFrameJpg
				setLogCof(*conf.LogrusLevel)

				logrus.Print("reload config file:\n", MarshalConfig())
			}
		}
	}()
}

func ReadConfig(configFile string) {
	if confText, err := ioutil.ReadFile(configFile); err != nil {
		panic(err)
	} else if err := json.Unmarshal(confText, &conf); err != nil {
		panic(err)
	}
}

func MarshalConfig() string {
	str, _ := json.MarshalIndent(conf, "", "  ")
	return string(str)
}

func GetConfig() *jsonInfo {
	return &conf
}
