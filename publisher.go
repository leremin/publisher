package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"time"
	"os"
	"path/filepath"
	"encoding/xml"
	"path"
)

const XmlFileName string = "Version.xml"

type Assembly struct {
	XMLName xml.Name `xml:"Assembly"`
	Version string `xml:"Version"`
	Files FileArray `xml:"Files"`
}

type FileArray struct {
	Files []File `xml:"Files"`
}

type File struct {
	XMLName xml.Name `xml:"File"`
	Path string `xml:"Path"`
	Hash string `xml:"Hash"`
	Size int64 `xml:"Size"`
}

func (v *FileArray) AddFile(path string, hash string, size int64) {
	newFile := File{Path: path, Hash: hash, Size: size}
	v.Files = append(v.Files, newFile)
}

func hashFileMd5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}

func version() string {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	seconds := int(now.Sub(midnight).Minutes())
	return fmt.Sprintf("%s.%d", now.Format("06.1.2"), seconds)
}

func writeXml(rootPath string, asm Assembly) {
	file, _ := os.Create(path.Join(rootPath, XmlFileName))
	file.WriteString(xml.Header)
	xmlWriter := io.Writer(file)
	encoder := xml.NewEncoder(xmlWriter)
	encoder.Indent("  ", "    ")
	encoder.Encode(asm)
}

func main() {
	if len(os.Args) > 1 {
		rootPath := os.Args[1]
		asm := Assembly{Version: version()}

		filepath.Walk(rootPath,
			func(path string, fileInfo os.FileInfo, err error) (e error) {
			if filepath.Ext(path) == ".exe" || filepath.Ext(path) == ".dll" {
				md5, _ := hashFileMd5(path)
				relPath, _ := filepath.Rel(rootPath, path)
				asm.Files.AddFile(relPath, md5, fileInfo.Size())
			}
			return nil
		})
	
		writeXml(rootPath, asm)
	} else {
		fmt.Println("Usage publisher.exe %rootDir%")
	}
}