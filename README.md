# hike

Requests multiple url based on a url

## Install 

```text
go install github.com/zblurx/hike@latest
```

## Usage

```bash
$ hike https://www.google.com/this/is/a/mistake
https://www.google.com/this/is/a/mistake [404] [Content-Length: 1578] [Error 404 (Not Found)!!1]
https://www.google.com/this/is/a [404] [Content-Length: 1570] [Error 404 (Not Found)!!1]
https://www.google.com/this/is [404] [Content-Length: 1568] [Error 404 (Not Found)!!1]
https://www.google.com/this [404] [Content-Length: 1565] [Error 404 (Not Found)!!1]
https://www.google.com [200] [Content-Length: 0] [Google]
```

*worst example but it can find nice things*

## Permute mode

```bash
$ hike https://www.google.com/this/is/a/mistake
https://www.google.com/this/is/mistake [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/this/is [404] [Content-Length: 1568] [Error 404 (Not Found)!!1]
https://www.google.com/this/mistake/is [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/this/mistake [404] [Content-Length: 1573] [Error 404 (Not Found)!!1]
https://www.google.com/this [404] [Content-Length: 1565] [Error 404 (Not Found)!!1]
https://www.google.com/is/this/mistake [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/is/this [404] [Content-Length: 1568] [Error 404 (Not Found)!!1]
https://www.google.com/is/mistake/this [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/is/mistake [404] [Content-Length: 1571] [Error 404 (Not Found)!!1]
https://www.google.com/is [200] [Content-Length: 0] [Google]
https://www.google.com/mistake/this/is [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/mistake/this [404] [Content-Length: 1573] [Error 404 (Not Found)!!1]
https://www.google.com/mistake/is/this [404] [Content-Length: 1576] [Error 404 (Not Found)!!1]
https://www.google.com/mistake/is [404] [Content-Length: 1571] [Error 404 (Not Found)!!1]
https://www.google.com/mistake [404] [Content-Length: 1568] [Error 404 (Not Found)!!1]
https://www.google.com [200] [Content-Length: 0] [Google]
```