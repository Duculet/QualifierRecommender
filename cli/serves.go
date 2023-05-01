package cli

import (
	"RecommenderServer/server"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func CommandWikiServes() *cobra.Command {

	var serveOnPort int     // used by serve
	var workflowFile string // used by serve
	var certFile string     // used by serve
	var keyFile string      // used by serve

	// subcommand serve
	cmdServe := &cobra.Command{
		Use:   "serves",
		Short: "Serve a SchemaTree model via an HTTP Server",
		Long: "Load the <model> (schematree binary) and the recommendation" +
			" endpoint using an HTTP Server.\nAvailable endpoints are stated in the server README.",
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			if (keyFile == "") != (certFile == "") {
				log.Panicln("Either both --cert and --key must be set, or neither of them")
			}

			// Initiate the HTTP server. Make it stop on <Enter> press.
			server.LoadAllModels()
			router := server.SetupNewEndpoints(500)

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
	cmdServe.Flags().IntVarP(&serveOnPort, "port", "p", 8080, "`port` of http server")
	cmdServe.Flags().StringVarP(&certFile, "cert", "c", "", "the location of the certificate file (for TLS)")
	cmdServe.Flags().StringVarP(&keyFile, "key", "k", "", "the location of the private key file (for TLS)")
	cmdServe.Flags().StringVarP(&workflowFile, "workflow", "w", "./configuration/Workflow.json", "`path` to config file that defines the workflow")
	return cmdServe
}
