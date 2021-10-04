FROM rust:1-buster

RUN git clone git@github.com:mobilecoinfoundation/mobilecoin.git /tmp/mobilecoin
WORKDIR /tmp/mobilecoin

RUN export SIGSTRUCT_URI=$(curl -s https://enclave-distribution.prod.mobilecoin.com/production.json | jq -r '.consensus.sigstruct')
RUN curl -O https://enclave-distribution.prod.mobilecoin.com/${SIGSTRUCT_URI}

RUN SGX_MODE=HW IAS_MODE=PROD CONSENSUS_ENCLAVE_CSS=$(pwd)/consensus-enclave.css cargo build --release -p mc-mobilecoind -p mc-mobilecoind-json

CMD ["./target/release/mobilecoind",\
	"--ledger-db /path/to/ledger-db",\
	"--mobilecoind-db /path/to/mobilecoind-db",\
	"--poll-interval 10",\
	"--mobilecoind-db /path/to/mobilecoind-db",\
	"--peer mc://node2.prod.mobilecoinww.com/",\
	"--tx-source-url https://ledger.mobilecoinww.com/node1.prod.mobilecoinww.com/",\
	"--tx-source-url https://ledger.mobilecoinww.com/node2.prod.mobilecoinww.com/",\
	"--listen-uri insecure-mobilecoind://127.0.0.1:4444/",\
	""]

# CMD ["./target/release/mobilecoind-json"]