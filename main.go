package main

import (
	"context"
	"log"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

const (
	// Per: https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/
	jwksURL = `https://srnd.cloudflareaccess.com/cdn-cgi/access/certs`
)

func main() {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		log.Printf("Received request: %s %s", c.Method(), c.Path())
		log.Println("Headers:")
		c.Request().Header.VisitAll(func(key, value []byte) {
			log.Printf("%s: %s", key, value)
		})
		log.Println("==================================================================================")
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/admin/", validateToken, func(c *fiber.Ctx) error {
		return c.SendString("Hello, Admin!")
	})

	log.Fatal(app.Listen(":3005"))
}

func validateToken(c *fiber.Ctx) error {
	cfAssertionToken := c.Get("CF-Access-JWT-Assertion")

	if cfAssertionToken == "" {
		return c.SendString("No assertion token present.")
	}

	// Create context and fetch JWK
	ctx := context.Background()
	options := keyfunc.Options{
		Ctx: ctx,
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.Keyfunc\nError: %s", err.Error())
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	}
	key, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		log.Fatalf("Failed to get JWK Key Function.\nError: %s", err.Error())
		return c.SendString("Error getting JWK Key Function.")
	}

	// Parse the JWT
	token, err := jwt.Parse(cfAssertionToken, key.Keyfunc)
	if err != nil {
		log.Fatalf("Failed to parse the JWT.\nError: %s", err.Error())
		return c.SendString("Error parsting JWT.")
	}

	// Validate the JWT
	if !token.Valid {
		log.Fatalf("The token is not valid.")
		return c.SendString("Invalid token.")
	}
	log.Println("The token is valid.")

	return c.Next()
}
