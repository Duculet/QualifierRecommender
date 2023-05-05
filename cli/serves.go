package cli

import (
	"RecommenderServer/server"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func CommandWikiServes() *cobra.Command {

	var serveOnPort int     // used by serve
	var workflowFile string // used by serve
	var certFile string     // used by serve
	var keyFile string      // used by serve
	var hardLimit int       // used by serve

	// subcommand serves
	cmdServes := &cobra.Command{
		Use:   "serves <models-dir>",
		Short: "Serve a multitude of SchemaTree models via an HTTP Server",
		Long: "Load the models (schematree binaries) from <models-dir> and the recommendation" +
			" endpoint using an HTTP Server.\nAvailable endpoints are stated in the server README.",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			models_dir := filepath.Clean(args[0])

			if (keyFile == "") != (certFile == "") {
				log.Panicln("Either both --cert and --key must be set, or neither of them")
			}

			// Initiate the HTTP server. Make it stop on <Enter> press.
			router := server.SetupNewEndpoints(models_dir, workflowFile, hardLimit)

			var server = &http.Server{
				Addr:              fmt.Sprintf("0.0.0.0:%v", serveOnPort),
				ReadHeaderTimeout: 5 * time.Second,
				Handler:           router,
			}

			log.Printf("Now listening for https requests on 0.0.0.0:%v\n", serveOnPort)

			if certFile != "" && keyFile != "" {
				err := server.ListenAndServeTLS(certFile, keyFile)
				if err != nil {
					log.Panicln(err)
				}
			} else {
				// we do not want semgrep to catch this because the option to use the server with TLS is provided, but not necessary in all environments
				err := server.ListenAndServe() // nosemgrep: go.lang.security.audit.net.use-tls.use-tls
				if err != nil {
					log.Panicln(err)
				}
			}
		},
	}
	cmdServes.Flags().IntVarP(&serveOnPort, "port", "p", 8080, "`port` of http server")
	cmdServes.Flags().StringVarP(&certFile, "cert", "c", "", "the location of the certificate file (for TLS)")
	cmdServes.Flags().StringVarP(&keyFile, "key", "k", "", "the location of the private key file (for TLS)")
	cmdServes.Flags().StringVarP(&workflowFile, "workflow", "w", "", "`path` to config file that defines the workflow")
	cmdServes.Flags().IntVarP(&hardLimit, "hardlimit", "l", 500, "hard limit for the number of recommendations")
	return cmdServes
}
