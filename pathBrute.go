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

var workersCount = 2
var timeoutSec = 15
var verboseMode = false
var intelligentMode = false
var CMSmode = false
var SpreadMode = false
var Statuscode = 0
var currentCount int = 0 
var ContinueNum int = 0 
var proxyMode = false

var totalListCount int = 0
var currentListCount int = 1

var currentFakeCount int32 = 0 
var currentProgressCount int32 = 0 

var Pathsource = ""
var tmpTitleList [][]string	
var tmpResultList [][]string	
var tmpResultList1 []string	
var tmpResultList4 []string

var wpFileList []string
var joomlaFileList []string	
var drupalFileList []string
var proxy_addr=""
var reachedTheEnd=false

var userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36"
        
func f(from string) {
    for i := 0; i < 3; i++ {
        fmt.Println(from, ":", i)
    }
}

func cleanup() {
	//var lastXCount=0
    fmt.Println("\nCTRL-C (interrupt signal)")
    //for {
    //	if lastXCount!=currentListCount {
	//		fmt.Println(currentListCount)
	//	}
	//}

	for _, v := range tmpResultList {
		if !stringInSlice(v[0],tmpResultList1) {
			tmpResultList1 = append(tmpResultList1, v[0])
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
	var tmpResultList3 []string
	if len(tmpResultList2)>0 {
		for _, v := range tmpResultList2 {
			timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
			client := http.Client{
				Timeout: timeout,
			}
			if proxyMode==true {
				url_i := url.URL{}
				url_proxy, _ := url_i.Parse(proxy_addr)
				http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
			}
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req, err := http.NewRequest("GET", v, nil)
			if err==nil {
				req.Header.Add("User-Agent", userAgent)
				resp, err := client.Do(req)		
				//resp, err := client.Get(v)
				if err == nil {
					s, err := goscraper.Scrape(v, 5)
					if err == nil {
						var tmpTitle=strings.TrimSpace(s.Preview.Title)
						var lenBody = 0
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							lenBody = len(body)
						}
						if (resp.StatusCode!=403 && resp.StatusCode!=503 && resp.StatusCode!=404 && resp.StatusCode!=406 && resp.StatusCode!=400 && resp.StatusCode!=500 && resp.StatusCode!=204) {
							var a = v+" ["+(strconv.Itoa(resp.StatusCode))+"] ["+strconv.Itoa(lenBody)+"] ["+tmpTitle+"]"
							tmpResultList3 = append(tmpResultList3,a)
							//fmt.Printf("%s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
							//log.Printf("%s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)	
						}
					}
				}
			}
		}
	}
	if len(tmpResultList3)>0 {
		//fmt.Println("\n[+] Results")
		//log.Printf("\n[+] Results")
		//for _, v := range tmpResultList3 {
		//	fmt.Println(v)
		//}		
		time.Sleep(5 * time.Second)
		fmt.Println("\n")
		log.Printf("\n")

		var wg sync.WaitGroup
		urlChan := make(chan string)
		wg.Add(workersCount)
		for i := 0; i < workersCount; i++ {
			go func() {
				checkURL(urlChan)
				wg.Done()
			}()
		}
		for _, each := range tmpResultList3 {
			urlChan <- each
		}
		close(urlChan)  
		wg.Wait()		
					
	} else {
		fmt.Println("\n[*] No results found")
	}
	os.Exit(3)
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

func getPage(newURL1 string) (string, string, int, int) {
	//u, err := url.Parse(newUrl)
	//if err != nil {
	//	panic(err)
	//}
	//var newURL1=u.Scheme+"://"+u.Host+"/NonExistence"		
	//var newURL1=u.Scheme+"://"+u.Host

	var tmpStatusCode = 0
	var tmpTitle = ""
	var lenBody=0
	var tmpFinalURL =""
	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", newURL1, nil)
	if err==nil {
		req.Header.Add("User-Agent", userAgent)
		resp, err := client.Do(req)			
		if err==nil{					
			tmpStatusCode = resp.StatusCode
			s, err := goscraper.Scrape(newURL1, 5)			
			if err==nil {
				tmpTitle = s.Preview.Title
				tmpTitle=strings.TrimSpace(tmpTitle)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err==nil {
				lenBody = len(body)
			}
			tmpFinalURL = resp.Request.URL.String()			
			return tmpFinalURL,tmpTitle,tmpStatusCode, lenBody
		}
	}
	return tmpFinalURL,tmpTitle,tmpStatusCode, lenBody
}

func testFakePath(urlChan chan string) {
	//var retryList []string	
    for newUrl := range urlChan {
		var getTmpTitle=""
		var getTmpStatusCode=0
		var getLenBody=0
		var getTmpFinalURL=""
		//getTmpFinalURL,getTmpTitle,getTmpStatusCode,getLenBody=getPage(newUrl)	

		u, err := url.Parse(newUrl)
		if err != nil {
			panic(err)
		}
		var newURL1=u.Scheme+"://"+u.Host+"/NonExistence"

		getTmpFinalURL,getTmpTitle,getTmpStatusCode,getLenBody=getPage(newURL1)	
		newUrl = strings.Replace(newUrl, "/NonExistence", "", -1)
		if strings.HasSuffix(getTmpFinalURL,"/") {
			getTmpFinalURL=getTmpFinalURL[0:len(getTmpFinalURL)-1]
		}			
		if verboseMode==true {
			if getTmpStatusCode==200{
				if verboseMode==true {
					fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
				}
				var a = [][]string{{newUrl, strconv.Itoa(getTmpStatusCode), "",""}}
				tmpResultList = append(tmpResultList,a...)
			} else if getTmpStatusCode==401{
				if verboseMode==true {
					fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
				}
				var a = [][]string{{newUrl, strconv.Itoa(getTmpStatusCode), "",""}}
				tmpResultList = append(tmpResultList,a...)
			} else {
				if verboseMode==true {
					fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
				}
			}
			var a = [][]string{{newUrl, getTmpTitle, strconv.Itoa(getLenBody), strconv.Itoa(getTmpStatusCode)}}
			tmpTitleList = append(tmpTitleList,a...)
			_ = a
		}
		//_ = err

		/*
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			//CheckRedirect: func(req *http.Request, via []*http.Request) error {
        	//	return http.ErrUseLastResponse
    		//},
		}
		if proxyMode==true {
			url_i := url.URL{}
			url_proxy, _ := url_i.Parse(proxy_addr)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		var tmpComplete=false
		for tmpComplete==false {
			req, err := http.NewRequest("GET", newUrl, nil)
			if err==nil {
				req.Header.Add("User-Agent", userAgent)
				resp, err := client.Do(req)
				if err==nil {
					var initialStatusCode = resp.StatusCode
					var initialTmpTitle = ""
					s, err := goscraper.Scrape(newUrl, 5)
					if err==nil {
						initialTmpTitle=strings.TrimSpace(s.Preview.Title)
						var lenBody = 0
						body, err := ioutil.ReadAll(resp.Body)
						
						finalURL := resp.Request.URL.String()
						if err==nil {
							lenBody = len(body)
							if strings.HasSuffix(finalURL,"/") {
								finalURL=finalURL[0:len(finalURL)-1]
							}			
							tmpStatusCode := strconv.Itoa(initialStatusCode)
							newUrl = strings.Replace(newUrl, "/NonExistence", "", -1)
							if verboseMode==true {
								if tmpStatusCode=="200"{
									if verboseMode==true {
										fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
										log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
									}
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								} else if tmpStatusCode=="401"{
									if verboseMode==true {
										fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
										log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
									}
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								} else {
									if verboseMode==true {
										fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
										log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(tmpStatusCode),strconv.Itoa(lenBody), initialTmpTitle)
									}
								}
								var a = [][]string{{newUrl, s.Preview.Title, strconv.Itoa(lenBody), tmpStatusCode}}
								tmpTitleList = append(tmpTitleList,a...)
								_ = a
							}
							//}
							_ = err
						}
					}
				resp.Body.Close()
				tmpComplete=true
				}
				_ = resp
			} 
			_ = err
			if tmpComplete==false {
				if stringInSlice(newUrl,retryList) {
					tmpComplete=true
				} else {
					retryList = append(retryList,newUrl)
					time.Sleep(5 * time.Second)
				}
			}
		}
		*/
		atomic.AddInt32(&currentFakeCount, 1)
    }
}
func checkURL(urlChan chan string) {
	var tmpResultList3 []string
    for v := range urlChan {    
    	//fmt.Println("A "+v)
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", v, nil)
		req.Header.Add("User-Agent", userAgent)
		resp, err := client.Do(req)		

		s, err := goscraper.Scrape(v, 5)
		var lenBody = 0
		var tmpTitle = ""
		if err==nil {
			tmpTitle=strings.TrimSpace(s.Preview.Title)						
			body, err3 := ioutil.ReadAll(resp.Body)
			if err3==nil {
				lenBody = len(body)
			}
		}
		if (resp.StatusCode!=403 && resp.StatusCode!=503 && resp.StatusCode!=404 && resp.StatusCode!=406 && resp.StatusCode!=400 && resp.StatusCode!=500 && resp.StatusCode!=204) {
			if intelligentMode==false {
				if resp.StatusCode==401 {
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				} else if (resp.StatusCode==200) {
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				} else {
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				}				
			} else {
				var initialStatusCode=resp.StatusCode
				var initialPageSize=lenBody

				u, err := url.Parse(v)
				if err != nil {
					panic(err)
				}
				numberOfa := strings.Count(u.Path, "/")				
				tmpSplit2 :=strings.Split(u.Path,"/")
				var counter1=numberOfa
				if numberOfa<3 {	
					var newURL3=u.Scheme+"://"+u.Host	
					getTmpFinalURL3,getTmpTitle3,getTmpStatusCode3,getLenBody3:=getPage(newURL3)
					//if getTmpStatusCode3!=resp.StatusCode {
					var getTmpFinalURL4=""
					var getTmpTitle4=""
					var getTmpStatusCode4=0
					var getLenBody4=0
					var newURL4=""
					if strings.HasSuffix(v,".php") {
						tmpSplit3 :=strings.Split(u.Path,".php")
						newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+".php"
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)
					} else if strings.HasSuffix(v,".aspx") {
						tmpSplit3 :=strings.Split(u.Path,".aspx")
						newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+".aspx"
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)
					} else if strings.HasSuffix(v,".asp") {
						tmpSplit3 :=strings.Split(u.Path,".asp")
						newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+".asp"
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)
					} else if strings.HasSuffix(v,".html") {
						tmpSplit3 :=strings.Split(u.Path,".html")
						newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+".html"
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)
					} else if strings.HasSuffix(v,".htm") {
						tmpSplit3 :=strings.Split(u.Path,".htm")
						newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+".htm"
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)
					} else {
						if  strings.HasSuffix(v,"/") {
							newURL4=newURL3+"/xxx/"
						} else { 						
							newURL4=newURL3+"/xxx"
						}
						getTmpFinalURL4,getTmpTitle4,getTmpStatusCode4,getLenBody4=getPage(newURL4)						
					}						
					if (initialStatusCode!=getTmpStatusCode4) {
						//fmt.Println("b1 "+v+" "+newURL4)
						//fmt.Println("b2 "+strconv.Itoa(resp.StatusCode))
						//fmt.Println("b3 "+strconv.Itoa(getTmpStatusCode3)+" "+strconv.Itoa(getTmpStatusCode4))
						tmpResultList3 = append(tmpResultList3, v)
					} 
					_=getTmpFinalURL4
					_=getTmpTitle4
					_=getTmpStatusCode4
					_=getLenBody4
					_=getTmpFinalURL3
					_=getTmpTitle3
					_=getTmpStatusCode3
					_=getLenBody3
				} else {
					for counter1>numberOfa-1 {							
						var uriPath1=""				
						if counter1==numberOfa {
							uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/",1)
						} else {
							uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/xxx",1)
						}
						var newURL=u.Scheme+"://"+u.Host+uriPath1

						req1, err := http.NewRequest("GET", newURL, nil)
						//fmt.Println("xx")
						if err==nil {
							req1.Header.Add("User-Agent", userAgent)
							resp1, err := client.Do(req1)		
							if err==nil {								
								body, err := ioutil.ReadAll(resp1.Body)
								if err==nil {
									lenBody = len(body)
									if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {
										u1, err := url.Parse(newURL)
										if err != nil {
											panic(err)
										}
										tmpSplit3 :=strings.Split(u1.Path,"/")
										if len(tmpSplit3)>3 {
											//fmt.Println("gg")
											var uriPath2=strings.Replace(u1.Path,"/"+tmpSplit3[2]+"/","",1)
											var newURL1=u.Scheme+"://"+u.Host+uriPath2
											req2, err := http.NewRequest("GET", newURL1, nil)
											req2.Header.Add("User-Agent", userAgent)
											resp2, err := client.Do(req2)														
											if err==nil {
												body2, err2 := ioutil.ReadAll(resp2.Body)
										
												var lenBody2 = len(body2)
												//fmt.Println("gg1 "+newURL1)
												//fmt.Println(strconv.Itoa(resp2.StatusCode)+" "+strconv.Itoa(initialStatusCode))
												//fmt.Println(strconv.Itoa(lenBody2)+" "+strconv.Itoa(initialPageSize))
												if resp2.StatusCode==initialStatusCode && initialPageSize==lenBody2 {
													if strings.HasSuffix(newURL1, "/") {
														newURL1=newURL1[0:len(newURL1)-1]
													}
													if !stringInSlice(newURL1,tmpResultList3) {
														//fmt.Println("gg2")
														tmpResultList3 = append(tmpResultList3, newURL1)
													} 
												} else {
													if !stringInSlice(v,tmpResultList3) {
														//fmt.Println("gg3")
														tmpResultList3 = append(tmpResultList3, v)
													}
												}												
												_=err2
											}
											_ = resp2
										} else {			
											//fmt.Println("yy")
											var newURL1 = newURL[0:len(newURL)-1]
											req2, err := http.NewRequest("GET", newURL1, nil)
											_=err
											req2.Header.Add("User-Agent", userAgent)
											resp2, err := client.Do(req2)	
											if resp2.StatusCode==resp1.StatusCode {
												newURL=newURL1
											}
											if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {																												
												if !stringInSlice(newURL,tmpResultList3) {
													//fmt.Println("c1 "+v+" "+newURL)
													//fmt.Println("b2 "+strconv.Itoa(resp.StatusCode))
													//fmt.Println("b3 "+strconv.Itoa(getTmpStatusCode3)+" "+strconv.Itoa(getTmpStatusCode4))

													tmpResultList3 = append(tmpResultList3, newURL)
												}
											}
										}
									} else {
										//here1
										u1, err := url.Parse(v)
										if err != nil {
											panic(err)
										}
										var newURL2=u1.Scheme+"://"+u1.Host
										req2, err := http.NewRequest("GET", newURL2, nil)
										req2.Header.Add("User-Agent", userAgent)
										resp2, err := client.Do(req2)														
										if resp2.StatusCode==resp1.StatusCode {
											tmpResultList3 = append(tmpResultList3, newURL2)
										} else {
											tmpResultList3 = append(tmpResultList3, v)
										}
									}
								}
							} else {
								fmt.Println(err)
							}
							_=err							
						} else {
							fmt.Println(err)
						}
						//counter1+=1
						counter1-=1
					}
					_ = initialStatusCode
					_ = initialPageSize
				}
			}
		}	
    }

	RemoveDuplicates(&tmpResultList3)
	sort.Strings(tmpResultList3)
	for _, v := range tmpResultList3 {
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		v1, err := url.Parse(v)
    	if err != nil {
    	    panic(err)
    	}		
    	if v1.Path!="/" {
			req2, err := http.NewRequest("GET", v, nil)
			req2.Header.Add("User-Agent", userAgent)
			resp2, err := client.Do(req2)														
			if (resp2.StatusCode!=403 && resp2.StatusCode!=503 && resp2.StatusCode!=404 && resp2.StatusCode!=406 && resp2.StatusCode!=400 && resp2.StatusCode!=500 && resp2.StatusCode!=204) {
				if err==nil {
					body2, err2 := ioutil.ReadAll(resp2.Body)				
					if err2==nil {
						s, err3 := goscraper.Scrape(v, 5)
						if err3==nil {
							var tmpTitle2 = ""
							tmpTitle2=strings.TrimSpace(s.Preview.Title)						
							var lenBody2 = len(body2)
							if !stringInSlice(v,tmpResultList4) {
								fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
								log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
								tmpResultList4 = append(tmpResultList4,v)
							}
						} else { 
							if !stringInSlice(v,tmpResultList4) {
								fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
								log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
								tmpResultList4 = append(tmpResultList4,v)
							}
						}
					} else {
						if !stringInSlice(v,tmpResultList4) {
							fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
							log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
							tmpResultList4 = append(tmpResultList4,v)
						}
					}
				}
			}
		}
	}
}

