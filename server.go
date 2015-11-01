package ip6tun

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"strconv"
	"time"
)

const HTTPAuthHeader = "X-IP6TUN-AUTH"

type (
	ServerConfig struct {
		Debug      bool          `default:"false"`
		Port       int           `default:"8080"`
		ServerName string        `default:"localhost"`
		APIKey     string        `required:"true"`
		CertFile   string        `envconfig:"tls_cert" required:"true"`
		KeyFile    string        `envconfig:"tls_key" required:"true"`
		MaxIdle    time.Duration `default:"86400"`
	}
	Server struct {
		config   *ServerConfig
		listener net.Listener
		router   *httprouter.Router
		broker   *Broker
	}
)

func (c *ServerConfig) String() string {
	return fmt.Sprintf(
		"Debug: %d, Port: %d, ServerName: %s, ApiKey: %s, CertFile: %s, KeyFile: %s, MaxIdle: %s",
		c.Debug,
		c.Port,
		c.ServerName,
		c.APIKey,
		c.CertFile,
		c.KeyFile,
		c.MaxIdle)
}

// NewServer returns a new initialized ip6tun server
func NewServer(config *ServerConfig) (*Server, error) {
	// set debug level is requested
	if config.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("Debugmode enabled. Config: %s", config)
	}
	// output config in debug mode
	// load keypair
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   config.ServerName,
	}
	// create listener
	l, err := tls.Listen("tcp6", fmt.Sprintf(":%d", config.Port), tlsConfig)
	if err != nil {
		return nil, err
	}
	// create the control server
	s := &Server{
		config:   config,
		listener: l,
		router:   httprouter.New(),
		broker:   NewBroker(config.MaxIdle * time.Second),
	}
	// router configuration
	s.router.POST("/", s.tunnelCreate)
	s.router.GET("/", s.tunnelList)
	s.router.GET("/:id", s.tunnelView)
	s.router.PUT("/:id", s.tunnelUpdate)
	s.router.DELETE("/:id", s.tunnelDelete)

	return s, nil
}

// Run starts the HTTP server
func (s *Server) Run() {
	log.Infof("Server running at https://%s:%d", s.config.ServerName, s.config.Port)
	http.Serve(s.listener, s.logger(s.auth(s.router)))
}

// Stop stops the underlying broker gracefully i.e. closing all tunnels
func (s *Server) Stop() {
	s.broker.Close()
}

// tunnelList outputs a json list of all tunnels
func (s *Server) tunnelList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get client list
	list := s.broker.Tunnels()
	response(w, http.StatusOK, list)
}

// tunnelView outputs a single tunnel identified by its id
func (s Server) tunnelView(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// get id from route
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// get tunnel
	tun, err := s.broker.Get(uint32(id))
	if err != nil {
		if err == ErrTunnelNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	response(w, http.StatusOK, tun)
}

// tunnelCreate handles creating of new tunnels
func (s *Server) tunnelCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// process request
	rtun, host, err := readRequest(w, r)
	if err != nil {
		return
	}
	// add new tunnel
	tun, err := s.broker.Add(rtun.Name, host, rtun.RemotePort, rtun.LocalPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response(w, http.StatusCreated, tun)
}

// tunnelUpdate handles updating of existing tunnels
func (s Server) tunnelUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// process request
	rtun, host, err := readRequest(w, r)
	if err != nil {
		return
	}
	// get id from route
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// update tunnel
	tun, err := s.broker.Update(uint32(id), host, rtun.RemotePort)
	if err != nil {
		if err == ErrTunnelNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	response(w, http.StatusOK, tun)
}

// tunnelDelete handles deletion of tunnels
func (s Server) tunnelDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// get id from route
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// get tunnel
	err = s.broker.Delete(uint32(id))
	if err != nil {
		if err == ErrTunnelNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// readRequest parses and the json body into *TunnelUpsert struct and validates it.
// The remote host is extracted from the request
func readRequest(w http.ResponseWriter, r *http.Request) (*TunnelRequest, string, error) {
	rtun := new(TunnelRequest)
	// read request body into struct
	lr := http.MaxBytesReader(w, r.Body, 1048576) // 1MB is more than enough
	d := json.NewDecoder(lr)
	err := d.Decode(rtun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, "", err
	}
	// validate request
	if err := rtun.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, "", err
	}
	// get remote host from request
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, "", err
	}
	// try to resolve remote host
	if _, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[%s]:%d", host, rtun.RemotePort)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, "", err
	}
	return rtun, host, nil
}

// shortcut to respond with any data
func response(w http.ResponseWriter, status int, v interface{}) error {
	// set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// encode response
	encoder := json.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

// auth middleware
func (s *Server) auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check auth header
		if r.Header.Get(HTTPAuthHeader) != s.config.APIKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// logger middleware
func (s *Server) logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := time.Now()
		h.ServeHTTP(w, r)
		log.Infof("Finished %s %s in %s", r.Method, r.URL, time.Now().Sub(n))
	})
}
