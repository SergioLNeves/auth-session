PRIVATE_KEY = private-key.pem
PUBLIC_KEY = public-key.pem
KEY_SIZE = 2048

.PHONY: run
run:
	go run ./cmd/api/main.go

.PHONY: gen-key
gen-key: $(PUBLIC_KEY)
$(PRIVATE_KEY):
	@echo "Gerando chave privada..."
	openssl genrsa -out $(PRIVATE_KEY) $(KEY_SIZE)
	@chmod 600 $(PRIVATE_KEY)
$(PUBLIC_KEY): $(PRIVATE_KEY)
	@echo "Extraindo chave pública..."
	openssl rsa -in $(PRIVATE_KEY) -pubout -out $(PUBLIC_KEY)
