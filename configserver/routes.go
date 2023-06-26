package configserver

import (
	"fmt"
	"log"
	"strconv"

	"github.com/google/gopacket/pcap"
	"github.com/psco-tech/gw-coach-recording-agent/uploader"
	"github.com/spf13/viper"

	"github.com/gofiber/fiber/v2"
	"github.com/psco-tech/gw-coach-recording-agent/models"
)

type AppOverview struct {
	AgentInfo                 uploader.AgentInfo
	AgentInfoErr              string
	UploadStorageDirectory    string
	UploadStorageDirectoryErr string
}

type YTUpload struct {
	YoutubeLink string
}

type NetworkInterface struct {
	Name        string
	Description string
	Addresses   []string
}

type ConnectAudioForm struct {
	NetworkInterfaces          []NetworkInterface
	MTU                        int
	ActiveNetworkInterfaceName string
}

func Start() {
	app := fiber.New(fiber.Config{})
	database, err := models.NewDatabase()
	if err != nil {
		log.Fatalf("Could not connect to DB: %s", err)
	}

	app.Use(appConfigExistsMiddleware(database))

	// Serve static files like CSS or images
	app.Static("/static", "configserver/static")

	app.Use("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Next()
	})

	app.Get("/init", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "init-config.html", "/init", nil)
	})

	appConfigPost := func(c *fiber.Ctx) error {
		var parsedConfig models.AppConfig
		appConfig, err := database.GetAppConfig()
		if err != nil {
			log.Printf("Could not get app Config: %s", err.Error())
			appConfig = *new(models.AppConfig)
		}
		log.Printf("FOUND CONFIG: %s", appConfig)

		if err := c.BodyParser(&parsedConfig); err != nil {
			log.Printf("Could not parse app config from body: %s", err.Error())
			return renderWithError(c.Response().BodyWriter(), "init-config.html", "/init", nil, fmt.Sprintf("Could not save connection: %s", err.Error()))
		}
		// Only update the token - if there are other configs we don't want to blow it out
		appConfig.AgentToken = parsedConfig.AgentToken
		database.Save(&appConfig)
		return c.Redirect("/")
	}

	app.Post("/init", appConfigPost)
	app.Post("/app", appConfigPost)

	overview := func(c *fiber.Ctx) error {
		appConfig, err := database.GetAppConfig()
		if err != nil {
			log.Printf("Could not get app Config: %s", err.Error())
			appConfig = *new(models.AppConfig)
		}
		viper.SetDefault("app_connect_host", "http://127.0.0.1:8080")
		appConnect := uploader.NewAppConnect(appConfig.AgentToken, viper.GetString("app_connect_host"))
		agentInfo, aeErr := appConnect.Info()

		var overview = *new(AppOverview)
		if aeErr != nil {
			overview.AgentInfoErr = aeErr.Error()
		} else {
			overview.AgentInfo = agentInfo
		}

		uploadDir, upDirErr := uploader.GetUploadsDirectory()
		if upDirErr != nil {
			overview.UploadStorageDirectoryErr = upDirErr.Error()
		} else {
			overview.UploadStorageDirectory = uploadDir
		}

		return render(c.Response().BodyWriter(), "overview.html", "/", overview)
	}
	app.Get("/overview", overview)
	app.Get("/", overview)

	app.Get("/avaya-pbx-connect", func(c *fiber.Ctx) error {
		pbxConn, _ := database.GetPBXConnectionCredentials()
		return render(c.Response().BodyWriter(), "avaya-pbx-connect.html", "/connection", pbxConn)
	})

	app.Get("/cad", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "cad.html", "/cad", nil)
	})

	app.Get("/app", func(c *fiber.Ctx) error {
		appConfig, err := database.GetAppConfig()
		if err != nil {
			log.Printf("Could not get app Config: %s", err.Error())
			appConfig = *new(models.AppConfig)
		}
		return render(c.Response().BodyWriter(), "app-config.html", "/app", appConfig)
	})

	app.Post("/avaya-pbx-connect", func(c *fiber.Ctx) error {
		pbxConn, err := database.GetPBXConnectionCredentials()
		if err != nil {
			log.Printf("Could not get PBX connection credentials: %s", err.Error())
			pbxConn = *new(models.PBXConnectionCredentials)
		}

		if err := c.BodyParser(&pbxConn); err != nil {
			log.Printf("Could not parse PBX connection credentials from body: %s", err.Error())
			return renderWithError(c.Response().BodyWriter(), "avaya-pbx-connect.html", "/connection", pbxConn, fmt.Sprintf("Could not save connection: %s", err.Error()))
		}
		log.Printf("Parse creds: %s", pbxConn)
		database.Save(&pbxConn)

		return render(c.Response().BodyWriter(), "avaya-pbx-connect.html", "/connection", pbxConn)
	})

	app.Get("/devices", func(c *fiber.Ctx) error {
		devices := database.GetAllDevices()
		log.Printf("Devices: %s", devices)
		return render(c.Response().BodyWriter(), "devices.html", "/devices", devices)
	})

	app.Get("/add-device", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "add-device.html", "/devices", nil)
	})

	app.Post("/add-device", func(c *fiber.Ctx) error {
		var dev models.Device
		if err := c.BodyParser(&dev); err != nil {
			log.Printf("Could not parse Add Device from body: %s", err.Error())
			return renderWithError(c.Response().BodyWriter(), "add-device.html", "/devices", nil, fmt.Sprintf("Could not add device: %s", err.Error()))
		}
		log.Printf("Parse Device: %s", dev)

		if dev.Extension == "" {
			return renderWithError(c.Response().BodyWriter(), "add-device.html", "/devices", nil, fmt.Sprintf("Must set an extension"))
		}

		database.Save(&dev)

		return c.Redirect("/devices")
	})

	app.Get("/del-device/:deviceId", func(c *fiber.Ctx) error {
		var devId, err = strconv.Atoi(c.Params("deviceId"))

		if err != nil {
			return c.Redirect("/devices")
		}

		var dev = database.GetDeviceById(devId)
		return render(c.Response().BodyWriter(), "del-device.html", "/devices", dev)
	})

	app.Post("/del-device/:deviceId", func(c *fiber.Ctx) error {
		var devId, err = strconv.Atoi(c.Params("deviceId"))

		if err != nil {
			return c.Redirect("/devices")
		}

		var dev = database.GetDeviceById(devId)
		if dev.ID < 1 {
			return c.Redirect("/devices")
		}

		database.Delete(&dev)

		return c.Redirect("/devices")
	})

	app.Get("/uploads", func(c *fiber.Ctx) error {
		uploads, _ := database.GetRecentUploads()
		return render(c.Response().BodyWriter(), "uploads.html", "/uploads", uploads)
	})

	app.Post("/uploads", func(c *fiber.Ctx) error {
		yt := *new(YTUpload)
		if err := c.BodyParser(&yt); err != nil {
			log.Printf("Could not parse youtube link: %s", err.Error())
			return c.Redirect("/uploads")
		}

		outputDir, _ := uploader.GetUploadsDirectory()
		outputFile, err := DownloadYouTubeVideo(yt.YoutubeLink, outputDir)
		if err != nil {
			log.Printf("Could not download youtube link: %s %s", yt.YoutubeLink, err.Error())
			return c.Redirect("/uploads")
		}

		ur := *new(models.UploadRecord)
		ur.Status = models.UploadStatusQueued
		ur.ContentType = "video/mp4"
		ur.FilePath = outputFile
		ur.Type = models.UploadRecordTypeCFS_AUDIO
		database.Save(&ur)

		cha := uploader.GetUploadRecordChannel()

		cha <- ur

		return c.Redirect("/uploads")
	})

	app.Post("/uploads/:recordId/retry", func(c *fiber.Ctx) error {
		var recId, _ = strconv.Atoi(c.Params("recordId"))
		ur := database.GetUploadRecordById(recId)
		cha := uploader.GetUploadRecordChannel()
		cha <- ur
		return c.Redirect("/uploads")
	})

	app.Get("/connect-audio", func(c *fiber.Ctx) error {

		connectAudioConfig, err := database.GetPassiveMonitoringConfig()
		if err != nil {
			log.Printf("Could not load audio connection config: %s", err.Error())
		}

		formBody := new(ConnectAudioForm)
		if connectAudioConfig.MTU != 0 {
			formBody.MTU = connectAudioConfig.MTU
		} else {
			formBody.MTU = 1500
		}
		formBody.NetworkInterfaces = loadNetworkInterfaces()

		if connectAudioConfig.NetworkInterfaceName != "" {
			formBody.ActiveNetworkInterfaceName = connectAudioConfig.NetworkInterfaceName
		} else {
			formBody.ActiveNetworkInterfaceName = ""
		}

		return render(c.Response().BodyWriter(), "connect-audio.html", "/connect-audio", formBody)
	})

	app.Post("/connect-audio", func(c *fiber.Ctx) error {

		var parsedConfig models.PassiveMonitoringConfig
		config, err := database.GetPassiveMonitoringConfig()
		if err != nil {
			log.Printf("Could not get connect audio Config: %s", err.Error())
			config = *new(models.PassiveMonitoringConfig)
		}

		if err := c.BodyParser(&parsedConfig); err != nil {
			log.Printf("Could not parse app config from body: %s", err.Error())
			return renderWithError(c.Response().BodyWriter(), "connect-audio.html", "/connect-audio", nil, fmt.Sprintf("Could not save configs: %s", err.Error()))
		}
		// Only update the token - if there are other configs we don't want to blow it out
		config.MTU = parsedConfig.MTU
		config.NetworkInterfaceName = parsedConfig.NetworkInterfaceName
		database.Save(&config)
		return c.Redirect("/connect-audio")
	})

	go func() {
		log.Fatal(app.Listen(":3002"))
	}()
}

func loadNetworkInterfaces() []NetworkInterface {
	interfaces := make([]NetworkInterface, 0)
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return interfaces
	}

	for _, device := range devices {
		if device.Flags&0x01 == 0x01 {
			// exclude loopback
			continue
		}
		if device.Flags&0x08 == 0x08 {
			// exclude wireless devices
			continue
		}

		if device.Flags&0x02 != 0x02 {
			// exclude devices that are not up
			continue
		}

		iface := NetworkInterface{
			Name:        device.Name,
			Description: device.Description,
			Addresses:   make([]string, 0),
		}

		for _, a := range device.Addresses {
			iface.Addresses = append(iface.Addresses, a.IP.String())
		}

		interfaces = append(interfaces, iface)
	}
	return interfaces
}

func appConfigExistsMiddleware(db *models.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if c.Path() == "/init" {
			return c.Next()
		}

		appConfig, err := db.GetAppConfig()
		if err != nil {
			log.Printf("Could load app config: %s", err.Error())
			return c.Redirect("/init")
		}

		if appConfig.AgentToken == "" {
			return c.Redirect("/init")
		}

		return c.Next()
	}
}
