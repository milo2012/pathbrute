package main

import (
    "fmt"
    "os"
    "time"
    "log"
    "bufio"
    "net/http"
    "io/ioutil"
	"github.com/mkideal/cli"
	"github.com/badoux/goscraper"
	"github.com/fatih/color"
	"sync"
	"strings"
	"strconv"
	"sort"
	"crypto/tls"
	"io"
	"sync/atomic"
	"net/url"
	"github.com/hashicorp/go-version"
	"os/signal"
	"syscall"
)


var verboseMode = false
var intelligentMode = false
var CMSmode = false
var SpreadMode = false
var Statuscode = 0
var currentCount int = 0 

var totalListCount int = 0
var currentListCount int = 1

var currentFakeCount int32 = 0 
var currentProgressCount int32 = 0 

var Pathsource = ""
var tmpTitleList [][]string	
var tmpResultList [][]string	
var tmpResultList1 []string	

var joomlaFileList []string	
var drupalFileList []string

func f(from string) {
    for i := 0; i < 3; i++ {
        fmt.Println(from, ":", i)
    }
}

func cleanup() {
    fmt.Println("\nCTRL-C (interrupt signal)")
	for _, v := range tmpResultList {
		if !stringInSlice(v[0],tmpResultList1) {
			//fmt.Println("xxx ",v[0])
			tmpResultList1 = append(tmpResultList1, v[0])
			//tmpResultList1 = append(tmpResultList1, v[0])
		}
	}
	var tmpResultList2 []string
	sort.Strings(tmpResultList1)
	for _, v := range tmpResultList1 {
		u, err := url.Parse(v)
		if err==nil {
			if len(u.Path)>0 {
				tmpResultList2 = append(tmpResultList2,v)
			}
		}
	}    
	if len(tmpResultList2)<1 {
		fmt.Println("No results found")
	} else {
		fmt.Println("[+] Results")
		for _, v := range tmpResultList2 {
			fmt.Printf("%s\n",v)
		}
	}
}

func removeCharacters(input string, characters string) string {
	 filter := func(r rune) rune {
		 if strings.IndexRune(characters, r) < 0 {
				 return r
		 }
		 return -1
	 }
	 return strings.Map(filter, input)
}

