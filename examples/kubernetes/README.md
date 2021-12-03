<p align="center">
<img src="https://github.com/ksrichard/easyraft/raw/main/logo.png" width="50%">
</p>

HTTP based key-value store example on Kubernetes (webkvs)
---
This example demonstrates, how easily you can implement an HTTP based in-memory distributed
key value store using EasyRaft running on Kubernetes

Usage
---
1. Build example:
   3. ``docker build -t webkvs:latest .``
2. Deploy to Kubernetes
   1. ``kubectl apply -f deployments.yaml``
3. Put value:
   1. ``curl --location --request POST 'http://localhost:5001/put?map=test&key=somekey&value=somevalue'``
4. Get value:
   1. ``curl --location --request GET 'http://localhost:5001/get?map=test&key=somekey'``

Testing
---
You can try scale up/down (by setting replicas for example in `deployments.yaml`)
and check if you are able to put/get data and all the nodes are okay.

**Notet:** It is recommended to use `StatefulSet` as it will nicely scale up/down one-by-one the nodes without any issue.  