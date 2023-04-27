package configserver



import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/psco-tech/gw-coach-recording-agent/models"
)

func Start() {
	app := fiber.New(fiber.Config{})
	database, err := models.NewDatabase()
	if err != nil {
		log.Fatalf("Could not connect to DB: #{err}", err)
	}

	// Serve static files like CSS or images
	app.Static("/static", "configserver/static")

	app.Use("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "overview.html", "/", nil)
	})

	app.Get("/connection", func(c *fiber.Ctx) error {
		pbxConn, _ := database.GetPBXConnectionCredentials()
		return render(c.Response().BodyWriter(), "connect.html", "/connection", pbxConn)
	})

	app.Get("/app", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "app-config.html", "/app", nil)
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

	go func() {
		log.Fatal(app.Listen(":3002"))
	}()
}
