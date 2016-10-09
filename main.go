package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"github.com/pkg/sftp" // SFTP connection
	"github.com/satori/go.uuid" // UUID Generation
	"github.com/shibukawa/configdir" // Configuration file
	"gopkg.in/alecthomas/kingpin.v2" // CLI helper
	"golang.org/x/crypto/ssh" // SSH Session
)

var (
	path = kingpin.Arg("path", "Path to file you want to share").Required().String()
	MAX_PACKET_SIZE = 1<<15
	configDirs = configdir.New("dictget", "goshare")
	usr, _ = user.Current()
)

func PublicKeyFile(file string) ssh.AuthMethod {
	// Check for ~ characters and replace accordingly
	if file[:2] == "~/" {
		dir := usr.HomeDir
		file = filepath.Join(dir, file[2:])
	}

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading SSH key")
		os.Exit(1)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Println("Error parsing SSH key")
		os.Exit(1)
	}
	return ssh.PublicKeys(key)
}

type SSHConfiguration struct {
	User string
	Host string
	Key string
	Port int
	RemotePath string
}

type Configuration struct {
	SSH SSHConfiguration
	ShareUrl string
}

func getBaseConfiguration() Configuration {
	// Creates the base configuration options
	config := Configuration{}
	config.ShareUrl = "http://share.example.com/%s"
	config.SSH = SSHConfiguration{
		Host: "share.example.com",
		User: "share_user",
		Key: "~/.ssh/id_rsa",
		Port: 22,
		RemotePath: "/var/www/",
	}
	return config
}

func main() {
	kingpin.Parse()

	// Configuration
	config := getBaseConfiguration()

	configDirs.LocalPath = filepath.Join(usr.HomeDir, ".config", "goshare")

	folder := configDirs.QueryFolderContainsFile("config.json")

	if folder != nil {
		data, _ := folder.ReadFile("config.json")
		json.Unmarshal(data, &config)
	} else {
		folders := configDirs.QueryFolders(configdir.Local)
		jsonConfig, _ := json.Marshal(config)
		folders[0].WriteFile("config.json", jsonConfig)
		log.Fatal("Unable to read config file, created base configuration.")
	}

	// Check if path exists and if it isn't a directory
	stat, err := os.Stat(*path)
	if os.IsNotExist(err) || stat.IsDir() {
		log.Fatalf("%s not found or it isn't a file\n", *path)
	}

	// Copy file to temp folder with new name
	newFileName := fmt.Sprintf("%s%s",
		uuid.NewV4().String(), filepath.Ext(*path))

	// Create ssh connection
	sshConfig := &ssh.ClientConfig{
		User: config.SSH.User,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(config.SSH.Key),
		},
	}

	connection, err := ssh.Dial("tcp",
		fmt.Sprintf("%s:%d", config.SSH.Host, config.SSH.Port),
		sshConfig)
	if err != nil {
		log.Fatal("Failed to stablish SSH connection")
	}
	defer connection.Close()

	client, err := sftp.NewClient(connection, sftp.MaxPacket(MAX_PACKET_SIZE))

	if err != nil {
		log.Fatal("Unable to start SFTP connection")
	}
	defer client.Close()

	destPath := fmt.Sprintf("%s/%s", config.SSH.RemotePath, newFileName)

	w, err := client.OpenFile(destPath, syscall.O_CREAT | syscall.O_RDWR)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	srcFile, err := ioutil.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(srcFile)
	if err != nil {
		log.Fatal(err)
	}

	// Show sharing URL
	fmt.Printf(config.ShareUrl, newFileName)

	os.Exit(0)
}
