package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

var CLI struct {
	Config string `short:"c" help:"Use the given configfile as input"`
}

type Crd struct {
	FileUrl   string `yaml:"fileUrl"`
	MergeAt   string `yaml:"mergeAt"`
	MergeFrom string `yaml:"mergeFrom"`
	Version   string `yaml:"version"`
}

func main() {
	cfg := struct {
		Template string `yaml:"template"`
		Crds     []Crd  `yaml:"crds"`
	}{}

	var (
		yamlFile []byte
		err      error

		xrd yaml.Node
	)

	kong.Parse(&CLI)
	yamlFile, err = os.ReadFile(CLI.Config)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v \n", err)
		return
	}

	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		fmt.Printf("unable to unmarshal config file or file is invalid  err #%v \n", err)
		return
	}

	if yamlFile, err = os.ReadFile(cfg.Template); err != nil {
		fmt.Printf("unable to open %s err   #%v \n", cfg.Template, err)
		return
	}

	if err = yaml.Unmarshal(yamlFile, &xrd); err != nil {
		fmt.Printf("unable to read xrd spec or file is invalid  err #%v \n", err)
		return
	}

	for _, crd := range cfg.Crds {
		// TODO: quick and dirty replace. should use go-template
		crd.FileUrl = strings.Replace(crd.FileUrl, "{{ .Version }}", crd.Version, 1)
		if yamlFile, err = readUrl(crd.FileUrl); err != nil {
			log.Printf("unable to read CRD URL %s  err  #%v \n", crd.FileUrl, err)
			continue
		}

		var (
			node               yaml.Node
			fromPath, toPath   *yamlpath.Path
			fromQuery, toQuery []*yaml.Node
		)

		if err = yaml.Unmarshal(yamlFile, &node); err != nil {
			fmt.Println(err)
			continue
		}

		if fromPath, err = yamlpath.NewPath("$" + crd.MergeFrom); err != nil {
			fmt.Println(err)
			continue
		}

		if toPath, err = yamlpath.NewPath("$" + crd.MergeAt); err != nil {
			fmt.Println(err)
			continue
		}

		if fromQuery, err = fromPath.Find(&node); err != nil {
			fmt.Println(err)
			continue
		}

		if toQuery, err = toPath.Find(&xrd); err != nil {
			fmt.Println(err)
			continue
		}

		if len(fromQuery) != 1 {
			fmt.Printf("%s - invalid result match for %s. must be exactly 1 match", crd.FileUrl, crd.MergeFrom)
			continue
		}

		if len(toQuery) != 1 {
			fmt.Printf("%s - invalid result match for %s. must be exactly 1 match", crd.FileUrl, crd.MergeAt)
			continue
		}

		found := false
		for i := 0; i < len(toQuery[0].Content); i += 2 {
			node := toQuery[0].Content[i]
			if node.Kind != yaml.ScalarNode {
				continue
			}
			switch node.Value {
			case "properties":
				toQuery[0].Content[i+1] = fromQuery[0]
				found = true
			}
		}

		if !found {
			var key yaml.Node
			key.SetString("properties")
			toQuery[0].Content = append(toQuery[0].Content, &key, fromQuery[0])
		}
	}

	var spec []byte
	if spec, err = encode(&xrd); err != nil {
		fmt.Printf("unable to marshal XRD back to Yaml  err  #%v\n", err)
		return
	}

	if err = os.WriteFile(cfg.Template, spec, 0644); err != nil {
		fmt.Printf("unable to write file  err  #%v", err)
	}
}

func readUrl(url string) (b []byte, err error) {
	var response *http.Response
	if response, err = http.Get(url); err != nil {
		return
	}
	defer response.Body.Close()

	b, err = io.ReadAll(response.Body)
	return
}

func encode(a *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()
	e.SetIndent(2)

	if err := e.Encode(a); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
