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
)


var verboseMode = false
var intelligentMode = false
var CMSmode = false
var SpreadMode = false
var Statuscode = 0
var currentCount int32 = 0 
var currentFakeCount int32 = 0 
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
		timeout := time.Duration(15 * time.Second)
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
						//fmt.Printf("%s [%s] [%s]\n",newUrl+"/xxx", color.BlueString(tmpStatusCode), tmpTitle)
						if tmpStatusCode=="200"{
							fmt.Printf("%s [%s] [%s]\n",newUrl+"/xxx", color.BlueString(tmpStatusCode), tmpTitle)
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
						} else if tmpStatusCode=="401"{
							fmt.Printf("%s [%s] [%s]\n",newUrl+"/xxx", color.GreenString(tmpStatusCode), tmpTitle)
							var a = [][]string{{newUrl, tmpStatusCode, "",""}}
							tmpResultList = append(tmpResultList,a...)
						} else {
							fmt.Printf("%s [%s] [%s]\n",newUrl+"/xxx", color.RedString(tmpStatusCode), tmpTitle)
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
		timeout := time.Duration(15 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := client.Get(newUrl)
		//currentCount+=1
		var tmpTitle = ""
		if err!=nil{			
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
			atomic.AddInt32(&currentCount, 1)
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
									//here
									if tmpTitle!="Error" && tmpTitle!="Request Rejected"{
										//fmt.Println("xxx ",finalURL,each[2],strconv.Itoa(lenBody),each[3],strconv.Itoa(resp.StatusCode))
										if (each[2]!=strconv.Itoa(lenBody) || each[3]!=strconv.Itoa(resp.StatusCode)){
											if resp.StatusCode!=403 && resp.StatusCode!=404 && resp.StatusCode!=500  {
												//fmt.Println("yyy ",finalURL, each[2], strconv.Itoa(lenBody),each[3], resp.StatusCode)
												//fmt.Println("yyy0 ",newUrl,finalURL)
												//fmt.Println("yyy1 ",each[1])
												//fmt.Println("yyy2 ",strings.TrimSpace(tmpTitle))
												if CMSmode==false {
													//fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle)					

													if tmpStatusCode=="200"{
														fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.BlueString(tmpStatusCode),  lenBody, tmpTitle)
														//fmt.Println(each[1])
														//fmt.Println(tmpTitle)
													} else if tmpStatusCode=="401"{
														fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.GreenString(tmpStatusCode),  lenBody, tmpTitle)
													} else {
														fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.RedString(tmpStatusCode),  lenBody, tmpTitle)
													}

												}
												//fmt.Println("add ",newUrl)
												var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
												//fmt.Println("here6 ",a)
												tmpResultList = append(tmpResultList,a...)
											}
										}
									}
								} else {
									//fmt.Printf("%s [%s] [%d] [%s] \n",newUrl, color.GreenString(tmpStatusCode), lenBody, tmpTitle)					
									if each[3]!=tmpStatusCode{
										var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
										//fmt.Println("here5 ",a)
										tmpResultList = append(tmpResultList,a...)
										//fmt.Println("add1 ",newUrl)
									}
									if tmpStatusCode=="200"{
										fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.BlueString(tmpStatusCode),  lenBody, tmpTitle)
									} else if tmpStatusCode=="401"{
										fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.GreenString(tmpStatusCode),  lenBody, tmpTitle)										
										//fmt.Println(each)
										//fmt.Println(strings.TrimSpace(tmpTitle))
									} else {
										fmt.Printf("%s [%s] [%d] [%s]\n",newUrl, color.RedString(tmpStatusCode),  lenBody, tmpTitle)
									}

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
			atomic.AddInt32(&currentCount, 1)
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
	timeout := time.Duration(15 * time.Second)
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
	Pathsource string `cli:"s,source" usage:"Path source (default | msf | RobotsDisallowed | SecLists)"`
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
				pFilename = filename1
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
		} else {
			if len(uriPath)<1 {			
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
				}		
				_ = err2
			} else {
				pathList = append(pathList, uriPath)
			}
		}
		//start

		var finalList []string

		if SpreadMode==false {
			for _, x := range contentList {
			  for _, v := range pathList {
				url := x      		
				path := v
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl := url+path
					finalList = append(finalList, newUrl)
				} else {		
					newUrl := url+"/"+path
					finalList = append(finalList, newUrl)
				}
			  }
			}
		} else {
 	 	    for _, v := range pathList {
			  for _, x := range contentList {
				url := x      		
				path := v
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl := url+path
					finalList = append(finalList, newUrl)
				} else {		
					newUrl := url+"/"+path
					finalList = append(finalList, newUrl)
				}
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


		fmt.Println("\n[*] Testing URI Paths")
		//real uripaths
		completed1 := 0
		for _, each := range finalList {
			urlChan <- each
			completed1++
		}
		close(urlChan)  
		for {
			time.Sleep(10 * time.Millisecond)
			//fmt.Println(len(finalList),currentCount )
			if len(finalList)==int(currentCount) {
				fmt.Println("\n[*] Processing results. Please wait...")
				break
			}
		}    
	
		fmt.Println("\n")
		if CMSmode==true {
			for _, v := range tmpResultList {
				var wpVer = ""
				timeout := time.Duration(15 * time.Second)
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
									//fmt.Println(v1)

									if strings.Contains(v1,"<version>") {
										//fmt.Println(v1)
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
			RemoveDuplicates(&tmpResultList1)
			sort.Strings(tmpResultList1)
			for _, v := range tmpResultList1 {
				fmt.Printf("%s\n",v)
			}
		}		
		//end
		return nil
	})
	
	//fmt.Scanln(&input)
}
