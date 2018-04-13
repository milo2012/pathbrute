# pathbrute
Pathbrute  
**Please check RELEASE section for copmiled executables***  
```
$ ./pathBrute -h
Options:

  -h, --help       display help information
  -U, --filename   File containing list of websites
  -u, --url        Url of website
  -P, --Paths      File containing list of URI paths
  -p, --path       URI path
  -s, --source     Path source (default | msf | RobotsDisallowed | SecLists)
  -n, --threads    No of concurrent threads
  -c               Status code
  -i               Intelligent mode
  -v, --verbose    Verbose mode
      --cms        Fingerprint CMS
  -x               Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host
```

#Compilation
```
go get github.com/mkideal/cli
go get github.com/badoux/goscraper
go get github.com/fatih/color
go github.com/hashicorp/go-version
go build pathBrute.go
```
  
#Example 
```
./pathBrute -s default -f urls.txt -v -i -n 25 
[*] Getting Default Page Title for Invalid URI Paths
http://xxxx.com/xxx [404] [404 Not Found]

[*] Testing URI Paths
http://xxxx.com/AdminRealm [404] [168] [404 Not Found]
http://xxxx.com/AddressBookJ2WE/services/AddressBook/wsdl/ [404] [168] [404 Not Found]
http://xxxx.com/AdminJDBC [404] [168] [404 Not Found]
http://xxxx.com/AdminMain [404] [168] [404 Not Found]
http://xxxx.com/Admin [404] [168] [404 Not Found]
http://xxxx.com/AdminProps [404] [168] [404 Not Found]
http://xxxx.com/AddressBookJ2WB [404] [168] [404 Not Found]
http://xxxx.com/AE/index.jsp [404] [168] [404 Not Found]
http://xxxx.com/.web [404] [168] [404 Not Found]
http://xxxx.com/ADS-EJB [200] [482] []
```
  
#Example using the --cms option
```
$ /git/pathbrute/pathBrute -U urls.txt --cms -i -v

[*] Testing URI Paths
http://xxxx.com/CHANGELOG.txt [404] [1118] [404 Not Found] [59 of 68]
http://yyyy.com/wp-links-opml.php [404] [2139] [404 - Error: 404] [61 of 68]
http://zzzz.com/wp-links-opml.php [200] [5930] [] [64 of 68]
http://zzzz.com/administrator/manifests/files/joomla.xml [200] [6154] [] [65 of 68]
http://zzzz.com/CHANGELOG.txt [200] [5898] [] [66 of 68]
http://zzzz.com/administrator/language/en-GB/en-GB.xml [200] [6139] [] [67 of 68]

-- redacted for brevity --- 

[*] Processing results. Please wait...
http://ffff.com [Joomla 3.8.6]
http://eeee.com/web [Wordpress 4.9.2]
http://xxxx.com [Joomla 2.5.28]
http://yyyy.com [Joomla 1.7.1]
http://gggg.com [Drupal 7.21]
http://hhhh.com [Wordpress 4.6.11]
http://iiii.com [Wordpress 4.9.5]
https://jjjj.com [Wordpress 4.9.3]
https://kkkk.com [Wordpress 4.9.5]
https://llll.com [Wordpress 4.9.5]
```