func testFakePath(urlChan chan string) {
    for newUrl := range urlChan {
		timeout := time.Duration(8 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := client.Get(newUrl)
		if err==nil{
			var initialStatusCode = resp.StatusCode
			//if resp.StatusCode==200 {
			finalURL := resp.Request.URL.String()
			s, err := goscraper.Scrape(finalURL, 5)
			if err==nil {
				var lenBody = 0
				body, err := ioutil.ReadAll(resp.Body)
				if err==nil {
					lenBody = len(body)
					if strings.HasSuffix(finalURL,"/") {
						finalURL=finalURL[0:len(finalURL)-1]
					}			
					//if finalURL==newUrl {
					tmpStatusCode := strconv.Itoa(initialStatusCode)
					newUrl = strings.Replace(newUrl, "/xxx", "", -1)
					//fmt.Println(newUrl,s.Preview.Title)
					if verboseMode==true {
						var tmpTitle=strings.TrimSpace(s.Preview.Title)
						//fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/xxx", color.BlueString(tmpStatusCode),strconv.Itoa(lenBody), tmpTitle)
						if tmpStatusCode=="200"{
							fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/xxx", color.BlueString(tmpStatusCode),strconv.Itoa(lenBody), tmpTitle)
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
						} else if tmpStatusCode=="401"{
							fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/xxx", color.GreenString(tmpStatusCode),strconv.Itoa(lenBody), tmpTitle)
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
						} else {
							fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/xxx", color.RedString(tmpStatusCode),strconv.Itoa(lenBody), tmpTitle)
						}

						var a = [][]string{{newUrl, s.Preview.Title, strconv.Itoa(lenBody), tmpStatusCode}}
						tmpTitleList = append(tmpTitleList,a...)
						_ = a
					}
					//}
					_ = err
				}
			//}
			resp.Body.Close()
			}
		} 
		_ = err
		_ = resp
		atomic.AddInt32(&currentFakeCount, 1)
    }
}

func getUrlWorker(urlChan chan string) {
	//red := color.New(color.FgRed).SprintFunc()
	//currentCount+=1

    for newUrl := range urlChan {
    	//if verboseMode == true {
		//	fmt.Printf("Checking: %s \n\r",newUrl)
		//}
		timeout := time.Duration(8 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := client.Get(newUrl)
		initialStatusCode := ""
		//currentCount+=1
		var tmpTitle = ""
		if err!=nil{			
			initialStatusCode = strconv.Itoa(resp.StatusCode)
			if strings.Contains(err.Error(),"Client.Timeout exceeded") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("Timeout"))						
			} else if strings.Contains(err.Error(),"connection refused") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("Connection Refused"))									
			} else if strings.Contains(err.Error(),"no such host") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("Unknown Host"))									
			} else if strings.Contains(err.Error(),"connection reset by peer") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("Connection Reset"))									
			} else if strings.Contains(err.Error(),"tls: no renegotiation") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("TLS Error"))	
			} else if strings.Contains(err.Error(),"TLS handshake timeout") {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString("Timeout"))													
			} else {
				fmt.Printf("%s [%s]\n",newUrl, color.RedString(err.Error()))
			}
			currentCount+=1
			currentListCount+=1
			//atomic.AddInt(&currentCount, 1)
		} else {
			if verboseMode==true {
				/*if moreData==false {
					tmpStatusCode := strconv.Itoa(resp.StatusCode)
					if Statuscode!=0 {
						if resp.StatusCode==Statuscode {
							fmt.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
							fmt.Println("add5 ",newUrl)
						}
					} else {						
						if tmpStatusCode=="200"{
							fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							fmt.Println("add3 ",newUrl)
							tmpResultList = append(tmpResultList,a...)
						} else if tmpStatusCode=="401"{
							fmt.Printf("%s [%s]\n",newUrl, color.GreenString(tmpStatusCode))
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							fmt.Println("add4 ",newUrl)
							tmpResultList = append(tmpResultList,a...)
						} else {
							fmt.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
						}
					}

				} else {
				*/
				var lenBody = 0
				body, err := ioutil.ReadAll(resp.Body)
				if err==nil {
					lenBody = len(body)
				}
				finalURL := resp.Request.URL.String()
				var tmpTitle = ""
				if finalURL==newUrl {
					s, err := goscraper.Scrape(finalURL, 5)
					if err==nil {
						tmpTitle = s.Preview.Title
						tmpTitle = strings.TrimSpace(tmpTitle)
					}
				}					
				
				if intelligentMode==true {
					tmpStatusCode := strconv.Itoa(resp.StatusCode)
					for _, each := range tmpTitleList { 
						if strings.Contains(finalURL,each[0]) {
							if newUrl==finalURL { 				
							//if each[0]==finalURL {   		
								if each[1]!=strings.TrimSpace(tmpTitle) {
									//fmt.Println(each[2])
									if tmpStatusCode=="200"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
									} else if tmpStatusCode=="401"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
									} else {
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}

									if tmpTitle!="Error" && tmpTitle!="Request Rejected" && tmpTitle!="Runtime Error"{
										if (each[2]!=strconv.Itoa(lenBody) || each[3]!=strconv.Itoa(resp.StatusCode)){
											if resp.StatusCode!=403 && resp.StatusCode!=404 && resp.StatusCode!=500 && resp.StatusCode!=204 {
												if CMSmode==false {
													if each[3]!=initialStatusCode && each[2]!=strconv.Itoa(lenBody){
														var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),tmpTitle}}
														tmpResultList = append(tmpResultList,a...)
													}
												}
											}
										}
									}
								} else {
									if each[3]!=initialStatusCode {
										var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
										tmpResultList = append(tmpResultList,a...)
									}
									if tmpStatusCode=="200"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
									} else if tmpStatusCode=="401"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
									} else {
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
									//currentListCount+=1

								}
							}
						}
					}
				} else {
					tmpStatusCode := strconv.Itoa(resp.StatusCode)
					//if CMSmode==false {
					if Statuscode!=0 {
						if resp.StatusCode==Statuscode {
							fmt.Printf("*** %s [%s] [%d] [%s] \n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle)					
							var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
							tmpResultList = append(tmpResultList,a...)
						} else {
								//fmt.Printf("yyy")
								// %s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
						}						
					} else {				
						if tmpStatusCode=="200"{
							fmt.Printf("%s [%s] [%d] [%s] \n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle)					
							var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
							tmpResultList = append(tmpResultList,a...)
						} else if tmpStatusCode=="401"{
							fmt.Printf("%s [%s]\n",newUrl, color.GreenString(tmpStatusCode))
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
						} else {
							fmt.Printf("%s [%s] [%d] [%s] \n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle)					
						}
					}
					//}
				}
				//fmt.Printf("Checking: %s\n",newUrl)
				//}
			} else {
				if Statuscode!=0 {
					tmpStatusCode := strconv.Itoa(resp.StatusCode)	
					if resp.StatusCode==Statuscode {	
						fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
						finalURL := resp.Request.URL.String()
						if strings.HasSuffix(finalURL,"/") {
							finalURL=finalURL[0:len(finalURL)-1]
						}
						if finalURL==newUrl {
							if resp.StatusCode!=403 {
								var a = [][]string{{newUrl, tmpStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							}
						}
					}
					// else {
					//	fmt.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
					//}				
				} else {
					tmpStatusCode := strconv.Itoa(resp.StatusCode)	
					if resp.StatusCode==200 {		
						fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
						finalURL := resp.Request.URL.String()
						if strings.HasSuffix(finalURL,"/") {
							finalURL=finalURL[0:len(finalURL)-1]
						}
						if finalURL==newUrl {
							if resp.StatusCode!=403 {
								var a = [][]string{{newUrl, tmpStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							}
						}
					} else {
						fmt.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
					}				
				}
			}
			resp.Body.Close()
			currentCount+=1
			currentListCount+=1
			//atomic.AddInt(&currentCount, 1)
		} 
		_ = err
		_ = resp
		_ = tmpTitle 
		//currentCount+=1
		//fmt.Printf("%d\n",currentCount)
    }
}

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

func testURL(newUrl string) {   
	timeout := time.Duration(8 * time.Second)
	client := http.Client{
	    Timeout: timeout,
	}

	fmt.Printf("Checking: %s \n\r",newUrl)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := client.Get(newUrl)
	if err == nil{
		fmt.Println("ooo %s [%s]",newUrl, resp.StatusCode)
		//if resp.StatusCode==200{
	    //    fmt.Println("Working "+newUrl)
	    //    s, err := goscraper.Scrape(newUrl, 5)
	    //    if err == nil {
		//        fmt.Printf("%s : %s\n", newUrl, s.Preview.Title)
	    //    }
		//}
		resp.Body.Close()
	} 
	//else {
	//	fmt.Println("%s\n",err)		
	//}
	_ = err
	_ = resp
}

func BytesToString(data []byte) string {
	return string(data[:])
}

func stringInSlice(str string, list []string) bool {
 	for _, v := range list {
 		if v == str {
 			return true
 		}
 	}
 	return false
 }

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func DownloadFile(filepath string, url string) error {
    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()
    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

type argT struct {
	cli.Helper
	Filename string `cli:"U,filename" usage:"File containing list of websites"`
	URLpath string `cli:"u,url" usage:"Url of website"`
	PFilename string `cli:"P,Paths" usage:"File containing list of URI paths"`
	Path string `cli:"p,path" usage:"URI path"`
	Pathsource string `cli:"s,source" usage:"Path source (default | msf | exploitdb | exploitdb-asp | exploitdb-aspx | exploitdb-cfm | exploitdb-cgi | exploitdb-cfm | exploitdb-jsp | exploitdb-perl | exploitdb-php  | RobotsDisallowed | SecLists)"`
	Threads int  `cli:"n,threads" usage:"No of concurrent threads"`
	Statuscode int  `cli:"c" usage:"Status code"`
	Intellimode bool `cli:"i" usage:"Intelligent mode"`
	Verbose bool `cli:"v,verbose" usage:"Verbose mode"`
	CMSmode bool `cli:"cms" usage:"Fingerprint CMS"`
	SpreadMode bool `cli:"x" usage:"Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host"`
	//SpreadMode bool `cli:"spread" usage:"Spread load across different hosts"`
	//Y    bool `cli:"y" usage:"boolean type, too"`
}

func main() {
	log.SetOutput(ioutil.Discard)
	//log.SetFlags(0)
	
	joomlaFileList = append(joomlaFileList,"/administrator/manifests/files/joomla.xml")
	joomlaFileList = append(joomlaFileList,"/administrator/language/en-GB/en-GB.xml")
	drupalFileList = append(drupalFileList,"/CHANGELOG.txt")
	//drupalFileList = append(drupalFileList,"/LICENSE.txt")
	//const workersCount = 1
	workersCount := 2
	
	filename1 := ""
	pFilename := ""
	uriPath := ""
	
	var contentList []string
	var pathList []string
	
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		filename1 = argv.Filename
		pFilename = argv.PFilename
		Pathsource = argv.Pathsource
		if argv.SpreadMode {
			SpreadMode = true
		}
		if argv.Statuscode>0 {
			Statuscode = argv.Statuscode
		}
		if argv.Intellimode {
			intelligentMode = true
		}
		if argv.Verbose {
			verboseMode = true
		}		
		if len(argv.Path)>0 { 
			uriPath = argv.Path
		}
		if argv.Threads>0 {
			workersCount = argv.Threads
		}
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanup()
			os.Exit(1)
		}()		
	
		if len(pFilename)>0 {		
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				log.Printf("File %s not exists", pFilename)
				os.Exit(3)
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
				_ = err2
			}		
		} 
		if len(uriPath)>0 {
			pathList = append(pathList, uriPath)
		}

		if Pathsource=="default" {
			pFilename = "defaultPaths.txt"
			_, err1 := os.Stat("defaultPaths.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/defaultPaths.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("defaultPaths.txt", fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines("defaultPaths.txt")
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}		
		if Pathsource=="msf" {
			pFilename = "pathList.txt"
			_, err1 := os.Stat("pathList.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/metasploitHelper/master/pathList.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("pathList.txt", fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines("pathList.txt")
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}
		if Pathsource=="exploitdb" {
			pFilename = "exploit_db.txt"
			_, err1 := os.Stat("exploit_db.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploit_db.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}		
		if Pathsource=="exploitdb-asp" {
			pFilename = "exploitdb_asp.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}		
		if Pathsource=="exploitdb-aspx" {
			pFilename = "exploitdb_aspx.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}		
		if Pathsource=="exploitdb-cfm" {
			pFilename = "exploitdb_cfm.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="exploitdb-cgi" {
			pFilename = "exploitdb_cgi.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="exploitdb-jsp" {
			pFilename = "exploitdb_jsp.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="exploitdb-jsp" {
			pFilename = "exploitdb_jsp.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="exploitdb-perl" {
			pFilename = "exploitdb_perl.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="exploitdb-php" {
			pFilename = "exploitdb_php.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}	
		if Pathsource=="SecLists" {
			pFilename = "SecLists-common.txt"
			_, err1 := os.Stat("SecLists-common.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("SecLists-common.txt", fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines("SecLists-common.txt")
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2

		}
		if Pathsource=="RobotsDisallowed" {
			pFilename = "RobotsDisallowed.txt"
			_, err1 := os.Stat("RobotsDisallowed.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/danielmiessler/RobotsDisallowed/master/Top100000-RobotsDisallowed.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("RobotsDisallowed.txt", fileUrl)
				_ = err
			}
			_ = err1
			lines, err2 := readLines("RobotsDisallowed.txt")
			for _, v := range lines {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					pathList = append(pathList, v)
				}
			}		
			_ = err2
		}
		if len(argv.URLpath)<1 && len(argv.Filename)<1 {
			fmt.Println("[!] Please set the -U or the -u argument")
			os.Exit(3)
		} else {
			if len(argv.Filename)>0 {
				_, err := os.Stat(filename1)
				if os.IsNotExist(err) {
					log.Printf("File %s not exists", filename1)
					os.Exit(3)
				}
				lines, err := readLines(filename1)
				for _, v := range lines {
					if strings.Contains(v,"http") {
						contentList = append(contentList, v)
					} else {
						if len(v)>0 {
							contentList = append(contentList, "https://"+v)
							contentList = append(contentList, "http://"+v)
						}
					}
					//fmt.Println("https://"+v)
				}		
				_ = err
			} else {
				if strings.Contains(argv.URLpath,"http") {
					contentList = append(contentList, argv.URLpath)
				} else {
					if len(argv.URLpath)>0 {
						contentList = append(contentList, "https://"+argv.URLpath)
						contentList = append(contentList, "http://"+argv.URLpath)
					}
				}
			}
		}


		var contentList1 []string
  	    for _, v := range contentList {
			if strings.HasSuffix(v,":443") {
				v=v[0:len(v)-4]
				v=strings.TrimSpace(v)
				if len(v)>0 {
					if !stringInSlice(v,contentList1) {
						contentList1 = append(contentList1, v)
					}
				}
			} else {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					contentList1 = append(contentList1, v)
				}
			}			
  	    }
		contentList=contentList1
		//_ = contentList1

		if argv.CMSmode {
			CMSmode = true
			pathList = append(pathList, "/wp-links-opml.php")
		    for _, v := range joomlaFileList {
		    	pathList = append(pathList,v)
		    }
		    for _, v := range drupalFileList {
		    	pathList = append(pathList,v)
		    }
		} 
		
		var finalList []string

		if SpreadMode==false {
			for _, x := range contentList {
			  for _, v := range pathList {
				url := x      		
				path := v
				newUrl := ""
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl = url+path
				} else {		
					newUrl = url+"/"+path
				}
				finalList = append(finalList, newUrl)
			  }
			}
		} else {
 	 	    for _, v := range pathList {
			  for _, x := range contentList {
				url := x      		
				path := v
				newUrl := ""
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl = url+path
				} else {		
					newUrl = url+"/"+path
					//fmt.Println(newUrl)
				}
				finalList = append(finalList, newUrl)
			  }
			}
		}

		urlChan := make(chan string)
		if intelligentMode==true {
			var wg1 sync.WaitGroup
			wg1.Add(workersCount)
	
			for i := 0; i < workersCount; i++ {
				go func() {
					testFakePath(urlChan)
					wg1.Done()
				}()
			}

			fmt.Println("[*] Getting Default Page Title for Invalid URI Paths")
			completed := 0
			for _, each := range contentList {
				urlChan <- each+"/xxx"
				completed++
			}
			close(urlChan)    
			for {
				time.Sleep(10 * time.Millisecond)
				if len(contentList)==int(currentFakeCount) {
					break
				}
			}
		}
	
		var wg sync.WaitGroup
		urlChan = make(chan string)
		wg.Add(workersCount)
	
		for i := 0; i < workersCount; i++ {
			go func() {
				getUrlWorker(urlChan)
				wg.Done()
			}()
		}


		totalListCount=len(finalList)

		fmt.Println("\n[*] Testing URI Paths")
		//real uripaths
		completed1 := 0
		for _, each := range finalList {
			urlChan <- each
			completed1++
		}
		close(urlChan)  
		
		var tmpLastCount = 0
		var lastTime = time.Now()

		for {
			time.Sleep(10 * time.Millisecond)
			if len(finalList)==int(currentCount) {
				//if len(finalList)==int(currentListCount) {
				fmt.Println("\n[*] Processing results. Please wait...")
				break
			} else {
				if int(currentCount)!=int(tmpLastCount) {
					tmpLastCount = int(currentCount)
					lastTime=time.Now()
				} else {
					t := time.Now()
					elapsed := t.Sub(lastTime)
					if elapsed.Seconds()>60 && currentCount>0 {
						fmt.Println("here\n")
						break 
					}
				}									
			} 
		}   
	
		fmt.Println("\n")
		if CMSmode==true {
			for _, v := range tmpResultList {
				var wpVer = ""
				timeout := time.Duration(8 * time.Second)
				client := http.Client{
					Timeout: timeout,
				}
				if strings.HasSuffix(v[0],"/administrator/language/en-GB/en-GB.xml") || strings.HasSuffix(v[0],"/administrator/manifests/files/joomla.xml") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							if strings.Contains(bodyStr,"_Incapsula_Resource") {
								wpVer="- Protected by Incapsula"
							} else {
								s := strings.Split(bodyStr,"\n")
								for _, v1 := range s {

									if strings.Contains(v1,"<version>") {
										v1=strings.Replace(v1,"</version>","",1)
										v1=strings.Replace(v1,"<version>","",1)
										v1=strings.TrimSpace(v1)
										wpVer = v1
									}
								}
							}
						}
						v[0]=strings.Replace(v[0],"/administrator/language/en-GB/en-GB.xml","",1)
						v[0]=strings.Replace(v[0],"/administrator/manifests/files/joomla.xml","",1)					
						if len(wpVer)>0 {
							var a = v[0]+color.BlueString(" [Joomla "+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}
				if strings.Contains(v[0],"/CHANGELOG.txt") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							var tmpFound = false
							for _, v1 := range s {
								if tmpFound==false {
									if strings.Contains(v1,"Drupal ") {
										v1=strings.TrimSpace(v1)
										wpVer = strings.Split(v1,",")[0]
										tmpFound=true
									}
								}
							}
						}
						v[0]=strings.Replace(v[0],"/CHANGELOG.txt","",1)					
						if len(wpVer)>0 {
							var a = v[0]+color.BlueString(" ["+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}				

				//if strings.HasPrefix(v[3],"Links for ") {			
				if strings.Contains(v[0],"/wp-links-opml.php") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							for _, v1 := range s {
								if strings.Contains(v1," generator=\"") {
									v1=removeCharacters(v1,"<!--  generator=\"WordPress\"/")
									v1=removeCharacters(v1,"<!-- generator=\"WordPress\"/")
									v1=removeCharacters(v1,"\" -->")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								}
							}
						}
					}
					v[0]=strings.Replace(v[0],"/wp-links-opml.php","",1)
					if len(wpVer)>0 {
						var a = v[0]+color.BlueString(" [Wordpress "+wpVer+"]")
						tmpResultList1 = append(tmpResultList1, a)
					}		
				}
			}
		} else {
			for _, v := range tmpResultList {
				if !stringInSlice(v[0],tmpResultList1) {
					//fmt.Println("xxx ",v[0])
					tmpResultList1 = append(tmpResultList1, v[0])
					//tmpResultList1 = append(tmpResultList1, v[0])
				}
			}
			fmt.Println("[+] Results")
			sort.Strings(tmpResultList1)
			for _, v := range tmpResultList1 {
				u, err := url.Parse(v)
				if err==nil {
					if len(u.Path)>0 {
						fmt.Printf("%s\n",v)
					}
				}
			}
		}

		if CMSmode==true {
			var joomlaKBList [][]string	
			var wpKBList [][]string	
			var drupalKBList [][]string	

			var a = [][]string{{"joomla","3.7.0","Joomla Component Fields SQLi Remote Code Execution","exploit/unix/webapp/joomla_comfields_sqli_rce"}}
			joomlaKBList = append(joomlaKBList,a...)
			var b = [][]string{{"joomla","3.4.4-3.6.3","Joomla Account Creation and Privilege Escalation","auxiliary/admin/http/joomla_registration_privesc"}}
			joomlaKBList = append(joomlaKBList,b...)
			var c = [][]string{{"joomla","1.5.0-3.4.5","Joomla HTTP Header Unauthenticated Remote Code Execution","exploit/multi/http/joomla_http_header_rce"}}
			joomlaKBList = append(joomlaKBList,c...)
			var d = [][]string{{"joomla","3.2-3.4.4","Joomla Content History SQLi Remote Code Execution","exploit/unix/webapp/joomla_contenthistory_sqli_rce"}}
			joomlaKBList = append(joomlaKBList,d...)
			var e = [][]string{{"joomla","2.5.0-2.5.13,3.0.0-3.1.4","Joomla Media Manager File Upload Vulnerability","exploit/unix/webapp/joomla_media_upload_exec"}}
			joomlaKBList = append(joomlaKBList,e...)

			a = [][]string{{"wordpress","4.6","WordPress PHPMailer Host Header Command Injection","exploit/unix/webapp/wp_phpmailer_host_header"}}
			wpKBList = append(wpKBList,a...)
			b = [][]string{{"wordpress","4.7-4.7.1","WordPress REST API Content Injection","auxiliary/dos/http/wordpress_long_password_dos"}}
			wpKBList = append(wpKBList,b...)
			c = [][]string{{"wordpress","3.7.5,3.9-3.9.3,4.0-4.0.1","WordPress Long Password DoS",""}}
			wpKBList = append(wpKBList,c...)
			d = [][]string{{"wordpress","3.5-3.9.2","Wordpress XMLRPC DoS","auxiliary/dos/http/wordpress_xmlrpc_dos"}}
			wpKBList = append(wpKBList,d...)
			e = [][]string{{"wordpress","0-1.5.1.3","WordPress cache_lastpostdate Arbitrary Code Execution","exploit/unix/webapp/wp_lastpost_exec"}}
			wpKBList = append(wpKBList,e...)
			var f = [][]string{{"wordpress","0-4.4.1","Wordpress XML-RPC system.multicall Credential Collector","auxiliary/scanner/http/wordpress_multicall_creds"}}
			wpKBList = append(wpKBList,f...)
			var g = [][]string{{"wordpress","0-4.6","WordPress Traversal Directory DoS",""}}
			wpKBList = append(wpKBList,g...)
			
			a = [][]string{{"drupal","7.0,7.31","Drupal HTTP Parameter Key/Value SQL Injection","exploit/multi/http/drupal_drupageddon"}}
			drupalKBList = append(drupalKBList,a...)
			b = [][]string{{"drupal","7.15,7.2","PHP XML-RPC Arbitrary Code Execution","exploit/unix/webapp/php_xmlrpc_eval"}}
			drupalKBList = append(drupalKBList,b...)
			c = [][]string{{"drupal","7.0-7.56,8.0<8.3.9,8.4.0<8.4.6,8.5.0-8.5.1","CVE-2018-7600 / SA-CORE-2018-002","https://github.com/a2u/CVE-2018-7600"}}
			drupalKBList = append(drupalKBList,c...)
			
			RemoveDuplicates(&tmpResultList1)
			sort.Strings(tmpResultList1)
			for _, v := range tmpResultList1 {
				fmt.Printf("%s\n",v)
				if strings.Contains(v,"Joomla") {
					tmpSplit1 :=strings.Split(v,"[Joomla ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					for _, v := range joomlaKBList {
						if strings.Contains(v[1],",") {
							s := strings.Split(string(v[1]),",")
							for _, s1 := range s {
								if strings.Contains(s1,"-") {
									s2 := strings.Split(s1,"-")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err		
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								}
							}	
						} else {
							if strings.Contains(v[1],"-") {
								s2 := strings.Split(v[1],"-")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else if strings.Contains(v[1],"<") {
								s2 := strings.Split(v[1],"<")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err		
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							}

						}			
					}
				}					
				if strings.Contains(v,"Wordpress") {
					tmpSplit1 :=strings.Split(v,"[Wordpress ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					for _, v := range wpKBList {
						if strings.Contains(v[1],",") {
							s := strings.Split(string(v[1]),",")
							for _, s1 := range s {
								if strings.Contains(s1,"-") {
									s2 := strings.Split(s1,"-")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err		
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								}
							}	
						} else {
							if strings.Contains(v[1],"-") {
								s2 := strings.Split(v[1],"-")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else if strings.Contains(v[1],"<") {
								s2 := strings.Split(v[1],"<")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err		
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							}

						}			
					}
				}		
				if strings.Contains(v,"Drupal") {
					tmpSplit1 :=strings.Split(v,"[Drupal ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					for _, v := range drupalKBList {
						if strings.Contains(v[1],",") {
							s := strings.Split(string(v[1]),",")
							for _, s1 := range s {
								if strings.Contains(s1,"-") {
									s2 := strings.Split(s1,"-")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err									
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								}
							}	
						} else {
							if strings.Contains(v[1],"-") {
								s2 := strings.Split(v[1],"-")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							}

						}			
					}
				}				
			}
		}		
		//end
		return nil
	})
	
	//fmt.Scanln(&input)
}
