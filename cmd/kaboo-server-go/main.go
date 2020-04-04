package main

// Import our dependencies. We'll use the standard HTTP library as well as the gorilla router for this app
import (
	"log"
	"os"

	"github.com/ngutman/kaboo-server-go/internal/kaboo"

	cli "github.com/urfave/cli/v2"
)

func main() {
	var restPort int
	var websocketPort int
	var auth0Domain string
	var auth0Audience string
	app := &cli.App{
		Name: "kaboo",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "rest-port",
				Value:       3001,
				Usage:       "API REST listen port",
				Destination: &restPort,
			},
			&cli.IntFlag{
				Name:        "websocket-port",
				Value:       3002,
				Usage:       "API WebSocket listen port",
				Destination: &websocketPort,
			},
			&cli.StringFlag{
				Name:        "auth0-domain",
				Usage:       "Auth0 Domain (e.g. \"dev-XXXXXX.auth0.com\")",
				Destination: &auth0Domain,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "auth0-audience",
				Usage:       "Auth0 Audience (e.g. \"https://myapp/api/\"",
				Destination: &auth0Audience,
				Required:    true,
			},
		},
		Usage: "Kaboo server FTW",
		Action: func(c *cli.Context) error {
			server := kaboo.NewServer(restPort, websocketPort, auth0Domain, auth0Audience)
			server.Start()
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}