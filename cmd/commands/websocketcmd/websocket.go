package websocketcmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/H3Cki/peerhub"
	"github.com/H3Cki/peerhub/internal/peer"
	sig "github.com/H3Cki/peerhub/internal/signal"

	//"github.com/H3Cki/peerhub/internal/inmemory"
	"github.com/urfave/cli/v2"
)

var (
	defaultPort           = 54321
	defaultMasterPassword = ""
)

var Command = &cli.Command{
	Name:    "websocket",
	Aliases: []string{"ws"},
	Action:  runWebsocket,
	Flags: []cli.Flag{
		&cli.IntFlag{Name: "port", Value: defaultPort, EnvVars: []string{"PH_PORT"}, Usage: "port to run the server on"},
		&cli.StringFlag{Name: "master-password", Value: defaultMasterPassword, EnvVars: []string{"PH_MASTER_PASSWORD"}, Usage: "master password for the server"},
	},
}

func runWebsocket(ctx *cli.Context) error {
	hub := peerhub.NewHub(peerhub.HubConfig{
		PeerService:   peer.NewInMemoryService(),
		SignalService: sig.NewInMemoryService(),
	})

	hndl := &handler{hub: hub, wc: newConnCache()}
	mux := http.NewServeMux()
	hndl.registerHandlers(mux)

	port := ctx.Int("port")

	addr := fmt.Sprintf("0.0.0.0:%d", port)

	fmt.Printf("serving connections at %s", addr)

	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	srvErrC := make(chan error)
	go func() {
		srvErrC <- srv.ListenAndServe()
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)

	var err error
	select {
	case err = <-srvErrC:
	case <-sigC:
		err = srv.Shutdown(context.Background())
	}

	return err
}
