```bash
sudo apt update
sudo apt install build-essential libssl-dev cpuid
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

```bash
ego-go build
ego sign doracle-poc

unset SGX_AESM_ADDR
# If not,
#
# ERROR: sgxquoteexprovider: failed to load libsgx_quote_ex.so.1: libsgx_quote_ex.so.1: cannot open shared object file: No such file or directory [openenclave-src/host/sgx/linux/sgxquoteexloader.c:oe_sgx_load_quote_ex_library:118]
# ERROR: Failed to load SGX quote-ex library (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/host/sgx/sgxquote.c:oe_sgx_qe_get_target_info:688]
# ERROR: SGX Plugin _get_report(): failed to get ecdsa report. OE_QUOTE_LIBRARY_LOAD_ERROR (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/enclave/sgx/attester.c:_get_report:320]

AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracle-poc
```