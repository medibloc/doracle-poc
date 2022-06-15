```bash
sudo apt update
sudo apt install build-essential libssl-dev cpuid
sudo snap install go --classic
sudo snap install ego-dev --classic
sudo ego install az-dcap-client

sudo usermod -a -G sgx_prv $USER
```

```bash
ego-go build
ego sign main

unset SGX_AESM_ADDR
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run server
```