func getUrlWorker(urlChan chan string) {
	//lastURL
    for newUrl := range urlChan {
    	var newUrl1 = strings.Split(newUrl," | ")
    	newUrl = newUrl1[0]
    	var currentListCount, _ = strconv.Atoi(newUrl1[1])
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		if ContinueNum==0 || ContinueNum<=currentListCount {			
			client := http.Client{
				Timeout: timeout,
			}
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			req, err := http.NewRequest("GET", newUrl, nil)
			if err==nil {
				req.Header.Add("User-Agent", userAgent)
				initialStatusCode := ""
				var tmpTitle = ""
				resp, err := client.Do(req)			
				if err!=nil{					
					if (strings.Contains(err.Error(),"i/o timeout") || strings.Contains(err.Error(),"Client.Timeout exceeded") || strings.Contains(err.Error(),"TLS handshake timeout")) {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)						
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)
					} else if strings.Contains(err.Error(),"connection refused") {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)									
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)
					} else if strings.Contains(err.Error(),"no such host") {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)									
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)	
					} else if strings.Contains(err.Error(),"connection reset by peer") {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)									
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)	
					} else if strings.Contains(err.Error(),"tls: no renegotiation") {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
					} else if strings.Contains(err.Error(),"stopped after 10 redirects") {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)	
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)							
					} else {
						fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
						log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
					}
					//currentCount+=1
					currentListCount+=1
				} else {
					initialStatusCode = strconv.Itoa(resp.StatusCode)
					initialTmpTitle := ""
					s, err := goscraper.Scrape(newUrl, 5)
					if err==nil {
						initialTmpTitle = s.Preview.Title
					}
					_ = s
					if verboseMode==true {
						var lenBody = 0
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							//errorFound=true
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
						if intelligentMode==true && CMSmode==false{
							tmpStatusCode := strconv.Itoa(resp.StatusCode)
							var tmpFound=false
							for _, each := range tmpTitleList { 
								var originalURL=""
								if strings.HasSuffix(each[0],"/") {
									originalURL=each[0]
								} else {
									originalURL=each[0]+"/"
								}
								if strings.Contains(finalURL,originalURL) {
									if newUrl==finalURL { 		
										tmpFound=true			
										if (strings.TrimSpace(each[1])!=strings.TrimSpace(tmpTitle) || len(tmpTitle)<1) {
											if tmpTitle!="Error" && tmpTitle!="Request Rejected" && tmpTitle!="Runtime Error"{
												if resp.StatusCode!=403 && resp.StatusCode!=503 && resp.StatusCode!=404 && resp.StatusCode!=400 && resp.StatusCode!=500 && resp.StatusCode!=204 {
													if (each[2]!=strconv.Itoa(lenBody)) {
														if CMSmode==false {
															if each[3]!=initialStatusCode && each[2]!=strconv.Itoa(lenBody){
																var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
																tmpResultList = append(tmpResultList,a...)
															}
														}
													} 												
												}
											}  
										} else {
											if (strings.TrimSpace(each[1])==strings.TrimSpace(tmpTitle)) {
												if initialStatusCode!=each[3] {
													var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
													tmpResultList = append(tmpResultList,a...)
												}
											}
										}
									}
									if tmpStatusCode=="200"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									} else if tmpStatusCode=="401"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									} else {
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								} 
							}
							if tmpFound==false {
								u, err := url.Parse(newUrl)
								if err != nil {
									panic(err)
								}				
								//tmpStatusCode := strconv.Itoa(resp.StatusCode)
								var newURL2=u.Scheme+"://"+u.Host				
								if resp.StatusCode==401 && initialStatusCode=="401" {
									fmt.Printf("%s [%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
									log.Printf("%s [%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
									var a = [][]string{{newURL2, initialStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								} else if (resp.StatusCode!=401 && initialStatusCode=="401") {
									fmt.Printf("%s [%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
									log.Printf("%s [%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
									var a = [][]string{{newURL2, initialStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								} else {
									if tmpStatusCode=="200"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									} else if tmpStatusCode=="401"{
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									} else {
										fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								}
							}
						} else {
							tmpStatusCode := strconv.Itoa(resp.StatusCode)
							//if CMSmode==false {
							//fmt.Println("aa")
							if Statuscode!=0 {
								if resp.StatusCode==Statuscode {
									fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)					
									log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)
									var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
									tmpResultList = append(tmpResultList,a...)
								} else {
										fmt.Printf("yy1 %s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
										log.Printf("yy2 %s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
								}						
							} else {				
								if tmpStatusCode=="200"{
									fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)					
									log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)
									var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
									tmpResultList = append(tmpResultList,a...)
								} else if tmpStatusCode=="401"{
									fmt.Printf("%s [%s]\n",newUrl, color.GreenString(tmpStatusCode))
									log.Printf("%s [%s]\n",newUrl, color.GreenString(tmpStatusCode))
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								} else {
									fmt.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)	
									log.Printf("%s [%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)				
								}
							}
							//}
						}
					} else {
						if Statuscode!=0 {
							tmpStatusCode := strconv.Itoa(resp.StatusCode)	
							if resp.StatusCode==Statuscode {	
								fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
								log.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
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
						} else {
							tmpStatusCode := strconv.Itoa(resp.StatusCode)	
							if resp.StatusCode==200 {		
								fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
								log.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
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
								log.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
							}				
						}
					}
					resp.Body.Close()
					//currentCount+=1
					//currentListCount+=1
					_ = resp
					_ = tmpTitle 
				} 
			}
			if currentListCount==totalListCount {
				reachedTheEnd=true
			}
			currentListCount+=1
			
			_ = err
		} else {
			currentListCount+=1
		}
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
    	fmt.Println(err)
        return err
    }
    defer resp.Body.Close()
    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
    	fmt.Println(err)
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
	Pathsource string `cli:"s,source" usage:"Path source (default | msf | exploitdb | exploitdb-asp | exploitdb-aspx | exploitdb-cfm | exploitdb-cgi | exploitdb-cfm | exploitdb-jsp | exploitdb-perl | exploitdb-php | exploitdb-others | RobotsDisallowed | SecLists)"`
	Threads int  `cli:"n,threads" usage:"No of concurrent threads (default: 2)"`
	Statuscode int  `cli:"c" usage:"Status code"`
	Intellimode bool `cli:"i" usage:"Intelligent mode"`
	Verbose bool `cli:"v,verbose" usage:"Verbose mode"`
	CMSmode bool `cli:"cms" usage:"Fingerprint CMS"`
	SpreadMode bool `cli:"x" usage:"Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host"`
	Logfilename string `cli:"l,log" usage:"Output to log file"`
	ContinueNum int  `cli:"r" usage:"Resume from x as in [x of 9999]"`	
	Proxyhost string `cli:"pHost" usage:"IP of HTTP proxy"`
	Proxyport string `cli:"pPort" usage:"Port of HTTP proxy (default 8080)"`
	Uagent string `cli:"ua" usage:"Set User-Agent"`
	Timeoutsec int `cli:"timeout" usage:"Set timeout to x seconds"`
}

func main() {
	//log.SetOutput(ioutil.Discard)
	//log.SetFlags(0)
	wpFileList	   = append(wpFileList,"/readme.html")
	joomlaFileList = append(joomlaFileList,"/administrator/manifests/files/joomla.xml")
	joomlaFileList = append(joomlaFileList,"/administrator/language/en-GB/en-GB.xml")
	drupalFileList = append(drupalFileList,"/CHANGELOG.txt")
	//drupalFileList = append(drupalFileList,"/LICENSE.txt")
	//const workersCount = 1
	
	filename1 := ""
	pFilename := ""
	uriPath := ""
	
	var contentList []string
	var pathList []string
	
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		if argv.Timeoutsec>0 {
			timeoutSec = argv.Timeoutsec
		}
		if len(argv.Uagent)>0 {
			userAgent=argv.Uagent
		}

		if len(argv.Proxyhost)>0 {
			if len(argv.Proxyport)>0 {
				proxy_addr="http://"+argv.Proxyhost+":"+argv.Proxyport
			} else {
				proxy_addr="http://"+argv.Proxyhost+":8080"
			}
			proxyMode=true
		}
		if argv.ContinueNum>0 {
			ContinueNum = argv.ContinueNum
		}
		if len(argv.Logfilename)>0 {
			logfileF, err := os.OpenFile(argv.Logfilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644)
			if err != nil {
					log.Fatal(err)
			}   
			defer logfileF.Close()
			log.SetOutput(logfileF)
		} else {
			logfileF, err := os.OpenFile("tmp.log", os.O_WRONLY|os.O_CREATE,0644)
			if err != nil {
					log.Fatal(err)
			}   
			defer logfileF.Close()
			log.SetOutput(logfileF)
		}
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
		//c := make(chan os.Signal, 2)
		//signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			//<-c
			sig := <-sigs
			fmt.Println(sig)
			//done <- true
			cleanup()
			os.Exit(3)
		}()		
		
		if len(pFilename)>0 {		
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fmt.Printf("[*] File %s not exists\n", pFilename)
				os.Exit(3)
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
				}
			}		
			_ = err2
		} 
		if len(uriPath)>0 {
			pathList = append(pathList, uriPath)
		}
		if len(Pathsource)>0 { 
			if Pathsource!="default" && Pathsource!="msf" && Pathsource!="exploitdb" && Pathsource!="exploitdb-asp" && Pathsource!="exploitdb-aspx" && Pathsource!="exploitdb-cfm" && Pathsource!="exploitdb-cgi" && Pathsource!="exploitdb-cfm" && Pathsource!="exploitdb-jsp" && Pathsource!="exploitdb-perl" && Pathsource!="exploitdb-php" && Pathsource!="RobotsDisallowed" && Pathsource!="SecLists" {
				fmt.Println("[*] Please select a valid Path source")
				os.Exit(3)
			}
		}
		if Pathsource=="default" {
			pFilename = "defaultPaths.txt"
			_, err1 := os.Stat("defaultPaths.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/defaultPaths.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("defaultPaths.txt", fileUrl)
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines("defaultPaths.txt")
			if err2==nil {
				for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines("pathList.txt")
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
				}		
			}
			_ = err2
		}
		if Pathsource=="exploitdb" {
			pFilename = "exploitdb_all.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
			for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
				}		
			}
			_ = err2
		}	
		if Pathsource=="exploitdb-others" {
			pFilename = "exploitdb_others.txt"
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/"+pFilename
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile(pFilename, fileUrl)
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines("SecLists-common.txt")
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
				if err!=nil {
					fmt.Println("[*] Error: ",err)
				}
				_ = err
			}
			_ = err1
			lines, err2 := readLines("RobotsDisallowed.txt")

			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
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
					fmt.Printf("[*] File %s not exists\n", filename1)
					os.Exit(3)
				}
				lines, err := readLines(filename1)
				if err!=nil {
					fmt.Println("Error: ",err)
				} else {
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
		    for _, v := range wpFileList {
		    	pathList = append(pathList,v)
		    }			
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
				}
				finalList = append(finalList, newUrl)
			  }
			}
		}

		//sigs1 := make(chan os.Signal, 1)
		//signal.Notify(sigs1, syscall.SIGINT, syscall.SIGTERM)

		urlChan := make(chan string)
		if intelligentMode==true {
			var wg1 sync.WaitGroup
			wg1.Add(workersCount)
	
			for i := 0; i < workersCount; i++ {
				go func() {
					//sig := <-sigs1
					testFakePath(urlChan)
					wg1.Done()
					//fmt.Println(sig)
					//done <- true
				}()
			}

			fmt.Println("[*] Getting Page Titles for Invalid URI Paths [Intelligent Mode]")
			log.Printf("[*] Getting Page Titles for Invalid URI Paths [Intelligent Mode]")
			completed := 0
			for _, each := range contentList {
				urlChan <- each+"/NonExistence"
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
				//sig := <-sigs
				getUrlWorker(urlChan)
				wg.Done()
				//fmt.Println(sig)
				//done <- true
			}()
		}

		totalListCount=len(finalList)

		fmt.Println("\n[*] Testing URI Paths: (Total: "+strconv.Itoa(totalListCount)+")")
		log.Printf("`\n[*] Testing URI Paths")
		//real uripaths
		completed1 := 0
		for _, each := range finalList {
			urlChan <- each+" | "+strconv.Itoa(completed1+1)
			completed1++
		}
		close(urlChan)  
		
		//var tmpLastCount = 0
		//var lastTime = time.Now()

		for {			
			time.Sleep(10 * time.Millisecond)
			if reachedTheEnd==true {
				time.Sleep(20 * time.Millisecond)
				break
			}
			if ContinueNum>len(finalList) || int(currentCount)>=len(finalList) {
				break
			}
			if len(finalList)==int(currentCount) {
				fmt.Println("\n[*] Processing results. Please wait...")
				log.Printf("\n[*] Processing results. Please wait...")
				break
			} 
			/*if int(currentCount)!=int(tmpLastCount) {
					tmpLastCount = int(currentCount)
					lastTime=time.Now()
			} else {
					if int(currentCount)>0 {
						fmt.Println(currentCount)
						t := time.Now()
						elapsed := t.Sub(lastTime)
						if elapsed.Seconds()>30 && currentCount>0 {
							break 
						}	
					}												
			} */
		}   
	
		//fmt.Println("\n")
		if CMSmode==true {
			for _, v := range tmpResultList {
				var wpVer = ""
				timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
				client := http.Client{
					Timeout: timeout,
				}
				if strings.HasSuffix(v[0],"/administrator/language/en-GB/en-GB.xml") || strings.HasSuffix(v[0],"/administrator/manifests/files/joomla.xml") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, err := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					resp, err := client.Do(req)		
					//resp, err := client.Get(v[0])
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
							var a = color.BlueString("[Found] ")+v[0]+color.BlueString(" [Joomla "+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}
				if strings.Contains(v[0],"/CHANGELOG.txt") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					req, err := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					resp, err := client.Do(req)		

					//resp, err := client.Get(v[0])
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
							var a = color.BlueString("[Found] ")+v[0]+color.BlueString(" ["+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}				

				if strings.Contains(v[0],"/readme.html") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, err := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					resp, err := client.Do(req)		
					//resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							for _, v1 := range s {
								if strings.Contains(v1,"<br /> Versão ") {
									v1=removeCharacters(v1,"<br /> Versão ")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								}
								if strings.Contains(v1,"<br /> Version ") {
									v1=removeCharacters(v1,"<br /> Version ")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								}
							}
						}
					}
					v[0]=strings.Replace(v[0],"/readme.html","",1)
					if len(wpVer)>0 {
						var a = color.BlueString("[Found] ")+v[0]+color.BlueString(" [Wordpress "+wpVer+"]")
						tmpResultList1 = append(tmpResultList1, a)
					}		
				}
				//if strings.HasPrefix(v[3],"Links for ") {			
				if strings.Contains(v[0],"/wp-links-opml.php") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, err := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					resp, err := client.Do(req)		
					//resp, err := client.Get(v[0])
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
						var a = color.BlueString("[Found] ")+v[0]+color.BlueString(" [Wordpress "+wpVer+"]")
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

			var tmpResultList2 []string	
			//var tmpResultList3 []string	

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
				fmt.Println("\n[*] No results found")
				log.Printf("\n[*] No results found")
			} else {
				time.Sleep(5 * time.Second)
				fmt.Println("\n")
				log.Printf("\n")

				var wg sync.WaitGroup
				urlChan = make(chan string)
				wg.Add(workersCount)
				for i := 0; i < workersCount; i++ {
					go func() {	
						checkURL(urlChan)
						wg.Done()
					}()
				}
				for _, each := range tmpResultList2 {
					urlChan <- each
				}
				close(urlChan)  
				wg.Wait()				
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
			if len(tmpResultList1)>0 {
				fmt.Println("\n")
			}
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
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err		
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
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
									log.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else if strings.Contains(v[1],"<") {
								s2 := strings.Split(v[1],"<")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
									log.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err		
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
									log.Printf("%s [%s]\n\n",v[2],v[3])
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
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err		
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
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
									log.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else if strings.Contains(v[1],"<") {
								s2 := strings.Split(v[1],"<")
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(s2[0])
								va2, err := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
									log.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err		
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("%s [%s]\n\n",v[2],v[3])
									log.Printf("%s [%s]\n\n",v[2],v[3])
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
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s2[0])
									va2, err := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
									_ = err									
								} else { 
									va0, err := version.NewVersion(selectedVer)
									va1, err := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
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
									log.Printf("%s [%s]\n\n",v[2],v[3])
								}
								_ = err
							} else { 
								va0, err := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if err==nil {
									if va0.Equal(va1) {
										fmt.Printf("%s [%s]\n\n",v[2],v[3])
										log.Printf("%s [%s]\n\n",v[2],v[3])
									}
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
