SUBJECT='/C=CN/ST=SC/L=CD/O=x/OU=Org/CN=\*.x.com/emailAddress=birdyfj@gmail.com'


echo "generate CA certificate and private key"
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj $SUBJECT


echo "generate server private key and certificate request(csr)"
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj $SUBJECT

echo "sign certification"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -subj $SUBJECT

