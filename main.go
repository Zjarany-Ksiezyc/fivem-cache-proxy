package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)


func main() {
	if os.Getenv("Origin") == "" {
		log.Print("Non origin server defined")
		os.Exit(1)
		return
	}

	http.HandleFunc("/", func(responde http.ResponseWriter, request *http.Request) {
		if !strings.Contains(request.URL.Path, "resource.rpf") {
			log.Print(request.RemoteAddr+" --> localhost"+request.URL.Path)
			io.WriteString(responde, "Caching proxy is up and running\n")

			return
		}

		hash := request.URL.Query().Get("hash")
		resource := UrlToResource(request.URL.Path)
		cache := ReadCacheFile(resource, hash)

		io.WriteString(responde, cache)
	})

	port := os.Getenv("Port")
	if port == "" {
		port = "80"
	}

	log.Print("Caching proxy started on port: ", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func UrlToResource(str string) string {
	semiResoult := strings.Replace(str, "/", "", 1)
	result := strings.Replace(semiResoult, "/resource.rpf", "", 1)

	return result
}

func ReadCacheFile(folder, hash string) string {
	mainPath := filepath.Join("cache", folder, hash)
	folderPath := filepath.Join("cache", folder)

	_, err := os.Stat(mainPath)
	if os.IsNotExist(err) {
		log.Print("Creating new cache row for resource: "+folder+" hash: ["+hash+"]")


		// deleting folder
		_, err := os.Stat(folderPath)
		if err == nil {
			err := os.RemoveAll(folderPath)
			if err != nil {
				log.Print(err)
			}
		}


		// creating new folder
		err = os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			log.Print(err)
		}

		originIp := os.Getenv("Origin")
		resp, err := http.Get("http://" + originIp + "/files/" + folder + "/resource.rpf?hash=" + hash)

		if err != nil {
			log.Print(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			httpErr := fmt.Sprintf("Origin server respond: %d ", resp.StatusCode)
			log.Print(httpErr)
		}


		// creating new file
		file, err := os.Create(mainPath)
		if err != nil {
			log.Print(err)
		}
		defer file.Close()


		// copy content to new file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Print(err)
		}
	}


	// sending content to client
	data, err := os.ReadFile(mainPath)
	if err != nil {
		log.Print(err)
	}

	return string(data)
}