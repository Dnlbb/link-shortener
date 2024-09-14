package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type ConfigFlags struct {
	Start  string
	Result string
	File   string
	DB     string
	Key    string
}

var Conf ConfigFlags

func validateAddress(address string) error {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("некорректный формат адреса: %s, ожидается формат host:port", address)
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return fmt.Errorf("некорректный порт: %s", port)
	}
	return nil
}

func validateBaseURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("некорректный формат URL: %s", rawURL)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("URL должен содержать протокол и хост: %s", rawURL)
	}

	if strings.HasSuffix(parsedURL.Path, "/") {
		return fmt.Errorf("URL не должен заканчиваться на '/'")
	}

	return nil

}

func ParseFlags() {
	flag.StringVar(&Conf.Start, "a", ":8080", "Address and port to run server.")
	flag.StringVar(&Conf.Result, "b", "http://localhost:8080", "The server address before the short url.")
	flag.StringVar(&Conf.File, "f", "./tmp/short-url-db.json", "The path to the file to save.")
	flag.StringVar(&Conf.DB, "d", "", "The path to the postgresql.")
	flag.Parse()

	if RunAddr := os.Getenv("SERVER_ADDRESS"); RunAddr != "" {
		Conf.Start = RunAddr
	}
	if ResAddr := os.Getenv("BASE_URL"); ResAddr != "" {
		Conf.Result = ResAddr
	}
	if FilePath := os.Getenv("FILE_STORAGE_PATH"); FilePath != "" {
		Conf.File = FilePath
	}
	if PathDB := os.Getenv("DATABASE_DSN"); PathDB != "" {
		Conf.DB = PathDB
	}
	Conf.Key = os.Getenv("KEY")

	if err := validateAddress(Conf.Start); err != nil {
		fmt.Println(err)
		flag.Usage()
		return
	}

	if err := validateBaseURL(Conf.Result); err != nil {
		fmt.Println(err)
		flag.Usage()
		return
	}
}
