package configserver



import (
	"fmt"
	"github.com/psco-tech/gw-coach-recording-agent/uploader"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/psco-tech/gw-coach-recording-agent/models"
)

type AppOverview struct {
	AgentInfo uploader.AgentInfo
	AgentInfoErr string
}

type YTUpload struct {
	YoutubeLink string
}

func Start() {
	app := fiber.New(fiber.Config{})
	database, err := models.NewDatabase()
	if err != nil {
		log.Fatalf("Could not connect to DB: #{err}", err)
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
		appConfig, err  := database.GetAppConfig()
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
		appConfig, err  := database.GetAppConfig()
		if err != nil {
			log.Printf("Could not get app Config: %s", err.Error())
			appConfig = *new(models.AppConfig)
		}

		appConnect := uploader.NewAppConnect(appConfig.AgentToken, "http://127.0.0.1:8080")
		agentInfo, aeErr := appConnect.Info()

		var overview = *new(AppOverview)
		if aeErr != nil {
			overview.AgentInfoErr = aeErr.Error()
		} else {
			overview.AgentInfo = agentInfo
		}

		return render(c.Response().BodyWriter(), "overview.html", "/", overview)
	}
	app.Get("/overview", overview)
	app.Get("/", overview)

	app.Get("/connection", func(c *fiber.Ctx) error {
		pbxConn, _ := database.GetPBXConnectionCredentials()
		return render(c.Response().BodyWriter(), "connect.html", "/connection", pbxConn)
	})

	app.Get("/app", func(c *fiber.Ctx) error {
		appConfig, err  := database.GetAppConfig()
		if err != nil {
			log.Printf("Could not get app Config: %s", err.Error())
			appConfig = *new(models.AppConfig)
		}
		return render(c.Response().BodyWriter(), "app-config.html", "/app", appConfig)
	})

	app.Post("/connection", func(c *fiber.Ctx) error {
		pbxConn, err  := database.GetPBXConnectionCredentials()
		if err != nil {
			log.Printf("Could not get PBX connection credentials: %s", err.Error())
			pbxConn = *new(models.PBXConnectionCredentials)
		}

		if err := c.BodyParser(&pbxConn); err != nil {
			log.Printf("Could not parse PBX connection credentials from body: %s", err.Error())
			return renderWithError(c.Response().BodyWriter(), "connect.html", "/connection", pbxConn, fmt.Sprintf("Could not save connection: %s", err.Error()))
		}
		log.Printf("Parse creds: %s", pbxConn)
		database.Save(&pbxConn)

		return render(c.Response().BodyWriter(), "connect.html", "/connection", pbxConn)
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
		database.Save(&ur)

		return c.Redirect("/uploads")
	})

	go func() {
		log.Fatal(app.Listen(":3002"))
	}()
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

		if(appConfig.AgentToken == "") {
			return c.Redirect("/init")
		}

		return c.Next()
	}
}
