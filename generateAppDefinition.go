package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type CmdArgs struct {
	pkgname string // "-p"
	action  string // "-a"
}

var cmdArgs = CmdArgs{}

func init() {
	generateCommand = &Command{
		UsageLine: "generateAppDefinition",
		Short:     "generate AppDefinition json file used for createAppDefinition",
		Run:       generateAppDefinition,
	}

	generateCommand.Fset.StringVar(&cmdArgs.pkgname, "p", "", "the android app's pkgname")
	generateCommand.Fset.StringVar(&cmdArgs.action, "a", "", "action for appdefinition")
	commands = append(commands, generateCommand)
}

type AppDefinitionInput struct {
	//	Name                string   `json:"name,omitempty"`
	//Version             string   `json:"version,omitempty"`
	Platform            string   `json:"platform,omitempty"`
	Logo                string   `json:"logo,omitempty"`
	Password            string   `json:"password,omitempty"`
	PkgUrl              string   `json:"pkg_url,omitempty"`
	PkgMd5              string   `json:"pkg_md5,omitempty"`
	PkgType             string   `json:"pkg_type,omitempty"`
	PreviewUrl          string   `json:"preview_url,omitempty"`
	ShortDescription    string   `json:"short_description,omitempty"`
	Developer           string   `json:"developer,omitempty"`
	AppStoreUrl         string   `json:"app_store_url,omitempty"`
	Visibility          string   `json:"visibility,omitempty"`
	ScreenShotUrls      []string `json:"screen_shot_urls,omitempty"`
	DetailedDescription string   `json:"detailed_description,omitempty"`
	SourceCodeUrl       string   `json:"source_code_url,omitempty"`
	Category            string   `json:"category,omitempty"`
	DevelopmentTime     string   `json:"development_time,omitempty"`
}

type AppDefinitionConfig struct {
	Token    string
	Endpoint string
}

var adconfig = AppDefinitionConfig{Token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTIyNzI3MzMsImlzcyI6ImxlYWZzYW5kIiwicnNjIjoidXNlci1hY2NvdW50Iiwic3ViIjoiMDFjNjhkOTAtNWZhNC00Zjc1LWI0N2QtNTAzYzk1ZmQ1OGNkIiwidHlwIjoic2lnbmluIn0.q24vZ34AH1t-rKoc2cHMHXJ6wqmvcUO2Uxk8cgy1i_8KIjKID4Fi87ewXxyBmCs09gudzGn74SZQUZ7Ig1b2TJFgupEEyve-f99-2ZN_AzfDqLth7lee-iafZIbNRZipLKFIXb1Kbn7TfzzY0mzgOAvnmG9madbTh4pmMRHgaDc",
	Endpoint: "http://13.230.247.221:8080/app/v1/definitions"}

var (
	generateCommand *Command
	pkgName         string
)

func generateAppDefinition(args []string) error {

	apps, err := loadIndexes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generateAppDefinition error [%v]", err)
		return err
	}

	if cmdArgs.pkgname == "" {
		fmt.Fprintf(os.Stderr, " pls input the pkgName\n")
		return fmt.Errorf("pls input pkgName")
	}

	for _, app := range apps {

		if app.PackageName == cmdArgs.pkgname {
			adi := AppDefinitionInput{
				Platform:            "android",
				Logo:                app.IconURL2(),
				PkgUrl:              app.Apks[0].URL(),
				Visibility:          "public",
				DetailedDescription: app.Description,
				SourceCodeUrl:       app.SourceCode,
				Category:            app.Categories[0],
				ShortDescription:    app.Summary,
				ScreenShotUrls:      app.ScreenshotURLs(),
				PkgMd5:              "-",
				PkgType:             "apk",
				Developer:           app.AuthorName,
			}

			if cmdArgs.action == "show-app-info" {
				appinfo, err := json.Marshal(&app)
				if err != nil {
					fmt.Fprintf(os.Stderr, "json parse error, [%v]\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "%s\n", string(appinfo))
				}
				return nil

			}

			if cmdArgs.action == "show-adi-info" {

				if adiinfo, err := json.Marshal(&adi); err != nil {
					fmt.Fprintf(os.Stderr, "json parse error, [%v]\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "%s\n", string(adiinfo))
				}
				return nil

			}

			adiJson := []byte{}
			adiJson, err := json.Marshal(&adi)
			if err != nil {
				fmt.Println("json parse error:", err)
				return err
			}
			//			fmt.Fprintf(os.Stderr, "adi :\n%s", string(adiJson))

			responseData, status, err := PostTeacoRequest(adiJson)
			if err != nil {
				return err
			}
			if status > 400 {
				return fmt.Errorf("PostTeacoRequest error with status code: %d, with data: %s ", status, string(responseData))
			}
			fmt.Println("response data:")
			fmt.Println(string(responseData))

			return nil
		}

	}
	fmt.Fprintln(os.Stderr, "no pkg found")
	return nil

}
func PostTeacoRequest(rd []byte) ([]byte, int, error) {

	rb := bytes.NewReader(rd)
	request, err := http.NewRequest("POST", adconfig.Endpoint, rb)
	if err != nil {
		return nil, 0, fmt.Errorf("http.NewRequest err [%v]", err)
	}
	request.Header.Add("Authorization", "Bearer "+adconfig.Token)
	request.Header.Add("Content-Type", "application/json")

	rsp, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, 0, err
	}
	defer rsp.Body.Close()

	rspData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return rspData, rsp.StatusCode, err
	}
	if rsp.StatusCode >= 200 && rsp.StatusCode < 300 {
		return rspData, rsp.StatusCode, nil
	} else {
		return rspData, rsp.StatusCode, fmt.Errorf("status code not 2XX but [%d]", rsp.StatusCode)
	}

}
