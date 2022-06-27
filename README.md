# PoC: Decentralized Oracle for sensitive data validation

This is a proof-of-concept of the decentralized oracle for off-chain data validation while preserving privacy.


## Hardware Requirements

The oracle only works on [SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)-[FLC](https://github.com/intel/linux-sgx/blob/master/psw/ae/ref_le/ref_le.md) environment with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.
You can check if your hardware supports SGX and it is enabled in the BIOS by following [EGo guide](https://docs.edgeless.systems/ego/#/getting-started/troubleshoot?id=hardware).


## Prerequisites

```bash
sudo apt update
sudo apt install build-essential libssl-dev

sudo snap install go --classic
sudo snap install ego-dev --classic
sudo ego install az-dcap-client

sudo usermod -a -G sgx_prv $USER
# If not
#
# [error_driver2api sgx_enclave_common.cpp:273] Enclave not authorized to run, .e.g. provisioning enclave hosted in app without access rights to /dev/sgx_provision. You need add the user id to group sgx_prv or run the app as root.
# [load_pce ../pce_wrapper.cpp:175] Error, call sgx_create_enclave for PCE fail [load_pce], SGXError:4004.
# ERROR: quote3_error_t=SGX_QL_INTERFACE_UNAVAILABLE (oe_result_t=OE_PLATFORM_ERROR) [openenclave-src/host/sgx/sgxquote.c:oe_sgx_qe_get_target_info:706]
# ERROR: SGX Plugin _get_report(): failed to get ecdsa report. OE_PLATFORM_ERROR (oe_result_t=OE_PLATFORM_ERROR) [openenclave-src/enclave/sgx/attester.c:_get_report:320]
```


## Build and sign

First of all, please prepare a signing key and a `enclave.json`.
```bash
openssl genrsa -out private.pem -3 3072
openssl rsa -in private.pem -pubout -out public.pem
```
```json
{
	"exe": "doracle-poc",
	"key": "private.pem",
	"debug": true,
	"heapSize": 512,
	"executableHeap": false,
	"productID": 1,
	"securityVersion": 1,
	"mounts": [
		{
			"source": "<a-directory-you-want>",
			"target": "/data",
			"type": "hostfs",
			"readOnly": false
		},
		{
			"target": "/tmp",
			"type": "memfs"
		}
	],
	"env": null,
	"files": null
}
```

Then, build a binary and sign it using the key that you generated.
```bash
ego-go build -o doracle-poc cmd/doracle-poc/main.go
ego sign doracle-poc
```


## Run

Before running the binary, the environment variable `SGX_AESM_ADDR` must be unset.
If not, the Azure DCAP client won't be used automatically.
```bash
unset SGX_AESM_ADDR
# If not,
#
# ERROR: sgxquoteexprovider: failed to load libsgx_quote_ex.so.1: libsgx_quote_ex.so.1: cannot open shared object file: No such file or directory [openenclave-src/host/sgx/linux/sgxquoteexloader.c:oe_sgx_load_quote_ex_library:118]
# ERROR: Failed to load SGX quote-ex library (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/host/sgx/sgxquote.c:oe_sgx_qe_get_target_info:688]
# ERROR: SGX Plugin _get_report(): failed to get ecdsa report. OE_QUOTE_LIBRARY_LOAD_ERROR (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/enclave/sgx/attester.c:_get_report:320]
```

Run the binary using `ego` so that it can be run in the secure enclave.
### peer to peer 
```bash
# For the first oracle that generates an oracle key,
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc start --init

# For an oracle that joins to the existing oracle group,
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc start --peer http://<ip>:<port>

# For restarting the oracle that already has the oracle key,
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc
```

### oracle and panacea communication

(First Oracle Node) Register oracle with panacea and generate oracle key.<br/>
After executing this command, `node-key.sealed` will be created in your local storage.
```bash
#####[First oracle node]#####
CHAIN_ID={your chain id}
GRPC_ADDR={panacea grpc address}
FIRST_ORACLE_MNEMONIC={your first oracle mnemonic}

# For the first oracle registration for panacea
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node register-node --chain-id $CHAIN_ID --grpcAddr $GRPC_ADDR --mnemonic "$FIRST_ORACLE_MNEMONIC"

# For oracle where it need to generate a oracle key
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node init-oracle-key --chain-id $CHAIN_ID --grpcAddr $GRPC_ADDR --mnemonic "$FIRST_ORACLE_MNEMONIC"
```

(Second Oracle Node) Register another oracles with panacea
```bash
#####[Second oracle node]#####
CHAIN_ID={your chain id}
GRPC_ADDR={panacea grpc address}
SECOND_ORACLE_MNEMONIC={your second oracle mnemonic}

# For the second oracle registration for panacea
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node register-node --chain-id $CHAIN_ID --grpcAddr $GRPC_ADDR --mnemonic "$SECOND_ORACLE_MNEMONIC"
```

(First Oracle Node) Voting a performed on the first Oracle node
```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node vote {voting target oracle address} --chain-id $CHAIN_ID --grpcAddr $GRPC_ADDR --mnemonic "$FIRST_ORACLE_MNEMONIC" --signer-id $(ego signerid doracle-poc)
```

(Second Oracle Node) If voting is successfully completed, Oracle can get oraclePrivKey<br/>
After executing this command, `oracle-key.sealed` will be created in your local storage.
```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node get-oracle-key --chain-id $CHAIN_ID --grpcAddr $GRPC_ADDR --mnemonic "$SECOND_ORACLE_MNEMONIC"
```

### Each oracle's ciphertext decryption test
Get a file encrypted with oraclePubKey in plaintext through the command below.<br/>
Since we get the publicKey from the panacea, we communicate with the panacea grpc. <br/>
After the command is executed, the cipher text specified as the file name exists in the local storage.
```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc client generate-encrypt-data "encrypt data sample" --grpcAddr $GRPC_ADDR --output /data/{fileName}
```

You can check that decryption is possible on any oracle node.
```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc node read-encrypt-data /data/{fileName}
```