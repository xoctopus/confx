
## # generate ca private key
## openssl genrsa -out ca.key 2048
## # generate certification sign request(csr)
## openssl req -new -key ca.key -out ca.csr -subj "/C=CN/ST=SC/L=CD/O=x/OU=Org/CN=*.x.com/emailAddress=birdyfj@gmail.com"
## # generate self-signed certificate
## openssl x509 -req -in ca.csr -signkey ca.key -out ca.cert

SUBJECT='/C=CN/ST=SC/L=CD/O=x/OU=Org/CN=\*.x.com/emailAddress=birdyfj@gmail.com'

CAKEY=ca.key
CACERT=ca.cert

openssl req -x509 -nodes -newkey rsa:2048 -keyout $CAKEY -out ca.cert -subj $SUBJECT

KEY=key
CSR=csr
CERT=cert

mkdir -p server

openssl genrsa -out server/$KEY 2048
openssl req -nodes -new -key server/$KEY -out server/$CSR -subj $SUBJECT
openssl x509 -req -in server/$CSR -CA $CACERT -CAkey $CAKEY -CAcreateserial -out server/$CERT

mkdir -p client

openssl genrsa -out client/$KEY 2048
openssl req -nodes -new -key client/$KEY -out client/$CSR -subj $SUBJECT
openssl x509 -req -in client/$CSR -CA $CACERT -CAkey $CAKEY -CAcreateserial -out client/$CERT

