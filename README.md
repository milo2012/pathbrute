# pathbrute
Pathbrute  
  
Pathbrute is a DirB/Dirbuster type of tool designed to brute force directories and files names on web/application servers.  However, it has some new tricks.  
Pathbrute has a number of wordlists from metasploit/exploit-database and other sources that it uses to discover interesting content on servers.  
  
pathBrute contains/uses a number of self compiled wordlists for identifying “interesting” content and potentially vulnerable websites.
1) More than 13924 URI paths from Exploit-Database 
2) URI paths from Metasploit Framework

pathBrute can also use wordlists from other sources if you prefer.  
pathBrute can also be used for identifying if any type of CMS (Joomla, WordPress and Drupal) is running on the target websites and fingerprint the versions of the CMS using the –cms option.  
  
Binaries for different platforms and architectures are available in the the release section.  
 
**Please check RELEASE section for compiled executables**    
  
```
$ ./pathBrute -h
Options:

  -h, --help       display help information
  -U, --filename   File containing list of websites
  -u, --url        Url of website
  -P, --Paths      File containing list of URI paths
  -p, --path       URI path
  -s, --source     Path source (default | msf | exploitdb | exploitdb-asp | exploitdb-aspx | exploitdb-cfm | exploitdb-cgi | exploitdb-cfm | exploitdb-jsp | exploitdb-perl | exploitdb-php  | RobotsDisallowed | SecLists)
  -n, --threads    No of concurrent threads (default: 2)
  -c               Status code
  -i               Intelligent mode
  -v, --verbose    Verbose mode
      --cms        Fingerprint CMS
  -x               Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host
  -l, --log        Output to log file
  -r               Resume from x as in [x of 9999]
      --pHost      IP of HTTP proxy
      --pPort      Port of HTTP proxy (default 8080)
      --ua         Set User-Agent
      --timeout    Set timeout to x seconds
```
 
#Docker
```
- Building from Dockerfile
docker build -t example-scratch -f Dockerfile
docker run -it 2af3eecdb017 /pathBrute_linux -u http://testphp.vulnweb.com/ -s default  -v -i -n 20

- Pull latest Docker image
docker pull milo2012/pathbrute
docker run -it 589606bdc12a /pathBrute_linux -u http://testphp.vulnweb.com/ -s default  -v -i -n 20

```
    
#Compilation  
```
#Manual Compilation  `
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

[Found] https://127.0.0.1/.gitignore [200] [28] []
[Found] https://127.0.0.1/.htaccess [200] [1164] []
[Found] https://127.0.0.1/PMA/ [200] [8575] [phpMyAdmin]
[Found] https://127.0.0.1/.htaccess [200] [1164] []
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
