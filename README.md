# pathbrute
Pathbrute
```
$ ./pathBrute -h
Options:

  -h, --help       display help information
  -U, --filename   File containing list of websites
  -u, --url        Url of website
  -P, --Paths      File containing list of URI paths
  -s, --source     Path source (default | msf | RobotsDisallowed | SecLists)
  -p, --path       URI path
  -n, --threads    No of concurrent threads
  -c               Status code
  -i               Intelligent mode
  -v, --verbose    Verbose mode
      --cms        Fingerprint CMS
  -x               Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host
```
  
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

