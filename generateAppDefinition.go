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
	force   bool   // "-f" , when ScreenShotUrls is empty, i also want to create it
	//	listCategories bool   // "-lc"
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
	generateCommand.Fset.BoolVar(&cmdArgs.force, "f", false, "force create appdefinition even though ScreenShotUrls nonexist")
	//	generateCommand.Fset.BoolVar(&cmdArgs.listCategories, "lc", false, "listCategories")
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

var adconfig = AppDefinitionConfig{Token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NTQ0ODE4OTcsImlzcyI6ImxlYWZzYW5kIiwicnNjIjoidXNlci1hY2NvdW50Iiwic3ViIjoiMDFjNjhkOTAtNWZhNC00Zjc1LWI0N2QtNTAzYzk1ZmQ1OGNkIiwidHlwIjoic2lnbmluIn0.hVGTdU1N4mA0OzACyCv9hKscuSeN-lyQjVElGK3ZQhgbeFx76d11MUvjuTj4_5Cao-x5KgsEcUvZ4Cvg0iAKgueuRD4uInIVaLJL4GZzJ22VDaz6Kn0-8BRzHyW-J-UmVW3N2m7B4cs-bNHcZb8hQvB_e_s49EhYmjYrC3B7qb8",
	Endpoint: "https://api.teaco.io/app/v1/definitions"}
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

	if cmdArgs.pkgname == "" && cmdArgs.action != "upload-all" {
		fmt.Fprintf(os.Stderr, " pls input the pkgName\n")
		return fmt.Errorf("pls input pkgName")
	}

	uploadItem, totalItem := 0, 0
	defer func() {
		fmt.Printf("TotalItem: [%d], UploadItem: [%d]\n", totalItem, uploadItem)
	}()
	for _, app := range apps {
		totalItem++

		if cmdArgs.action == "upload-all" {

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
			if len(adi.ScreenShotUrls) > 4 {
				adi.ScreenShotUrls = adi.ScreenShotUrls[0:4]
			}
			_, err := checkPackage(&adi)

			if err != nil {

				fmt.Println("json parse error:", err)
				//return err
				continue
			}
			adiJson := []byte{}
			adiJson, err = json.Marshal(&adi)
			if err != nil {
				fmt.Println("json parse error:", err)
				//return err
				continue
			}
			//			fmt.Fprintf(os.Stderr, "adi :\n%s", string(adiJson))

			responseData, status, err := PostTeacoRequest(adiJson)
			if err != nil {
				//return err
				fmt.Println("PostTeacoRequest error")
				continue
			}
			if status > 400 {
				//return fmt.Errorf("PostTeacoRequest error with status code: %d, with data: %s ", status, string(responseData))
				fmt.Printf("PostTeacoRequest error with status code: %d, with data: %s ", status, string(responseData))
				continue
			}
			fmt.Println("response data:")
			fmt.Println(string(responseData))
			uploadItem++
			/*
				if uploadItem > 10 {
					break
				}
			*/

		} else {

			if cmdArgs.pkgname != "" {
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

					if len(adi.ScreenShotUrls) > 4 {
						adi.ScreenShotUrls = adi.ScreenShotUrls[0:4]
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

					_, err := checkPackage(&adi)

					if err != nil {

						fmt.Println("json parse error:", err)
						return err
					}
					adiJson := []byte{}
					adiJson, err = json.Marshal(&adi)
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

					uploadItem++
					return nil
				}
			}
		}

	}
	if cmdArgs.pkgname != "" {
		fmt.Fprintln(os.Stderr, "no pkg found")
	}
	return nil

}

func checkPackage(adi *AppDefinitionInput) (bool, error) {

	if adi.Logo == "" {
		return false, fmt.Errorf("logoURL is nil")
	}
	if len(adi.Logo) < 8 {
		return false, fmt.Errorf("logoURL is not legal [%s]", adi.Logo)
	}
	if adi.Logo[:7] != "http://" && adi.Logo[:8] != "https://" {
		return false, fmt.Errorf("logoURL is not legal [%s], must start with [http://] or [https://]", adi.Logo)
	}
	/*
		if len(adi.ScreenShotUrls) == 0 {
			if !cmdArgs.force {
				fmt.Println(cmdArgs.pkgname, ":\n", adi.DetailedDescription)
				return false, fmt.Errorf("ScreenShotUrls Is Empty!!")
			}
		}
	*/

	if len(adi.ShortDescription) == 0 {
		return false, fmt.Errorf("app ShortDescription is empty!!")
	}

	return true, nil
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
