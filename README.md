# A User-NamespacePrefix Limit Kubernetes Admission Webhook

This webhook only allow user operator object in specific namespace-prefix namespaces which is defined in the policy file


## BUILD
mkdir -p webhook/src && git clone https://github.com/lovejoy/user-namespace-prefix-limit-admission-webhook
cd webhook  && export GOPATH=$(pwd)
CGO_ENABLED=0 GOOS=linux go build  -a -installsuffix cgo -o webhook  user-namespace-prefix-limit-admission-webhook


## USAGE

Generate ca cert file & server cert file & key file

```bash
openssl genrsa -out ca.key 2048
openssl genrsa -out server.key 2048
openssl req -new -x509 -key ca.key -out ca.crt
openssl req -new -key server.key -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
```

Notice: The domain when you create server.crt must same as the domain you use in ValidatingWebhookConfiguration

Modify the policy.json as you want 

Deploy the webhook on your way

Edit all-object.yaml

kubectl create -f all-object.yaml to create the  ValidatingWebhookConfiguration 

