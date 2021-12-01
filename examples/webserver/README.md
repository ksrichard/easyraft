<p align="center">
<img src="https://github.com/ksrichard/easyraft/raw/main/logo.png" width="50%">
</p>

Simple HTTP based key-value store example
---
This example demonstrates, how easily you can implement an HTTP based in-memory distributed
key value store using EasyRaft.

Usage
---
1. Build example:
   2. ``go build -o webserver``
2. Run nodes locally:
   1. ``EASYRAFT_PORT=5000 DISCOVERY_PORT=5001 HTTP_PORT=5002 DATA_DIR="s1" ./webserver``
   2. ``EASYRAFT_PORT=5003 DISCOVERY_PORT=5004 HTTP_PORT=5005 DATA_DIR="s2" ./webserver``
   3. ``EASYRAFT_PORT=5006 DISCOVERY_PORT=5007 HTTP_PORT=5008 DATA_DIR="s3" ./webserver``
3. Put value on any node:
   1. ``curl --location --request POST 'http://localhost:5008/put?map=test&key=somekey&value=somevalue'``
4. Get value from all the nodes:
   1. ``curl --location --request GET 'http://localhost:5002/get?map=test&key=somekey'``
   2. ``curl --location --request GET 'http://localhost:5005/get?map=test&key=somekey'``
   3. ``curl --location --request GET 'http://localhost:5008/get?map=test&key=somekey'``
