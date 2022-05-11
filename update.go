package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/jkt-check/fdroid-getter/fdroid"
)

func init() {
	commands = append(commands, &Command{
		UsageLine: "update",
		Short:     "update the index",
		Run:       runUpdate,
	})

}

func runUpdate(args []string) error {
	anyModified := false

	for _, r := range config.Repos {
		if !r.Enable {
			continue
		}

		if err := r.updateIndex(); err == errNotModified {

		} else if err != nil {
			return fmt.Errorf("could not update index: %v", err)
		}
	}
	if anyModified {
		cachePath := filepath.Join(mustCache(), "cache-gob")
		os.Remove(cachePath)
	}
	return nil
}

const jarFile = "index-v1.jar"

var errNotModified = fmt.Errorf("not modified")
var httpClient = &http.Client{}

func (r *repo) updateIndex() error {
	url := fmt.Sprintf("%s/%s", r.URL, jarFile)
	return downloadEtag(url, indexPath(r.ID), nil)
}

func indexPath(name string) string {
	return filepath.Join(mustData(), name+".jar")
}

func downloadEtag(url, path string, sum []byte) error {
	fmt.Printf("Downloading %s........\n", url)
	defer fmt.Println()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	etagPath := path + "-etag"
	if _, err := os.Stat(path); err == nil {
		etag, _ := ioutil.ReadFile(etagPath)
		req.Header.Add("If-Non-Match", string(etag))
	}
	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Download failed: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if resp.StatusCode == http.StatusNotModified {
		fmt.Printf("not modified")
		return errNotModified
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if sum == nil {
		_, err := io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		got := sha256.Sum256(data)
		if !bytes.Equal(sum, got[:]) {
			return fmt.Errorf("sha256 mismatch")
		}
		if _, err := f.Write(data); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(etagPath, []byte(respEtag(resp)), 0o644); err != nil {
		return err
	}
	fmt.Printf("done")
	return nil
}

func respEtag(resp *http.Response) string {
	etags, e := resp.Header["Etag"]
	if !e || len(etags) == 0 {
		return ""
	}
	return etags[0]
}

type apkPtrList []*fdroid.Apk

func (al apkPtrList) Len() int           { return len(al) }
func (al apkPtrList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al apkPtrList) Less(i, j int) bool { return al[i].VersCode > al[j].VersCode }

const cacheVersion = 2

type cache struct {
	Version int
	Apps    []fdroid.App
}

func loadIndexes() ([]fdroid.App, error) {
	cachePath := filepath.Join(mustCache(), "cache-gob")
	if f, err := os.Open(cachePath); err == nil {
		defer f.Close()
		var c cache
		if err := gob.NewDecoder(f).Decode(&c); err == nil && c.Version == cacheVersion {
			return c.Apps, nil
		}
	}

	m := make(map[string]*fdroid.App)

	for _, r := range config.Repos {
		if !r.Enable {
			continue
		}
		index, err := r.loadIndex()
		if err != nil {
			return nil, fmt.Errorf("error while loading %s: %v", r.ID, err)
		}
		for i := range index.Apps {
			app := index.Apps[i]
			orig, e := m[app.PackageName]
			if !e {
				m[app.PackageName] = &app
				continue
			}
			apks := append(orig.Apks, app.Apks...)
			sort.Stable(apkPtrList(apks))
			m[app.PackageName].Apks = apks
		}
	}

	apps := make([]fdroid.App, 0, len(m))
	for _, a := range m {
		apps = append(apps, *a)
	}

	sort.Sort(fdroid.AppList(apps))

	if f, err := os.Create(cachePath); err == nil {
		defer f.Close()
		gob.NewEncoder(f).Encode(cache{
			Version: cacheVersion,
			Apps:    apps,
		})
	}
	return apps, nil
}

func (r *repo) loadIndex() (*fdroid.Index, error) {

	p := indexPath(r.ID)
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("index does not exist; try 'fdroid-get update'")
	} else if err != nil {
		return nil, fmt.Errorf("could not open index: %v", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat index: %v", err)
	}
	return fdroid.LoadIndexJar(f, stat.Size(), nil)
}
