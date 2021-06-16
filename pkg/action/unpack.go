package action

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/itxaka/luet-mtree/pkg/helpers"
	"github.com/itxaka/luet-mtree/pkg/log"
	"github.com/mudler/luet/pkg/helpers/docker"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type YamlMtree struct {
	Mtree string
}

// excludes is a list of dirs that are excluded from the mtree check, usually those that are modified
var excludes = []string{"var/cache/luet", "oem/", "usr/local/cloud-config", "usr/local/lost+found", "lost+found", "tmp/", "mnt/"}

func UnpackAndMtree(image string, destination string) (map[string]string, error) {
	// Create temp dir for extracting the metadata
	tmpDirMetadata := fmt.Sprintf("%s/luet-mtree-metadata-%d", os.TempDir(), rand.Int())

	defer os.RemoveAll(tmpDirMetadata)

	// create temp dir for storing the pure mtree validation file
	mtreeTmpDir, err := os.MkdirTemp("", "luet-mtree-metadata-mtree")
	if err != nil {
		log.Log("Error Creating temp dir")
		return helpers.WrapErrorMap(err)
	}

	defer os.RemoveAll(mtreeTmpDir)

	// metadataImage is the name of the image unpacked + .metadata.yaml
	metadataImage := fmt.Sprintf("%s.metadata.yaml", image)

	// mtreeFile is the file where we gonna extract the mtree checksums
	mtreeFileNameSplit := strings.Split(image, ":")
	mtreeFileName := mtreeFileNameSplit[len(mtreeFileNameSplit)-1]
	mtreeFile := fmt.Sprintf("%s/%s.mtree", mtreeTmpDir, mtreeFileName)

	// Empty auth for now, this probably need some flags added to the command?
	auth := &types.AuthConfig{
		Username:      "",
		Password:      "",
		ServerAddress: "",
		Auth:          "",
		IdentityToken: "",
		RegistryToken: "",
	}

	// Dir for temporarily extracting the docker image
	extractTempDir, _ := os.MkdirTemp("", "luet-mtree-docker-extract")
	info, err := docker.DownloadAndExtractDockerImage(extractTempDir, metadataImage, tmpDirMetadata, auth, false)
	if err != nil {
		log.Log("Error while downloading docker image: %s", err.Error())
		return helpers.WrapErrorMap(err)
	}
	log.Log("Downloaded image %s with digest: %v", info.Name, info.Target.Digest)

	// find extracted metadata
	var metadataFilesFound []string

	err = filepath.WalkDir(tmpDirMetadata, func(path string, dir os.DirEntry, err error) error {
		log.Log("Walking path: %s", path)
		if strings.HasSuffix(path, ".metadata.yaml") {
			metadataFilesFound = append(metadataFilesFound, path)
		}
		return nil
	})
	// Should have only find one file!
	if err != nil || len(metadataFilesFound) > 1 {
		return helpers.WrapErrorMap(err)
	}

	metadataFile := metadataFilesFound[0]

	// read from the yaml and extract the mtree key
	mFile, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		log.Log("Error reading from metadata file %s: %s", metadataFile, err.Error())
		return helpers.WrapErrorMap(err)
	}
	metadataValues := YamlMtree{}
	err = yaml.Unmarshal(mFile, &metadataValues)
	if err != nil {
		log.Log("Error unmarshalling metadata file: %s", err.Error())
		return helpers.WrapErrorMap(err)
	}
	// save the mtree values to the mtreeFile
	mtreeFileWrite, err := os.Create(mtreeFile)
	if err != nil {
		log.Log("Error creating the mtree file %s: %s", mtreeFile, err.Error())
		return helpers.WrapErrorMap(err)
	}

	_, err = mtreeFileWrite.Write([]byte(metadataValues.Mtree))
	if err != nil {
		log.Log("Error saving the mtree file %s: %s", mtreeFile, err.Error())
		return helpers.WrapErrorMap(err)
	}
	err = mtreeFileWrite.Close()
	if err != nil {
		log.Log("Error closing the mtree file %s: %s", mtreeFile, err.Error())
		return helpers.WrapErrorMap(err)
	}
	// call the check action with the mtree validation and the dir destination where the files are
	checkAction := NewCheckAction(destination, mtreeFile, "bsd", excludes)
	out, err := checkAction.Run()
	if err != nil {
		if err.Error() == "validation failed" {
			// This is a special error case as its specific to the checks so we want to save
			// the check failure data and mention the log file so it can be investigated further
			log.Log("Validation failed :(")
			mtreeFailedOutput, _ := os.Create("/tmp/luet_mtree_failures.log")
			_, _ = mtreeFailedOutput.Write([]byte(out))
			_ = mtreeFailedOutput.Close()
			returnData := map[string]string{
				"data":  "",
				"error": "Validation failed, check /tmp/luet_mtree_failures.log for the full failures",
				"state": "Checks failed",
			}

			return returnData, err
		} else {
			log.Log("Unknown error found when checking the mtree values: %s", err.Error())
			return helpers.WrapErrorMap(err)
		}
	}
	log.Log("Validation succeeded :)")
	return helpers.WrapErrorMap(err)
}
