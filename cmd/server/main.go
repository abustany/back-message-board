package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/endpoint"
	"github.com/abustany/back-message-board/pkg/postservice"
	"github.com/abustany/back-message-board/pkg/poststore"
)

func die(logger log.Logger, err error) {
	logger.Log("startup_error", err)
	os.Exit(1)
}

func main() {
	listenAddress := flag.String("listen", "127.0.0.1:1412", "Address on which to start the HTTP server")
	adminUser := flag.String("adminUser", "", "Username of the admin user")
	adminPassword := flag.String("adminPassword", "", "Password of the admin user")

	flag.Parse()

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	mainLogger := log.With(logger, "module", "main")

	store, err := poststore.NewMemoryPostStore()

	if err != nil {
		die(mainLogger, errors.Wrap(err, "Error while creating post store"))
	}

	if *adminUser == "" {
		die(mainLogger, errors.New("You didn't provide an admin user, accessing the admin API will not be possible!"))
	}

	adminUsers := map[string]string{
		*adminUser: *adminPassword,
	}

	ep := endpoint.NewHttpEndpoint(logger, postservice.New(store), adminUsers)

	mainLogger.Log("listen", *listenAddress)
	err = http.ListenAndServe(*listenAddress, ep)

	if err != nil {
		die(mainLogger, errors.Wrap(err, "Error while starting HTTP server"))
	}
}
