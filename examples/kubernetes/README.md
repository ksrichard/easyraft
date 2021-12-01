<p align="center">
<img src="https://github.com/ksrichard/easyraft/raw/main/logo.png" width="50%">
</p>

HTTP based key-value store example on Kubernetes (webkvs)
---
This example demonstrates, how easily you can implement an HTTP based in-memory distributed
key value store using EasyRaft running on Kubernetes

Usage
---
0. (Optional, Minikube only) Start Minikube tunnel
   1. ``minikube tunnel``
1. Build example (using minikube):
   2. ``eval $(minikube docker-env)``
   3. ``docker build -t webkvs:latest .``
2. Deploy to Kubernetes
   1. ``kubectl apply -f deployments.yaml``
3. Put value:
   1. ``curl --location --request POST 'http://localhost:5001/put?map=test&key=somekey&value=somevalue'``
4. Get value:
   1. ``curl --location --request GET 'http://localhost:5001/get?map=test&key=somekey'``