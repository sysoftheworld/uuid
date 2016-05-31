# uuid

uuid generates a Universally Unique IDentifier based on the standards set in [https://tools.ietf.org/html/rfc4122][RFC4122] and 
[http://pubs.opengroup.org/onlinepubs/9629399/apdxa.htm][DCE1.1].  

Version 1, 2, and 4 return just a UUID object
```Go
v1 := uuid.NewV1()
v2 := uuid.NewV2()
v4 := uuid.NewV4()
```

Version 3 and 5 return a UUID object along with an error. This is incase something went wrong while hashing. Additionally, they 
require a UUID compliant [https://tools.ietf.org/html/rfc4122#section-4.3][Namespace] and a name. This package provides 4 namespaces
for use (DNSNamespace, URLNamespace, IODNamespace, and X500Namespace), but any UUID may be used. 

```Go
v3, err := uuid.NewV3(uuid.DNSNamespace, "name")
v5, err := uuid.NewV5(uuid.DNSNamespace, "name")
```

The package also provides a String() func to convert the bytes to hex format

```Go
v1 := uuid.NewV1()
v1.String()
```
