// Server serves the REST API of the message board.
package main

import (
	"flag"
	"net/http"
	"os"
	"time"

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

func loadCSV(logger log.Logger, store poststore.Store, filename string) {
	fd, err := os.Open(filename)

	if err != nil {
		die(logger, errors.Wrapf(err, "Error while opening %s", filename))
	}

	defer fd.Close()

	var counter uint

	defer func(start time.Time) {
		logger.Log("event", "load_from_csv", "success", err == nil, "elapsed", time.Since(start), "n_records", counter)
	}(time.Now())

	counter, err = poststore.LoadFromCSV(store, fd, true)

	if err != nil {
		die(logger, errors.Wrap(err, "Error while loading data from CSV file"))
	}
}

func main() {
	listenAddress := flag.String("listen", "127.0.0.1:1412", "Address on which to start the HTTP server")
	adminUser := flag.String("adminUser", "", "Username of the admin user")
	adminPassword := flag.String("adminPassword", "", "Password of the admin user")
	csvFile := flag.String("loadCSV", "", "Optional, path of a CSV to load into the store after starting. The first record is considered as a header and is skipped.")

	flag.Parse()

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	mainLogger := log.With(logger, "module", "main")

	store, err := poststore.NewMemoryPostStore()

	if err != nil {
		die(mainLogger, errors.Wrap(err, "Error while creating post store"))
	}

	if *csvFile != "" {
		loadCSV(mainLogger, store, *csvFile)
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
