package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"github.com/pkg/sftp"
	"github.com/satori/go.uuid"
	"gopkg.in/alecthomas/kingpin.v2"
	"golang.org/x/crypto/ssh"
)

func PublicKeyFile(file string) ssh.AuthMethod {
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
}

type Configuration struct {
	SSH SSHConfiguration
	RemotePath string
	ShareUrl string
}

var (
	path = kingpin.Arg("path", "Path to file you want to share").Required().String()
	MAX_PACKET_SIZE = 1<<15
)

func main() {
	kingpin.Parse()

	// Configuration
	config := Configuration{}
	config.ShareUrl = "http://share.example.com/%s"
	config.RemotePath = "/var/www/"
	config.SSH = SSHConfiguration{
		Host: "share.example.com",
		User: "share_user",
		Key: "~/.ssh/id_rsa",
		Port: 22,
	}

	// Check if path exists and if it isn't a directory
	stat, err := os.Stat(*path)
	if os.IsNotExist(err) || stat.IsDir() {
		log.Fatalf("%s not found or it isn't a file\n", *path)
	}

	// Copy file to temp folder with new name
	newFileName := uuid.NewV4().String()

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

	destPath := fmt.Sprintf("%s/%s", config.RemotePath, newFileName)

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
	fmt.Println("Here's the URL to share your file:")
	fmt.Printf(config.ShareUrl, newFileName)
	fmt.Println()

	os.Exit(0)
}
