## Erasure Codes

Encode a file:

```shell
./lotus bftdsn encode -k 10 -m 3 [path of the file to encode]
# For example
# ./lotus bftdsn encode -k 2 -m 2 x.apk
# Then 2+2 chunks will be output: x.apk.0 x.apk.1 x.apk.2 x.apk.3
```

Decode chunks to get a file:
```shell
./lotus bftdsn decode -k 10 -m 3 -out [path of the output file] [path of the chunks]
# For example
# ./lotus bftdsn decode -k 2 -m 2 -out y.apk x.apk
# The program will find the 2+2 chunks and decode them to get y.apk
```