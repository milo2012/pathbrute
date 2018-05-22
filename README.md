# pathbrute
Pathbrute  
  
Pathbrute is a DirB/Dirbuster type of tool designed to brute force directories and files names on web/application servers.  
However, it has some new tricks. It is no longer a dumb directories/files brute force tool if you use the -v and -i option.  
  
It runs on Windows/Linux/OSX operating systems and on ARM/ARM64/x86/x64 processors.  
    
Some of it includes:    
1) Wordlists from Exploit databases and Metasploit  
2) Identify interesting URLs eventhough websites return HTTP status code 200 for all URI paths.  
3) Identify valid  paths that require authentication (HTTP status code 401)  
4) Reduce the number of results for wordlists with URI paths with nested directories (See https://github.com/milo2012/pathbrute/issues/1 for more information)  
  
Pathbrute has a number of wordlists from metasploit/exploit-database and other sources that it uses to discover interesting content on servers.  
  
pathBrute contains/uses a number of self compiled wordlists for identifying “interesting” content and potentially vulnerable websites.
1) More than 18603 URI paths from Exploit-Database 
2) More than 355 URI paths from Metasploit Framework

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
  -s, --source     Path source (default | msf | exploitdb | exploitdb-asp | exploitdb-aspx | exploitdb-cfm | exploitdb-cgi | exploitdb-cfm | exploitdb-jsp | exploitdb-perl | exploitdb-php | exploitdb-others | RobotsDisallowed | SecLists)
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
      --update     Update URI path wordlists from Github
```
***
 
#Docker
```
- Building from Dockerfile
docker build -t example-scratch -f Dockerfile
docker run -it 2af3eecdb017 /pathBrute_linux -u http://testphp.vulnweb.com/ -s default  -v -i -n 20

- Pull latest Docker image
docker pull milo2012/pathbrute
docker run -it 589606bdc12a /pathBrute_linux -u http://testphp.vulnweb.com/ -s default  -v -i -n 20

```
***
    
#Compilation  
```
#Manual Compilation  `
go get github.com/mkideal/cli
go get github.com/badoux/goscraper
go get github.com/fatih/color
go github.com/hashicorp/go-version
go build pathBrute.go  
```
***
  
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
***
  
#Explanation of the output from pathBrute  
```
https://208.88.199.241/sap/bc/webdynpro/sap/wdr_test_gantt [401] [458] [File or directory not found] [27736 of 38988]
```
Below is a description of the output from pathBrute  
  
**[401]** - refers to the HTTP status code  
**[458]** - refers to the size of the HTTP response  
**[File or directory not found]** - refers to the title of the page  
**[27736 of 38988]** - refers to the current position in the list   
         
***
    
#Example using the --cms option  
Below is a sample output when using the --cms option to fingerprint the CMS on the target hosts.  It also returns the Metasploit modules based on the version of the CMS software.  
 
```
$ /git/pathbrute/pathBrute -U urls.txt --cms -i -v
[...redacted for brevity...]

[Found] https://[redacted] [Wordpress 4.8.6]

[Found] https://[redacted] [Wordpress 3.0]
Wordpress XML-RPC system.multicall Credential Collector [auxiliary/scanner/http/wordpress_multicall_creds]
WordPress Traversal Directory DoS [auxiliary/dos/http/wordpress_directory_traversal_dos]

[Found] https://[redacted] [Wordpress 3.8.26]
Wordpress XMLRPC DoS [auxiliary/dos/http/wordpress_xmlrpc_dos]
Wordpress XML-RPC system.multicall Credential Collector [auxiliary/scanner/http/wordpress_multicall_creds]
WordPress Traversal Directory DoS [auxiliary/dos/http/wordpress_directory_traversal_dos]

[Found] https://[redacted] [Wordpress 3.9.9]
Wordpress XML-RPC system.multicall Credential Collector [auxiliary/scanner/http/wordpress_multicall_creds]
WordPress Traversal Directory DoS [auxiliary/dos/http/wordpress_directory_traversal_dos]

[Found] https://[redacted] [Wordpress 3.3]
Wordpress XML-RPC system.multicall Credential Collector [auxiliary/scanner/http/wordpress_multicall_creds]
WordPress Traversal Directory DoS [auxiliary/dos/http/wordpress_directory_traversal_dos]
```

