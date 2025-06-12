package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type Config struct {
	ImagesDir string `json:"images_dir"`
}

func main() {

	home := os.Getenv("HOME")

	conf_file_path := filepath.Join(home, ".background", "conf.json")
	conf_data, err := os.ReadFile(conf_file_path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	var conf_vals Config
	err = json.Unmarshal(conf_data, &conf_vals)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
	images_dir := strings.Replace(conf_vals.ImagesDir, "$HOME", home, -1)

	current_path := filepath.Join(home, ".background", "current.txt")
	image := ""

	if len(os.Args) == 1 {
		image, err = setCurrent()
		if err != nil {
			log.Fatalf("Failed to read current image: %v", err)
		}
	} else {
		switch os.Args[1] {
		case "current", "-c":
			image, err = setCurrent()
			if err != nil {
				log.Fatalf("Failed to read current image: %v", err)
			}
		case "--list", "--ls":
			println("")
			paths, err := os.ReadDir(images_dir)
			if err != nil {
				log.Fatalf("Failed to read images directory: %v", err)
			}
			var images []string
			for _, path := range paths {
				if !path.IsDir() && isImageFile(path.Name()) {
					images = append(images, path.Name())
				}
			}
			sort.Strings(images)
			if len(images) == 0 {
				fmt.Println("No images found in directory.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
			for _, name := range images {
				fileStem := getFileStem(name)
				fileName := "[" + name + "]"
				fmt.Fprintf(w, "%-25s\t%-35s\n", fileStem, fileName)
			}
			w.Flush()
			println("")
			return
		case "--random", "--rand", "-r":
			println("")
			paths, err := os.ReadDir(images_dir)
			if err != nil {
				log.Fatalf("Failed to read current image: %v", err)
			}
			var images []string
			for _, path := range paths {
				if !path.IsDir() && isImageFile(path.Name()) {
					images = append(images, path.Name())
				}
			}
			if len(images) == 0 {
				fmt.Println("No images found in directory.")
				return
			}
			random := rand.Intn(len(images))
			image_name := images[random]
			image = getFileStem(image_name)
			fmt.Printf("Randomly selected image: %s\n", image)
			println("")
		default:
			image = os.Args[1]
		}
	}

	image, err = findFile(images_dir, image)
	if err != nil {
		log.Fatalf("Failed to find image: %v", err)
	}

	logFile, err := os.OpenFile("/dev/null", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open /dev/null: %v", err)
	}
	defer logFile.Close()

	cmd := exec.Command("swaybg", "-i", filepath.Join(images_dir, image), "-m", "fill")
	cmd.Stdout = io.Discard
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start swaybg: %v", err)
	}

	time.Sleep(500 * time.Millisecond) // Wait for initialization
	if cmd.Process == nil || cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		log.Fatalf("swaybg failed to run or exited immediately")
	}

	if err := exec.Command("pkill", "-o", "swaybg").Run(); err != nil {
		log.Printf("Warning: pkill -o swaybg failed: %v", err) // Non-fatal
	}

	//fmt.Printf("Background set to: %s\n", image)

	current, err := os.Create(current_path)
	if err != nil {
		log.Fatalf("Failed to create current.txt: %v", err)
	}
	defer current.Close()
	if _, err := current.WriteString(image); err != nil {
		log.Fatalf("Failed to write to current.txt: %v", err)
	}

}

func setCurrent() (string, error) {
	home := os.Getenv("HOME")
	current_path := filepath.Join(home, ".background", "current.txt")
	image := ""
	current, err := os.Open(current_path)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(current)
	for scanner.Scan() {
		image = scanner.Text()
	}
	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	return image, nil
}

func findFile(dir, file string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Default().Fatalf("Could not open directory '%s': %v\n", dir, err)
		return "", err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			return findFile(path, file)
		} else {
			fileName := entry.Name()
			fileStem := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			if file == fileStem || file == fileName {
				return fileName, nil
			}
		}
	}
	return "", nil
}

func getFileStem(filename string) (string) {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func isImageFile(name string) bool {
    ext := strings.ToLower(filepath.Ext(name))
    return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" || ext == ".gif"
}

type filteredWriter struct {
    out io.Writer
}

func (w *filteredWriter) Write(p []byte) (n int, err error) {
    if !bytes.Contains(p, []byte("Found config")) {
        return w.out.Write(p)
    }
    return len(p), nil // Discard "Found config" messages
}
