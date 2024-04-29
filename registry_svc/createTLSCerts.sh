#!/usr/bin/env bash
shopt -s nullglob globstar
set -x # have bash print command been ran
set -e # fail if any command fails

setup_certs(){
  { # create CA.
    openssl \
      req \
      -new \
      -newkey rsa:4096 \
      -days 1024 \
      -nodes \
      -x509 \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -keyout configs/v1/certificates/rootCA.key \
      -out configs/v1/certificates/rootCA.crt
  }

  { # create server certs.
    openssl \
      req \
      -new \
      -newkey rsa:2048 \
      -days 372 \
      -nodes \
      -x509 \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -addext "subjectAltName=DNS:*.asonrythme.com,DNS:*.asonrythme,DNS:localhost" \
      -CA configs/v1/certificates/rootCA.crt \
      -CAkey configs/v1/certificates/rootCA.key  \
      -keyout configs/v1/certificates/server.key \
      -out configs/v1/certificates/server.crt
  }

  { # create client certs.
    openssl \
      req \
      -new \
      -newkey rsa:2048 \
      -days 372 \
      -nodes \
      -x509 \
      -subj "/C=US/ST=CA/O=MyOrg/CN=myOrgCA" \
      -addext "subjectAltName=DNS:*.asonrythme.com,DNS:*.asonrythme,DNS:localhost" \
      -CA configs/v1/certificates/rootCA.crt \
      -CAkey configs/v1/certificates/rootCA.key  \
      -keyout configs/v1/certificates/client.key \
      -out configs/v1/certificates/client.crt
  }

  { # clean
    rm -rf configs/v1/certificates/*.csr
    rm -rf configs/v1/certificates/*.srl

    chmod 666 configs/v1/certificates/server.crt configs/v1/certificates/server.key configs/v1/certificates/rootCA.crt
  }
}
setup_certs

