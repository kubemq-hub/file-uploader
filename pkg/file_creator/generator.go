package file_creator

import (
	"fmt"
	"github.com/kubemq-io/file-uploader/pkg/uuid"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var randNum = rand.New(rand.NewSource(time.Now().UnixNano()))

type GeneratorRequest struct {
	Dir   string `json:"dir"`
	Size  int    `json:"size"`
	Items int    `json:"items"`
}

func randName() string {
	str := uuid.New().String()
	return strings.Replace(str, "-", "", -1)
}

func randomBytes(n int) []byte {
	r := make([]byte, n)
	if _, err := randNum.Read(r); err != nil {
		panic("rand.Read failed: " + err.Error())
	}
	return r
}
func (g *GeneratorRequest) Do() error {
	newDirName := randName()
	err := os.Mkdir(filepath.Join(g.Dir, newDirName), 0600)
	if err != nil {
		return err
	}
	for i := 0; i < g.Items; i++ {
		data := randomBytes(g.Size)
		fileName := filepath.Join(g.Dir, newDirName, randName())
		err := ioutil.WriteFile(fileName, data, 0600)
		if err != nil {
			return fmt.Errorf("error creating file %s,%s", fileName, err.Error())
		}
		fmt.Println(fmt.Sprintf("file %s created", fileName))
	}
	return nil
}
