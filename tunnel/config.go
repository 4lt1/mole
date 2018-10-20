package tunnel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
)

// SSHConfigFile finds specific attributes of a ssh server configured on a
// ssh config file.
type SSHConfigFile struct {
	sshConfig *ssh_config.Config
}

// NewSSHConfigFile creates a new instance of SSHConfigFile based on the given
// ssh config file path.
func NewSSHConfigFile() (*SSHConfigFile, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	f, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	log.Debugf("using ssh config file from: %s", configPath)

	return &SSHConfigFile{sshConfig: cfg}, nil
}

// Get consults a ssh config file to extract some ssh server attributes
// from it, returning a SSHHost. Any attribute which its value is an empty
// string is an attribute that could not be found in the ssh config file.
func (r SSHConfigFile) Get(host string) *SSHHost {
	hostname := r.getHostname(host)

	port, err := r.sshConfig.Get(host, "Port")
	if err != nil {
		port = ""
	}

	user, err := r.sshConfig.Get(host, "User")
	if err != nil {
		user = ""
	}

	localForward, err := r.getLocalForward(host)
	if err != nil {
		localForward = &LocalForward{Local: "", Remote: ""}
		log.Debugf("error to read LocalForward configuration from ssh config file")
	}

	key := r.getKey(host)

	return &SSHHost{
		Hostname:     hostname,
		Port:         port,
		User:         user,
		Key:          key,
		LocalForward: localForward,
	}
}

func (r SSHConfigFile) getHostname(host string) string {
	hostname, err := r.sshConfig.Get(host, "Hostname")

	if err != nil {
		return host
	}

	if hostname == "" {
		hostname = host
	}

	return hostname
}

func (r SSHConfigFile) getLocalForward(host string) (*LocalForward, error) {

	var local, remote string
	c, err := r.sshConfig.Get(host, "LocalForward")
	if err != nil {
		return nil, err
	}
	l := strings.Split(c, " ")
	if len(l) == 2 {
		local = l[0]
		remote = l[1]
	}

	return &LocalForward{Local: local, Remote: remote}, nil

}

func (r SSHConfigFile) getKey(host string) string {
	id, err := r.sshConfig.Get(host, "IdentityFile")

	if err != nil {
		return ""
	}

	if id != "" {
		if strings.HasPrefix(id, "~") {
			return filepath.Join(os.Getenv("HOME"), id[1:])
		}

		return id
	}

	return ""
}

// SSHHost represents a host configuration extracted from a ssh config file.
type SSHHost struct {
	Hostname     string
	Port         string
	User         string
	Key          string
	LocalForward *LocalForward
}

// LocalForward represents a LocalForward configuration for SSHHost.
type LocalForward struct {
	Local  string
	Remote string
}

// String returns a string representation of a SSHHost.
func (h SSHHost) String() string {
	return fmt.Sprintf("[hostname=%s, port=%s, user=%s, key=%s]", h.Hostname, h.Port, h.User, h.Key)
}
