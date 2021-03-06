package server

import (
	"fmt"
	"github.com/coreos/go-systemd/daemon"
	"log"
	"net/http"
	//"strconv"
	"time"
	"strconv"
)

// Server stores the hostname and port number
type Server struct {
	Hostname  string `json:"Hostname"`  // Server name
	UseHTTP   bool   `json:"UseHTTP"`   // Listen on HTTP
	UseHTTPS  bool   `json:"UseHTTPS"`  // Listen on HTTPS
	HTTPPort  int    `json:"HTTPPort"`  // HTTP port
	HTTPSPort int    `json:"HTTPSPort"` // HTTPS port
	CertFile  string `json:"CertFile"`  // HTTPS certificate
	KeyFile   string `json:"KeyFile"`   // HTTPS private key
}

// Run starts the HTTP and/or HTTPS listener
func Run(httpHandlers http.Handler, httpsHandlers http.Handler, s Server) {
	if s.UseHTTP && s.UseHTTPS {
		go func() {
			startHTTPS(httpsHandlers, s)
		}()

		startHTTP(httpHandlers, s)
	} else if s.UseHTTP {
		startHTTP(httpHandlers, s)
	} else if s.UseHTTPS {
		startHTTPS(httpsHandlers, s)
	} else {
		log.Println("Config file does not specify a listener to start")
	}
}

// startHTTP starts the HTTP listener
func startHTTP(handlers http.Handler, s Server) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), "Running HTTP "+httpAddress(s))

	daemon.SdNotify(false, "READY=1")
	interval, err := daemon.SdWatchdogEnabled(false)
	if err == nil && interval > 0 {
		go func() {
			for {
				res, err := http.Get("http://127.0.0.1:"+ strconv.Itoa(s.HTTPPort)) // ❸
				if err == nil {
					daemon.SdNotify(false, "WATCHDOG=1")
				}
				res.Body.Close()
				time.Sleep(interval / 3)
			}
		}()

	}

	// Start the HTTP listener
	log.Fatal(http.ListenAndServe(httpAddress(s), handlers))
}

// startHTTPs starts the HTTPS listener
func startHTTPS(handlers http.Handler, s Server) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), "Running HTTPS "+httpsAddress(s))

	daemon.SdNotify(false, "READY=1")
	interval, err := daemon.SdWatchdogEnabled(false)
	if err == nil && interval > 0 {
		go func() {
			for {
				res, err := http.Get("https://127.0.0.1:"+ strconv.Itoa(s.HTTPSPort)) // ❸
				if err == nil {
					daemon.SdNotify(false, "WATCHDOG=1")
				}
				res.Body.Close()
				time.Sleep(interval / 3)
			}
		}()

	}

	// Start the HTTPS listener
	log.Fatal(http.ListenAndServeTLS(httpsAddress(s), s.CertFile, s.KeyFile, handlers))
}

// httpAddress returns the HTTP address
func httpAddress(s Server) string {
	//appendix := ""
	//if s.HTTPPort != 80{

	appendix := ":" + fmt.Sprintf("%d", s.HTTPPort)
	//}
	return s.Hostname + appendix
}

// httpsAddress returns the HTTPS address
func httpsAddress(s Server) string {
	//appendix := ""
	//if s.HTTPPort != 443{

	appendix := ":" + fmt.Sprintf("%d", s.HTTPSPort)
	//}
	return s.Hostname + appendix
}
